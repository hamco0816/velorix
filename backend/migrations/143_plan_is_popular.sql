-- 给 subscription_plans 加 is_popular 主推标记字段
-- 用途：admin 标记某档套餐为"主推"，前端订阅卡角上显示 ⭐ 徽章，提升主推档转化率
-- 不影响排序逻辑（sort_order 仍然是显示顺序的唯一来源）
ALTER TABLE subscription_plans ADD COLUMN IF NOT EXISTS is_popular BOOLEAN NOT NULL DEFAULT FALSE;
