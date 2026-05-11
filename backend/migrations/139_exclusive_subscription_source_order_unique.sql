-- 给 exclusive_subscriptions.source_order_id 加部分唯一索引：
-- 一笔订单（source_order_id）只能创建一份独享名额。
--
-- 业务背景：
--   migration 136 加了 source_order_id 普通索引（用于退款时反查）；
--   但只有索引没有唯一约束，一旦 fulfillment 链路被并发触发或重试时未走 CAS（实际不太可能但理论存在），
--   可能在数据库层重复创建名额。
--
-- 设计要点：
--   - 部分唯一索引（WHERE source_order_id IS NOT NULL AND deleted_at IS NULL）
--     让管理员手工赠送（source_order_id NULL）和软删除记录互不干扰
--   - 跟现有 idx_exclusive_subscription_source_order 不冲突（CREATE 语句是 IF NOT EXISTS）


CREATE UNIQUE INDEX IF NOT EXISTS uk_exclusive_subscription_source_order_active
    ON exclusive_subscriptions (source_order_id)
    WHERE source_order_id IS NOT NULL AND deleted_at IS NULL;

