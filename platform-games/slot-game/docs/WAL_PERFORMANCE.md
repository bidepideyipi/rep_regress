# WAL性能优化方案

## 问题：WAL在高并发下的性能瓶颈

### 核心瓶颈
```
fsync系统调用：等待磁盘确认数据落盘
- HDD硬盘：~5-10ms
- SSD硬盘：~1-2ms
- NVMe SSD：~0.1-0.5ms
```

单次WAL写入包含fsync时，QPS通常只有200-500

## 优化方案

### 方案1：批量fsync（推荐）

**思路**：积累多个写入后一次性fsync

```go
type WALWriter struct {
    buffer      chan []byte      // 缓冲通道
    file        *os.File
    fsyncTicker *time.Ticker
    batch       [][]byte        // 待fsync的批次
    batchSize   int             // 批次大小
}

func (w *WALWriter) Write(data []byte) error {
    // 1. 快速写入buffer（无fsync）
    w.buffer <- data
    return nil
}

func (w *WALWriter) flushLoop() {
    for {
        select {
        case data := <-w.buffer:
            w.batch = append(w.batch, data)
            // 达到批量大小，触发fsync
            if len(w.batch) >= w.batchSize {
                w.doFsync()
            }
        case <-w.fsyncTicker.C:
            // 定时fsync（防止数据积压过多）
            w.doFsync()
        }
    }
}

func (w *WALWriter) doFsync() {
    for _, data := range w.batch {
        w.file.Write(data)
    }
    w.file.Sync()  // 批量fsync，分摊成本
    w.batch = w.batch[:0]
}
```

**效果**：100个写入共享1次fsync，平均延迟从5ms降到0.05ms

### 方案2：使用高效存储

**Open-onothing O_DIRECT / Direct I/O**

```go
// 使用mmap映射文件
type MmapWAL struct {
    data   []byte
    offset int
    mutex  sync.Mutex
}

func (w *MmapWAL) Write(data []byte) error {
    w.mutex.Lock()
    copy(w.data[w.offset:], data)
    w.offset += len(data)
    w.mutex.Unlock()
    // 后台定时msync
    return nil
}
```

### 方案3：无fsync + 崩溃扫描（平衡方案）

**思路**：不强制fsync，崩溃时扫描完整文件

```go
// 写入时不fsync，依赖OS定期刷盘
func (w *WALWriter) Write(data []byte) error {
    w.mutex.Lock()
    defer w.mutex.Unlock()
    w.file.Write(data)  // 只写不fsync
    return nil
}

// 崩溃恢复时扫描完整文件
func (w *WALWriter) Recover() {
    // 从文件头扫描，跳过已确认的记录
    // 通过"已确认标记"判断哪些数据需要重放
}
```

### 方案4：分区WAL（减少锁竞争）

```go
type ShardedWAL struct {
    shards []*WALWriter  // 多个WAL分区
    count  int
}

func (w *ShardedWAL) Write(key string, data []byte) error {
    // 根据key选择分区，减少锁竞争
    shardIndex := hash(key) % w.count
    return w.shards[shardIndex].Write(data)
}
```

**效果**：4个分区，理论上QPS可提升4倍

## 性能对比

| 方案 | 延迟 | QPS | 可靠性 | 复杂度 |
|------|------|-----|--------|--------|
| 无WAL | 0.1ms | 10000+ | 易丢失 | 低 |
| WAL+fsync | 5ms | 200-500 | 不丢失 | 低 |
| 批量fsync | 1ms | 1000+ | 不丢失 | 中 |
| 无fsync扫描 | 0.5ms | 2000+ | 可能丢少量 | 中 |
| 分区WAL | 0.3ms | 3000+ | 不丢失 | 高 |

## 推荐方案

### 对于游戏日志场景

**方案A：批量fsync WAL**（推荐）
```go
// 配置
wal:
  enabled: true
  batch_size: 100      // 100条批量fsync
  flush_interval: 100ms // 最长100ms刷新
  sharding: 4          // 4个分区
```
- QPS：2000+
- 最多丢失100ms的数据
- 实现复杂度中等

**方案B：双写 + 降级**（生产级）
```go
// 优先写ClickHouse，失败时写本地WAL
func Write(entry GameLogDetail) error {
    // 先尝试直接写ClickHouse
    if err := clickhouse.Write(entry); err == nil {
        return nil
    }

    // 失败时写WAL作为降级
    return wal.Write(entry)
}
```
- 正常情况无额外开销
- 故障时自动降级
- 兼顾性能和可靠性

**方案C：纯内存 + 监控**（低成本）
```go
// 接受少量丢失，通过监控发现异常
- 无WAL开销
- 监控ClickHouse写入延迟
- 发现积压时告警
```
- QPS最高
- 适合内网环境
- 需要运维配合

## 结论

1. **WAL不是银弹**：确实有性能开销
2. **批量fsync**：是性价比最高的方案
3. **按需选择**：根据业务容忍度选择方案
4. **监控先行**：无论选什么方案，都要做好监控

对于游戏日志场景，建议：
- 开发/测试：纯内存方案
- 预发布：批量fsync WAL
- 生产：双写+降级或引入消息队列
