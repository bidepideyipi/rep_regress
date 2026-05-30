# Platform Games Gateway

游戏平台统一网关服务，提供认证授权、请求路由、安全防护等功能。

## 功能特性

- **OAuth 2.0 风格认证**: API Key + HMAC-SHA256 签名验证
- **JWT 令牌管理**: RS256 算法，支持令牌生成和验证
- **AES 加密**: AES-256-GCM 用于敏感数据加密
- **安全防护**:
  - 重放攻击防护
  - 速率限制
  - IP 白名单
  - 输入验证
- **RESTful API**: 统一的响应格式

## 项目结构

```
gateway/
├── cmd/generate-keys/    # 密钥生成工具
├── config/                # 配置文件
├── internal/              # 内部包
│   ├── auth/             # 认证模块 (HMAC, JWT, AES)
│   ├── middleware/       # 中间件
│   ├── handlers/         # HTTP 处理器
│   ├── services/        # 业务服务
│   ├── models/          # 数据模型
│   ├── repository/      # 数据访问层
│   ├── security/        # 安全模块
│   └── router/          # 路由配置
├── pkg/response/         # 统一响应格式
├── scripts/              # 脚本工具
├── deployments/         # 部署配置
└── tests/               # 测试
```

## 快速开始

### 1. 安装依赖

```bash
make deps
```

### 2. 生成密钥

```bash
make keys
```

或使用密钥生成工具：

```bash
go run cmd/generate-keys/main.go -type rsa -output ./keys
```

### 3. 配置

编辑 `config/config.yaml` 文件，配置服务器、Redis、数据库等参数。

### 4. 运行

```bash
make run
```

### 5. 使用 Docker

```bash
cd deployments
docker-compose up
```

## API 接口

### 认证接口

#### 游戏访问授权
```http
POST /v1/auth/game-access
Content-Type: application/json

{
  "merchant_id": "merchant_001",
  "user_id": "user_12345",
  "game_id": "slot_game_v1",
  "language": "zh-CN",
  "currency": "CNY",
  "timestamp": 1699123456789,
  "nonce": "a1b2c3d4e5f6g7h8",
  "signature": "..."
}
```

#### 令牌验证
```http
POST /v1/auth/verify-token
Content-Type: application/json

{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "timestamp": 1699123456789,
  "signature": "..."
}
```

### 用户接口

#### 获取用户信息
```http
GET /v1/user/info
Authorization: Bearer <access_token>
```

#### 更新用户偏好
```http
PUT /v1/user/info
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "preferences": {
    "language": "zh-CN",
    "sound_enabled": true
  }
}
```

### 游戏接口

#### 游戏操作
```http
POST /v1/game/action
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "action": "spin",
  "game_id": "slot_game_v1",
  "game_data": {
    "bet_amount": 1.0,
    "bet_lines": 5
  }
}
```

### 财务接口

#### 查询余额
```http
GET /v1/finance/balance
Authorization: Bearer <access_token>
```

#### 交易记录
```http
GET /v1/finance/transactions?limit=10&offset=0
Authorization: Bearer <access_token>
```

## 签名计算

```javascript
// 1. 构建签名字符串
const params = {
  merchant_id: "merchant_001",
  user_id: "user_12345",
  game_id: "slot_game_v1"
};
const sortedKeys = Object.keys(params).sort();
const paramString = sortedKeys.map(k => `${k}=${params[k]}`).join('&');
const timestamp = Date.now();
const nonce = generateNonce(16);
const signString = `${paramString}&timestamp=${timestamp}&nonce=${nonce}`;

// 2. 使用 HMAC-SHA256 计算签名
const crypto = require('crypto');
const signature = crypto.createHmac('sha256', api_secret)
  .update(signString)
  .digest('hex');
```

## 测试

```bash
# 运行单元测试
make test

# 运行 API 测试
curl -X POST http://localhost:8080/v1/auth/game-access \
  -H "Content-Type: application/json" \
  -d '{...}'
```

## 构建部署

```bash
# 构建
make build

# Docker 镜像
make docker-build
```

## 配置说明

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| server.host | 服务器地址 | 0.0.0.0 |
| server.port | 服务器端口 | 8080 |
| server.mode | 运行模式 | release |
| auth.token_expires_in | 令牌有效期(秒) | 1800 |
| auth.timestamp_ttl | 时间戳有效期(秒) | 300 |
| security.rate_limit.requests_per_minute | 每分钟请求数 | 100 |
| security.enable_replay_protection | 重放攻击防护 | true |

## 许可证

Copyright © 2024 Platform Games Team
