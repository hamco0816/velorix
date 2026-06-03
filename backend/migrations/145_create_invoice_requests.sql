-- 发票功能：用户对「已完成」付费订单申请开票，管理员开票后留存发票号码等元数据。
-- 金额口径：amount 取所选订单 pay_amount（用户实付现金）之和，不含赠送、不含折扣。

-- 发票申请单表
CREATE TABLE IF NOT EXISTS invoice_requests (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT        NOT NULL,
    user_email      VARCHAR(255)  NOT NULL,
    user_name       VARCHAR(100)  NOT NULL DEFAULT '',
    recipient_email VARCHAR(255)  NOT NULL,            -- 接收发票的邮箱
    title_type      VARCHAR(20)   NOT NULL DEFAULT 'personal', -- personal 个人 / company 企业
    title_name      VARCHAR(255)  NOT NULL,            -- 抬头名称
    tax_id          VARCHAR(64),                       -- 纳税人识别号（企业必填）
    user_remark     TEXT,                              -- 用户备注（开票内容说明等）
    amount          DECIMAL(20,2) NOT NULL,            -- 申请开票总金额（所选订单 pay_amount 之和）
    status          VARCHAR(20)   NOT NULL DEFAULT 'pending', -- pending/issued/rejected/cancelled
    invoice_number  VARCHAR(64),                       -- 发票号码（开票后留存，PDF 文件本身不入库）
    invoice_date    TIMESTAMPTZ,                       -- 开票日期
    invoice_amount  DECIMAL(20,2),                     -- PDF 识别出的开票金额（与 amount 比对核验）
    issued_at       TIMESTAMPTZ,                       -- 标记已开票时间
    issued_by       BIGINT,                            -- 开票操作管理员 ID
    reject_reason   TEXT,                              -- 驳回原因
    email_sent      BOOLEAN       NOT NULL DEFAULT FALSE, -- 发票邮件是否发送成功
    email_error     TEXT,                              -- 邮件发送失败原因
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_invoice_requests_user_id ON invoice_requests (user_id);
CREATE INDEX IF NOT EXISTS idx_invoice_requests_status ON invoice_requests (status);
CREATE INDEX IF NOT EXISTS idx_invoice_requests_created_at ON invoice_requests (created_at);

-- 订单关联发票申请单：非空表示该订单已被占用（待开票或已开票），不能重复开票。
-- 申请单被驳回/取消时置回 NULL，订单可重新申请。
ALTER TABLE payment_orders ADD COLUMN IF NOT EXISTS invoice_request_id BIGINT;
CREATE INDEX IF NOT EXISTS idx_payment_orders_invoice_request_id ON payment_orders (invoice_request_id);
