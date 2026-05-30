# Slot Game Service

基于Golang的Slot游戏服务，使用Gin框架、Viper配置管理和Nacos配置中心。

## 功能特性

- 🎰 完整的Slot游戏逻辑实现
- 🚀 高性能HTTP服务
- 🔄 Nacos动态配置管理
- 📊 游戏统计和分析
- 🎯 支持多线路下注
- 🔮 支持Wild和Scatter符号
- 🎁 支持特殊功能触发

## 技术栈

- **Web框架**: Gin
- **配置管理**: Viper + Nacos
- **数据库**: MySQL + ClickHouse
- **语言**: Go 1.25.0

## 项目结构

```
platform-games/slot-game/
├── main.go                    # 主程序入口
├── config.yaml                # 应用配置文件
├── go.mod                     # Go模块定义
├── config/                    # 配置管理
│   └── app_config.go
├── controllers/               # 控制器层
│   └── game_controller.go
├── models/                    # 数据模型
│   ├── game_config.go
│   └── game_logic.go
├── nacos/                     # Nacos集成
│   └── config_manager.go
└── routers/                   # 路由配置
    └── router.go
```

## 快速开始

### 1. 环境准备

- Go 1.25.0+
- Nacos Server
- MySQL Server
- ClickHouse Server

### 2. 安装依赖

```bash
cd platform-games/slot-game
go mod download
go mod tidy
```

### 3. 配置Nacos

将游戏配置JSON上传到Nacos：
- **DataID**: `slot_game_config`
- **Group**: `SLOT_GAME_GROUP`
- **Namespace**: `public`

示例配置文件：`../../script/nacos_game_config.json`

### 4. 修改配置

编辑 `config.yaml` 文件，修改数据库和Nacos连接信息。

### 5. 启动服务

```bash
go run main.go
```

## 游戏配置说明

### 符号类型

- **normal**: 普通符号，需要按线路匹配
- **wild**: 万能符号，可替代普通符号
- **scatter**: 散布符号，任意位置触发

### 重量配置

每个卷轴的符号权重决定了符号出现的概率：
- 权重越大，出现概率越高
- 建议总和为100或其他数值

### 赔付倍数

每个符号可以配置不同连击数的赔付倍数：
- `match_count`: 连击数（2-5）
- `multiplier`: 赔付倍数
- `is_bet_line`: 是否需要匹配下注线

## 性能优化

1. **本地缓存**: Nacos配置自动缓存到本地
2. **动态更新**: 支持配置热更新
3. **并发处理**: Gin框架支持高并发
4. **轻量级**: 无状态设计，支持水平扩展

## 监控和日志

- 日志输出到 `logs/app.log`
- 支持JSON格式日志
- 自动日志轮转和压缩

## 开发建议

1. **配置管理**: 使用Nacos集中管理游戏配置
2. **数据持久化**: 游戏数据需要异步写入MySQL和ClickHouse
3. **限流控制**: 建议添加用户级限流
4. **监控告警**: 集成Prometheus进行监控

## 许可证

MIT License