# 游戏服务接口文档

## 文档说明
本文档描述游戏服务模块的所有API接口，包括接口路径、请求参数、响应参数、错误码等信息。

---

## 1. 游戏旋转接口

### 1.1 接口基本信息

| 项目 | 内容 |
|------|------|
| 接口名称 | 游戏旋转接口 |
| 接口路径 | `/api/v1/game/spin` |
| 请求方法 | `POST` |
| 接口描述 | 处理游戏旋转请求，生成游戏结果并计算赔付金额 |
| 认证方式 | Bearer Token |
| 权限要求 | 需要有效的access_token |

### 1.2 请求参数

#### 1.2.1 请求头

| 参数名 | 类型 | 必填 | 说明 | 示例值 |
|--------|------|------|------|--------|
| Content-Type | String | 是 | 请求内容类型 | application/json |
| Authorization | String | 是 | 认证令牌 | Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9... |

#### 1.2.2 请求体

| 参数名 | 类型 | 必填 | 说明 | 示例值 | 约束条件 |
|--------|------|------|------|--------|----------|
| user_id | String | 是 | 用户ID | user_001 | 长度1-32字符 |
| game_id | String | 是 | 游戏ID | game_001 | 长度1-32字符 |
| bet_amount | Decimal | 是 | 下注金额 | 100.00 | 最小值0.01 |
| bet_lines | Integer | 是 | 下注线数 | 10 | 范围1-20 |

#### 1.2.3 请求示例

```json
{
  "user_id": "user_001",
  "game_id": "game_001",
  "bet_amount": 100.00,
  "bet_lines": 10
}
```

### 1.3 响应参数

#### 1.3.1 成功响应

| 参数名 | 类型 | 说明 | 示例值 |
|--------|------|------|--------|
| code | Integer | 响应状态码 | 200 |
| message | String | 响应消息 | success |
| data | Object | 响应数据 | - |
| data.game_session_id | String | 游戏会话ID | gs_001 |
| data.reels | Array | 卷轴结果 | - |
| data.total_win | Decimal | 总赔付金额 | 500.00 |
| data.net_result | Decimal | 净结果（赔付-下注） | 400.00 |
| data.win_lines | Array | 中奖线路 | - |

#### 1.3.2 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "game_session_id": "gs_001",
    "reels": [
      ["symbol_1", "symbol_2", "symbol_3"],
      ["symbol_4", "symbol_5", "symbol_6"],
      ["symbol_7", "symbol_8", "symbol_9"]
    ],
    "total_win": 500.00,
    "net_result": 400.00,
    "win_lines": [
      {
        "line_id": 1,
        "win_amount": 200.00,
        "symbols": ["symbol_1", "symbol_5", "symbol_9"],
        "multiplier": 1
      }
    ]
  }
}
```

#### 1.3.3 错误响应

| 错误码 | 说明 | HTTP状态码 | 处理建议 |
|--------|------|------------|----------|
| 4001 | 参数校验失败 | 400 | 检查请求参数格式和取值范围 |
| 4002 | 用户ID不存在 | 400 | 确认用户ID是否正确 |
| 4003 | 游戏ID不存在 | 400 | 确认游戏ID是否正确 |
| 4004 | 游戏已停用 | 400 | 联系管理员或选择其他游戏 |
| 4005 | 余额不足 | 400 | 用户余额不足，无法下注 |
| 4006 | 下注金额超出范围 | 400 | 调整下注金额到合理范围 |
| 5001 | 游戏配置加载失败 | 500 | 联系技术支持 |
| 5002 | 游戏计算异常 | 500 | 联系技术支持 |
| 5003 | 数据库操作失败 | 500 | 联系技术支持 |

---

## 2. 玩家信息查询接口

### 2.1 接口基本信息

| 项目 | 内容 |
|------|------|
| 接口名称 | 玩家信息查询接口 |
| 接口路径 | `/api/v1/player/{user_id}` |
| 请求方法 | `GET` |
| 接口描述 | 查询玩家基本信息和余额 |
| 认证方式 | Bearer Token |
| 权限要求 | 需要有效的access_token |

### 2.2 请求参数

#### 2.2.1 请求头

| 参数名 | 类型 | 必填 | 说明 | 示例值 |
|--------|------|------|------|--------|
| Authorization | String | 是 | 认证令牌 | Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9... |

#### 2.2.2 路径参数

| 参数名 | 类型 | 必填 | 说明 | 示例值 | 约束条件 |
|--------|------|------|------|--------|----------|
| user_id | String | 是 | 用户ID | user_001 | 长度1-32字符 |

### 2.3 响应参数

#### 2.3.1 成功响应

| 参数名 | 类型 | 说明 | 示例值 |
|--------|------|------|--------|
| code | Integer | 响应状态码 | 200 |
| message | String | 响应消息 | success |
| data | Object | 响应数据 | - |
| data.user_id | String | 用户ID | user_001 |
| data.user_name | String | 用户名称 | 张三 |
| data.balance | Decimal | 账户余额 | 1000.00 |
| data.total_bet | Decimal | 总下注金额 | 50000.00 |
| data.total_win | Decimal | 总赔付金额 | 48500.00 |
| data.total_games | Integer | 总游戏次数 | 500 |
| data.last_play_time | DateTime | 最后游戏时间 | 2025-01-01T12:00:00Z |

#### 2.3.2 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "user_id": "user_001",
    "user_name": "张三",
    "balance": 1000.00,
    "total_bet": 50000.00,
    "total_win": 48500.00,
    "total_games": 500,
    "last_play_time": "2025-01-01T12:00:00Z"
  }
}
```

#### 2.3.3 错误响应

| 错误码 | 说明 | HTTP状态码 | 处理建议 |
|--------|------|------------|----------|
| 4011 | 用户不存在 | 404 | 确认用户ID是否正确 |
| 4012 | 用户已禁用 | 403 | 联系管理员 |

---

## 3. 游戏配置获取接口

### 3.1 接口基本信息

| 项目 | 内容 |
|------|------|
| 接口名称 | 游戏配置获取接口 |
| 接口路径 | `/api/v1/game/{game_id}/config` |
| 请求方法 | `GET` |
| 接口描述 | 获取指定游戏的配置信息 |
| 认证方式 | Bearer Token |
| 权限要求 | 需要有效的access_token |

### 3.2 请求参数

#### 3.2.1 请求头

| 参数名 | 类型 | 必填 | 说明 | 示例值 |
|--------|------|------|------|--------|
| Authorization | String | 是 | 认证令牌 | Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9... |

#### 3.2.2 路径参数

| 参数名 | 类型 | 必填 | 说明 | 示例值 | 约束条件 |
|--------|------|------|------|--------|----------|
| game_id | String | 是 | 游戏ID | game_001 | 长度1-32字符 |

### 3.3 响应参数

#### 3.3.1 成功响应

| 参数名 | 类型 | 说明 | 示例值 |
|--------|------|------|--------|
| code | Integer | 响应状态码 | 200 |
| message | String | 响应消息 | success |
| data | Object | 响应数据 | - |
| data.game_id | String | 游戏ID | game_001 |
| data.game_name | String | 游戏名称 | 幸运水果 |
| data.game_type | String | 游戏类型 | slot_3x3 |
| data.symbols | Array | 符号定义 | - |
| data.reel_weights | Array | 卷轴权重 | - |
| data.pay_table | Object | 赔付表 | - |
| data.min_bet | Decimal | 最小下注金额 | 0.10 |
| data.max_bet | Decimal | 最大下注金额 | 1000.00 |
| data.min_lines | Integer | 最小下注线数 | 1 |
| data.max_lines | Integer | 最大下注线数 | 20 |

#### 3.3.2 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "game_id": "game_001",
    "game_name": "幸运水果",
    "game_type": "slot_3x3",
    "symbols": [
      {"id": "symbol_1", "name": "樱桃", "type": "normal"},
      {"id": "symbol_2", "name": "柠檬", "type": "normal"},
      {"id": "symbol_3", "name": "橙子", "type": "normal"},
      {"id": "scatter", "name": "免费旋转", "type": "scatter"}
    ],
    "reel_weights": [
      [20, 15, 10, 5],
      [20, 15, 10, 5],
      [20, 15, 10, 5]
    ],
    "pay_table": {
      "symbol_1": {
        "3": {"multiplier": 5},
        "2": {"multiplier": 2}
      },
      "symbol_2": {
        "3": {"multiplier": 10},
        "2": {"multiplier": 3}
      }
    },
    "min_bet": 0.10,
    "max_bet": 1000.00,
    "min_lines": 1,
    "max_lines": 20
  }
}
```

#### 3.3.3 错误响应

| 错误码 | 说明 | HTTP状态码 | 处理建议 |
|--------|------|------------|----------|
| 4021 | 游戏不存在 | 404 | 确认游戏ID是否正确 |
| 4022 | 游戏已停用 | 403 | 联系管理员或选择其他游戏 |
| 4023 | 游戏配置加载失败 | 500 | 联系技术支持 |

---

## 4. 玩家余额更新接口

### 4.1 接口基本信息

| 项目 | 内容 |
|------|------|
| 接口名称 | 玩家余额更新接口 |
| 接口路径 | `/api/v1/player/balance` |
| 请求方法 | `POST` |
| 接口描述 | 更新玩家余额（内部接口，由结算服务调用） |
| 认证方式 | 内部服务认证 |
| 权限要求 | 内部服务权限 |

### 4.2 请求参数

#### 4.2.1 请求头

| 参数名 | 类型 | 必填 | 说明 | 示例值 |
|--------|------|------|------|--------|
| Content-Type | String | 是 | 请求内容类型 | application/json |
| X-Service-Token | String | 是 | 内部服务令牌 | internal_service_token |

#### 4.2.2 请求体

| 参数名 | 类型 | 必填 | 说明 | 示例值 | 约束条件 |
|--------|------|------|------|--------|----------|
| user_id | String | 是 | 用户ID | user_001 | 长度1-32字符 |
| amount | Decimal | 是 | 变动金额 | 100.00 | 可为正数或负数 |
| type | String | 是 | 变动类型 | bet/win/refund | 枚举值 |
| game_session_id | String | 否 | 游戏会话ID | gs_001 | 关联游戏会话 |

#### 4.2.3 请求示例

```json
{
  "user_id": "user_001",
  "amount": -100.00,
  "type": "bet",
  "game_session_id": "gs_001"
}
```

### 4.3 响应参数

#### 4.3.1 成功响应

| 参数名 | 类型 | 说明 | 示例值 |
|--------|------|------|--------|
| code | Integer | 响应状态码 | 200 |
| message | String | 响应消息 | success |
| data | Object | 响应数据 | - |
| data.user_id | String | 用户ID | user_001 |
| data.old_balance | Decimal | 原余额 | 1000.00 |
| data.new_balance | Decimal | 新余额 | 900.00 |
| data.transaction_id | String | 交易ID | tx_001 |

#### 4.3.2 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "user_id": "user_001",
    "old_balance": 1000.00,
    "new_balance": 900.00,
    "transaction_id": "tx_001"
  }
}
```

#### 4.3.3 错误响应

| 错误码 | 说明 | HTTP状态码 | 处理建议 |
|--------|------|------------|----------|
| 4031 | 用户不存在 | 404 | 确认用户ID是否正确 |
| 4032 | 余额不足 | 400 | 用户余额不足 |
| 4033 | 余额变动异常 | 500 | 联系技术支持 |

---

## 5. 游戏结果记录接口

### 5.1 接口基本信息

| 项目 | 内容 |
|------|------|
| 接口名称 | 游戏结果记录接口 |
| 接口路径 | `/api/v1/game/result` |
| 请求方法 | `POST` |
| 接口描述 | 记录游戏结果到数据库（内部接口，由游戏服务调用） |
| 认证方式 | 内部服务认证 |
| 权限要求 | 内部服务权限 |

### 5.2 请求参数

#### 5.2.1 请求头

| 参数名 | 类型 | 必填 | 说明 | 示例值 |
|--------|------|------|------|--------|
| Content-Type | String | 是 | 请求内容类型 | application/json |
| X-Service-Token | String | 是 | 内部服务令牌 | internal_service_token |

#### 5.2.2 请求体

| 参数名 | 类型 | 必填 | 说明 | 示例值 | 约束条件 |
|--------|------|------|------|--------|----------|
| game_session_id | String | 是 | 游戏会话ID | gs_001 | 长度1-32字符 |
| integrator_id | String | 是 | 集成商ID | integrator_001 | 长度1-32字符 |
| user_id | String | 是 | 用户ID | user_001 | 长度1-32字符 |
| game_id | String | 是 | 游戏ID | game_001 | 长度1-32字符 |
| bet_amount | Decimal | 是 | 下注金额 | 100.00 | 大于0 |
| win_amount | Decimal | 是 | 赔付金额 | 500.00 | 大于等于0 |
| net_result | Decimal | 是 | 净结果 | 400.00 | 可为正数或负数 |
| game_result | JSON | 是 | 游戏详细结果 | - | JSON格式 |

#### 5.2.3 请求示例

```json
{
  "game_session_id": "gs_001",
  "integrator_id": "integrator_001",
  "user_id": "user_001",
  "game_id": "game_001",
  "bet_amount": 100.00,
  "win_amount": 500.00,
  "net_result": 400.00,
  "game_result": {
    "reels": [
      ["symbol_1", "symbol_2", "symbol_3"],
      ["symbol_4", "symbol_5", "symbol_6"],
      ["symbol_7", "symbol_8", "symbol_9"]
    ],
    "win_lines": [
      {
        "line_id": 1,
        "win_amount": 200.00,
        "symbols": ["symbol_1", "symbol_5", "symbol_9"],
        "multiplier": 1
      }
    ]
  }
}
```

### 5.3 响应参数

#### 5.3.1 成功响应

| 参数名 | 类型 | 说明 | 示例值 |
|--------|------|------|--------|
| code | Integer | 响应状态码 | 200 |
| message | String | 响应消息 | success |
| data | Object | 响应数据 | - |
| data.game_session_id | String | 游戏会话ID | gs_001 |
| data.record_id | Long | 记录ID | 123456789 |

#### 5.3.2 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "game_session_id": "gs_001",
    "record_id": 123456789
  }
}
```

#### 5.3.3 错误响应

| 错误码 | 说明 | HTTP状态码 | 处理建议 |
|--------|------|------------|----------|
| 5041 | 游戏会话ID已存在 | 409 | 检查会话ID是否重复 |
| 5042 | 集成商不存在 | 400 | 确认集成商ID是否正确 |
| 5043 | 数据库写入失败 | 500 | 联系技术支持 |

---

## 6. 接口性能指标

### 6.1 性能要求

| 接口名称 | 响应时间目标 | 并发能力要求 | 可用性要求 |
|----------|--------------|--------------|------------|
| 游戏旋转接口 | < 200ms | 10000 QPS | 99.9% |
| 玩家信息查询接口 | < 100ms | 5000 QPS | 99.9% |
| 游戏配置获取接口 | < 50ms | 2000 QPS | 99.9% |
| 玩家余额更新接口 | < 100ms | 10000 QPS | 99.9% |
| 游戏结果记录接口 | < 150ms | 10000 QPS | 99.9% |

### 6.2 限流配置

| 接口名称 | 限流策略 | 限流阈值 | 时间窗口 |
|----------|----------|----------|----------|
| 游戏旋转接口 | 用户级限流 | 10次/秒 | 1秒 |
| 玩家信息查询接口 | 用户级限流 | 20次/秒 | 1秒 |
| 游戏配置获取接口 | 缓存优先 | 50次/秒 | 1秒 |
| 玩家余额更新接口 | 内部接口 | 无限制 | - |
| 游戏结果记录接口 | 内部接口 | 无限制 | - |

---

## 7. 接口安全说明

### 7.1 认证方式

| 接口类型 | 认证方式 | 令牌有效期 | 刷新机制 |
|----------|----------|------------|----------|
| 外部接口 | Bearer Token | 1小时 | 通过集成商重新获取 |
| 内部接口 | Service Token | 永久 | 服务启动时生成 |

### 7.2 数据加密

| 数据类型 | 加密方式 | 加密强度 | 备注 |
|----------|----------|----------|------|
| 传输数据 | HTTPS/TLS 1.3 | AES-256 | 全链路加密 |
| 敏感字段 | AES-256 | 256位 | 用户余额等 |
| 游戏地址 | 自定义加密 | 256位 | 防止直接访问 |

### 7.3 审计日志

所有接口调用都会记录审计日志，包括：
- 请求时间戳
- 请求方信息（集成商ID、用户ID）
- 请求参数（敏感信息脱敏）
- 响应状态码
- 处理时长

---

## 8. 版本记录

| 版本号 | 日期 | 修改内容 | 修改人 |
|--------|------|----------|--------|
| v1.0.0 | 2025-01-01 | 初始版本发布 | - |