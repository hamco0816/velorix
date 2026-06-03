-- In-app online support chat. Single-agent flow: every user has at most one
-- open conversation, and admins handle all conversations from one workbench.

CREATE TABLE IF NOT EXISTS support_conversations (
    id                 BIGSERIAL PRIMARY KEY,
    user_id            BIGINT       NOT NULL,
    status             VARCHAR(20)  NOT NULL DEFAULT 'open',
    subject            VARCHAR(128) NOT NULL DEFAULT '',
    last_message       TEXT         NOT NULL DEFAULT '',
    last_message_at    TIMESTAMPTZ,
    user_unread_count  INTEGER      NOT NULL DEFAULT 0,
    admin_unread_count INTEGER      NOT NULL DEFAULT 0,
    user_last_read_at  TIMESTAMPTZ,
    admin_last_read_at TIMESTAMPTZ,
    closed_at          TIMESTAMPTZ,
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at         TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_support_conversations_user_open
    ON support_conversations (user_id)
    WHERE status = 'open' AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_support_conversations_user_id
    ON support_conversations (user_id);
CREATE INDEX IF NOT EXISTS idx_support_conversations_status_recent
    ON support_conversations (status, COALESCE(last_message_at, updated_at) DESC);
CREATE INDEX IF NOT EXISTS idx_support_conversations_admin_unread
    ON support_conversations (admin_unread_count);

CREATE TABLE IF NOT EXISTS support_messages (
    id              BIGSERIAL PRIMARY KEY,
    conversation_id BIGINT      NOT NULL,
    sender_type     VARCHAR(20) NOT NULL,
    sender_id       BIGINT,
    content         TEXT        NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_support_messages_conversation_id
    ON support_messages (conversation_id, id DESC);
CREATE INDEX IF NOT EXISTS idx_support_messages_created_at
    ON support_messages (created_at);
