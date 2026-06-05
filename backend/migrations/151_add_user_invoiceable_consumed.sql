-- 用户累计「支持开票分组」的余额消费额度（与 balance 同单位：站内额度/USD）。
-- 记账时实时累加，不受用量日志清理影响。开票时按当前充值倍率折算成人民币，
-- 并用真实充值付费金额封顶，作为余额消费的可开票上限。
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS invoiceable_consumed DECIMAL(20,8) NOT NULL DEFAULT 0;
