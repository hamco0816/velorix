-- Preset color tone for the custom subscription plan badge (热门 / 巨量 etc.).
-- Stores a palette key (gold/obsidian/purple/emerald/sapphire/rose); only effective
-- when badge_text is non-empty. Existing badges default to 'gold' to keep the historic amber look.
ALTER TABLE subscription_plans
    ADD COLUMN IF NOT EXISTS badge_color VARCHAR(20) NOT NULL DEFAULT 'gold';
