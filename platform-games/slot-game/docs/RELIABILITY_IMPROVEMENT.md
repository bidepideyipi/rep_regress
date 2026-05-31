# ClickHouse写入可靠性改进方案

## 当前风险分析

### 数据丢失场景
| 场景 | 风险级别 | 原因 |
|------|----------|------|
| SIGKILL (kill -9) | 高 | 无法捕获信号，缓冲区数据丢失 |
| 程序panic/崩溃 | 中 | defer可能不执行 |
| ClickHouse宕机 | 中 | 重试失败后数据丢失 |
| 网络中断 | 中 | 超时后数据丢失 |
| 优雅关闭 | 低 | 已实现刷新机制 |

## 改进方案

### 方案一：本地WAL预写日志（推荐）

在写入ClickHouse前先写本地WAL，保证数据不丢失：

```go
// 结构：内存缓冲 + 本地WAL + ClickHouse
// 写入路径：Spin请求 → 内存缓冲 → 本地WAL → ClickHouse
//                           ↑___________↑
//                              异步读取重发
```

**优点：**
- 进程崩溃后可从WAL恢复
- ClickHouse故障时可继续接收请求
- 可实现"至少一次"语义

**缺点：**
- 需要磁盘IO
- 需要WAL清理机制

### 方案二：同步写入（低吞吐）

```go
// 每次spin请求同步写入ClickHouse
func (gc *GameController) Spin(c *gin.Context) {
    // ... 游戏逻辑 ...
    // 同步写入日志
    if err := gc.batchWriter.WriteSync(logEntry); err != nil {
        // 返回错误或降级处理
    }
    c.JSON(http.StatusOK, response)
}
```

**优点：**
- 数据可靠性最高
- 实现简单

**缺点：**
- 严重影响响应时间
- ClickHouse故障影响游戏服务

### 方案三：双写/消息队列（生产级）

```go
// Spin请求 → Kafka → ClickHouse（消费者）
// 或 Spin请求 → 本地队列 + ClickHouse双写
```

**优点：**
- 高可用
- 解耦服务

**缺点：**
- 需要额外组件
- 复杂度高

## 推荐实现：方案一

### 目录结构
```
slot-game/
├── wal/                    # WAL目录
│   ├── 20250530.wal       # 按天分片
│   └── 20250531.wal
├── models/
│   └── wal_writer.go      # WAL写入器
└── main.go                # 启动时恢复WAL
```

### 核心代码结构

```go
// WAL预写日志
type WALWriter struct {
    file      *os.File
    encoder   *json.Encoder
    mutex     sync.Mutex
    walDir    string
}

func (w *WALWriter) Write(entry GameLogDetail) error {
    w.mutex.Lock()
    defer w.mutex.Unlock()

    data, _ := json.Marshal(entry)
    _, err := w.file.Write(append(data, '\n'))
    return err
}

// 批量写入器集成WAL
type BatchWriter struct {
    conn        driver.Conn
    config      *BatchConfig
    buffer      []GameLogDetail
    wal         *WALWriter      // 新增
    // ...
}

// 写入时先写WAL
func (bw *BatchWriter) Write(entry GameLogDetail) error {
    // 1. 先写WAL
    if err := bw.wal.Write(entry); err != nil {
        return err  // WAL失败则拒绝
    }

    // 2. 再写内存缓冲
    bw.buffer = append(bw.buffer, entry)
    return nil
}

// 启动时恢复WAL
func RecoverFromWAL(walDir string, bw *BatchWriter) {
    files, _ := os.ReadDir(walDir)
    for _, file := range files {
        // 读取WAL文件
        // 重放未写入的数据
        // 确认写入成功后删除WAL
    }
}
```

### 配置建议

```yaml
clickhouse:
  wal:
    enabled: true
    dir: "./wal"
    max_file_size: 100MB
    retention_days: 7
  batch:
    batch_size: 100
    flush_interval: 5s
    sync_on_critical: true  # 关键操作同步flush
```

## 监控指标

建议添加以下监控：

1. **缓冲区大小**：`clickhouse_buffer_size`
2. **WAL积压**：`clickhouse_wal_size`
3. **写入成功率**：`clickhouse_write_success_rate`
4. **WAL恢复耗时**：`clickhouse_wal_recovery_duration`

## 结论

当前实现适合**内网环境 + 可接受少量丢失**的场景。

如果需要**严格不丢失**，建议：
1. 短期：添加本地WAL（方案一）
2. 长期：引入消息队列（方案三）
