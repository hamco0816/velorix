-- 独享池：新建 exclusive_subscriptions 表 + 扩展 subscription_plans / accounts
--
-- 一份 exclusive_subscription 代表「一个用户独占一个上游账号一段时间」的名额。
-- 同一账号同时只能被一份 active 名额占用（部分唯一索引保证）。
-- 用户可以同时持有多份独享名额（同池或跨池），调度器在用户名下的多个名额之间负载均衡。

BEGIN;

-- 1. 主表
CREATE TABLE IF NOT EXISTS exclusive_subscriptions (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL,
    group_id        BIGINT NOT NULL,
    plan_id         BIGINT NOT NULL,
    account_id      BIGINT NOT NULL,
    -- 状态：active / expired / refunded / cancelled
    status          VARCHAR(20) NOT NULL DEFAULT 'active',
    starts_at       TIMESTAMPTZ NOT NULL,
    expires_at      TIMESTAMPTZ NOT NULL,
    assigned_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- 最近一次续费时间，便于审计与界面展示
    last_renewal_at TIMESTAMPTZ,
    -- 累计用量统计（USD），用于后台展示和统计分析
    usage_usd       DECIMAL(20, 10) NOT NULL DEFAULT 0,
    notes           TEXT,
    -- 管理员赠送时记录管理员 ID；正常支付分配为 NULL
    assigned_by     BIGINT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

-- 同账号同时只能被一份 active 名额占用：核心数据完整性约束
CREATE UNIQUE INDEX IF NOT EXISTS uk_excl_sub_active_account
    ON exclusive_subscriptions (account_id)
    WHERE status = 'active' AND deleted_at IS NULL;

-- 调度热路径：根据 user_id + group_id 查活跃 seat
CREATE INDEX IF NOT EXISTS idx_excl_sub_user_active_group
    ON exclusive_subscriptions (user_id, status, group_id)
    WHERE deleted_at IS NULL;

-- 过期回收任务扫描
CREATE INDEX IF NOT EXISTS idx_excl_sub_expiry
    ON exclusive_subscriptions (status, expires_at)
    WHERE deleted_at IS NULL;

-- 普通查询索引
CREATE INDEX IF NOT EXISTS idx_excl_sub_user_id
    ON exclusive_subscriptions (user_id)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_excl_sub_group_id
    ON exclusive_subscriptions (group_id)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_excl_sub_plan_id
    ON exclusive_subscriptions (plan_id)
    WHERE deleted_at IS NULL;

-- 2. subscription_plans 加 kind 字段
-- 区分「共享套餐」与「独享套餐」；存量数据默认 shared，行为不变
ALTER TABLE subscription_plans
    ADD COLUMN IF NOT EXISTS kind VARCHAR(20) NOT NULL DEFAULT 'shared';

-- 3. accounts 加 assigned_seat_id 反向索引（冗余，加速后台库存看板）
-- 取值含义：NULL = 该账号空闲、可被分配；非 NULL = 当前正被某份独享名额占用
ALTER TABLE accounts
    ADD COLUMN IF NOT EXISTS assigned_seat_id BIGINT;

CREATE INDEX IF NOT EXISTS idx_accounts_assigned_seat
    ON accounts (assigned_seat_id)
    WHERE assigned_seat_id IS NOT NULL AND deleted_at IS NULL;

COMMIT;
