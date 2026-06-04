# 技术决策文档 (Architecture Decision Records)

本文档记录项目中的重要技术决策及其背景。

---

## ADR-001: Jackpot 累计奖池的并发更新策略

**状态**: 已采用
**日期**: 2024-06-03
**影响模块**: `rtp-processor/jackpot`

### 背景

Jackpot 累计奖池需要在高并发场景下保证数据一致性，同时考虑性能优化。每次玩家下注时，需要按固定比例累加到对应的奖池中。

### 考虑的方案

#### 方案 A：SELECT FOR UPDATE（悲观锁）

```go
// 需要两次数据库交互
BEGIN;
SELECT current_amount FROM jackpot_pool
WHERE game_id = '1' AND pool_type = 'grand'
FOR UPDATE;  -- 加行锁

-- 应用层计算新金额
newAmount := currentAmount + betAmount * ratio

UPDATE jackpot_pool SET current_amount = ?
WHERE game_id = '1' AND pool_type = 'grand';

COMMIT;
```

#### 方案 B：INSERT ... ON DUPLICATE KEY UPDATE（原子操作）

```sql
-- 单次数据库交互，MySQL 保证原子性
INSERT INTO jackpot_pool (game_id, pool_type, current_amount, win_count)
VALUES ('1', 'grand', 0.5, 0)
ON DUPLICATE KEY UPDATE
    current_amount = current_amount + 0.5,
    win_count = win_count;
```

### 方案对比

| 维度 | 方案 A (SELECT FOR UPDATE) | 方案 B (ON DUPLICATE KEY UPDATE) |
|------|---------------------------|----------------------------------|
| **数据库交互次数** | 2 次（SELECT + UPDATE） | 1 次 |
| **行锁持有时间** | 长（整个事务期间） | 短（仅 SQL 执行瞬间） |
| **并发性能** | 较低（锁竞争严重） | 较高 |
| **网络开销** | 较高 | 较低 |
| **代码复杂度** | 较高（需手动管理事务） | 较低 |
| **适用场景** | 需要读取当前值做复杂判断 | 简单增量操作 |

### 决策

**采用方案 B：INSERT ... ON DUPLICATE KEY UPDATE**

### 理由

1. **业务特点匹配**
   - Jackpot 累加是纯粹的增量操作
   - 不需要读取当前值进行业务判断
   - 只需执行 `current_amount = current_amount + amount`

2. **性能优势**
   - 单次数据库交互，减少网络往返
   - 行锁时间极短，高并发下锁竞争更少
   - MySQL 内部保证原子性，无需应用层额外处理

3. **代码简洁性**
   - 一条 SQL 完成操作，代码更少
   - 无需手动管理事务边界
   - 降低了出错概率

4. **MySQL 保证**
   - `ON DUPLICATE KEY UPDATE` 是 MySQL 原子操作
   - 主键/唯一索引冲突时自动转为 UPDATE
   - 并发安全由数据库引擎保证

### 实现

代码位置：`rtp-processor/jackpot/manager.go:246-252`

```go
updateQuery := `
    INSERT INTO jackpot_pool (game_id, pool_type, current_amount, win_count)
    VALUES (?, ?, ?, 0)
    ON DUPLICATE KEY UPDATE
        current_amount = current_amount + ?,
        win_count = win_count
`
```

### 注意事项

1. **幂等性**：批量更新采用累加方式，天然幂等
2. **精度控制**：四舍五入到分，避免浮点误差
3. **池不存在处理**：首次会自动 INSERT 创建记录

### 对比：用户余额扣减为什么用 SELECT FOR UPDATE

用户余额操作采用 `SELECT FOR UPDATE`（代码位置：`platform-games/slot-game/dao/user_dao.go`），原因：

1. **需要读取当前值**
   - 检查余额是否充足：`if user.Balance < amount`
   - 计算扣减后余额：`newBalance = user.Balance - amount`

2. **业务判断复杂**
   - 检查用户状态
   - 余额不足时需返回错误
   - 需要记录扣减前后的余额

3. **需要事务回滚**
   - 业务逻辑失败时回滚
   - 余额不足时不执行扣减

**总结**：
- **Jackpot 累加** → 简单增量操作 → `ON DUPLICATE KEY UPDATE`
- **用户余额** → 需要读取判断 → `SELECT FOR UPDATE`

### 参考资料

- MySQL 官方文档：[INSERT ... ON DUPLICATE KEY UPDATE](https://dev.mysql.com/doc/refman/8.0/en/insert-on-duplicate.html)
- Go实现：`rtp-processor/jackpot/manager.go`

---

## ADR-002: RocketMQ 消费模式选择

**状态**: 已采用
**日期**: 2024-06-03
**影响模块**: `rtp-processor/consumer`

### 背景

RTP 处理服务需要从 RocketMQ 消费游戏日志消息，写入 ClickHouse 并计算 Jackpot。

### 考虑的方案

| 方案 | 特点 | 适用场景 |
|------|------|----------|
| Push 模式 | RocketMQ 主动推送消息 | 消息处理简单，消费速率稳定 |
| Pull 模式 | 消费者主动拉取消息 | 需要精细控制消费速率 |

### 决策

**采用 Pull 模式**

### 理由

1. **控制数据写入节奏**
   - ClickHouse 批量写入效率更高
   - 可控制 `Flush()` 时机，避免频繁写入

2. **与偏移量管理配合**
   - 成功 `Flush()` 后才持久化 offset
   - 保证数据不丢失：`Flush() → PersistOffset()`

3. **流量控制**
   - 可根据 ClickHouse 写入压力调整拉取频率
   - 避免消息堆积

### 实现

代码位置：`rtp-processor/consumer/consumer.go`

```go
// 拉取消息
pullResult, err := gc.rocketConsumer.Pull(ctx, 100)

// 写入缓冲区
gc.chWriter.AddToBuffer(logEntry)

// 刷新到 ClickHouse
gc.chWriter.Flush()

// 成功后才持久化 offset
gc.rocketConsumer.PersistOffset(ctx, topic)
```

---

## ADR-003: RocketMQ Broker Pull 批量大小配置

**状态**: 已采用
**日期**: 2024-06-03
**影响模块**: `rtp-processor/consumer`

### 背景

RTP 处理服务使用 RocketMQ Pull 模式消费消息，实际拉取到 32 条消息，但代码中指定了 `Pull(ctx, 100)`。需要找到正确的 Broker 配置参数。

### 问题分析

1. **客户端请求**: `Pull(ctx, 100)` - 期望拉取 100 条
2. **实际返回**: 32 条消息
3. **错误配置尝试**:
   - ❌ `pullBatchSize=64` - 非有效 Broker 参数
   - ❌ `defaultQueryMaxNum=64` - 控制**查询**消息，非 Pull 消息
   - ✅ `maxTransferCountOnMessageInMemory=64` - 正确参数

### 根本原因

**RocketMQ 4.9.8 有多个消息传输限制参数**，不同参数控制不同场景：

### 正确的 Broker 配置参数

| 参数 | 默认值 | 作用 |
|------|--------|------|
| **maxTransferCountOnMessageInMemory** | **32** | 单次 Pull 最大消息数（内存消息）← **控制 Pull 批量的关键参数** |
| maxTransferBytesOnMessageInMemory | 262144 (256KB) | 单次 Pull 最大字节数（内存消息） |
| maxTransferCountOnMessageInDisk | 8 | 单次 Pull 最大消息数（磁盘消息） |
| maxTransferBytesOnMessageInDisk | 65536 (64KB) | 单次 Pull 最大字节数（磁盘消息） |

### 决策

**修改 `maxTransferCountOnMessageInMemory` 参数**

### 实现

在 `broker.conf` 中添加：

```properties
# 单次 Pull 最大消息数量（内存消息）- 这是控制 Pull 批量的关键参数
maxTransferCountOnMessageInMemory=64

# 可选：单次 Pull 最大字节数（内存消息）
maxTransferBytesOnMessageInMemory=262144

# 可选：单次 Pull 最大消息数量（磁盘消息）
maxTransferCountOnMessageInDisk=8

# 可选：单次 Pull 最大字节数（磁盘消息）
maxTransferBytesOnMessageInDisk=65536
```

重启 Broker：
```bash
cd ~/Env/rocketmq-all-4.9.8-bin-release/bin
sh mqshutdown broker
nohup sh mqbroker -c ~/Env/rocketmq-all-4.9.8-bin-release/conf/broker.conf -n localhost:9876 > ~/logs/mqbroker.log 2>&1 &
```

### 参数说明

1. **maxTransferCountOnMessageInMemory**
   - 控制消费者单次 Pull 能获取的最大消息数
   - 仅当消息在内存中时生效
   - 默认 32 是平衡性能和内存使用的合理值

2. **maxTransferBytesOnMessageInMemory**
   - 控制单次 Pull 的最大字节数
   - 与 `maxTransferCountOnMessageInMemory` 共同限制实际返回量

### 客户端参数说明

Go 客户端 `Pull(ctx, numbers)` 中的 `numbers` 参数：
- **含义**: 客户端期望的最大值（MaxMsgNums）
- **实际返回**: 由 Broker 的 `maxTransferCountOnMessageInMemory` 限制
- **关系**: `实际返回 = min(客户端请求, Broker限制)`

### 验证方法

```bash
# 1. 修改 broker.conf
maxTransferCountOnMessageInMemory=64

# 2. 重启 Broker
sh mqshutdown broker && nohup sh mqbroker -c .../broker.conf -n localhost:9876 > ~/logs/mqbroker.log 2>&1 &

# 3. 重启 rtp-processor
go run main.go

# 4. 查看日志确认
# 应该看到 "拉取到 64 条消息" 而不是 32 条
```

### 生产环境建议

| 场景 | 推荐值 | 理由 |
|------|--------|------|
| **默认** | 32 | 平衡性能和内存 |
| **高吞吐** | 64-128 | 减少 Broker 往返 |
| **低延迟** | 16-32 | 单次处理快，内存压力小 |

### 参考资料

- [GitHub Issue #2945 - broker busy or system busy](https://github.com/apache/rocketmq/issues/2945)
- [Server Configuration - Apache RocketMQ](https://rocketmq.apache.org/docs/4.x/parameterConfiguration/02server/)
- [RocketMQ 4.9.8 Broker Configuration](https://github.com/apache/rocketmq/releases/tag/rocketmq-4.9.8)

---

## 变更历史

| 日期 | 版本 | 变更说明 |
|------|------|----------|
| 2024-06-03 | 1.0 | 初始版本，记录 ADR-001 和 ADR-002 |
| 2024-06-03 | 1.1 | 添加 ADR-003，记录 RocketMQ Pull 批量大小配置的排查过程和正确参数 |
