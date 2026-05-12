-- Group 限时倍率（promo rate）：admin 在某段时间内给分组配置不同倍率，
-- 用户端展示划线价 + 倒计时，billing 在限时区间内自动用 promo 倍率计费。
--
-- 4 个新字段都是可空：
--   promo_rate_multiplier  限时倍率（如 0.50 = 5 折），仅在 starts_at <= now < ends_at 时生效
--   promo_starts_at        限时开始时间（NULL 表示立即生效）
--   promo_ends_at          限时结束时间（NULL 表示永久 — 但通常应配置）
--   promo_label            活动名称（如 "618 大促"），用于前端展示

ALTER TABLE groups
    ADD COLUMN IF NOT EXISTS promo_rate_multiplier decimal(10,4),
    ADD COLUMN IF NOT EXISTS promo_starts_at       timestamptz,
    ADD COLUMN IF NOT EXISTS promo_ends_at         timestamptz,
    ADD COLUMN IF NOT EXISTS promo_label           varchar(100);

-- 查询索引：billing 流程要快速判断当前是否在限时窗口内
CREATE INDEX IF NOT EXISTS idx_groups_promo_window
    ON groups (promo_starts_at, promo_ends_at)
    WHERE promo_rate_multiplier IS NOT NULL;
