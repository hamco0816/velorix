-- Tier style preset for the subscription plan (basic/standard/advanced/flagship/luxury/supreme).
-- Drives the per-tier color + icon in the comparison table (higher = more premium).
-- Existing plans default to 'basic'.
ALTER TABLE subscription_plans
    ADD COLUMN IF NOT EXISTS tier_style VARCHAR(20) NOT NULL DEFAULT 'basic';
