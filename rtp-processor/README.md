# RTP Batch Processor

独立 RTP 批处理服务，用于定时聚合游戏日志数据并生成 RTP 指标和告警。

## 特性

- 独立部署，不依赖 slot-game 服务
- 支持水平扩展（多实例通过配置协调）
- **配置从 Nacos 获取**（与 slot-game 共享配置）
- 定时聚合用户/游戏 RTP 指标
- 自动检测异常 RTP 并生成告警
- 支持优雅启停

## 架构

```
┌─────────────────────────────────────────────────────────────┐
│                    rtp-processor                             │
│                   (独立服务，单实例)                           │
│                                                              │
│  ┌──────────────┐    ┌──────────────────────────────────┐   │
│  │ 定时调度器    │───▶│ 聚合任务                          │   │
│  │ 5分钟/10分钟  │    │ - 用户实时聚合 (5分钟)            │   │
│  └──────────────┘    │ - 用户小时聚合 (5分钟)             │   │
│                      │ - 游戏实时聚合 (5分钟)            │   │
│                      │ - 游戏小时聚合 (5分钟)            │   │
│                      │ - 告警检测 (10分钟)               │   │
│                      └──────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                        │
                        ▼
              ┌─────────────────┐
              │   ClickHouse    │
              │ rtp_analytics   │
              └─────────────────┘
```

## 配置

配置通过 Nacos 获取，data_id 为 `app-config`，group 为 `DEFAULT_GROUP`。

Nacos 配置示例：

```json
{
  "clickhouse": {
    "host": "127.0.0.1",
    "port": 9000,
    "username": "default",
    "password": "your_password",
    "database": "rtp_analytics"
  },
  "aggregate_interval": "5m",
  "alert_interval": "10m"
}
```

### 命令行参数

```bash
./rtp-processor -nacos_addr 127.0.0.1:8848 \
                -namespace your_namespace \
                -group DEFAULT_GROUP \
                -data_id app-config
```

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-nacos_addr` | 127.0.0.1:8848 | Nacos 服务器地址 |
| `-namespace` | (空) | Nacos 命名空间 |
| `-group` | DEFAULT_GROUP | Nacos 配置组 |
| `-data_id` | app-config | Nacos 配置 data id |

## 部署

### 1. 初始化数据库

```bash
clickhouse-client -u default --password 'your_password' -d rtp_analytics --queries-file /Users/anthony/Documents/github/rtp_regress/script/clickhouse_init.sql
```

### 2. 构建服务

```bash
cd /Users/anthony/Documents/github/rtp_regress/rtp-processor
go mod tidy
go build -o rtp-processor .
```

### 3. 启动服务

```bash
./rtp-processor -namespace your_namespace
```

### 4. 使用 systemd 管理（可选）

```ini
[Unit]
Description=RTP Batch Processor
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/rtp-processor
ExecStart=/opt/rtp-processor/rtp-processor -nacos_addr 127.0.0.1:8848 -namespace your_namespace
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

## 部署架构

```
                    ┌─────────────────┐
                    │     Nacos       │
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
                             │
                            ▼
                 ┌─────────────────────┐
                 │   rtp-processor     │
                 │   (独立服务，单实例)   │
                 └─────────────────────┘
```

## 监控

服务启动后会输出以下日志：

```
从 Nacos 加载配置成功
ClickHouse 连接成功: 127.0.0.1:9000
RTP 批处理服务启动
聚合间隔: 5m0s, 告警间隔: 10m0s
[BatchService] 定时批处理服务已启动
[BatchService] 开始聚合批次...
[BatchService] 用户实时聚合完成
[BatchService] 聚合批次完成，耗时: 1.2s
```

## 故障排查

### Nacos 连接失败

```
从 Nacos 获取配置失败: ...
```

检查：
1. Nacos 服务是否运行
2. 命令行参数是否正确
3. namespace 是否存在

### ClickHouse 连接失败

```
Ping ClickHouse 失败: ...
```

检查：
1. ClickHouse 服务是否运行
2. Nacos 配置中的 clickhouse 账号密码是否正确
