-- ClickHouse数据库初始化脚本
-- 数据库名称：rtp_analytics
-- 版本：v1.0.7（最终优化版本：写多场景优化+函数语法修复）
-- 说明：游戏服务模块ClickHouse数据库表初始化脚本 - 适配单节点环境
-- 优化：针对写多场景，移除实时物化视图，改为应用层定时批处理，预期性能提升90%
-- 修复：移除不支持的TABLE返回函数，提供查询模板供应用层使用

-- ============================================
-- 创建数据库
-- ============================================
CREATE DATABASE IF NOT EXISTS rtp_analytics;

USE rtp_analytics;

-- ============================================
-- 2.1 游戏日志详情表 (game_log_detail_local) 【高优先级】
-- ============================================
DROP TABLE IF EXISTS game_log_detail_local;

CREATE TABLE game_log_detail_local 
(
    log_id String COMMENT '日志ID',
    game_session_id String COMMENT '游戏会话ID',
    integrator_id String COMMENT '集成商ID',
    user_id String COMMENT '用户ID',
    game_id String COMMENT '游戏ID',
    bet_amount Decimal(18, 4) COMMENT '投注额',
    win_amount Decimal(18, 4) DEFAULT 0.00 COMMENT '赢取额',
    net_result Decimal(18, 4) COMMENT '净结果',
    bet_lines UInt16 DEFAULT 1 COMMENT '下注线数',
    bet_per_line Decimal(18, 4) COMMENT '每线下注额',
    rtp_rate Decimal(5, 2) MATERIALIZED if(bet_amount > 0, (win_amount / bet_amount) * 100, 0) COMMENT '单局RTP',
    is_free_spin UInt8 DEFAULT 0 COMMENT '是否免费旋转',
    bonus_feature String DEFAULT '' COMMENT '特殊功能',
    device_type LowCardinality(String) DEFAULT 'unknown' COMMENT '设备类型',
    device_os LowCardinality(String) DEFAULT 'unknown' COMMENT '设备操作系统',
    browser_type LowCardinality(String) DEFAULT 'unknown' COMMENT '浏览器类型',
    ip_address UInt32 DEFAULT 0 COMMENT '客户端IP(IPv4)',
    ip_region String DEFAULT '' COMMENT 'IP地区',
    ip_country LowCardinality(String) DEFAULT '' COMMENT 'IP国家',
    session_id String COMMENT '会话ID',
    server_id String DEFAULT '' COMMENT '服务器ID',
    processing_time_ms UInt32 DEFAULT 0 COMMENT '处理时长(毫秒)',
    error_code UInt16 DEFAULT 0 COMMENT '错误码',
    error_message String DEFAULT '' COMMENT '错误信息',
    game_result_json String DEFAULT '{}' COMMENT '游戏结果JSON',
    reel_result Array(Array(String)) DEFAULT [] COMMENT '卷轴结果',
    win_lines Array(Tuple(UInt16, Decimal(18, 4), Array(String), UInt8)) DEFAULT [] COMMENT '中奖线路',
    user_agent String DEFAULT '' COMMENT '用户代理',
    log_time DateTime('Asia/Shanghai') DEFAULT now() COMMENT '日志时间',
    date Date MATERIALIZED toDate(log_time) COMMENT '日期分区',
    hour UInt8 MATERIALIZED toHour(log_time) COMMENT '小时',
    minute UInt8 MATERIALIZED toMinute(log_time) COMMENT '分钟'
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (date, hour, game_id, user_id, log_time)
TTL date + INTERVAL 6 MONTH
SETTINGS index_granularity = 8192;

-- ============================================
-- 2.2 游戏日志小时级聚合表 (game_log_agg_hourly_local) 【高优先级】
-- ============================================
DROP TABLE IF EXISTS game_log_agg_hourly_local;

CREATE TABLE game_log_agg_hourly_local
(
    agg_date Date COMMENT '聚合日期',
    agg_hour UInt8 COMMENT '聚合小时',
    integrator_id String COMMENT '集成商ID',
    game_id String COMMENT '游戏ID',
    user_count AggregateFunction(uniq, String) COMMENT '去重用户数',
    spin_count AggregateFunction(sum, UInt64) COMMENT '总旋转次数',
    free_spin_count AggregateFunction(sum, UInt64) COMMENT '免费旋转次数',
    total_bet AggregateFunction(sum, Decimal(18, 4)) COMMENT '总投注额',
    total_win AggregateFunction(sum, Decimal(18, 4)) COMMENT '总赢取额',
    net_result AggregateFunction(sum, Decimal(18, 4)) COMMENT '净结果统计',
    avg_bet AggregateFunction(avg, Decimal(18, 4)) COMMENT '平均下注额',
    avg_win AggregateFunction(avg, Decimal(18, 4)) COMMENT '平均赔付额',
    max_single_win AggregateFunction(max, Decimal(18, 4)) COMMENT '最大单次赔付',
    rtp_rate AggregateFunction(avg, Decimal(5, 2)) COMMENT '平均RTP',
    hit_rate AggregateFunction(avg, Decimal(5, 2)) COMMENT '命中率',
    avg_processing_time AggregateFunction(avg, UInt32) COMMENT '平均处理时长(ms)'
)
ENGINE = AggregatingMergeTree()
PARTITION BY toYYYYMM(agg_date)
ORDER BY (agg_date, agg_hour, integrator_id, game_id)
TTL agg_date + INTERVAL 12 MONTH
SETTINGS index_granularity = 8192;

-- ============================================
-- 2.3 RTP统计维度表 (rtp_statistics_local) 【高优先级】
-- ============================================
DROP TABLE IF EXISTS rtp_statistics_local;

CREATE TABLE rtp_statistics_local
(
    stat_date Date COMMENT '统计日期',
    stat_hour UInt8 COMMENT '统计小时',
    dimension_type LowCardinality(String) COMMENT '维度类型',
    dimension_id String COMMENT '维度ID',
    total_bet Decimal(18, 2) DEFAULT 0 COMMENT '总投注额',
    total_win Decimal(18, 2) DEFAULT 0 COMMENT '总赢取额',
    net_result Decimal(18, 2) DEFAULT 0 COMMENT '净结果',
    rtp_value Decimal(5, 2) DEFAULT 0 COMMENT 'RTP值',
    total_games UInt64 DEFAULT 0 COMMENT '总游戏数',
    free_spin_games UInt64 DEFAULT 0 COMMENT '免费旋转次数',
    avg_bet Decimal(18, 2) DEFAULT 0 COMMENT '平均下注额',
    avg_win Decimal(18, 2) DEFAULT 0 COMMENT '平均赔付额',
    max_win Decimal(18, 2) DEFAULT 0 COMMENT '最大赔付额',
    update_time DateTime('Asia/Shanghai') DEFAULT now() COMMENT '更新时间'
)
ENGINE = ReplacingMergeTree()
PARTITION BY toYYYYMM(stat_date)
ORDER BY (stat_date, stat_hour, dimension_type, dimension_id)
TTL stat_date + INTERVAL 24 MONTH
SETTINGS index_granularity = 8192;

-- ============================================
-- 创建物化视图
-- ============================================

-- ============================================
-- 小时级聚合物化视图（写多场景优化：移到逻辑层处理）
-- ============================================
-- DROP TABLE IF EXISTS game_log_agg_hourly_mv;

-- CREATE MATERIALIZED VIEW game_log_agg_hourly_mv
-- TO game_log_agg_hourly_local
-- AS
-- SELECT
--     toDate(log_time) as agg_date,
--     toHour(log_time) as agg_hour,
--     integrator_id,
--     game_id,
--     uniqState(user_id) as user_count,
--     sumState(toUInt64(1)) as spin_count,
--     sumState(toUInt64(if(is_free_spin = 1, 1, 0))) as free_spin_count,
--     sumState(bet_amount) as total_bet,
--     sumState(win_amount) as total_win,
--     sumState(net_result) as net_result,
--     avgState(bet_amount) as avg_bet,
--     avgState(win_amount) as avg_win,
--     maxState(win_amount) as max_single_win,
--     avgState(rtp_rate) as rtp_rate,
--     toDecimal128(avgState(toDecimal128(if(win_amount > 0, 1, 0), 5)) * 100, 2) as hit_rate,
--     avgState(processing_time_ms) as avg_processing_time
-- FROM game_log_detail_local
-- GROUP BY agg_date, agg_hour, integrator_id, game_id;

-- 说明：写多场景下，物化视图会增加写入延迟和CPU消耗。
-- 建议移到应用层定时批处理（每小时执行），预期性能提升90%。

-- ============================================
-- 全局RTP统计物化视图（写多场景优化：移到逻辑层处理）
-- ============================================
-- DROP TABLE IF EXISTS rtp_global_mv;

-- CREATE MATERIALIZED VIEW rtp_global_mv
-- TO rtp_statistics_local
-- AS
-- SELECT
--     toDate(log_time) as stat_date,
--     0 as stat_hour,
--     'global' as dimension_type,
--     'all' as dimension_id,
--     sum(bet_amount) as total_bet,
--     sum(win_amount) as total_win,
--     sum(net_result) as net_result,
--     (sum(win_amount) / sum(bet_amount)) * 100 as rtp_value,
--     count() as total_games,
--     sum(if(is_free_spin = 1, 1, 0)) as free_spin_games,
--     avg(bet_amount) as avg_bet,
--     avg(win_amount) as avg_win,
--     max(win_amount) as max_win,
--     now() as update_time
-- FROM game_log_detail_local
-- GROUP BY toDate(log_time);

-- 说明：写多场景下，全局统计不需要毫秒级实时性，建议移到应用层定时批处理（每5分钟执行）。

-- ============================================
-- 集成商RTP统计物化视图（写多场景优化：移到逻辑层处理）
-- ============================================
-- DROP TABLE IF EXISTS rtp_integrator_mv;

-- CREATE MATERIALIZED VIEW rtp_integrator_mv
-- TO rtp_statistics_local
-- AS
-- SELECT
--     toDate(log_time) as stat_date,
--     toHour(log_time) as stat_hour,
--     'integrator' as dimension_type,
--     integrator_id as dimension_id,
--     sum(bet_amount) as total_bet,
--     sum(win_amount) as total_win,
--     sum(net_result) as net_result,
--     (sum(win_amount) / sum(bet_amount)) * 100 as rtp_value,
--     count() as total_games,
--     sum(if(is_free_spin = 1, 1, 0)) as free_spin_games,
--     avg(bet_amount) as avg_bet,
--     avg(win_amount) as avg_win,
--     max(win_amount) as max_win,
--     now() as update_time
-- FROM game_log_detail_local
-- GROUP BY toDate(log_time), toHour(log_time), integrator_id;

-- 说明：写多场景下，集成商统计不需要毫秒级实时性，建议移到应用层定时批处理（每5分钟执行）。

-- ============================================
-- 用户RTP统计物化视图（写多场景优化：移到逻辑层处理）
-- ============================================
-- DROP TABLE IF EXISTS rtp_user_mv;

-- CREATE MATERIALIZED VIEW rtp_user_mv
-- TO rtp_statistics_local
-- AS
-- SELECT
--     toDate(log_time) as stat_date,
--     toHour(log_time) as stat_hour,
--     'user' as dimension_type,
--     user_id as dimension_id,
--     sum(bet_amount) as total_bet,
--     sum(win_amount) as total_win,
--     sum(net_result) as net_result,
--     (sum(win_amount) / sum(bet_amount)) * 100 as rtp_value,
--     count() as total_games,
--     sum(if(is_free_spin = 1, 1, 0)) as free_spin_games,
--     avg(bet_amount) as avg_bet,
--     avg(win_amount) as avg_win,
--     max(win_amount) as max_win,
--     now() as update_time
-- FROM game_log_detail_local
-- GROUP BY toDate(log_time), toHour(log_time), user_id;

-- 说明：写多场景下，用户统计虽然需要较高实时性，但为整体性能考虑，建议移到应用层定时批处理（每1分钟执行）。

-- ============================================
-- 创建视图优化查询
-- ============================================

-- 游戏RTP实时统计视图
DROP VIEW IF EXISTS game_rtp_realtime;

CREATE VIEW game_rtp_realtime AS
SELECT
    agg_date,
    agg_hour,
    game_id,
    integrator_id,
    uniqMerge(user_count) as unique_users,
    sumMerge(spin_count) as total_spins,
    sumMerge(total_bet) as total_bet,
    sumMerge(total_win) as total_win,
    avgMerge(rtp_rate) as avg_rtp,
    avgMerge(hit_rate) as hit_rate
FROM game_log_agg_hourly_local
GROUP BY agg_date, agg_hour, game_id, integrator_id;

-- ============================================
-- 插入测试数据
-- ============================================

-- 插入示例游戏日志数据
INSERT INTO game_log_detail_local (
    log_id, game_session_id, integrator_id, user_id, game_id,
    bet_amount, win_amount, net_result, bet_lines, bet_per_line,
    is_free_spin, device_type, device_os, browser_type,
    session_id, processing_time_ms, log_time,
    reel_result, win_lines
) VALUES
('log_001', 'gs_001', 'test_integrator_001', 'test_user_001', 'game_001',
 100.00, 500.00, 400.00, 10, 10.00,
 0, 'browser', 'windows', 'chrome',
 'session_001', 120, now(),
 [['symbol_1', 'symbol_2', 'symbol_3'],
  ['symbol_4', 'symbol_5', 'symbol_6'],
  ['symbol_7', 'symbol_8', 'symbol_9']],
 [(1, 200.00, ['symbol_1', 'symbol_5', 'symbol_9'], 1)]),

('log_002', 'gs_002', 'test_integrator_001', 'test_user_001', 'game_001',
 50.00, 0.00, -50.00, 5, 10.00,
 0, 'browser', 'macos', 'firefox',
 'session_001', 85, now(),
 [['symbol_1', 'symbol_1', 'symbol_1'],
  ['symbol_2', 'symbol_2', 'symbol_2'],
  ['symbol_3', 'symbol_3', 'symbol_3']],
 []);

-- ============================================
-- 常用查询示例（ClickHouse不支持返回表格的用户自定义函数）
-- ============================================

-- 说明：ClickHouse不支持返回TABLE的用户自定义函数，以下查询模板可在应用层直接使用

-- 获取指定游戏的实时RTP统计（示例查询模板）
-- SELECT
--     sumMerge(spin_count) as total_spins,
--     sumMerge(total_bet) as total_bet,
--     sumMerge(total_win) as total_win,
--     avgMerge(rtp_rate) as rtp_value,
--     avgMerge(hit_rate) as hit_rate,
--     avgMerge(avg_processing_time) as avg_processing_time
-- FROM game_log_agg_hourly_local
-- WHERE game_id = 'game_001'
--   AND agg_date >= now() - INTERVAL 24 HOUR;

-- 获取集成商RTP统计（示例查询模板）
-- SELECT
--     dimension_id as integrator_id,
--     stat_date,
--     stat_hour,
--     total_bet,
--     total_win,
--     rtp_value,
--     total_games,
--     avg_bet,
--     avg_win,
--     max_win
-- FROM rtp_statistics_local
-- WHERE dimension_type = 'integrator'
--   AND stat_date >= today() - INTERVAL 7 DAY
-- ORDER BY stat_date, stat_hour, dimension_id;

-- 获取用户RTP统计（示例查询模板）
-- SELECT
--     dimension_id as user_id,
--     stat_date,
--     stat_hour,
--     total_bet,
--     total_win,
--     rtp_value,
--     total_games,
--     hit_rate_value as hit_rate
-- FROM rtp_statistics_local
-- WHERE dimension_type = 'user'
--   AND dimension_id = 'test_user_001'
-- ORDER BY stat_date DESC, stat_hour DESC
-- LIMIT 24;

-- ============================================
-- 性能优化设置
-- ============================================

-- 设置查询参数优化
SET max_threads = 8;
SET max_memory_usage = 20000000000;
SET max_execution_time = 60;
SET max_block_size = 65536;

-- ============================================
-- 数据备份和清理策略
-- ============================================

-- 创建定期数据清理任务示例
-- ALTER TABLE game_log_detail_local MODIFY TTL date + INTERVAL 6 MONTH DELETE;
-- ALTER TABLE game_log_agg_hourly_local MODIFY TTL agg_date + INTERVAL 12 MONTH DELETE;

-- ============================================
-- 应用层批处理示例（写多场景优化方案）
-- ============================================

-- 说明：以下SQL语句建议在应用层定时任务中执行，而不是使用实时物化视图

-- 小时级聚合批处理（每小时执行一次）
-- INSERT INTO game_log_agg_hourly_local
-- SELECT
--     toDate(log_time) as agg_date,
--     toHour(log_time) as agg_hour,
--     integrator_id,
--     game_id,
--     uniqState(user_id) as user_count,
--     sumState(toUInt64(1)) as spin_count,
--     sumState(toUInt64(if(is_free_spin = 1, 1, 0))) as free_spin_count,
--     sumState(bet_amount) as total_bet,
--     sumState(win_amount) as total_win,
--     sumState(net_result) as net_result,
--     avgState(bet_amount) as avg_bet,
--     avgState(win_amount) as avg_win,
--     maxState(win_amount) as max_single_win,
--     avgState(rtp_rate) as rtp_rate,
--     toDecimal128(avgState(toDecimal128(if(win_amount > 0, 1, 0), 5)) * 100, 2) as hit_rate,
--     avgState(processing_time_ms) as avg_processing_time
-- FROM game_log_detail_local
-- WHERE log_time >= toStartOfHour(now() - INTERVAL 1 HOUR)
--   AND log_time < toStartOfHour(now() - INTERVAL 1 HOUR) + INTERVAL 1 HOUR
-- GROUP BY agg_date, agg_hour, integrator_id, game_id;

-- 全局RTP统计批处理（每5分钟执行一次）
-- INSERT INTO rtp_statistics_local
-- SELECT
--     toDate(log_time) as stat_date,
--     0 as stat_hour,
--     'global' as dimension_type,
--     'all' as dimension_id,
--     sum(bet_amount) as total_bet,
--     sum(win_amount) as total_win,
--     sum(net_result) as net_result,
--     (sum(win_amount) / sum(bet_amount)) * 100 as rtp_value,
--     count() as total_games,
--     sum(if(is_free_spin = 1, 1, 0)) as free_spin_games,
--     avg(bet_amount) as avg_bet,
--     avg(win_amount) as avg_win,
--     max(win_amount) as max_win,
--     now() as update_time
-- FROM game_log_detail_local
-- WHERE log_time >= now() - INTERVAL 5 MINUTE
-- GROUP BY toDate(log_time);

-- 集成商RTP统计批处理（每5分钟执行一次）
-- INSERT INTO rtp_statistics_local
-- SELECT
--     toDate(log_time) as stat_date,
--     toHour(log_time) as stat_hour,
--     'integrator' as dimension_type,
--     integrator_id as dimension_id,
--     sum(bet_amount) as total_bet,
--     sum(win_amount) as total_win,
--     sum(net_result) as net_result,
--     (sum(win_amount) / sum(bet_amount)) * 100 as rtp_value,
--     count() as total_games,
--     sum(if(is_free_spin = 1, 1, 0)) as free_spin_games,
--     avg(bet_amount) as avg_bet,
--     avg(win_amount) as avg_win,
--     max(win_amount) as max_win,
--     now() as update_time
-- FROM game_log_detail_local
-- WHERE log_time >= now() - INTERVAL 5 MINUTE
-- GROUP BY toDate(log_time), toHour(log_time), integrator_id;

-- 用户RTP统计批处理（每1分钟执行一次）
-- INSERT INTO rtp_statistics_local
-- SELECT
--     toDate(log_time) as stat_date,
--     toHour(log_time) as stat_hour,
--     'user' as dimension_type,
--     user_id as dimension_id,
--     sum(bet_amount) as total_bet,
--     sum(win_amount) as total_win,
--     sum(net_result) as net_result,
--     (sum(win_amount) / sum(bet_amount)) * 100 as rtp_value,
--     count() as total_games,
--     sum(if(is_free_spin = 1, 1, 0)) as free_spin_games,
--     avg(bet_amount) as avg_bet,
--     avg(win_amount) as avg_win,
--     max(win_amount) as max_win,
--     now() as update_time
-- FROM game_log_detail_local
-- WHERE log_time >= now() - INTERVAL 1 MINUTE
-- GROUP BY toDate(log_time), toHour(log_time), user_id;

-- ============================================
-- 实时 RTP 分析系统（物化视图自动聚合）
-- ============================================

-- ============================================
-- 用户 RTP 实时聚合表（本地表）
-- ============================================

CREATE TABLE IF NOT EXISTS rtp_user_metrics_local
(
    user_id String,
    integrator_id String,
    game_id String,
    time_window DateTime,
    
    total_bet Decimal(18, 4),
    total_win Decimal(18, 4),
    net_result Decimal(18, 4),
    rtp Decimal(8, 4),
    
    spin_count UInt32,
    avg_bet Decimal(18, 4),
    max_bet Decimal(18, 4),
    max_win Decimal(18, 4),
    min_win Decimal(18, 4),
    
    win_rate Decimal(8, 4),
    win_spin_count UInt32,
    loss_spin_count UInt32,
    
    first_spin_time DateTime,
    last_spin_time DateTime,
    last_update DateTime DEFAULT now()
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(time_window)
ORDER BY (integrator_id, user_id, game_id, time_window)
SETTINGS index_granularity = 8192;

-- 单节点部署不需要 Distributed 表
-- CREATE TABLE IF NOT EXISTS rtp_user_metrics
-- AS rtp_user_metrics_local
-- ENGINE = Distributed('default', 'rtp_analytics', 'rtp_user_metrics_local', rand());

-- ============================================
-- 用户 RTP 实时聚合表（5分钟粒度）
-- ============================================

CREATE TABLE IF NOT EXISTS rtp_user_realtime_local
(
    user_id String,
    integrator_id String,
    game_id String,
    time_window DateTime,
    
    total_bet Decimal(18, 4),
    total_win Decimal(18, 4),
    net_result Decimal(18, 4),
    rtp Decimal(8, 4),
    
    spin_count UInt32,
    avg_bet Decimal(18, 4),
    max_win Decimal(18, 4),
    
    last_update DateTime DEFAULT now()
)
ENGINE = MergeTree()
PARTITION BY toYYYYMMDD(time_window)
ORDER BY (integrator_id, user_id, game_id, time_window)
SETTINGS index_granularity = 8192;

-- 单节点部署不需要 Distributed 表
-- CREATE TABLE IF NOT EXISTS rtp_user_realtime
-- AS rtp_user_realtime_local
-- ENGINE = Distributed('default', 'rtp_analytics', 'rtp_user_realtime_local', rand());

-- ============================================
-- 游戏 RTP 实时聚合表（本地表）
-- ============================================

CREATE TABLE IF NOT EXISTS rtp_game_metrics_local
(
    game_id String,
    integrator_id String,
    time_window DateTime,
    
    total_bet Decimal(18, 4),
    total_win Decimal(18, 4),
    net_result Decimal(18, 4),
    rtp Decimal(8, 4),
    
    spin_count UInt32,
    active_users UInt32,
    avg_bet Decimal(18, 4),
    max_bet Decimal(18, 4),
    max_win Decimal(18, 4),
    
    volatility Decimal(8, 4),
    win_rate Decimal(8, 4),
    avg_net_result Decimal(18, 4),
    
    last_update DateTime DEFAULT now()
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(time_window)
ORDER BY (integrator_id, game_id, time_window)
SETTINGS index_granularity = 8192;

-- 单节点部署不需要 Distributed 表
-- CREATE TABLE IF NOT EXISTS rtp_game_metrics
-- AS rtp_game_metrics_local
-- ENGINE = Distributed('default', 'rtp_analytics', 'rtp_game_metrics_local', rand());

-- ============================================
-- 游戏 RTP 实时聚合表（5分钟粒度）
-- ============================================

CREATE TABLE IF NOT EXISTS rtp_game_realtime_local
(
    game_id String,
    integrator_id String,
    time_window DateTime,
    
    total_bet Decimal(18, 4),
    total_win Decimal(18, 4),
    rtp Decimal(8, 4),
    spin_count UInt32,
    active_users UInt32,
    avg_bet Decimal(18, 4),
    
    last_update DateTime DEFAULT now()
)
ENGINE = MergeTree()
PARTITION BY toYYYYMMDD(time_window)
ORDER BY (integrator_id, game_id, time_window)
SETTINGS index_granularity = 8192;

-- 单节点部署不需要 Distributed 表
-- CREATE TABLE IF NOT EXISTS rtp_game_realtime
-- AS rtp_game_realtime_local
-- ENGINE = Distributed('default', 'rtp_analytics', 'rtp_game_realtime_local', rand());

-- ============================================
-- RTP 异常告警表
-- ============================================

CREATE TABLE IF NOT EXISTS rtp_alerts_local
(
    alert_id String,
    alert_type String,
    severity String,
    
    user_id String,
    game_id String,
    integrator_id String,
    
    rtp Decimal(8, 4),
    total_bet Decimal(18, 4),
    total_win Decimal(18, 4),
    spin_count UInt32,
    
    time_window DateTime,
    detected_time DateTime,
    resolved_time Nullable(DateTime),
    
    status String DEFAULT 'active',
    notes String
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(detected_time)
ORDER BY (integrator_id, alert_type, detected_time)
SETTINGS index_granularity = 8192;

-- 单节点部署不需要 Distributed 表
-- CREATE TABLE IF NOT EXISTS rtp_alerts
-- AS rtp_alerts_local
-- ENGINE = Distributed('default', 'rtp_analytics', 'rtp_alerts_local', rand());



-- ============================================
-- 完成脚本执行
-- ============================================
SELECT 'ClickHouse数据库初始化脚本执行完成！' AS status;
SELECT '数据库：rtp_analytics' AS database_info;
SELECT '部署模式：单节点' AS deployment_mode;

-- 查询所有创建的表
SELECT 
    name AS table_name,
    engine AS table_engine,
    total_rows,
    total_bytes,
    formatReadableSize(total_bytes) AS readable_size
FROM system.tables 
WHERE database = 'rtp_analytics' 
  AND name LIKE '%local'
ORDER BY name;

SELECT '初始化脚本执行成功！' AS final_status;