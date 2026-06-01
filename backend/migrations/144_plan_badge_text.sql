-- Custom display badge for subscription plans.
-- Existing "popular" plans keep is_popular=true and let clients render a localized fallback,
-- while new plans can use arbitrary short labels such as 热门 / 最多 / 最新.
ALTER TABLE subscription_plans
    ADD COLUMN IF NOT EXISTS badge_text VARCHAR(48) NOT NULL DEFAULT '';
