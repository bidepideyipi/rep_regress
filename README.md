# RTP Regress 游戏平台

## 项目概述

高性能、高并发的 SLOT 游戏平台，支持浏览器端游戏体验，提供统一的认证网关服务，实现实时的 RTP（Return to Player）计算和监控。

## 项目结构

```
rtp_regress/
├── platform-admin/      # 管理后台（GoAdmin 框架）
├── platform-games/      # 游戏服务
│   ├── gateway/        # API 网关服务
│   └── slot-game/      # SLOT 游戏服务
├── rtp-processor/      # RTP 计算服务
├── tools/              # 工具集
├── script/             # 脚本文件
└── doc/                # 文档目录
```

## 技术栈

- **语言**: Go 1.24.0
- **服务注册/配置中心**: Nacos 2.x
- **消息队列**: RocketMQ 5.x
- **数据存储**: ClickHouse、MySQL、Redis
- **Web 框架**: GoAdmin

## Mac 本地开发环境搭建

### 环境要求

- macOS 12.0+
- Homebrew
- Go 1.24.0+
- JDK 8+ (用于 RocketMQ)
- MySQL 8.0+
- Redis 7.x

### 1. 安装基础依赖

```bash
# 安装 Go
brew install go

# 安装 MySQL
brew install mysql
brew services start mysql

# 安装 Redis
brew install redis
brew services start redis

# 安装 JDK（RocketMQ 需要）
brew install openjdk@8
sudo ln -sfn /opt/homebrew/opt/openjdk@8/libexec/openjdk.jdk /Library/Java/JavaVirtualMachines/openjdk-8.jdk
```

### 2. 安装 ClickHouse

```bash
# 安装 ClickHouse
brew install clickhouse

# 启动 ClickHouse
brew services start clickhouse

# 验证安装
clickhouse-client
```

### 3. 安装 Nacos

```bash
# 下载 Nacos 2.x
cd /opt
sudo mkdir -p nacos
cd nacos
sudo curl -L https://github.com/alibaba/nacos/releases/download/2.2.3/nacos-server-2.2.3.tar.gz -o nacos-server.tar.gz
sudo tar -xzf nacos-server.tar.gz
sudo mv nacos nacos-server

# 单机模式启动
cd nacos-server/bin
sudo sh startup.sh -m standalone

# 验证安装
# 访问 http://localhost:8848/nacos
# 默认用户名/密码: nacos/nacos
```

### 4. 安装 RocketMQ

```bash
# 下载 RocketMQ 5.x
cd /opt
sudo mkdir -p rocketmq
cd rocketmq
sudo curl -L https://archive.apache.org/dist/rocketmq/5.1.0/rocketmq-all-5.1.0-bin-release.zip -o rocketmq.zip
sudo unzip rocketmq.zip
sudo mv rocketmq-all-5.1.0-bin-release rocketmq

# 配置环境变量
echo 'export ROCKETMQ_HOME=/opt/rocketmq/rocketmq' >> ~/.zshrc
echo 'export PATH=$PATH:$ROCKETMQ_HOME/bin' >> ~/.zshrc
source ~/.zshrc

# 调整 JVM 参数（本地开发环境）
cd $ROCKETMQ_HOME/bin
# 编辑 runbroker.sh，将 JVM_XSS 调整为 256k
# 编辑 runserver.sh，将 JAVA_OPT 调整为较小的值

# 启动 NameServer
cd $ROCKETMQ_HOME/bin
sh mqnamesrv &
# 或使用 nohup 在后台运行
nohup sh mqnamesrv > ~/logs/mqnamesrv.log 2>&1 &

# 验证 NameServer 启动
tail -f ~/logs/mqnamesrv.log

# 启动 Broker
cd $ROCKETMQ_HOME/bin
sh mqbroker -n localhost:9876 &
# 或使用 nohup 在后台运行
nohup sh mqbroker -n localhost:9876 > ~/logs/mqbroker.log 2>&1 &

# 验证 Broker 启动
tail -f ~/logs/mqbroker.log

# 测试 RocketMQ
# 生产消息测试
sh tools.sh org.apache.rocketmq.example.quickstart.Producer

# 消费消息测试
sh tools.sh org.apache.rocketmq.example.quickstart.Consumer
```

## 服务启动顺序

⚠️ **重要**: 必须按照以下顺序启动服务，否则会出现连接失败的问题。

### 阶段一：基础存储启动

1. **启动 MySQL**
   ```bash
   brew services start mysql
   # 或
   mysql.server start

   # 验证
   mysql -u root -p -e "SELECT 1"
   ```

2. **启动 Redis**
   ```bash
   brew services start redis
   # 或
   redis-server /opt/homebrew/etc/redis.conf

   # 验证
   redis-cli ping
   # 应返回 PONG
   ```

### 阶段二：中间件启动

3. **启动 ClickHouse**
   ```bash
   brew services start clickhouse

   # 验证连接
   clickhouse-client --query "SELECT 1"

   # 创建数据库
   clickhouse-client --query "CREATE DATABASE IF NOT EXISTS rtp"

   # 查看版本
   clickhouse-client --query "SELECT version()"
   ```

4. **启动 Nacos**（必须先于应用启动）
   ```bash
   cd /opt/nacos/nacos-server/bin
   sh startup.sh -m standalone

   # 验证 Nacos 启动
   curl http://localhost:8848/nacos/v1/console/health/readiness

   # 访问控制台
   open http://localhost:8848/nacos
   ```

5. **启动 RocketMQ NameServer**（必须先于 Broker）
   ```bash
   cd /opt/rocketmq/rocketmq/bin
   nohup sh mqnamesrv > ~/logs/mqnamesrv.log 2>&1 &

   # 验证启动
   tail -f ~/logs/mqnamesrv.log
   # 看到 "The Name Server boot success" 即启动成功
   ```

6. **启动 RocketMQ Broker**（依赖 NameServer）
   ```bash
   cd /opt/rocketmq/rocketmq/bin
   nohup sh mqbroker -n localhost:9876 > ~/logs/mqbroker.log 2>&1 &

   # 验证启动
   tail -f ~/logs/mqbroker.log
   # 看到 "The broker boot success" 即启动成功
   ```

### 阶段三：应用服务启动

7. **初始化数据库**
   ```bash
   # 创建数据库
   mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS goadmin;"

   # 导入初始化脚本（如果有）
   mysql -u root -p goadmin < script/init.sql
   ```

8. **在 Nacos 中创建配置**

   访问 http://localhost:8848/nacos，进入「配置管理」→「配置列表」，点击「+」创建配置：

   **Data ID**: `app-config`
   **Group**: `DEFAULT_GROUP`
   **命名空间**: `public`

   **配置内容**:
   ```json
   {
     "clickhouse": {
       "host": "127.0.0.1",
       "port": 9000,
       "username": "default",
       "password": "",
       "database": "rtp"
     },
     "mysql": {
       "host": "127.0.0.1",
       "port": 3306,
       "username": "root",
       "password": "China!1234",
       "database": "goadmin"
     },
     "rocket_mq": {
       "name_servers": ["127.0.0.1:9876"],
       "producer": {
         "group_name": "rtp_producer_group",
         "topic": "game_events"
       },
       "consumer": {
         "group_name": "rtp_consumer_group",
         "topic": "game_events",
         "batch_size": 100
       }
     },
     "aggregate_user_interval": "5m",
     "aggregate_game_interval": "5m",
     "alert_interval": "10m"
   }
   ```

9. **初始化 ClickHouse 数据库**
   ```bash
   clickhouse-client --query "CREATE DATABASE IF NOT EXISTS rtp"
   clickhouse-client --query "CREATE TABLE IF NOT EXISTS rtp.game_events (..."  # 根据 script 目录中的 DDL 脚本执行
   ```

10. **启动 RTP Processor**
    ```bash
    cd rtp-processor
    go mod download
    go run main.go --nacos_addr=127.0.0.1
    ```

11. **启动 Platform Admin**
    ```bash
    cd platform-admin
    go mod download
    go run main.go
    # 访问 http://localhost:8080/pladmin
    ```

12. **启动游戏服务**
    ```bash
    # Gateway
    cd platform-games/gateway
    go mod download
    go run main.go

    # Slot Game
    cd platform-games/slot-game
    go mod download
    go run main.go
    ```

## 服务关闭顺序

关闭时应该按照相反的顺序进行：

```bash
# 1. 关闭应用服务
# 停止各个 Go 进程（Ctrl+C 或 kill）

# 2. 关闭 RocketMQ Broker
cd /opt/rocketmq/rocketmq/bin
sh mqshutdown broker

# 3. 关闭 RocketMQ NameServer
sh mqshutdown namesrv

# 4. 关闭 Nacos
cd /opt/nacos/nacos-server/bin
sh shutdown.sh

# 5. 关闭 ClickHouse
brew services stop clickhouse

# 6. 关闭 Redis
brew services stop redis

# 7. 关闭 MySQL
brew services stop mysql
```

## 常见问题

### RocketMQ 启动失败

**问题**: 内存不足导致 RocketMQ 启动失败

**解决**: 修改 JVM 参数
```bash
# 编辑 $ROCKETMQ_HOME/bin/runbroker.sh
# 找到 JAVA_OPT，调整为：
JAVA_OPT="${JAVA_OPT} -Xms256m -Xmx256m"

# 编辑 $ROCKETMQ_HOME/bin/runserver.sh
# 调整为：
JAVA_OPT="${JAVA_OPT} -Xms256m -Xmx256m -XX:MetaspaceSize=128m -XX:MaxMetaspaceSize=128m"
```

### Nacos 无法连接

**问题**: 应用启动时无法连接到 Nacos

**检查**:
1. 确认 Nacos 已启动: `ps aux | grep nacos`
2. 检查端口占用: `lsof -i:8848`
3. 查看日志: `tail -f /opt/nacos/nacos-server/logs/start.out`

### ClickHouse 连接被拒绝

**问题**: 连接 ClickHouse 时提示连接被拒绝

**解决**:
```bash
# 检查 ClickHouse 状态
brew services list

# 重启 ClickHouse
brew services restart clickhouse

# 检查配置文件
cat /opt/homebrew/etc/clickhouse-server/config.xml
```

### RocketMQ Broker 无法连接 NameServer

**问题**: Broker 日志显示连接 NameServer 失败

**解决**:
1. 确认 NameServer 已启动
2. 检查网络连接: `telnet localhost 9876`
3. 确认 Broker 启动参数正确: `-n localhost:9876`

## 开发指南

### 代码风格

项目遵循 Go 官方代码风格指南，使用 `gofmt` 格式化代码：

```bash
gofmt -w -s .
```

### 测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./rtp-processor/...

# 带覆盖率的测试
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 构建生产版本

```bash
# 构建 RTP Processor
cd rtp-processor
go build -ldflags="-s -w" -o rtp-processor main.go

# 构建 Platform Admin
cd platform-admin
go build -ldflags="-s -w" -o platform-admin main.go
```

## 相关文档

- [PRD.md](./PRD.md) - 产品需求文档
- [DEPLOYMENT_ARCHITECTURE.md](./DEPLOYMENT_ARCHITECTURE.md) - 部署架构文档
- [GAME_SERVICE_API.md](./GAME_SERVICE_API.md) - 游戏服务 API 文档
- [CK_VS_MONGODB.md](./CK_VS_MONGODB.md) - ClickHouse vs MongoDB 对比
- [TECH_DECISIONS.md](./TECH_DECISIONS.md) - 技术决策文档 (架构决策记录)

## License

Copyright © 2024 RTP Platform Team
