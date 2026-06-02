package middleware

import (
	"context"
	"errors"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/pkg/googleapi"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ip"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// APIKeyAuthGoogle is a Google-style error wrapper for API key auth.
func APIKeyAuthGoogle(apiKeyService *service.APIKeyService, cfg *config.Config) gin.HandlerFunc {
	return APIKeyAuthWithSubscriptionGoogle(apiKeyService, nil, cfg)
}

// APIKeyAuthWithSubscriptionGoogle behaves like ApiKeyAuthWithSubscription but returns Google-style errors:
// {"error":{"code":401,"message":"...","status":"UNAUTHENTICATED"}}
//
// It is intended for Gemini native endpoints (/v1beta) to match Gemini SDK expectations.
func APIKeyAuthWithSubscriptionGoogle(apiKeyService *service.APIKeyService, subscriptionService *service.SubscriptionService, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if v := strings.TrimSpace(c.Query("api_key")); v != "" {
			abortWithGoogleError(c, 400, "Query parameter api_key is deprecated. Use Authorization header or key instead.")
			return
		}
		apiKeyString := extractAPIKeyForGoogle(c)
		if apiKeyString == "" {
			abortWithGoogleError(c, 401, "API key is required")
			return
		}

		apiKey, err := apiKeyService.GetByKey(c.Request.Context(), apiKeyString)
		if err != nil {
			if errors.Is(err, service.ErrAPIKeyNotFound) {
				abortWithGoogleError(c, 401, "Invalid API key")
				return
			}
			abortWithGoogleError(c, 500, "Failed to validate API key")
			return
		}

		// 状态语义与 api_key_auth.go 对齐：disabled → 401；expired → 403；quota_exhausted → 429。
		// 只把"未知/已禁用"的 key 当 401 disabled，避免给 Gemini 客户端返回错误的 disabled 语义。
		if !apiKey.IsActive() {
			switch apiKey.Status {
			case service.StatusAPIKeyExpired:
				abortWithGoogleError(c, 403, "API key has expired")
				return
			case service.StatusAPIKeyQuotaExhausted:
				abortWithGoogleError(c, 429, "API key quota has been exhausted")
				return
			default:
				abortWithGoogleError(c, 401, "API key is disabled")
				return
			}
		}
		// 即使状态是 active，也要做运行时过期/配额检查（与 api_key_auth.go 第 178/182 行同款）
		if apiKey.IsExpired() {
			abortWithGoogleError(c, 403, "API key has expired")
			return
		}
		if apiKey.IsQuotaExhausted() {
			abortWithGoogleError(c, 429, "API key quota has been exhausted")
			return
		}

		// IP 白/黑名单检查：与 api_key_auth.go 第 89~98 行同款。
		// /v1beta 与 /antigravity/v1beta 路径之前没挂这个检查，导致用户给 API Key 配的 IP 限制
		// 在 Gemini 原生 SDK 路径上被绕过（GPT round 21 #2）。错误信息故意模糊。
		if len(apiKey.IPWhitelist) > 0 || len(apiKey.IPBlacklist) > 0 {
			clientIP := ip.GetTrustedClientIP(c)
			allowed, _ := ip.CheckIPRestrictionWithCompiledRules(clientIP, apiKey.CompiledIPWhitelist, apiKey.CompiledIPBlacklist)
			if !allowed {
				abortWithGoogleError(c, 403, "Access denied")
				return
			}
		}

		if apiKey.User == nil {
			abortWithGoogleError(c, 401, "User associated with API key not found")
			return
		}
		if !apiKey.User.IsActive() {
			abortWithGoogleError(c, 401, "User account is not active")
			return
		}

		// 简易模式：跳过余额和订阅检查
		if cfg.RunMode == config.RunModeSimple {
			c.Set(string(ContextKeyAPIKey), apiKey)
			c.Set(string(ContextKeyUser), AuthSubject{
				UserID:      apiKey.User.ID,
				Concurrency: apiKey.User.Concurrency,
			})
			c.Set(string(ContextKeyUserRole), apiKey.User.Role)
			setGroupContext(c, apiKey.Group)
			// 注入 Sub2APIUserID：simple 模式下也可能走 Gemini 独享调度，
			// 不写入会让 trySelectExclusiveSeatAccount 在 ctx 中拿不到 user，导致独享 seat 失效
			setSub2APIUserContext(c, apiKey.User.ID)
			_ = apiKeyService.TouchLastUsed(c.Request.Context(), apiKey.ID)
			c.Next()
			return
		}

		isSubscriptionType := apiKey.Group != nil && apiKey.Group.IsSubscriptionType()
		var hasActiveExclusiveSeat bool
		if isSubscriptionType && subscriptionService != nil {
			// 优先级：独享 seat > 共享订阅（与 api_key_auth.go 同步）
			// 调度层 DP2C 优先独享，鉴权/计费层必须保持一致语义，避免共享 sub 限额拦截独享请求
			if subscriptionService.HasActiveExclusiveSeat(c.Request.Context(), apiKey.User.ID, apiKey.Group.ID) {
				hasActiveExclusiveSeat = true
				ctx := context.WithValue(c.Request.Context(), ctxkey.ExclusiveSeatActive, true)
				c.Request = c.Request.WithContext(ctx)
			} else {
				subscription, err := subscriptionService.GetActiveSubscription(
					c.Request.Context(),
					apiKey.User.ID,
					apiKey.Group.ID,
				)
				if err != nil {
					abortWithGoogleError(c, 403, "No active subscription found for this group")
					return
				}
				needsMaintenance, validateErr := subscriptionService.ValidateAndCheckLimits(subscription, apiKey.Group)
				if validateErr != nil {
					// 套餐额度用完：返回 RESOURCE_EXHAUSTED + 明确中文文案 + Retry-After，避免客户端当成临时限流反复重试
					if service.IsUsageLimitError(validateErr) {
						detail := service.BuildSubscriptionLimitDetail(subscription, apiKey.Group, validateErr)
						abortWithGoogleRateLimit(c, detail.Message, detail.RetryAfterSeconds)
						return
					}
					abortWithGoogleError(c, 403, validateErr.Error())
					return
				}

				c.Set(string(ContextKeySubscription), subscription)

				if needsMaintenance {
					maintenanceCopy := *subscription
					subscriptionService.DoWindowMaintenance(&maintenanceCopy)
				}
			}
		}

		// 余额检查：独享 seat 已付费跳过
		if !isSubscriptionType && !hasActiveExclusiveSeat {
			if apiKey.User.Balance <= 0 {
				abortWithGoogleError(c, 403, "Insufficient account balance")
				return
			}
		}

		c.Set(string(ContextKeyAPIKey), apiKey)
		c.Set(string(ContextKeyUser), AuthSubject{
			UserID:      apiKey.User.ID,
			Concurrency: apiKey.User.Concurrency,
		})
		c.Set(string(ContextKeyUserRole), apiKey.User.Role)
		setGroupContext(c, apiKey.Group)
		// 注入 Sub2APIUserID：让 /v1beta/models 等只走鉴权的端点也能命中独享 seat 调度
		setSub2APIUserContext(c, apiKey.User.ID)
		_ = apiKeyService.TouchLastUsed(c.Request.Context(), apiKey.ID)
		c.Next()
	}
}

// extractAPIKeyForGoogle extracts API key for Google/Gemini endpoints.
// Priority: x-goog-api-key > Authorization: Bearer > x-api-key > query key
// This allows OpenClaw and other clients using Bearer auth to work with Gemini endpoints.
func extractAPIKeyForGoogle(c *gin.Context) string {
	// 1) preferred: Gemini native header
	if k := strings.TrimSpace(c.GetHeader("x-goog-api-key")); k != "" {
		return k
	}

	// 2) fallback: Authorization: Bearer <key>
	auth := strings.TrimSpace(c.GetHeader("Authorization"))
	if auth != "" {
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			if k := strings.TrimSpace(parts[1]); k != "" {
				return k
			}
		}
	}

	// 3) x-api-key header (backward compatibility)
	if k := strings.TrimSpace(c.GetHeader("x-api-key")); k != "" {
		return k
	}

	// 4) query parameter key (for specific paths)
	if allowGoogleQueryKey(c.Request.URL.Path) {
		if v := strings.TrimSpace(c.Query("key")); v != "" {
			return v
		}
	}

	return ""
}

func allowGoogleQueryKey(path string) bool {
	return strings.HasPrefix(path, "/v1beta") || strings.HasPrefix(path, "/antigravity/v1beta")
}

func abortWithGoogleError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"error": gin.H{
			"code":    status,
			"message": message,
			"status":  googleapi.HTTPStatusToGoogleStatus(status),
		},
	})
	c.Abort()
}
