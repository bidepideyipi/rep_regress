# Slot Game API 文档

## 基本信息

- **Base URL**: `http://localhost:8081`
- **Content-Type**: `application/json`
- **Character Set**: UTF-8

## 接口列表

### 1. 健康检查接口

#### 1.1 健康检查

- **接口**: `GET /health`
- **描述**: 检查服务健康状态
- **认证**: 无需认证

**响应示例**:

```json
{
  "status": "healthy",
  "time": "2024-01-01T10:00:00Z",
  "service": "slot-game",
  "game_initialized": true,
  "game_id": "game_001"
}
```

#### 1.2 Ping接口

- **接口**: `GET /health/ping`
- **描述**: 简单的ping检查
- **认证**: 无需认证

**响应示例**:

```json
{
  "status": "pong"
}
```

### 2. 游戏接口

#### 2.1 执行旋转

- **接口**: `POST /api/game/spin`
- **描述**: 执行游戏旋转操作
- **认证**: 需要认证（建议添加JWT）

**请求参数**:

```json
{
  "user_id": "string",       // 必填，用户ID
  "bet_amount": "number",    // 必填，下注金额，最小0.1
  "bet_lines": "integer",    // 必填，下注线数，1-20
  "session_id": "string"     // 必填，会话ID
}
```

**响应示例**:

```json
{
  "success": true,
  "data": {
    "session_id": "session_123",
    "user_id": "user_001",
    "game_id": "game_001",
    "bet_amount": 1.0,
    "bet_lines": 5,
    "bet_per_line": 0.2,
    "win_amount": 10.0,
    "net_result": 9.0,
    "is_free_spin": false,
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
        "positions": [0, 3, 6],
        "is_wild": false
      }
    ],
    "bonus_feature": "free_spins_10",
    "processing_time_ms": 5,
    "timestamp": "2024-01-01T10:00:00Z"
  }
}
```

**错误响应**:

```json
{
  "success": false,
  "message": "下注金额低于最小值 0.10"
}
```

#### 2.2 获取游戏配置

- **接口**: `GET /api/game/config`
- **描述**: 获取当前游戏配置信息
- **认证**: 需要认证（建议添加JWT）

**响应示例**:

```json
{
  "success": true,
  "data": {
    "game_id": "game_001",
    "game_name": "Classic Slot",
    "version": "1.0.0",
    "description": "经典老虎机游戏配置",
    "last_updated": "2024-01-01T00:00:00Z",
    "symbols_count": 10,
    "reels_count": 3,
    "pay_lines": 20,
    "min_bet": 0.1,
    "max_bet": 1000.0,
    "rtp": 96.5
  }
}
```

## 错误代码

| 状态码 | 说明      |
| --- | ------- |
| 200 | 成功      |
| 400 | 请求参数错误  |
| 401 | 未授权     |
| 403 | 禁止访问    |
| 404 | 资源不存在   |
| 500 | 服务器内部错误 |

## 游戏逻辑说明

### 符号类型

- **normal**: 普通符号，需要按线路匹配
- **wild**: 万能符号，可替代普通符号
- **scatter**: 散布符号，任意位置触发特殊功能

### 旋转算法

1. **权重计算**: 根据每个卷轴的符号权重计算出现概率
2. **符号生成**: 使用随机算法生成每个位置的符号
3. **线路匹配**: 按照20条预设线路检查匹配情况
4. **赔付计算**: 根据符号倍数和下注金额计算奖金

### Wild符号规则

- Wild可以替代所有normal符号
- Wild不能替代Scatter符号
- Wild参与匹配时按普通符号计算
- Wild本身的赔付倍数独立计算

### Scatter符号规则

- Scatter不需要按线路匹配
- 任意位置出现3个或以上触发特殊功能
- Scatter通常触发免费旋转或额外游戏

### RTP计算

```
RTP = (总奖金 / 总下注) * 100%
```

## 性能指标

| 指标     | 目标值           |
| ------ | ------------- |
| 响应时间   | < 50ms        |
| 并发处理   | > 10000 req/s |
| CPU使用率 | < 80%         |
| 内存使用   | < 500MB       |

## 使用示例

### cURL示例

```bash
# 健康检查
curl http://localhost:8081/health

# 执行旋转
curl -X POST http://localhost:8081/api/game/spin \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user_001",
    "bet_amount": 1.0,
    "bet_lines": 5,
    "session_id": "session_123"
  }'

# 获取配置
curl http://localhost:8081/api/game/config
```

## 注意事项

1. **并发控制**: 建议添加用户级别的并发限制
2. **数据持久化**: 游戏结果需要异步写入数据库
3. **监控告警**: 集成Prometheus和Grafana监控
4. **负载均衡**: 生产环境建议使用负载均衡器
5. **缓存策略**: 合理使用缓存提高性能
6. **日志记录**: 详细记录游戏日志用于审计