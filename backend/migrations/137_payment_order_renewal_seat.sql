-- 给 payment_orders 增加 renewal_seat_id 字段：标记一笔订单是否是「续费已有独享名额」。
--
-- 业务背景：v1 续费按钮直接调 RenewSeat 没经过支付（漏洞，可无限免费续期）。
-- v2 续费流程：用户点续费 → 创建续费订单 → 走支付 → fulfillment 检测此字段
-- 调 RenewSeat（保留账号、延期）而不是 AssignSeat（不消耗库存）。
--
-- 该字段为空（NULL）的订单是普通新购订单（独享/共享均可）。


ALTER TABLE payment_orders
    ADD COLUMN IF NOT EXISTS renewal_seat_id BIGINT;

CREATE INDEX IF NOT EXISTS idx_payment_orders_renewal_seat
    ON payment_orders (renewal_seat_id)
    WHERE renewal_seat_id IS NOT NULL;

