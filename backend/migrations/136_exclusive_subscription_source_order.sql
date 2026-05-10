-- 给独享名额增加 source_order_id 字段：精确反查"哪笔订单分配的这个 seat"
--
-- 背景：v1 退款释放 seat 用 (user_id, plan_id) 反查最近一份 active seat，
-- 用户同时买多份同 plan 退一份时可能释放错 seat。
-- 加上 source_order_id 后，退款时通过订单 ID 精确定位对应 seat。

BEGIN;

ALTER TABLE exclusive_subscriptions
    ADD COLUMN IF NOT EXISTS source_order_id BIGINT;

CREATE INDEX IF NOT EXISTS idx_excl_sub_source_order
    ON exclusive_subscriptions (source_order_id)
    WHERE source_order_id IS NOT NULL AND deleted_at IS NULL;

COMMIT;
