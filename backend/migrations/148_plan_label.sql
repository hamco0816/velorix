-- Tier label for the subscription plan comparison table column header (e.g. Lite / Pro).
-- Empty means the frontend derives it from the plan name (common-prefix stripping).
ALTER TABLE subscription_plans
    ADD COLUMN IF NOT EXISTS plan_label VARCHAR(48) NOT NULL DEFAULT '';
