# ClickHouse数据库设计文档

## 文档说明
本文档描述游戏服务模块的ClickHouse数据库设计，主要用于游戏日志存储、RTP统计分析、大数据分析等场景。ClickHouse采用列式存储，适合海量数据的OLAP查询。

---

## 1. 数据库基础信息

| 项目 | 内容 |
|------|------|
| 数据库名称 | rtp_analytics |
| 数据库版本 | ClickHouse 24.0+ |
| 集群名称 | slot_cluster |
| 字符集 | UTF8 |
| 时区 | Asia/Shanghai |
| 存储引擎 | MergeTree系列 |

---

## 2. 表结构设计

### 2.0 表优先级别说明

| 优先级别 | 表名 | 说明 | 实时性要求 | 性能要求 |
|----------|------|------|------------|----------|
| **高优先级** | game_log_detail_local | 游戏日志详情表 | 实时写入 | 查询性能要求高 |
| **高优先级** | game_log_agg_hourly_local | 游戏日志小时级聚合表 | 自动聚合 | 聚合计算性能要求高 |
| **高优先级** | rtp_statistics_local | RTP统计维度表 | 实时统计 | 统计查询性能要求高 |
| **低优先级** | user_behavior_analysis_local | 用户行为分析表 | 定期分析 | 批量分析，性能要求相对较低 |
| **低优先级** | game_performance_monitor_local | 游戏性能监控表 | 定期监控 | 监控频率较低，性能要求相对较低 |

**优先级别定义**：
- **高优先级**：核心业务表，需要高可用性、高性能、实时性，优先分配资源
- **低优先级**：辅助分析表，可以批量处理，对实时性要求较低，资源分配相对弹性

---

### 2.1 游戏日志详情表 (game_log_detail_local) 【高优先级】

#### 2.1.1 表基本信息

| 项目 | 内容 |
|------|------|
| 表名 | game_log_detail_local |
| 中文名 | 游戏日志详情本地表 |
| 用途 | 存储详细的游戏日志记录，支持RTP统计和大数据分析 |
| 引擎类型 | MergeTree |
| 分区策略 | 按月分区 |
| TTL策略 | 保留6个月 |

#### 2.1.2 字段设计

| 字段名 | 数据类型 | 允许NULL | 默认值 | 说明 | 注释 |
|--------|----------|----------|--------|------|------|
| log_id | String | NO | - | 日志ID | 唯一标识 |
| game_session_id | String | NO | - | 游戏会话ID | 关联业务主键 |
| integrator_id | String | NO | - | 集成商ID | 业务维度 |
| user_id | String | NO | - | 用户ID | 业务维度 |
| game_id | String | NO | - | 游戏ID | 业务维度 |
| bet_amount | Decimal(18,4) | NO | - | 投注额 | 精度到4位小数 |
| win_amount | Decimal(18,4) | NO | 0.00 | 赢取额 | 精度到4位小数 |
| net_result | Decimal(18,4) | NO | - | 净结果 | 赔付-下注 |
| bet_lines | UInt16 | NO | 1 | 下注线数 | 范围1-20 |
| bet_per_line | Decimal(18,4) | NO | - | 每线下注额 | bet_amount/bet_lines |
| rtp_rate | Decimal(5,2) | MATERIALIZED | (win_amount / bet_amount) * 100 | 单局RTP | 自动计算列 |
| is_free_spin | UInt8 | NO | 0 | 是否免费旋转 | 0-否，1-是 |
| bonus_feature | String | YES | NULL | 特殊功能 | 免费旋转、额外游戏等 |
| device_type | LowCardinality(String) | YES | unknown | 设备类型 | browser/mobile/tablet |
| device_os | LowCardinality(String) | YES | unknown | 设备操作系统 | windows/mac/ios/android |
| browser_type | LowCardinality(String) | YES | unknown | 浏览器类型 | chrome/firefox/safari |
| ip_address | IPv4 | YES | NULL | 客户端IP | IPv4格式 |
| ip_region | String | YES | NULL | IP地区 | 地理区域信息 |
| ip_country | LowCardinality(String) | YES | NULL | IP国家 | 国家代码 |
| session_id | String | NO | - | 会话ID | 用户会话标识 |
| server_id | String | YES | NULL | 服务器ID | 处理服务器标识 |
| processing_time_ms | UInt32 | NO | 0 | 处理时长(毫秒) | 性能监控 |
| error_code | UInt16 | YES | NULL | 错误码 | 异常监控 |
| error_message | String | YES | NULL | 错误信息 | 异常详情 |
| game_result_json | String | YES | NULL | 游戏结果JSON | 详细游戏数据 |
| reel_result | Array(Array(String)) | YES | [] | 卷轴结果 | 3x3或5x3的符号矩阵 |
| win_lines | Array(Tuple(UInt16, Decimal(18,4), Array(String), UInt8)) | YES | [] | 中奖线路 | 线路ID、金额、符号、倍数 |
| user_agent | String | YES | NULL | 用户代理 | 浏览器User-Agent |
| log_time | DateTime('Asia/Shanghai') | NO | now() | 日志时间 | 业务时间 |
| date | Date | MATERIALIZED | toDate(log_time) | 日期分区 | 用于分区和查询优化 |
| hour | UInt8 | MATERIALIZED | toHour(log_time) | 小时 | 用于小时级统计 |
| minute | UInt8 | MATERIALIZED | toMinute(log_time) | 分钟 | 用于分钟级统计 |

#### 2.1.3 索引和分区设计

| 设计项 | 配置内容 | 说明 |
|--------|----------|------|
| 分区键 | toYYYYMM(date) | 按月分区，便于数据归档和查询 |
| 排序键 | (date, hour, game_id, user_id, log_time) | 支持时间范围和用户维度查询 |
| 主键 | (date, hour, game_id, user_id, log_time) | 与排序键一致 |
| 数据粒度 | 8192 | ClickHouse默认值 |
| TTL策略 | date + INTERVAL 6 MONTH | 保留6个月数据 |
| 副本数 | 2 | 高可用配置 |

#### 2.1.4 分布式表设计

```sql
-- 分布式表
CREATE TABLE game_log_detail_distributed ON CLUSTER slot_cluster AS game_log_detail_local
ENGINE = Distributed(slot_cluster, rtp_analytics, game_log_detail_local, rand());
```

---

### 2.2 游戏日志小时级聚合表 (game_log_agg_hourly_local) 【高优先级】

#### 2.2.1 表基本信息

| 项目 | 内容 |
|------|------|
| 表名 | game_log_agg_hourly_local |
| 中文名 | 游戏日志小时级聚合表 |
| 用途 | 存储按小时聚合的游戏统计数据，支持快速RTP查询 |
| 引擎类型 | AggregatingMergeTree |
| 分区策略 | 按月分区 |
| TTL策略 | 保留12个月 |

#### 2.2.2 字段设计

| 字段名 | 数据类型 | 允许NULL | 默认值 | 说明 | 注释 |
|--------|----------|----------|--------|------|------|
| agg_date | Date | NO | - | 聚合日期 | 统计日期 |
| agg_hour | UInt8 | NO | - | 聚合小时 | 0-23 |
| integrator_id | String | NO | - | 集成商ID | 业务维度 |
| game_id | String | NO | - | 游戏ID | 业务维度 |
| user_count | AggregateFunction(uniq, String) | NO | - | 去重用户数 | 唯一用户统计 |
| spin_count | AggregateFunction(sum, UInt64) | NO | - | 总旋转次数 | 游戏次数统计 |
| free_spin_count | AggregateFunction(sum, UInt64) | NO | - | 免费旋转次数 | 特殊功能统计 |
| total_bet | AggregateFunction(sum, Decimal(18,4)) | NO | - | 总投注额 | 累计投注金额 |
| total_win | AggregateFunction(sum, Decimal(18,4)) | NO | - | 总赢取额 | 累计赔付金额 |
| net_result | AggregateFunction(sum, Decimal(18,4)) | NO | - | 净结果统计 | 累计净结果 |
| avg_bet | AggregateFunction(avg, Decimal(18,4)) | NO | - | 平均下注额 | 平均下注统计 |
| avg_win | AggregateFunction(avg, Decimal(18,4)) | NO | - | 平均赔付额 | 平均赔付统计 |
| max_single_win | AggregateFunction(max, Decimal(18,4)) | NO | - | 最大单次赔付 | 高价值统计 |
| rtp_rate | AggregateFunction(avg, Decimal(5,2)) | NO | - | 平均RTP | RTP统计 |
| rtp_p50 | AggregateFunction(quantile(0.5), Decimal(5,2)) | NO | - | RTP中位数 | 分布统计 |
| rtp_p90 | AggregateFunction(quantile(0.9), Decimal(5,2)) | NO | - | RTP 90分位 | 分布统计 |
| rtp_p95 | AggregateFunction(quantile(0.95), Decimal(5,2)) | NO | - | RTP 95分位 | 分布统计 |
| rtp_p99 | AggregateFunction(quantile(0.99), Decimal(5,2)) | NO | - | RTP 99分位 | 分布统计 |
| hit_rate | AggregateFunction(avg, Decimal(5,2)) | NO | - | 命中率 | 中奖概率统计 |
| avg_processing_time | AggregateFunction(avg, UInt32) | NO | - | 平均处理时长(ms) | 性能统计 |

#### 2.2.3 索引和分区设计

| 设计项 | 配置内容 | 说明 |
|--------|----------|------|
| 分区键 | toYYYYMM(agg_date) | 按月分区 |
| 排序键 | (agg_date, agg_hour, integrator_id, game_id) | 支持多维度查询 |
| 主键 | (agg_date, agg_hour, integrator_id, game_id) | 与排序键一致 |
| TTL策略 | agg_date + INTERVAL 12 MONTH | 保留12个月数据 |

#### 2.2.4 物化视图设计

```sql
-- 创建物化视图自动聚合
CREATE MATERIALIZED VIEW game_log_agg_hourly_mv ON CLUSTER slot_cluster
TO game_log_agg_hourly_local
AS
SELECT
    toDate(log_time) as agg_date,
    toHour(log_time) as agg_hour,
    integrator_id,
    game_id,
    uniqState(user_id) as user_count,
    sumState(1) as spin_count,
    sumState(if(is_free_spin = 1, 1, 0)) as free_spin_count,
    sumState(bet_amount) as total_bet,
    sumState(win_amount) as total_win,
    sumState(net_result) as net_result,
    avgState(bet_amount) as avg_bet,
    avgState(win_amount) as avg_win,
    maxState(win_amount) as max_single_win,
    avgState(rtp_rate) as rtp_rate,
    quantileState(0.5)(rtp_rate) as rtp_p50,
    quantileState(0.9)(rtp_rate) as rtp_p90,
    quantileState(0.95)(rtp_rate) as rtp_p95,
    quantileState(0.99)(rtp_rate) as rtp_p99,
    avgState(if(win_amount > 0, 1, 0)) * 100 as hit_rate,
    avgState(processing_time_ms) as avg_processing_time
FROM game_log_detail_local
GROUP BY agg_date, agg_hour, integrator_id, game_id;
```

---

### 2.3 RTP统计维度表 (rtp_statistics_local) 【高优先级】

#### 2.3.1 表基本信息

| 项目 | 内容 |
|------|------|
| 表名 | rtp_statistics_local |
| 中文名 | RTP统计维度表 |
| 用途 | 存储全局、集成商、用户三个维度的RTP统计数据 |
| 引擎类型 | SummingMergeTree |
| 分区策略 | 按月分区 |
| TTL策略 | 保留24个月 |

#### 2.3.2 字段设计

| 字段名 | 数据类型 | 允许NULL | 默认值 | 说明 | 注释 |
|--------|----------|----------|--------|------|------|
| stat_date | Date | NO | - | 统计日期 | 数据日期 |
| stat_hour | UInt8 | NO | - | 统计小时 | 0-23，0表示全天 |
| dimension_type | LowCardinality(String) | NO | - | 维度类型 | global/integrator/user |
| dimension_id | String | NO | - | 维度ID | 维度标识 |
| total_bet | Decimal(18,2) | NO | 0 | 总投注额 | 累计投注 |
| total_win | Decimal(18,2) | NO | 0 | 总赢取额 | 累计赔付 |
| net_result | Decimal(18,2) | NO | 0 | 净结果 | 赔付-下注 |
| rtp_value | Decimal(5,2) | NO | 0 | RTP值 | 赔付/投注*100 |
| total_games | UInt64 | NO | 0 | 总游戏数 | 游戏次数 |
| free_spin_games | UInt64 | NO | 0 | 免费旋转次数 | 特殊游戏次数 |
| avg_bet | Decimal(18,2) | NO | 0 | 平均下注额 | 总投注/总游戏数 |
| avg_win | Decimal(18,2) | NO | 0 | 平均赔付额 | 总赔付/总游戏数 |
| max_win | Decimal(18,2) | NO | 0 | 最大赔付额 | 单次最大赔付 |
| update_time | DateTime('Asia/Shanghai') | NO | now() | 更新时间 | 数据更新时间 |

#### 2.3.3 索引和分区设计

| 设计项 | 配置内容 | 说明 |
|--------|----------|------|
| 分区键 | toYYYYMM(stat_date) | 按月分区 |
| 排序键 | (stat_date, stat_hour, dimension_type, dimension_id) | 支持多维度查询 |
| 主键 | (stat_date, stat_hour, dimension_type, dimension_id) | 与排序键一致 |
| TTL策略 | stat_date + INTERVAL 24 MONTH | 保留24个月数据 |

#### 2.3.4 物化视图设计

```sql
-- 全局RTP统计物化视图
CREATE MATERIALIZED VIEW rtp_global_mv ON CLUSTER slot_cluster
TO rtp_statistics_local
AS
SELECT
    toDate(log_time) as stat_date,
    0 as stat_hour,
    'global' as dimension_type,
    'all' as dimension_id,
    sum(bet_amount) as total_bet,
    sum(win_amount) as total_win,
    sum(net_result) as net_result,
    (sum(win_amount) / sum(bet_amount)) * 100 as rtp_value,
    count() as total_games,
    sum(if(is_free_spin = 1, 1, 0)) as free_spin_games,
    avg(bet_amount) as avg_bet,
    avg(win_amount) as avg_win,
    max(win_amount) as max_win,
    now() as update_time
FROM game_log_detail_local
GROUP BY toDate(log_time);

-- 集成商RTP统计物化视图
CREATE MATERIALIZED VIEW rtp_integrator_mv ON CLUSTER slot_cluster
TO rtp_statistics_local
AS
SELECT
    toDate(log_time) as stat_date,
    toHour(log_time) as stat_hour,
    'integrator' as dimension_type,
    integrator_id as dimension_id,
    sum(bet_amount) as total_bet,
    sum(win_amount) as total_win,
    sum(net_result) as net_result,
    (sum(win_amount) / sum(bet_amount)) * 100 as rtp_value,
    count() as total_games,
    sum(if(is_free_spin = 1, 1, 0)) as free_spin_games,
    avg(bet_amount) as avg_bet,
    avg(win_amount) as avg_win,
    max(win_amount) as max_win,
    now() as update_time
FROM game_log_detail_local
GROUP BY toDate(log_time), toHour(log_time), integrator_id;

-- 用户RTP统计物化视图
CREATE MATERIALIZED VIEW rtp_user_mv ON CLUSTER slot_cluster
TO rtp_statistics_local
AS
SELECT
    toDate(log_time) as stat_date,
    toHour(log_time) as stat_hour,
    'user' as dimension_type,
    user_id as dimension_id,
    sum(bet_amount) as total_bet,
    sum(win_amount) as total_win,
    sum(net_result) as net_result,
    (sum(win_amount) / sum(bet_amount)) * 100 as rtp_value,
    count() as total_games,
    sum(if(is_free_spin = 1, 1, 0)) as free_spin_games,
    avg(bet_amount) as avg_bet,
    avg(win_amount) as avg_win,
    max(win_amount) as max_win,
    now() as update_time
FROM game_log_detail_local
GROUP BY toDate(log_time), toHour(log_time), user_id;
```

---

### 2.4 用户行为分析表 (user_behavior_analysis_local) 【低优先级】

#### 2.4.1 表基本信息

| 项目 | 内容 |
|------|------|
| 表名 | user_behavior_analysis_local |
| 中文名 | 用户行为分析表 |
| 用途 | 存储用户行为分析数据，支持用户画像和个性化推荐 |
| 引擎类型 | SummingMergeTree |
| 分区策略 | 按月分区 |
| TTL策略 | 保留12个月 |

#### 2.4.2 字段设计

| 字段名 | 数据类型 | 允许NULL | 默认值 | 说明 | 注释 |
|--------|----------|----------|--------|------|------|
| analysis_date | Date | NO | - | 分析日期 | 数据日期 |
| user_id | String | NO | - | 用户ID | 用户标识 |
| integrator_id | String | NO | - | 集成商ID | 所属集成商 |
| total_sessions | UInt32 | NO | 0 | 总会话数 | 用户活跃度 |
| total_duration | UInt32 | NO | 0 | 总时长(秒) | 用户停留时间 |
| avg_session_duration | UInt32 | NO | 0 | 平均会话时长 | 单次访问平均时长 |
| peak_hour | UInt8 | NO | 0 | 峰值小时 | 用户最活跃时段 |
| favorite_game | String | YES | NULL | 偏好游戏 | 用户最常玩的游戏 |
| favorite_bet_range | String | YES | NULL | 偏好下注范围 | 用户习惯下注区间 |
| risk_level | LowCardinality(String) | NO | low | 风险等级 | 用户风险偏好 |
| loyalty_level | LowCardinality(String) | NO | new | 忠诚度等级 | 用户忠诚度 |
| churn_risk | Decimal(5,2) | NO | 0 | 流失风险 | 流失概率预测 |
| lifetime_value | Decimal(18,2) | NO | 0 | 生命周期价值 | 用户LTV预测 |
| update_time | DateTime('Asia/Shanghai') | NO | now() | 更新时间 | 数据更新时间 |

---

### 2.5 游戏性能监控表 (game_performance_monitor_local) 【低优先级】

#### 2.5.1 表基本信息

| 项目 | 内容 |
|------|------|
| 表名 | game_performance_monitor_local |
| 中文名 | 游戏性能监控表 |
| 用途 | 存储游戏服务性能指标，支持实时监控和告警 |
| 引擎类型 | SummingMergeTree |
| 分区策略 | 按日分区 |
| TTL策略 | 保留90天 |

#### 2.5.2 字段设计

| 字段名 | 数据类型 | 允许NULL | 默认值 | 说明 | 注释 |
|--------|----------|----------|--------|------|------|
| monitor_time | DateTime('Asia/Shanghai') | NO | now() | 监控时间 | 时间戳 |
| server_id | String | NO | - | 服务器ID | 服务实例标识 |
| game_id | String | NO | - | 游戏ID | 游戏标识 |
| request_count | UInt64 | NO | 0 | 请求次数 | 接口调用次数 |
| success_count | UInt64 | NO | 0 | 成功次数 | 成功响应次数 |
| error_count | UInt64 | NO | 0 | 错误次数 | 错误响应次数 |
| avg_response_time | UInt32 | NO | 0 | 平均响应时间(ms) | 性能指标 |
| max_response_time | UInt32 | NO | 0 | 最大响应时间(ms) | 性能指标 |
| p50_response_time | UInt32 | NO | 0 | 50分位响应时间(ms) | 性能指标 |
| p99_response_time | UInt32 | NO | 0 | 99分位响应时间(ms) | 性能指标 |
| cpu_usage | Decimal(5,2) | NO | 0 | CPU使用率(%) | 系统资源 |
| memory_usage | Decimal(5,2) | NO | 0 | 内存使用率(%) | 系统资源 |
| disk_io | UInt32 | NO | 0 | 磁盘IO操作数 | 系统资源 |
| network_io | UInt64 | NO | 0 | 网络IO字节数 | 网络资源 |

---

## 3. 数据流设计

### 3.1 数据写入流程

```
1. 游戏服务处理用户请求
2. 游戏结果同时写入MySQL和ClickHouse
3. MySQL写入核心业务数据（实时性要求高）
4. ClickHouse写入详细日志（异步写入，批量处理）
5. 物化视图自动触发聚合计算
6. 聚合数据实时更新
```

### 3.2 数据查询流程

```
1. 实时查询：查询MySQL核心业务数据
2. 统计分析：查询ClickHouse聚合数据
3. 大数据分析：查询ClickHouse详细日志
4. 性能优化：根据查询特点选择合适的数据源
```

---

## 4. 查询优化策略

### 4.1 常用查询模式

| 查询类型 | 查询模式 | 优化策略 | 响应时间目标 |
|----------|----------|----------|--------------|
| 实时RTP查询 | 查询最近1小时的RTP数据 | 使用小时级聚合表 | < 100ms |
| 历史RTP分析 | 查询过去7天的RTP趋势 | 使用小时级聚合表 | < 500ms |
| 用户行为分析 | 查询用户游戏行为模式 | 使用详细日志表 | < 1s |
| 性能监控 | 查询系统性能指标 | 使用性能监控表 | < 100ms |
| 大数据分析 | 全量数据统计分析 | 使用详细日志表 | < 10s |

### 4.2 查询优化技巧

1. **时间范围过滤**：在WHERE子句中始终包含时间范围条件
2. **分区裁剪**：利用分区键减少扫描数据量
3. **预聚合**：使用物化视图聚合数据，避免实时计算
4. **索引利用**：合理设计排序键，提高查询效率
5. **并发查询**：使用分布式查询能力，提高查询速度

### 4.3 写多场景优化方案

#### 4.3.1 物化视图评估

| 物化视图 | 当前状态 | 写入频率 | 实时性需求 | 评估结果 | 优化建议 |
|----------|----------|----------|------------|----------|----------|
| `game_log_agg_hourly_mv` | 每次写入触发小时级聚合 | 高 | 低（分钟级可接受） | **移到逻辑层** | 改为每小时批处理 |
| `rtp_global_mv` | 每次写入触发全局RTP统计 | 高 | 低（分钟级可接受） | **移到逻辑层** | 改为每5分钟批处理 |
| `rtp_integrator_mv` | 每次写入触发集成商RTP统计 | 高 | 低（分钟级可接受） | **移到逻辑层** | 改为每5分钟批处理 |
| `rtp_user_mv` | 每次写入触发用户RTP统计 | 高 | 中（可能需要实时查看） | **移到逻辑层** | 改为每1分钟批处理 |

#### 4.3.2 优化方案

**问题分析**：
- 高频写入时，每次插入都会触发所有物化视图的实时计算
- 导致大量CPU资源消耗在聚合计算上，增加写入延迟
- 大部分统计查询不需要毫秒级实时性，可以接受一定延迟

**优化措施**：

1. **移除实时物化视图**
   - 删除所有CREATE MATERIALIZED VIEW语句
   - 保留聚合表结构（`game_log_agg_hourly_local`、`rtp_statistics_local`）
   - 原始数据仅写入`game_log_detail_local`表

2. **应用层定时批处理**
   ```sql
   -- 小时级聚合批处理（每小时执行）
   INSERT INTO game_log_agg_hourly_local
   SELECT
       toDate(log_time) as agg_date,
       toHour(log_time) as agg_hour,
       integrator_id,
       game_id,
       uniqState(user_id) as user_count,
       sumState(toUInt64(1)) as spin_count,
       -- ... 其他聚合字段
   FROM game_log_detail_local
   WHERE log_time >= now() - INTERVAL 1 HOUR
   GROUP BY agg_date, agg_hour, integrator_id, game_id;
   
   -- RTP统计批处理（每5分钟执行）
   INSERT INTO rtp_statistics_local
   SELECT
       toDate(log_time) as stat_date,
       0 as stat_hour,
       'global' as dimension_type,
       'all' as dimension_id,
       sum(bet_amount) as total_bet,
       -- ... 其他统计字段
   FROM game_log_detail_local
   WHERE log_time >= now() - INTERVAL 5 MINUTE
   GROUP BY toDate(log_time);
   ```

3. **批处理频率建议**
   - 小时级聚合：每小时执行一次
   - 全局RTP统计：每5分钟执行一次
   - 集成商RTP统计：每5分钟执行一次
   - 用户RTP统计：每1分钟执行一次（高优先级）

#### 4.3.3 性能提升预期

| 指标 | 优化前 | 优化后 | 提升幅度 |
|------|--------|--------|----------|
| 写入延迟 | ~100ms | ~10ms | **90%提升** |
| 写入CPU使用 | 高（含聚合计算） | 低（仅数据插入） | **70%降低** |
| 吞吐量 | ~10K/s | ~50K/s | **5倍提升** |
| 统计查询延迟 | 实时 | 1-5分钟延迟 | 可接受 |

#### 4.3.4 实施建议

1. **渐进式迁移**
   - 先移除用户级RTP物化视图，观察性能提升
   - 再移除其他物化视图，逐步优化
   - 监控系统性能，调整批处理频率

2. **监控和告警**
   - 监控批处理任务执行状态
   - 设置数据延迟告警（如超过预期延迟时间）
   - 监控聚合表数据完整性

3. **应急方案**
   - 保留物化视图创建脚本，需要时可快速恢复
   - 设置手动触发聚合任务的接口
   - 批处理失败时的补偿机制

---

## 5. 数据质量和一致性

### 5.1 数据验证规则

| 验证项 | 验证规则 | 异常处理 |
|--------|----------|----------|
| 投注金额 | bet_amount >= 0 | 拒绝写入，记录错误日志 |
| 赔付金额 | win_amount >= 0 | 拒绝写入，记录错误日志 |
| RTP计算 | rtp_rate = (win_amount / bet_amount) * 100 | 自动计算，保证准确性 |
| 时间字段 | log_time <= now() | 拒绝写入，记录错误日志 |
| 唯一性 | log_id全局唯一 | 拒绝重复数据 |

### 5.2 数据一致性保证

1. **双写机制**：MySQL和ClickHouse同时写入，确保数据一致性
2. **校验机制**：定期对比MySQL和ClickHouse数据，发现差异及时修复
3. **事务支持**：使用分布式事务保证数据写入的原子性
4. **容错机制**：写入失败时重试，超过重试次数后告警

---

## 6. 性能调优参数

### 6.1 表参数配置

| 参数名 | 默认值 | 推荐值 | 说明 |
|--------|--------|--------|------|
| index_granularity | 8192 | 8192 | 索引粒度 |
| max_bytes_to_merge_at_once | 161061273600 | 322122547200 | 合并时的最大字节数 |
| max_bytes_to_merge_at_min_space_in_pool | 1048576000 | 2097152000 | 合并时的最小空间要求 |
| max_insert_block_size | 1048576 | 1048576 | 插入块大小 |
| min_insert_block_size_rows | 1048576 | 1048576 | 插入块行数 |

### 6.2 查询参数配置

| 参数名 | 默认值 | 推荐值 | 说明 |
|--------|--------|--------|------|
| max_threads | 1 | CPU核心数 | 最大线程数 |
| max_memory_usage | 10000000000 | 20000000000 | 最大内存使用量 |
| max_execution_time | 300 | 60 | 最大执行时间(秒) |
| max_block_size | 65536 | 65536 | 最大块大小 |

---

## 7. 监控和告警

### 7.1 性能监控指标

| 监控指标 | 阈值 | 告警级别 | 处理建议 |
|----------|------|----------|----------|
| 查询响应时间 | > 1s | 警告 | 检查查询性能 |
| 查询响应时间 | > 10s | 严重 | 优化查询或扩容 |
| 写入延迟 | > 5s | 警告 | 检查写入性能 |
| 写入失败率 | > 1% | 严重 | 检查系统状态 |
| 存储空间使用率 | > 80% | 警告 | 规划扩容 |
| 存储空间使用率 | > 90% | 严重 | 立即扩容 |

### 7.2 数据质量监控

| 监控指标 | 阈值 | 告警级别 | 处理建议 |
|----------|------|----------|----------|
| 数据一致性差异 | > 0.01% | 警告 | 检查数据同步 |
| 数据写入延迟 | > 60s | 警告 | 检查写入队列 |
| 聚合计算延迟 | > 5min | 警告 | 检查物化视图状态 |

---

## 8. MySQL与ClickHouse数据分工

### 8.1 数据存储分工

| 数据类型 | 存储位置 | 用途 | 特点 |
|----------|----------|------|------|
| 核心业务数据 | MySQL | 实时业务操作 | 强一致性、事务支持 |
| 详细游戏日志 | ClickHouse | 大数据分析 | 列式存储、高性能查询 |
| 统计分析数据 | ClickHouse | 实时统计 | 自动聚合、快速查询 |
| 用户基本信息 | MySQL | 基础信息管理 | 关系查询、事务支持 |

### 8.2 查询分工

| 查询场景 | 数据源 | 响应时间 | 准确性 |
|----------|--------|----------|--------|
| 游戏结果记录查询 | MySQL | < 100ms | 100% |
| 余额查询 | MySQL | < 50ms | 100% |
| 实时RTP统计 | ClickHouse | < 200ms | 99.9% |
| 历史数据分析 | ClickHouse | < 2s | 99.9% |
| 用户行为分析 | ClickHouse | < 1s | 99.9% |

---

## 9. 版本记录

| 版本号 | 日期 | 修改内容 | 修改人 |
|--------|------|----------|--------|
| v1.0.0 | 2025-01-01 | 初始版本发布，包含核心表设计 | - |
| v1.0.1 | 2025-01-01 | 添加物化视图设计，优化查询性能 | - |