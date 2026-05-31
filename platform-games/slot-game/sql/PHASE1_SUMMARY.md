# Phase 1 完成总结

## ✅ 已完成的工作

### 1. ClickHouse SQL 脚本（已创建）

| 文件 | 说明 | 状态 |
|------|------|------|
| `/script/clickhouse_init.sql` | 数据库表结构初始化脚本 | ✅ 完成 |
| `/script/rtp_batch_aggregate.sql` | 定时批处理 SQL 脚本（参考用） | ✅ 完成 |

### 2. RTP 服务代码

| 文件 | 说明 | 状态 |
|------|------|------|
| `slot-game/rtp/rtp_service.go` | RTP 查询服务核心代码 | ✅ 完成 |
| `rtp-processor/` | **独立批处理服务** | ✅ 完成 |

## 📊 创建的 ClickHouse 表结构

### 核心聚合表

#### 用户维度

1. **rtp_user_realtime_local** - 用户实时 RTP（5分钟粒度）
2. **rtp_user_metrics_local** - 用户小时级 RTP（1小时粒度）

#### 游戏维度

3. **rtp_game_realtime_local** - 游戏实时 RTP（5分钟粒度）
4. **rtp_game_metrics_local** - 游戏小时级 RTP（1小时粒度）

#### 异常告警

5. **rtp_alerts_local** - RTP 异常告警表

## 🔄 独立批处理架构

### 为什么独立部署？

| 方案 | 多副本冲突 | 扩展性 | 复杂度 |
|------|-----------|--------|--------|
| 集成到 slot-game | ❌ 有冲突 | ⚠️ 受限 | 低 |
| **独立批处理服务** | ✅ 无冲突 | ✅ 好 | 中 |

### 架构图

```
                    ┌─────────────────┐
                    │   Nacos         │
                    │  (配置中心)      │
                    └────────┬────────┘
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
        ▼                    ▼                    ▼
┌───────────────┐   ┌───────────────┐   ┌───────────────┐
│  slot-game-1  │   │  slot-game-2  │   │  slot-game-3  │
│   (多副本)     │   │   (多副本)     │   │   (多副本)     │
└───────┬───────┘   └───────┬───────┘   └───────┬───────┘
        │                   │                   │
        └───────────────────┼───────────────────┘
                            │ 写入游戏日志
                            ▼
                    ┌─────────────────┐
                    │   ClickHouse    │
                    │game_log_detail  │
                    └────────┬────────┘
                             │ 读取
                            ▼
                 ┌─────────────────────┐
                 │   rtp-processor     │
                 │   (独立服务，单实例)   │
                 └─────────────────────┘
                            │ 写入聚合结果
                            ▼
                    ┌─────────────────┐
                    │ ClickHouse       │
                    │ rtp_user_*       │
                    │ rtp_game_*       │
                    │ rtp_alerts       │
                    └─────────────────┘
```

### 批处理任务配置

| 任务 | 执行频率 | 说明 |
|------|---------|------|
| 用户实时聚合 | 每 5 分钟 | 聚合 5 分钟内的用户 RTP |
| 用户小时聚合 | 每 5 分钟 | 聚合 1 小时内的用户 RTP |
| 游戏实时聚合 | 每 5 分钟 | 聚合 5 分钟内的游戏 RTP |
| 游戏小时聚合 | 每 5 分钟 | 聚合 1 小时内的游戏 RTP |
| 高 RTP 告警 | 每 10 分钟 | 检测 RTP > 98% 的用户 |
| 低 RTP 告警 | 每 10 分钟 | 检测 RTP < 50% 的用户 |
| 游戏异常告警 | 每 10 分钟 | 检测游戏整体 RTP 异常 |

## 🎯 RTP API 服务功能

| 方法 | 说明 | 参数 |
|------|------|------|
| `GetUserRealtimeRTP()` | 查询用户实时 RTP | userID, gameID, minutes |
| `GetUserHourlyRTP()` | 查询用户小时级 RTP | userID, gameID, hours |
| `GetGameRealtimeRTP()` | 查询游戏实时 RTP | gameID, minutes |
| `GetGameHourlyRTP()` | 查询游戏小时级 RTP | gameID, hours |
| `GetActiveAlerts()` | 获取活跃告警 | limit |
| `DetectAbnormalRTP()` | 手动检测异常 RTP | highThreshold, lowThreshold, minSpins |
| `GetTopHighRTPUsers()` | 高 RTP 用户排行 | hours, minSpins, limit |
| `ResolveAlert()` | 解决告警 | alertID, notes |

## 🚀 部署步骤

### Step 1: 初始化数据库

```bash
clickhouse-client -d rtp_analytics --queries-file /Users/anthony/Documents/github/rtp_regress/script/clickhouse_init.sql
```

### Step 2: 部署 slot-game（多副本）

```bash
cd /Users/anthony/Documents/github/rtp_regress/platform-games/slot-game
go build -o slot-game .
./slot-game -config config.yaml
```

### Step 3: 部署 rtp-processor（单实例）

```bash
cd /Users/anthony/Documents/github/rtp_regress/rtp-processor
go mod tidy
go build -o rtp-processor .
./rtp-processor -config config.yaml
```

## 📁 相关文件路径

```
/Users/anthony/Documents/github/rtp_regress/
├── script/
│   ├── clickhouse_init.sql      # 数据库初始化 SQL
│   └── rtp_batch_aggregate.sql  # 批处理 SQL（参考）
│
├── platform-games/slot-game/
│   ├── main.go                  # 主程序（无批处理逻辑）
│   ├── rtp/
│   │   └── rtp_service.go       # RTP 查询服务
│   └── ...
│
└── rtp-processor/              # ⭐ 独立批处理服务
    ├── main.go                  # 服务入口
    ├── config/
    │   └── config.go            # 配置加载
    ├── processor/
    │   └── batch_service.go     # 批处理核心逻辑
    ├── config.yaml              # 配置文件
    └── README.md                # 使用文档
```

## 🔧 运维

### 扩缩容

| 组件 | 扩缩方式 |
|------|---------|
| slot-game | 增加/减少副本数（无状态） |
| rtp-processor | **保持单实例**，如需提升性能可优化 SQL 或升级硬件 |

### 监控

rtp-processor 日志输出：
```
[BatchService] 定时批处理服务已启动
[BatchService] 开始聚合批次...
[BatchService] 用户实时聚合完成
[BatchService] 聚合批次完成，耗时: 1.2s
```

### 故障排查

1. **聚合数据为空**：检查 game_log_detail_local 是否有数据
2. **告警未生成**：检查 rtp_user_realtime_local 是否有符合条件的数据
3. **性能慢**：增加 ClickHouse 并发连接数或优化 SQL
