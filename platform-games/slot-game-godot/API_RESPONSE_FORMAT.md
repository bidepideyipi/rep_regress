# Slot Game Spin API 响应格式

## 顶层结构

```json
{
  "success": boolean,   // 请求是否成功
  "message": string,    // 错误消息（成功时为空）
  "data": {...}         // 实际数据
}
```

## data 字段说明

### 基本信息字段

| 字名 | 类型 | 示例值 | 说明 |
|------|------|--------|------|
| `user_id` | string | "user_godot_001" | 用户ID |
| `session_id` | string | "session_1780062048_6601" | 会话ID |
| `game_id` | string | "game_001" | 游戏ID |
| `timestamp` | string | "2026-05-29T21:40:49.433691+08:00" | 服务器时间戳 |

### 下注相关字段

| 字名 | 类型 | 示例值 | 说明 |
|------|------|--------|------|
| `bet_amount` | float | 1.0 | 总下注额 |
| `bet_lines` | float | 5.0 | 赔线数量 |
| `bet_per_line` | float | 0.2 | 每条线的下注 (bet_amount ÷ bet_lines) |
| `is_free_spin` | boolean | false | 是否为免费旋转 |

### 结果相关字段

| 字名 | 类型 | 示例值 | 说明 |
|------|------|--------|------|
| `win_amount` | float | 3.0 | 赢取金额 |
| `net_result` | float | 2.0 | 净收益 (win_amount - bet_amount) |
| `rtp_rate` | float | 300.0 | RTP比率 (本次回报率) |
| `processing_time_ms` | float | 0.0 | 服务器处理时间（毫秒） |
| `bonus_feature` | string | "" | 特殊奖励功能名称 |

## reel_result - 卷轴结果

3个卷轴，每个卷轴3行符号：

```
Reel 0: ["lemon",  "orange",  "seven"]
Reel 1: ["grape",  "lemon",   "watermelon"]
Reel 2: ["lemon",  "cherry",  "lemon"]
```

**位置索引**（0-8）：
```
| 0 | 3 | 6 |
| 1 | 4 | 7 |
| 2 | 5 | 8 |
```

## win_lines - 中奖线路数组

```json
[
  {
    "line_id": float,        // 赔线ID
    "symbol_id": string,     // 匹配的符号
    "match_count": float,    // 匹配数量
    "multiplier": float,     // 倍数
    "win_amount": float,    // 该线路赢取金额
    "is_wild": boolean,      // 是否包含万能符号
    "positions": [float]    // 符号位置索引数组
  }
]
```

### win_lines 示例分析

```json
{
  "line_id": 4.0,
  "symbol_id": "lemon",
  "match_count": 3.0,
  "multiplier": 15.0,
  "win_amount": 3.0,
  "is_wild": false,
  "positions": [0.0, 4.0, 8.0]
}
```

**解读**：
- 位置 [0, 4, 8] 是一条对角线
- 匹配了 3 个 lemon
- 倍数 15 × bet_per_line(0.2) = 3.0

## 完整示例

```json
{
  "success": true,
  "message": "",
  "data": {
    "bet_amount": 1.0,
    "bet_lines": 5.0,
    "bet_per_line": 0.2,
    "bonus_feature": "",
    "game_id": "game_001",
    "is_free_spin": false,
    "net_result": 2.0,
    "processing_time_ms": 0.0,
    "reel_result": [
      ["lemon", "orange", "seven"],
      ["grape", "lemon", "watermelon"],
      ["lemon", "cherry", "lemon"]
    ],
    "rtp_rate": 300.0,
    "session_id": "session_1780062048_6601",
    "timestamp": "2026-05-29T21:40:49.433691+08:00",
    "user_id": "user_godot_001",
    "win_amount": 3.0,
    "win_lines": [
      {
        "is_wild": false,
        "line_id": 3.0,
        "match_count": 2.0,
        "multiplier": 5.0,
        "positions": [6.0, 7.0, 8.0],
        "symbol_id": "lemon",
        "win_amount": 0.0
      },
      {
        "is_wild": false,
        "line_id": 4.0,
        "match_count": 3.0,
        "multiplier": 15.0,
        "positions": [0.0, 4.0, 8.0],
        "symbol_id": "lemon",
        "win_amount": 3.0
      },
      {
        "is_wild": false,
        "line_id": 5.0,
        "match_count": 2.0,
        "multiplier": 5.0,
        "positions": [6.0, 4.0, 2.0],
        "symbol_id": "lemon",
        "win_amount": 0.0
      }
    ]
  }
}
```

## GDScript 解析代码

代码实现见 `scripts/spin_response.gd`，主要方法：

```gdscript
var response = SpinResponse.new()
response._from_json(json_dict)

# 获取数据
var win_amount = response.get_win_amount()
var reel_result = response.get_reel_result()
var win_lines = response.get_win_lines()
```
