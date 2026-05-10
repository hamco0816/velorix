-- 给独享名额(exclusive_subscriptions)增加日/周/月用量窗口字段，配合 daily/weekly/monthly_limit_usd
-- 在请求前真正拦截超额，避免限额仅是"展示"。
--
-- 业务背景：
--   - migration 138 给 exclusive_subscriptions 加了 daily/weekly/monthly_limit_usd 快照
--   - 但缺乏对应的 daily_usage_usd / window_start 字段，billing_cache 只能按总 usage_usd 累加，
--     导致即使 seat 上限是日 $30，用户跑到 $100 仍然能继续请求
--   - 这里补上窗口用量字段，service 层 IncrementSeatUsage 时按窗口边界 lazy reset 并累加
--
-- 设计要点：
--   - 窗口用量默认 0；窗口起点 nullable，service 第一次写入时设置
--   - 与 user_subscriptions 同款字段命名/类型，保持一致

BEGIN;

ALTER TABLE exclusive_subscriptions
    ADD COLUMN IF NOT EXISTS daily_window_start   TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS weekly_window_start  TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS monthly_window_start TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS daily_usage_usd      DECIMAL(20, 10) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS weekly_usage_usd     DECIMAL(20, 10) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS monthly_usage_usd    DECIMAL(20, 10) NOT NULL DEFAULT 0;

COMMIT;
