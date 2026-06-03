package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/handler/admin"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// RegisterInvoiceRoutes 注册发票（开票申请）相关路由：用户端 + 管理端。
func RegisterInvoiceRoutes(
	v1 *gin.RouterGroup,
	invoiceHandler *handler.InvoiceHandler,
	adminInvoiceHandler *admin.InvoiceHandler,
	jwtAuth middleware.JWTAuthMiddleware,
	adminAuth middleware.AdminAuthMiddleware,
	settingService *service.SettingService,
) {
	// --- 用户端发票接口（需登录）---
	authenticated := v1.Group("/invoices")
	authenticated.Use(gin.HandlerFunc(jwtAuth))
	authenticated.Use(middleware.BackendModeUserGuard(settingService))
	{
		authenticated.GET("/invoiceable-orders", invoiceHandler.GetInvoiceableOrders)
		authenticated.GET("/my", invoiceHandler.GetMyInvoices)
		authenticated.POST("", invoiceHandler.ApplyInvoice)
		authenticated.GET("/:id", invoiceHandler.GetMyInvoice)
		authenticated.POST("/:id/cancel", invoiceHandler.CancelInvoice)
	}

	// --- 管理端发票管理（admin 鉴权）---
	adminGroup := v1.Group("/admin/invoices")
	adminGroup.Use(gin.HandlerFunc(adminAuth))
	{
		adminGroup.GET("", adminInvoiceHandler.List)
		adminGroup.GET("/:id", adminInvoiceHandler.GetDetail)
		adminGroup.POST("/:id/parse-pdf", adminInvoiceHandler.ParsePDF)
		adminGroup.POST("/:id/issue", adminInvoiceHandler.Issue)
		adminGroup.POST("/:id/reject", adminInvoiceHandler.Reject)
	}
}
