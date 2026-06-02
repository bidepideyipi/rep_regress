# Jackpot 设计方案（混合池模式）

## 1. 池子结构

```
┌─────────────────────────────────────────────────────────────┐
│                  共享 Grand Pool（全局大奖）                 │
│  game_id = "0"，所有游戏 0.2% 注入 → 高额大奖，制造爆款     │
└─────────────────────────────────────────────────────────────┘

┌──────────────┬──────────────┬──────────────┬──────────────┐
│ Game A       │ Game B       │ Game C       │    ...       │
│ Mini/Minor/ │ Mini/Minor/ │ Mini/Minor/ │              │
│ Major       │ Major       │ Major       │              │
│ 各游戏0.8%  │ 独立运营    │             │              │
└──────────────┴──────────────┴──────────────┴──────────────┘
```

## 2. 层级配置

┌───────┬──────────┬──────────┬──────────┬──────────┐
│ 层级  │ game_id  │ 初始金额 │ 触发概率 │ 分配比例 │
├───────┼──────────┼──────────┼──────────┼──────────┤
│ Mini  │ 具体游戏 │ 100      │ 高       │ 30%      │
├───────┼──────────┼──────────┼──────────┼──────────┤
│ Minor │ 具体游戏 │ 500      │ 中高     │ 25%      │
├───────┼──────────┼──────────┼──────────┼──────────┤
│ Major │ 具体游戏 │ 2,000    │ 中       │ 25%      │
├───────┼──────────┼──────────┼──────────┼──────────┤
│ Grand │ "0"      │ 10,000   │ 低       │ 20%      │
└───────┴──────────┴──────────┴──────────┴──────────┘

## 3. 触发机制

**A：符号触发**
- 同一符号在任意位置出现五次 → 触发一次幸运转盘 Mini 或Grand

**B：随机触发**
- 每次 Spin 有极小概率（如 1/10000）触发一次幸运转盘 Mini 或Grand

## 4. 累积机制

```
每笔下注 100 元：
├── 0.3 元 → game_id=当前游戏, pool_type="mini"  （30%）
├── 0.25 元 → game_id=当前游戏, pool_type="minor" （25%）
├── 0.25 元 → game_id=当前游戏, pool_type="major" （25%）
└── 0.2 元 → game_id="0", pool_type="grand"       （20% 共享）
```

## 5. 数据库设计

```sql
-- Jackpot 配置表
CREATE TABLE jackpot_config (
    game_id VARCHAR(32) PRIMARY KEY,
    enabled TINYINT(1) DEFAULT 1,
    contribution_rate DECIMAL(5,4) DEFAULT 0.0100, -- 1%
    mini_seed DECIMAL(18,2) DEFAULT 100,
    minor_seed DECIMAL(18,2) DEFAULT 500,
    major_seed DECIMAL(18,2) DEFAULT 2000,
    grand_seed DECIMAL(18,2) DEFAULT 10000
);

-- Jackpot 实时池（Redis 为主，DB 备份）
-- game_id = "0" 表示全局共享池
CREATE TABLE jackpot_pool (
    game_id VARCHAR(32),
    pool_type VARCHAR(16), -- mini/minor/major/grand
    current_amount DECIMAL(18,2),
    last_win_time DATETIME,
    win_count BIGINT DEFAULT 0,
    PRIMARY KEY (game_id, pool_type)
);

-- Jackpot 中奖记录
CREATE TABLE jackpot_win_record (
    transaction_id VARCHAR(64) PRIMARY KEY,
    integrator_id VARCHAR(32),
    user_id VARCHAR(32),
    game_id VARCHAR(32),
    pool_type VARCHAR(16),
    win_amount DECIMAL(18,2),
    session_id VARCHAR(64),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

## 6. Redis 结构

```
# 实时奖池金额
jackpot:{game_id}:{pool_type}:amount → DECIMAL

# 示例：
# jackpot:0:grand:amount     # 共享 Grand 池
# jackpot:game_a:mini:amount  # 游戏 A 的 Mini 池
# jackpot:game_a:minor:amount # 游戏 A 的 Minor 池
# jackpot:game_a:major:amount # 游戏 A 的 Major 池
```

## 7. 查询逻辑

```python
def get_jackpots(game_id):
    return {
        'mini': redis.get(f'jackpot:{game_id}:mini:amount'),
        'minor': redis.get(f'jackpot:{game_id}:minor:amount'),
        'major': redis.get(f'jackpot:{game_id}:major:amount'),
        'grand': redis.get(f'jackpot:0:grand:amount')
    }
```

## 8. 设计优势

| 优势 | 说明 |
|------|------|
| **大奖吸引力** | Grand 跨游戏累积，容易出 10 万+ 爆款 |
| **多层次参与** | Mini/Minor/Major 满足不同玩家期待 |
| **风险可控** | 单个游戏出问题，只影响其独立池 |
| **结构简洁** | game_id="0" 统一表示共享池，无需额外表 |
