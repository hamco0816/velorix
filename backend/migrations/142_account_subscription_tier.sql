-- 给 accounts 表加 subscription_tier 字段，用于订阅定价助手按档位聚合用量统计
-- 取值约定（按 platform 分）：
--   openai: free / plus / pro_5x / pro_20x / team
--   anthropic: free / pro / max_5x / max_20x / team
--   gemini: free / pro / ultra
-- 空字符串表示未标记，统计时单独归类为"未分类"
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS subscription_tier VARCHAR(32) NOT NULL DEFAULT '';
