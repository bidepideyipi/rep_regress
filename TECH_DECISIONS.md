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

| 维度          | 方案 A (SELECT FOR UPDATE) | 方案 B (ON DUPLICATE KEY UPDATE) |
| ----------- | ------------------------ | ------------------------------ |
| **数据库交互次数** | 2 次（SELECT + UPDATE）     | 1 次                            |
| **行锁持有时间**  | 长（整个事务期间）                | 短（仅 SQL 执行瞬间）                  |
| **并发性能**    | 较低（锁竞争严重）                | 较高                             |
| **网络开销**    | 较高                       | 较低                             |
| **代码复杂度**   | 较高（需手动管理事务）              | 较低                             |
| **适用场景**    | 需要读取当前值做复杂判断             | 简单增量操作                         |

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

RTP 处理服务需要从 RocketMQ 批量消费游戏日志消息写入 ClickHouse ，批量消费是为了提高ClickHouse写入效率（理论支持，尚未技术论证，标记为技术债务）。

注意：

1、依赖MQ的批量消费特性，不允许自定义实现程序的内存缓冲区。

2、保证在ClickHouse批量写入后再ACK，保证消费不会丢失。（理论支持：ClickHouse可以保证原子性和幂等性，论证工作，标记为技术债务）

3、Push模式无法在顺序消费的模式下实现并发（批量）

4、Pull模式在borker变化后需要手工切换指向，所以不采纳。

#### 测试方法

1、在 RocketMQ 中创建一个主题（Topic），并配置一个消费者组（Consumer Group）。
2、启动 RTP 处理服务，确保消费者成功连接到主题。
3、在 RocketMQ 中发送测试消息到该主题，验证消费者是否成功接收并处理消息。

堆积10万条消息后再启动消息者

```shell
// 发送测试消息
cd tools
go run rmq_bench.go -count 100000 -rate 0
`````

```text
=== 发送完成 ===
2026/06/06 11:07:02 总耗时: 31.812s
2026/06/06 11:07:02 成功: 100000
2026/06/06 11:07:02 失败: 0
2026/06/06 11:07:02 平均速率: 3143 msg/s
```

通过mqadmin命令查看消费速率

```shell
./mqadmin consumerProgress -n localhost:9876
```````

```text
#Group                       #Count  #Version   #Type  #Model        #TPS
slot_game_consumer_group       1     V4_5_2     PUSH    CLUSTERING    1666
```

## 变更历史

| 日期         | 版本  | 变更说明                                         |
| ---------- | --- | -------------------------------------------- |
| 2024-06-03 | 1.0 | 初始版本，记录 ADR-001 和 ADR-002                    |
| 2024-06-03 | 1.1 | 添加 ADR-003，记录 RocketMQ Pull 批量大小配置的排查过程和正确参数 |
| 2024-06-06 | 1.2 | 矫正了 ADR-002 和 ADR-003 等智谱AI的幻觉生成内容 |

