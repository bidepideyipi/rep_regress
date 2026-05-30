# 游戏访问机制 API 接口详细文档

## 📡 API 基础信息

### 基本信息

| 项目 | 说明 |
|------|------|
| **API版本** | v1.0 |
| **基础URL** | `https://api.platform.com/v1` |
| **认证方式** | API Key + HMAC-SHA256签名 |
| **数据格式** | JSON |
| **字符编码** | UTF-8 |
| **时间格式** | ISO 8601 (YYYY-MM-DDTHH:mm:ssZ) |

### 通用请求头

```http
Content-Type: application/json
X-API-Key: your_api_key_here
X-Timestamp: 1699123456789
X-Signature: calculated_signature_here
X-Request-ID: unique_request_id
```

### 通用响应格式

```json
{
  "success": true,
  "code": 200,
  "message": "Success",
  "data": {},
  "timestamp": 1699123456789,
  "request_id": "req_123456"
}
```

## 🔐 认证相关接口

### 1. 游戏访问授权

#### 接口信息

| 项目 | 内容 |
|------|------|
| **接口路径** | `/auth/game-access` |
| **请求方法** | POST |
| **接口描述** | 获取游戏访问权限，返回加密的游戏URL |
| **权限要求** | 需要有效的API Key |

#### 请求参数

```typescript
interface GameAccessRequest {
  merchant_id: string;        // 商户ID
  user_id: string;           // 用户ID (商户端的用户标识)
  game_id: string;           // 游戏ID
  language?: string;         // 语言偏好 (默认: zh-CN)
  currency?: string;         // 货币类型 (默认: CNY)
  return_url?: string;       // 游戏结束后返回的URL
  custom_data?: object;      // 自定义数据 (可选)
  timestamp: number;         // 当前时间戳 (毫秒)
  nonce: string;             // 随机字符串 (16位)
  signature: string;         // HMAC-SHA256签名
}
```

#### 请求示例

```json
{
  "merchant_id": "merchant_001",
  "user_id": "user_12345",
  "game_id": "slot_game_v1",
  "language": "zh-CN",
  "currency": "CNY",
  "return_url": "https://merchant.com/game-callback",
  "timestamp": 1699123456789,
  "nonce": "a1b2c3d4e5f6g7h8",
  "signature": "3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0"
}
```

#### 响应参数

```typescript
interface GameAccessResponse {
  success: boolean;
  code: number;
  message: string;
  data: {
    game_url: string;         // 加密的游戏URL
    access_token: string;     // 访问令牌 (JWT)
    expires_in: number;       // 令牌过期时间 (秒)
    session_id: string;       // 会话ID
    timestamp: number;
  };
  request_id: string;
}
```

#### 成功响应示例

```json
{
  "success": true,
  "code": 200,
  "message": "Access granted",
  "data": {
    "game_url": "https://games.platform.com/slot-game?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...&data=k4w8n...&timestamp=1699123456789&signature=1a2b3c...",
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJwbGF0Zm9ybSIsInN1YiI6InVzZXJfMTIzNDUiLCJhdWQiOiJzbG90X2dhbWVfdjEiLCJleHAiOjE2OTkxMzc1Njc4OSwiaWF0IjoxNjk5MTIzNDU2Nzg5LCJtZXJjaGFudF9pZCI6Im1lcmNoYW50XzAwMSIsInBlcm1pc3Npb25zIjpbInBsYXkiXX0.signature",
    "expires_in": 1800,
    "session_id": "session_1699123456789",
    "timestamp": 1699123456789
  },
  "request_id": "req_123456"
}
```

#### 错误响应示例

```json
{
  "success": false,
  "code": 401,
  "message": "Invalid signature",
  "data": null,
  "request_id": "req_123456"
}
```

#### 签名计算方法

```javascript
// 1. 构建签名字符串
const signString = `merchant_id=${merchant_id}&user_id=${user_id}&game_id=${game_id}&timestamp=${timestamp}&nonce=${nonce}`;

// 2. 使用HMAC-SHA256计算签名
const signature = crypto.createHmac('sha256', api_secret).update(signString).digest('hex');
```

#### 错误码说明

| 错误码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 401 | 签名验证失败 |
| 403 | 权限不足 |
| 404 | 游戏不存在 |
| 429 | 请求过于频繁 |
| 500 | 服务器内部错误 |

---

### 2. 令牌验证接口

#### 接口信息

| 项目 | 内容 |
|------|------|
| **接口路径** | `/auth/verify-token` |
| **请求方法** | POST |
| **接口描述** | 验证访问令牌的有效性 |
| **权限要求** | 无 |

#### 请求参数

```typescript
interface VerifyTokenRequest {
  token: string;             // 访问令牌
  timestamp: number;         // 时间戳
  signature: string;         // 签名
}
```

#### 响应参数

```typescript
interface VerifyTokenResponse {
  success: boolean;
  code: number;
  message: string;
  data: {
    is_valid: boolean;       // 令牌是否有效
    user_info: {
      user_id: string;
      merchant_id: string;
      permissions: string[];
    };
    expires_at: number;
  };
}
```

---

## 👤 用户信息接口

### 3. 获取用户信息

#### 接口信息

| 项目 | 内容 |
|------|------|
| **接口路径** | `/user/info` |
| **请求方法** | GET |
| **接口描述** | 获取用户基本信息（用户名、头像、余额等） |
| **权限要求** | 需要有效的访问令牌 |

#### 请求头

```http
Authorization: Bearer access_token_here
X-Timestamp: 1699123456789
X-Signature: calculated_signature_here
```

#### 响应参数

```typescript
interface UserInfoResponse {
  success: boolean;
  code: number;
  message: string;
  data: {
    user_id: string;         // 用户ID
    username: string;        // 用户名
    avatar_url: string;      // 头像URL
    balance: number;         // 当前余额
    currency: string;        // 货币类型
    vip_level: number;       // VIP等级
    total_play_time: number; // 总游戏时长 (秒)
    last_play_time: number;  // 最后游戏时间
    preferences: {
      language: string;
      sound_enabled: boolean;
      music_enabled: boolean;
    };
  };
  timestamp: number;
}
```

#### 成功响应示例

```json
{
  "success": true,
  "code": 200,
  "message": "User info retrieved",
  "data": {
    "user_id": "user_platform_12345",
    "username": "Player001",
    "avatar_url": "https://cdn.platform.com/avatars/user_12345.png",
    "balance": 999.50,
    "currency": "CNY",
    "vip_level": 3,
    "total_play_time": 7200,
    "last_play_time": 1699123400000,
    "preferences": {
      "language": "zh-CN",
      "sound_enabled": true,
      "music_enabled": true
    }
  },
  "timestamp": 1699123456789
}
```

---

### 4. 更新用户信息

#### 接口信息

| 项目 | 内容 |
|------|------|
| **接口路径** | `/user/info` |
| **请求方法** | PUT |
| **接口描述** | 更新用户偏好设置 |
| **权限要求** | 需要有效的访问令牌 |

#### 请求参数

```typescript
interface UpdateUserInfoRequest {
  preferences: {
    language?: string;
    sound_enabled?: boolean;
    music_enabled?: boolean;
  };
  timestamp: number;
  signature: string;
}
```

---

## 🎮 游戏操作接口

### 5. 游戏操作

#### 接口信息

| 项目 | 内容 |
|------|------|
| **接口路径** | `/game/action` |
| **请求方法** | POST |
| **接口描述** | 执行游戏操作（下注、旋转等） |
| **权限要求** | 需要有效的访问令牌 |

#### 请求参数

```typescript
interface GameActionRequest {
  action: string;            // 操作类型: spin, bet, win, collect
  game_id: string;           // 游戏ID
  game_data: {
    bet_amount?: number;     // 下注金额
    bet_lines?: number;      // 下注线数
    session_id?: string;     // 游戏会话ID
    // 其他游戏特定参数
  };
  timestamp: number;
  signature: string;
}
```

#### 请求示例 (Slot游戏旋转)

```json
{
  "action": "spin",
  "game_id": "slot_game_v1",
  "game_data": {
    "bet_amount": 1.0,
    "bet_lines": 5,
    "session_id": "session_1699123456789"
  },
  "timestamp": 1699123460789,
  "signature": "4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1"
}
```

#### 响应参数

```typescript
interface GameActionResponse {
  success: boolean;
  code: number;
  message: string;
  data: {
    action: string;
    result: {
      game_result: any;       // 游戏结果 (游戏特定)
      balance_change: number; // 余额变化
      balance: number;        // 更新后的余额
      win_amount: number;     // 赢取金额
      timestamp: number;
    };
  };
  timestamp: number;
}
```

#### 成功响应示例 (Slot游戏旋转结果)

```json
{
  "success": true,
  "code": 200,
  "message": "Spin completed",
  "data": {
    "action": "spin",
    "result": {
      "game_result": {
        "reel_result": [
          ["cherry", "lemon", "orange"],
          ["plum", "grape", "watermelon"],
          ["bell", "seven", "wild"]
        ],
        "win_lines": [
          {
            "line_id": 1,
            "symbol_id": "cherry",
            "match_count": 3,
            "win_amount": 10.0,
            "multiplier": 10.0,
            "positions": [0, 3, 6]
          }
        ]
      },
      "balance_change": 9.0,
      "balance": 1008.5,
      "win_amount": 10.0,
      "timestamp": 1699123461000
    }
  },
  "timestamp": 1699123461000
}
```

---

### 6. 游戏状态同步

#### 接口信息

| 项目 | 内容 |
|------|------|
| **接口路径** | `/game/sync` |
| **请求方法** | POST |
| **接口描述** | 同步游戏状态到服务器 |
| **权限要求** | 需要有效的访问令牌 |

#### 请求参数

```typescript
interface GameSyncRequest {
  game_id: string;
  session_id: string;
  game_state: {
    current_bet: number;
    auto_spin: boolean;
    sound_enabled: boolean;
    // 其他游戏状态
  };
  timestamp: number;
  signature: string;
}
```

---

## 💰 财务操作接口

### 7. 余额查询

#### 接口信息

| 项目 | 内容 |
|------|------|
| **接口路径** | `/finance/balance` |
| **请求方法** | GET |
| **接口描述** | 查询用户当前余额 |
| **权限要求** | 需要有效的访问令牌 |

#### 响应参数

```typescript
interface BalanceResponse {
  success: boolean;
  code: number;
  message: string;
  data: {
    balance: number;
    currency: string;
    last_update: number;
    pending_bets: number;
    available_balance: number;
  };
}
```

#### 成功响应示例

```json
{
  "success": true,
  "code": 200,
  "message": "Balance retrieved",
  "data": {
    "balance": 1008.50,
    "currency": "CNY",
    "last_update": 1699123461000,
    "pending_bets": 0.0,
    "available_balance": 1008.50
  },
  "timestamp": 1699123470789
}
```

---

### 8. 交易记录查询

#### 接口信息

| 项目 | 内容 |
|------|------|
| **接口路径** | `/finance/transactions` |
| **请求方法** | GET |
| **接口描述** | 查询用户交易记录 |
| **权限要求** | 需要有效的访问令牌 |

#### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| limit | number | 否 | 返回记录数量 (默认: 10, 最大: 100) |
| offset | number | 否 | 偏移量 (默认: 0) |
| type | string | 否 | 交易类型 (bet, win, deposit, withdraw) |
| start_time | number | 否 | 开始时间戳 |
| end_time | number | 否 | 结束时间戳 |

#### 响应参数

```typescript
interface TransactionsResponse {
  success: boolean;
  code: number;
  message: string;
  data: {
    transactions: Array<{
      transaction_id: string;
      type: string;
      amount: number;
      balance_before: number;
      balance_after: number;
      description: string;
      timestamp: number;
    }>;
    total: number;
    page: number;
    page_size: number;
  };
}
```

---

## 🔔 通知接口

### 9. 游戏回调通知

#### 接口信息

| 项目 | 内容 |
|------|------|
| **接口路径** | `/callback/game-event` |
| **请求方法** | POST |
| **接口描述** | 游戏事件回调通知 |
| **权限要求** | 需要有效的API Key |

#### 回调事件类型

| 事件类型 | 说明 |
|----------|------|
| `game_start` | 游戏开始 |
| `game_end` | 游戏结束 |
| `big_win` | 大额中奖 |
| `jackpot` | 中大奖 |
| `error` | 游戏错误 |

#### 回调参数

```typescript
interface GameEventCallback {
  event_type: string;        // 事件类型
  merchant_id: string;       // 商户ID
  user_id: string;           // 用户ID
  game_id: string;           // 游戏ID
  event_data: {
    session_id: string;
    // 事件特定数据
  };
  timestamp: number;
  signature: string;
}
```

#### 回调示例 (大额中奖)

```json
{
  "event_type": "big_win",
  "merchant_id": "merchant_001",
  "user_id": "user_12345",
  "game_id": "slot_game_v1",
  "event_data": {
    "session_id": "session_1699123456789",
    "win_amount": 500.0,
    "bet_amount": 1.0,
    "multiplier": 500.0
  },
  "timestamp": 1699123470000,
  "signature": "5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2"
}
```

---

## 🛡️ 安全相关接口

### 10. API密钥刷新

#### 接口信息

| 项目 | 内容 |
|------|------|
| **接口路径** | `/auth/refresh-key` |
| **请求方法** | POST |
| **接口描述** | 刷新API密钥 |
| **权限要求** | 需要有效的管理员权限 |

#### 请求参数

```typescript
interface RefreshKeyRequest {
  merchant_id: string;
  key_type: 'api_key' | 'api_secret' | 'both';
  reason: string;             // 刷新原因
  timestamp: number;
  signature: string;
}
```

#### 响应参数

```typescript
interface RefreshKeyResponse {
  success: boolean;
  data: {
    api_key?: string;
    api_secret?: string;
    expires_at: number;
  };
}
```

---

## 📊 统计接口

### 11. 游戏统计

#### 接口信息

| 项目 | 内容 |
|------|------|
| **接口路径** | `/stats/game` |
| **请求方法** | GET |
| **接口描述** | 获取游戏统计数据 |
| **权限要求** | 需要有效的API Key |

#### 响应参数

```typescript
interface GameStatsResponse {
  success: boolean;
  data: {
    total_players: number;
    active_sessions: number;
    total_bet_amount: number;
    total_win_amount: number;
    rtp: number;              // 实际RTP
    game_distribution: {
      [game_id: string]: {
        player_count: number;
        total_spins: number;
        total_bet: number;
        total_win: number;
      };
    };
    time_range: {
      start_time: number;
      end_time: number;
    };
  };
}
```

---

## 📝 错误处理

### 标准错误响应

```json
{
  "success": false,
  "code": 400,
  "message": "Bad Request",
  "errors": [
    {
      "field": "merchant_id",
      "message": "Invalid merchant ID format"
    }
  ],
  "request_id": "req_123456",
  "timestamp": 1699123456789
}
```

### 常见错误码

| 错误码 | HTTP状态码 | 说明 |
|--------|------------|------|
| 400 | 400 | 请求参数错误 |
| 401 | 401 | 认证失败 |
| 403 | 403 | 权限不足 |
| 404 | 404 | 资源不存在 |
| 429 | 429 | 请求过于频繁 |
| 500 | 500 | 服务器内部错误 |
| 503 | 503 | 服务暂时不可用 |

---

## 🔧 开发测试

### 测试环境

| 环境 | 基础URL | 说明 |
|------|----------|------|
| **开发环境** | `https://dev-api.platform.com/v1` | 开发和测试 |
| **测试环境** | `https://test-api.platform.com/v1` | 集成测试 |
| **生产环境** | `https://api.platform.com/v1` | 正式环境 |

### 测试账号

```json
{
  "merchant_id": "test_merchant_001",
  "api_key": "test_api_key_12345",
  "api_secret": "test_api_secret_67890"
}
```

### Postman集合

提供完整的Postman API集合，包含所有接口的示例请求和响应。

---

## 📞 技术支持

- **API文档**: https://docs.platform.com/api
- **SDK下载**: https://sdk.platform.com
- **技术支持**: api-support@platform.com
- **问题反馈**: https://github.com/platform/api-issues

---

**API版本**: v1.0  
**最后更新**: 2024-01-15  
**维护团队**: 平台技术团队