-- 给套餐(plan)增加可覆盖 group 限额/倍率的字段，并把对应快照位置加到订阅记录上。
--
-- 业务背景：
--   - group 上原有 daily/weekly/monthly_limit_usd + rate_multiplier，是"分组一刀切"
--   - 现在要做"同 group 下不同套餐有不同额度"（如 Lite ¥90 月限 $30 / Pro ¥240 月限 $100 / Max ¥700 月限 $400）
--   - 解法：plan 上加可空字段；用户购买时把 plan 当时的值快照到订阅记录（user_subscriptions / exclusive_subscriptions）
--   - 调度时优先读订阅记录里的快照值，NULL 才回落到 group
--
-- 设计要点：
--   - 全部 nullable，NULL = "使用 group 默认值"，向后兼容存量
--   - 快照在订阅记录上，避免日后 plan 改限额影响已购买用户

BEGIN;

-- subscription_plans: 套餐自身可覆盖的限额/倍率
ALTER TABLE subscription_plans
    ADD COLUMN IF NOT EXISTS daily_limit_usd   DECIMAL(20, 8),
    ADD COLUMN IF NOT EXISTS weekly_limit_usd  DECIMAL(20, 8),
    ADD COLUMN IF NOT EXISTS monthly_limit_usd DECIMAL(20, 8),
    ADD COLUMN IF NOT EXISTS rate_multiplier   DECIMAL(10, 4);

-- user_subscriptions: 快照购买时的 plan 限额；NULL 时回落到 group
ALTER TABLE user_subscriptions
    ADD COLUMN IF NOT EXISTS daily_limit_usd   DECIMAL(20, 8),
    ADD COLUMN IF NOT EXISTS weekly_limit_usd  DECIMAL(20, 8),
    ADD COLUMN IF NOT EXISTS monthly_limit_usd DECIMAL(20, 8),
    ADD COLUMN IF NOT EXISTS rate_multiplier   DECIMAL(10, 4);

-- exclusive_subscriptions: 独享池同样支持
ALTER TABLE exclusive_subscriptions
    ADD COLUMN IF NOT EXISTS daily_limit_usd   DECIMAL(20, 8),
    ADD COLUMN IF NOT EXISTS weekly_limit_usd  DECIMAL(20, 8),
    ADD COLUMN IF NOT EXISTS monthly_limit_usd DECIMAL(20, 8),
    ADD COLUMN IF NOT EXISTS rate_multiplier   DECIMAL(10, 4);

COMMIT;
