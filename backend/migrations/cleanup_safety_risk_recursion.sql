-- 清理 AI 审核递归 BUG 产生的脏数据
--
-- 背景：safety_risk_ai_review.callAIReviewLLM 把待审核 prompt 当 user content 发回
-- 本地 gateway，body 必然含敏感词。中间件不识别这条内部回环 → 再写事件 → 再触发审核
-- → 形成递归。线上一次测试点击撑出 1.4M 条记录。
--
-- 代码侧已修复（X-Internal-AI-Review header + 中间件放行）。这个脚本只清历史脏数据。
--
-- !!! 执行前注意 !!!
--   1. 先备份：pg_dump -t safety_risk_events <db> > safety_risk_events_backup.sql
--   2. 检查 SELECT 计数对得上预期，再换成 DELETE 执行
--   3. 建议在低峰期跑，1.4M 行 DELETE 会持锁数秒

-- ============================================================
-- 第一步：先 SELECT 看清要删多少行
-- ============================================================

-- 1.1 看总数 & 各规则命中分布（确认是哪些规则在递归）
SELECT rule_word, action, COUNT(*) AS hits
FROM safety_risk_events
WHERE created_at >= NOW() - INTERVAL '7 days'
GROUP BY rule_word, action
ORDER BY hits DESC;

-- 1.2 看是不是同一秒内连续 ID 暴增（递归特征）
SELECT
  date_trunc('minute', created_at) AS minute,
  COUNT(*) AS hits,
  COUNT(DISTINCT user_id) AS users
FROM safety_risk_events
WHERE created_at >= NOW() - INTERVAL '24 hours'
GROUP BY minute
ORDER BY hits DESC
LIMIT 20;

-- ============================================================
-- 第二步：清理（按递增激进程度提供 3 种方案，选一种执行）
-- ============================================================

-- 方案 A（推荐）：清掉受影响规则的 warn/block 事件
-- 这三条规则是 builtin，且都被 AI 审核 few-shot 样本带过敏感词命中过
-- BEGIN;
-- DELETE FROM safety_risk_events
-- WHERE rule_word IN (
--   'prompt injection',
--   'ignore all previous instructions',
--   'ransomware'
-- )
-- AND created_at < NOW();   -- 全部历史；如果想保留最近的真实事件，把 < NOW() 改成更早的时间界
-- -- 看下删了多少行：
-- SELECT 'deleted rows: ' || COUNT(*) FROM safety_risk_events;
-- COMMIT;

-- 方案 B：清掉某个用户的所有事件（你说"用户都是我们自己"，可以按 user_id 清）
-- 先查 user_id：
-- SELECT id, email FROM users WHERE email = '3030758482@qq.com';
-- 然后：
-- BEGIN;
-- DELETE FROM safety_risk_events WHERE user_id = <填上面查出的 id>;
-- COMMIT;

-- 方案 C（最激进）：整张表清空，重新开始
-- 1.4M 行 DELETE 较慢，TRUNCATE 秒级；但 TRUNCATE 不可回滚，且会重置自增 ID
-- BEGIN;
-- TRUNCATE TABLE safety_risk_events RESTART IDENTITY;
-- COMMIT;

-- ============================================================
-- 第三步：（可选）回收磁盘空间
-- ============================================================
-- DELETE 不会立刻释放表空间，autovacuum 会逐步回收。要立即回收：
-- VACUUM FULL safety_risk_events;
-- 注意：VACUUM FULL 期间表会被独占锁，建议低峰期执行。
