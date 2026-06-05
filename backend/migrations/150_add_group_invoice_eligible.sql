-- 分组开票开关：该分组的消费是否可开票。
-- 默认 false：自对接的低利润分组不开票；只有自建号池等高利润分组手动开启后，
-- 其消费（余额按量 + 套餐购买）才计入用户的可开票额度。
ALTER TABLE groups
    ADD COLUMN IF NOT EXISTS invoice_eligible BOOLEAN NOT NULL DEFAULT FALSE;
