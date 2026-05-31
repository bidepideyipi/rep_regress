# ClickHouse RTP 实时分析系统

## 📊 系统架构

```
Game Logs → RocketMQ → ClickHouse (game_log_detail_local)
                              ↓
                    Materialized Views (自动聚合)
                              ↓
                    ├─ rtp_user_realtime (用户实时RTP)
                    ├─ rtp_user_metrics (用户小时级RTP)
                    ├─ rtp_game_realtime (游戏实时RTP)
                    ├─ rtp_game_metrics (游戏小时级RTP)
                    └─ rtp_alerts (异常告警)
                              ↓
                    RTP API Service (查询服务)
```

## 🚀 快速部署

### 1. 执行 SQL 脚本

```bash
cd /Users/anthony/Documents/github/rtp_regress/platform-games/slot-game/sql

# 一键部署（需要 clickhouse-client）
chmod +x deploy.sh
./deploy.sh

# 或手动执行
clickhouse-client -d rtp_analytics --queries-file sql/01_user_rtp_tables.sql
clickhouse-client -d rtp_analytics --queries-file sql/02_game_rtp_tables.sql
clickhouse-client -d rtp_analytics --queries-file sql/03_rtp_alerts_tables.sql
```

### 2. 验证部署

```bash
# 查看已创建的表
clickhouse-client -d rtp_analytics -q "SHOW TABLES"

# 执行测试查询
clickhouse-client -d rtp_analytics --queries-file sql/04_test_queries.sql
```

## 📋 创建的表和视图

### 核心表

| 表名 | 说明 | 聚合粒度 |
|------|------|---------|
| `rtp_user_realtime_local` | 用户实时 RTP 本地表 | 5分钟 |
| `rtp_user_realtime` | 用户实时 RTP 分布式表 | 5分钟 |
| `rtp_user_metrics_local` | 用户小时级 RTP 本地表 | 1小时 |
| `rtp_user_metrics` | 用户小时级 RTP 分布式表 | 1小时 |
| `rtp_game_realtime_local` | 游戏实时 RTP 本地表 | 5分钟 |
| `rtp_game_realtime` | 游戏实时 RTP 分布式表 | 5分钟 |
| `rtp_game_metrics_local` | 游戏小时级 RTP 本地表 | 1小时 |
| `rtp_game_metrics` | 游戏小时级 RTP 分布式表 | 1小时 |
| `rtp_alerts_local` | 异常告警本地表 | - |
| `rtp_alerts` | 异常告警分布式表 | - |

### 物化视图

| 视图名 | 说明 | 触发条件 |
|--------|------|---------|
| `rtp_user_realtime_mv` | 用户实时聚合 | 每5分钟自动聚合 |
| `rtp_user_hourly_mv` | 用户小时聚合 | 每小时自动聚合 |
| `rtp_game_realtime_mv` | 游戏实时聚合 | 每5分钟自动聚合 |
| `rtp_game_hourly_mv` | 游戏小时聚合 | 每小时自动聚合 |
| `rtp_high_alert_mv` | 高 RTP 告警 | RTP > 98% 且旋转 ≥20次 |
| `rtp_low_alert_mv` | 低 RTP 告警 | RTP < 50% 且旋转 ≥20次 |
| `rtp_game_alert_mv` | 游戏异常告警 | 游戏整体 RTP 偏离预期 10%+ |

## 🔍 常用查询示例

### 1. 查看用户实时 RTP

```sql
SELECT
    user_id, game_id, time_window,
    total_bet, total_win, rtp, spin_count,
    avg_bet, max_win, last_update
FROM rtp_user_realtime
WHERE user_id = 'test_user_001'
  AND game_id = 'slot_game_001'
  AND time_window >= now() - INTERVAL 30 MINUTE
ORDER BY time_window DESC;
```

### 2. 查看游戏整体 RTP

```sql
SELECT
    game_id, time_window,
    total_bet, total_win, rtp, spin_count,
    active_users, volatility, win_rate
FROM rtp_game_metrics
WHERE game_id = 'slot_game_001'
  AND time_window >= now() - INTERVAL 24 HOUR
ORDER BY time_window DESC;
```

### 3. 查看当前活跃告警

```sql
SELECT
    alert_id, alert_type, severity,
    user_id, game_id, rtp, spin_count,
    detected_time, notes
FROM rtp_alerts
WHERE status = 'active'
ORDER BY detected_time DESC;
```

### 4. 高 RTP 用户排行

```sql
SELECT
    user_id, game_id,
    sum(total_bet) AS total_bet,
    sum(total_win) AS total_win,
    if(sum(total_bet) > 0, round(sum(total_win) / sum(total_bet), 4), 0) AS rtp,
    sum(spin_count) AS spin_count
FROM rtp_user_realtime_local
WHERE time_window >= now() - INTERVAL 1 HOUR
GROUP BY user_id, game_id
HAVING spin_count >= 10
ORDER BY rtp DESC
LIMIT 20;
```

## 🎯 异常检测阈值

### 用户 RTP 告警

| 告警类型 | RTP 阈值 | 最小旋转次数 | 最小总投注 |
|----------|---------|-------------|-----------|
| 高 RTP 告警 | > 98% | ≥ 20次 | ≥ 100 |
| 低 RTP 告警 | < 50% | ≥ 20次 | ≥ 100 |

### 游戏 RTP 告警

| 告警类型 | RTP 阈值 | 最小旋转次数 |
|----------|---------|-------------|
| 游戏异常告警 | > 105% 或 < 85% | ≥ 100次 |

## 📊 RTP API 服务

### 初始化服务

```go
import "platform-games/slot-game/rtp"

rtpService, err := rtp.NewRTPService("127.0.0.1", 9000, "default", "", "rtp_analytics")
if err != nil {
    log.Fatal(err)
}
defer rtpService.Close()
```

### 查询用户实时 RTP

```go
ctx := context.Background()
metrics, err := rtpService.GetUserRealtimeRTP(ctx, "user_001", "game_001", 30)
for _, m := range metrics {
    fmt.Printf("Time: %s, RTP: %.4f, Spins: %d\n", m.TimeWindow, m.RTP, m.SpinCount)
}
```

### 查询游戏整体 RTP

```go
metrics, err := rtpService.GetGameHourlyRTP(ctx, "game_001", 24)
for _, m := range metrics {
    fmt.Printf("Time: %s, RTP: %.4f, Active Users: %d\n", m.TimeWindow, m.RTP, m.ActiveUsers)
}
```

### 获取活跃告警

```go
alerts, err := rtpService.GetActiveAlerts(ctx, 20)
for _, alert := range alerts {
    fmt.Printf("Alert: %s, User: %s, RTP: %.4f, Severity: %s\n", 
        alert.AlertType, alert.UserID, alert.RTP, alert.Severity)
}
```

### 手动检测异常 RTP

```go
abnormalUsers, err := rtpService.DetectAbnormalRTP(ctx, 0.98, 0.50, 20)
for _, user := range abnormalUsers {
    fmt.Printf("User: %s, Game: %s, RTP: %.4f, Spins: %d\n", 
        user.UserID, user.GameID, user.RTP, user.SpinCount)
}
```

## ⚙️ 配置说明

### ClickHouse 连接配置

在 `config.yaml` 中添加：

```yaml
clickhouse:
  host: 127.0.0.1
  port: 9000
  username: default
  password: ""
  database: rtp_analytics
```

### 物化视图更新频率

- **实时聚合**: 每5分钟自动更新
- **小时聚合**: 每小时自动更新
- **告警检测**: 每5分钟检测一次

## 🔧 维护操作

### 清理历史数据

```sql
-- 清理30天前的用户实时数据
ALTER TABLE rtp_user_realtime_local DELETE WHERE time_window < now() - INTERVAL 30 DAY;

-- 清理90天前的用户小时数据
ALTER TABLE rtp_user_realtime_local DELETE WHERE time_window < now() - INTERVAL 90 DAY;
```

### 优化表性能

```sql
-- 优化用户实时表
OPTIMIZE TABLE rtp_user_realtime_local FINAL;

-- 优化游戏小时表
OPTIMIZE TABLE rtp_game_metrics_local FINAL;
```

### 查看物化视图状态

```sql
SELECT
    database, name, engine,
    total_rows, total_bytes
FROM system.tables
WHERE database = 'rtp_analytics'
  AND engine LIKE '%Materialized%'
ORDER BY name;
```

## 📈 性能指标

| 指标 | 值 |
|------|-----|
| 实时聚合延迟 | < 5秒 |
| 小时聚合延迟 | < 1分钟 |
| 查询响应时间 | < 100ms |
| 数据保留时间 | 实时: 30天, 小时: 90天 |

## 🚨 监控建议

1. **监控物化视图更新**: 每小时检查 `last_update` 字段
2. **监控告警数量**: 每小时统计活跃告警数量
3. **监控数据量**: 每天检查表行数和存储大小
4. **监控查询性能**: 记录慢查询日志

## 📚 参考资料

- [ClickHouse Materialized Views](https://clickhouse.com/docs/en/sql-reference/statements/create/view/)
- [ClickHouse Aggregation Functions](https://clickhouse.com/docs/en/sql-reference/functions/aggregation-functions/)
- [RTP Calculation Best Practices](https://example.com/rtp-guide)