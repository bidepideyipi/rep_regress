-- ============================================
-- RTP 定时批处理聚合脚本
-- 说明：替代物化视图，在应用层定时执行聚合
-- 建议执行频率：每 5 分钟执行一次
-- ============================================

USE rtp_analytics;

-- ============================================
-- 1. 用户 RTP 实时聚合（5分钟粒度）
-- ============================================
INSERT INTO rtp_user_realtime_local
SELECT
    user_id,
    integrator_id,
    game_id,
    toStartOfFiveMinutes(log_time) AS time_window,

    sum(bet_amount) AS total_bet,
    sum(win_amount) AS total_win,
    sum(win_amount - bet_amount) AS net_result,
    if(sum(bet_amount) > 0, round(sum(win_amount) / sum(bet_amount), 4), 0) AS rtp,

    count() AS spin_count,
    avg(bet_amount) AS avg_bet,
    max(win_amount) AS max_win,

    now() AS last_update
FROM game_log_detail_local
WHERE error_code = 0
  AND log_time >= now() - INTERVAL 5 MINUTE
GROUP BY
    user_id,
    integrator_id,
    game_id,
    toStartOfFiveMinutes(log_time);

-- ============================================
-- 2. 用户 RTP 小时聚合
-- ============================================
INSERT INTO rtp_user_metrics_local
SELECT
    user_id,
    integrator_id,
    game_id,
    toStartOfHour(log_time) AS time_window,

    sum(bet_amount) AS total_bet,
    sum(win_amount) AS total_win,
    sum(win_amount - bet_amount) AS net_result,
    if(sum(bet_amount) > 0, round(sum(win_amount) / sum(bet_amount), 4), 0) AS rtp,

    count() AS spin_count,
    avg(bet_amount) AS avg_bet,
    max(bet_amount) AS max_bet,
    max(win_amount) AS max_win,
    min(win_amount) AS min_win,

    if(count() > 0, round(countIf(win_amount > 0) / count(), 4), 0) AS win_rate,
    countIf(win_amount > 0) AS win_spin_count,
    countIf(win_amount = 0) AS loss_spin_count,

    min(log_time) AS first_spin_time,
    max(log_time) AS last_spin_time,
    now() AS last_update
FROM game_log_detail_local
WHERE error_code = 0
  AND log_time >= now() - INTERVAL 1 HOUR
GROUP BY
    user_id,
    integrator_id,
    game_id,
    toStartOfHour(log_time);

-- ============================================
-- 3. 游戏 RTP 实时聚合（5分钟粒度）
-- ============================================
INSERT INTO rtp_game_realtime_local
SELECT
    game_id,
    integrator_id,
    toStartOfFiveMinutes(log_time) AS time_window,

    sum(bet_amount) AS total_bet,
    sum(win_amount) AS total_win,
    if(sum(bet_amount) > 0, round(sum(win_amount) / sum(bet_amount), 4), 0) AS rtp,
    count() AS spin_count,
    uniqExact(user_id) AS active_users,
    avg(bet_amount) AS avg_bet,

    now() AS last_update
FROM game_log_detail_local
WHERE error_code = 0
  AND log_time >= now() - INTERVAL 5 MINUTE
GROUP BY
    game_id,
    integrator_id,
    toStartOfFiveMinutes(log_time);

-- ============================================
-- 4. 游戏 RTP 小时聚合
-- ============================================
INSERT INTO rtp_game_metrics_local
SELECT
    game_id,
    integrator_id,
    toStartOfHour(log_time) AS time_window,

    sum(bet_amount) AS total_bet,
    sum(win_amount) AS total_win,
    sum(win_amount - bet_amount) AS net_result,
    if(sum(bet_amount) > 0, round(sum(win_amount) / sum(bet_amount), 4), 0) AS rtp,

    count() AS spin_count,
    uniqExact(user_id) AS active_users,
    avg(bet_amount) AS avg_bet,
    max(bet_amount) AS max_bet,
    max(win_amount) AS max_win,

    stddevPop(win_amount - bet_amount) AS volatility,
    if(count() > 0, round(countIf(win_amount > 0) / count(), 4), 0) AS win_rate,
    avg(win_amount - bet_amount) AS avg_net_result,

    now() AS last_update
FROM game_log_detail_local
WHERE error_code = 0
  AND log_time >= now() - INTERVAL 1 HOUR
GROUP BY
    game_id,
    integrator_id,
    toStartOfHour(log_time);

-- ============================================
-- 5. RTP 异常告警检测（每10分钟执行一次）
-- ============================================
INSERT INTO rtp_alerts_local
SELECT
    concat('alert_', toString(now()), '_', user_id, '_', game_id) AS alert_id,
    'user_high_rtp' AS alert_type,
    'high' AS severity,

    user_id,
    game_id,
    integrator_id,

    rtp,
    total_bet,
    total_win,
    spin_count,

    time_window,
    now() AS detected_time,
    CAST(NULL AS Nullable(DateTime)) AS resolved_time,
    'active' AS status,
    concat('用户 RTP 过高: ', toString(rtp), ' (阈值: 0.98)') AS notes
FROM rtp_user_realtime_local
WHERE rtp > 0.98
  AND spin_count >= 20
  AND total_bet >= 100
  AND time_window >= now() - INTERVAL 10 MINUTE;

INSERT INTO rtp_alerts_local
SELECT
    concat('alert_', toString(now()), '_', user_id, '_', game_id) AS alert_id,
    'user_low_rtp' AS alert_type,
    'medium' AS severity,

    user_id,
    game_id,
    integrator_id,

    rtp,
    total_bet,
    total_win,
    spin_count,

    time_window,
    now() AS detected_time,
    CAST(NULL AS Nullable(DateTime)) AS resolved_time,
    'active' AS status,
    concat('用户 RTP 过低: ', toString(rtp), ' (阈值: 0.50)') AS notes
FROM rtp_user_realtime_local
WHERE rtp < 0.50
  AND spin_count >= 20
  AND total_bet >= 100
  AND time_window >= now() - INTERVAL 10 MINUTE;

INSERT INTO rtp_alerts_local
SELECT
    concat('alert_', toString(now()), '_', game_id) AS alert_id,
    'game_abnormal' AS alert_type,
    'high' AS severity,

    '' AS user_id,
    game_id,
    integrator_id,

    rtp,
    total_bet,
    total_win,
    spin_count,

    time_window,
    now() AS detected_time,
    CAST(NULL AS Nullable(DateTime)) AS resolved_time,
    'active' AS status,
    concat('游戏整体 RTP 异常: ', toString(rtp), ' (预期: 0.95)') AS notes
FROM rtp_game_realtime_local
WHERE (rtp > 1.05 OR rtp < 0.85)
  AND spin_count >= 100
  AND time_window >= now() - INTERVAL 10 MINUTE;

SELECT 'RTP 批处理聚合完成！' AS status, now() AS execution_time;
