# Flink RocketMQ Processor

基于 Flink 2.2.1 的流批一体 RTP 计算服务。消费 RocketMQ 游戏日志，写入明细数据并实时计算 RTP，结果存入 ClickHouse。

## 数据流

```
游戏服务 → RocketMQ → Flink Stream
                          │
                          ├─→ Sink1: game_log_detail_local (明细)
                          │
                          └─→ Window Aggregate → Sink2: rtp_statistics_local (统计)
```

## 架构设计

本架构属于现代 Kappa 架构的演进版本，采用 Flink + ClickHouse 的流批一体方案：

```
游戏服务 → RocketMQ → Flink Stream → ClickHouse (明细 + 统计)
                        ↑
                   历史数据重放（批处理模式：数据修正、重算）
```

### 与传统架构对比

| 特征 | Lambda | Kappa | 本架构 |
|------|--------|-------|--------|
| 代码路径 | 双套（批+流） | 单套 | 单套 |
| 批处理引擎 | Spark | 流引擎重放 | Flink 批处理 |
| 主存储 | HDFS | Kafka 日志 | ClickHouse |
| 数据修正 | 批处理重跑 | 消息重放 | 消息重放 |

### 核心优势

- **代码复用**：流批共用一套 Flink Job，避免维护双套代码
- **实时性**：Flink 流处理实现秒级 RTP 计算
- **数据修正**：重放 RocketMQ 历史消息，随时重算任意时间段数据
- **存储分离**：ClickHouse 专注列式存储，计算与存储解耦
- **查询灵活**：保留明细数据，支持任意维度查询

## 技术栈

- **Flink**: 2.2.1
- **RocketMQ**: 4.9.8
- **ClickHouse**: 24.0
- **Nacos**: 2.0
- **Java**: 11

## 功能

| 功能 | 输出表 | 说明 |
|------|--------|------|
| 明细写入 | game_log_detail_local | 原始游戏日志，支持任意维度查询 |
| 全局RTP | rtp_statistics_local | 1分钟滚动窗口 |
| 终端RTP | rtp_statistics_local | 1分钟滚动窗口 |
| 玩家RTP | rtp_statistics_local | 5分钟滚动窗口 |

## 构建

```bash
mvn clean package
```

## 运行

### 本地模式（开发测试）

```bash
mvn exec:java -Dexec.mainClass="com.rtp.flink.RocketMQFlinkJob"
```

### Flink 集群模式（生产）

```bash
flink run -Dnacos.serverAddr=127.0.0.1:8848 \
         -c com.rtp.flink.RocketMQFlinkJob \
         target/flink-rmg-processor-1.0-SNAPSHOT.jar
```

## 消息格式

```json
{
  "gameId": "game-123",
  "integratorId": "integrator-001",
  "playerId": "player-001",
  "terminal": "web",
  "betAmount": 100.00,
  "winAmount": 95.00
}
```

## ClickHouse 表

### game_log_detail_local（明细）

| 字段 | 类型 | 说明 |
|------|------|------|
| game_id | String | 游戏局ID |
| player_id | String | 玩家ID |
| terminal | String | 终端类型 |
| bet_amount | Decimal(18,2) | 下注金额 |
| win_amount | Decimal(18,2) | 赔付金额 |
| created_at | DateTime64 | 创建时间 |

### rtp_statistics_local（统计）

| 字段 | 类型 | 说明 |
|------|------|------|
| dimension_type | String | global/terminal/player |
| dimension_key | String | 终端ID或玩家ID |
| total_bet_amount | Decimal(18,2) | 总下注 |
| total_win_amount | Decimal(18,2) | 总赔付 |
| rtp | Decimal(10,2) | RTP百分比 |
| window_start | DateTime64 | 窗口开始 |
| window_end | DateTime64 | 窗口结束 |
