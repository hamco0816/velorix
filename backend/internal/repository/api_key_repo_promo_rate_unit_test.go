package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

// 回归测试：GetByKeyForAuth 的 WithGroup 字段白名单必须包含限时倍率 4 字段，
// 否则鉴权快照里 promo 永远是 nil，计费时 PromoActiveAt 恒为 false，限时倍率不生效。
func TestAPIKeyRepository_GetByKeyForAuth_LoadsGroupPromoRate(t *testing.T) {
	repo, client := newAPIKeyRepoSQLite(t)
	ctx := context.Background()
	user := mustCreateAPIKeyRepoUser(t, ctx, client, "getbykey-auth-promo@test.com")

	promoRate := 0.2
	promoStartsAt := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)
	promoEndsAt := time.Now().UTC().Add(time.Hour).Truncate(time.Second)

	group, err := client.Group.Create().
		SetName("g-auth-promo").
		SetPlatform(service.PlatformAnthropic).
		SetStatus(service.StatusActive).
		SetSubscriptionType(service.SubscriptionTypeStandard).
		SetRateMultiplier(1).
		SetPromoRateMultiplier(promoRate).
		SetPromoStartsAt(promoStartsAt).
		SetPromoEndsAt(promoEndsAt).
		SetPromoLabel("限时优惠").
		Save(ctx)
	require.NoError(t, err)

	key := &service.APIKey{
		UserID:  user.ID,
		Key:     "sk-getbykey-auth-promo",
		Name:    "Promo Key",
		GroupID: &group.ID,
		Status:  service.StatusActive,
	}
	require.NoError(t, repo.Create(ctx, key))

	got, err := repo.GetByKeyForAuth(ctx, key.Key)
	require.NoError(t, err)
	require.NotNil(t, got.Group)

	require.NotNil(t, got.Group.PromoRateMultiplier, "promo_rate_multiplier 必须被查出，否则限时倍率不生效")
	require.Equal(t, promoRate, *got.Group.PromoRateMultiplier)
	require.NotNil(t, got.Group.PromoStartsAt)
	require.True(t, promoStartsAt.Equal(*got.Group.PromoStartsAt))
	require.NotNil(t, got.Group.PromoEndsAt)
	require.True(t, promoEndsAt.Equal(*got.Group.PromoEndsAt))
	require.Equal(t, "限时优惠", got.Group.PromoLabel)

	// 直接验证业务效果：当前时刻在促销窗口内，计费应取限时倍率而非原倍率。
	require.True(t, got.Group.PromoActiveAt(time.Now()))
	require.Equal(t, promoRate, got.Group.EffectiveRateMultiplier(time.Now()))
}
