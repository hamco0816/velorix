-- Safety risk events captured before requests are forwarded upstream.
-- Store only a short prompt preview, never the full raw request body.

CREATE TABLE IF NOT EXISTS safety_risk_events (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    user_id BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    api_key_id BIGINT NULL REFERENCES api_keys(id) ON DELETE SET NULL,
    api_key_name TEXT NOT NULL DEFAULT '',
    group_id BIGINT NULL REFERENCES groups(id) ON DELETE SET NULL,
    group_name TEXT NOT NULL DEFAULT '',

    request_id TEXT NOT NULL DEFAULT '',
    client_request_id TEXT NOT NULL DEFAULT '',
    method TEXT NOT NULL DEFAULT '',
    path TEXT NOT NULL DEFAULT '',
    client_ip TEXT NOT NULL DEFAULT '',
    user_agent TEXT NOT NULL DEFAULT '',

    rule_source TEXT NOT NULL DEFAULT 'local',
    rule_word TEXT NOT NULL DEFAULT '',
    rule_path TEXT NOT NULL DEFAULT '',
    category TEXT NOT NULL DEFAULT 'content_safety',
    severity TEXT NOT NULL DEFAULT 'warning',
    action TEXT NOT NULL DEFAULT 'blocked',

    ai_reviewed BOOLEAN NOT NULL DEFAULT FALSE,
    ai_review_provider TEXT NOT NULL DEFAULT '',
    ai_review_result TEXT NOT NULL DEFAULT 'not_used',

    status TEXT NOT NULL DEFAULT 'pending',
    prompt_preview TEXT NOT NULL DEFAULT '',
    reviewed_by_user_id BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMPTZ NULL,
    review_note TEXT NOT NULL DEFAULT '',
    cleared_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_safety_risk_events_created_at
    ON safety_risk_events (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_safety_risk_events_status_created_at
    ON safety_risk_events (status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_safety_risk_events_user_created_at
    ON safety_risk_events (user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_safety_risk_events_api_key_created_at
    ON safety_risk_events (api_key_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_safety_risk_events_group_created_at
    ON safety_risk_events (group_id, created_at DESC);
