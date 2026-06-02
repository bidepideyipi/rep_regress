方案A：单池子（简化）

  Jackpot 池（唯一）
  ├─ 种子：1000元
  ├─ 累积：每笔下注的 1%
  ├─ 触发：5个任意相同符号
  └─ 中奖：全部拿走，重置为种子

  方案B：双池子（平衡）

  Mini Jackpot（小奖）
  ├─ 触发概率高
  └─ 5个任意低价值符号

  Grand Jackpot（大奖）
  ├─ 触发概率低
  └─ 5个 Wild/Seven

  方案C：四池子（完整）

  保留原有的4层设计

  ---
  我的建议

  用 2 个池子就够了：

  ┌───────┬────────────────┬────────────────────┐
  │ 池子  │    触发条件    │        奖励        │
  ├───────┼────────────────┼────────────────────┤
  │ Mini  │ 5个普通符号    │ 频繁触发，保持活跃 │
  ├───────┼────────────────┼────────────────────┤
  │ Grand │ 5个 Wild/Seven │ 稀有触发，制造惊喜 │
  └───────┴────────────────┴────────────────────┘

  理由：
  - Mini 满足日常参与感
  - Grand 制造爆款话题
  - 减少复杂度，便于理解和运营

  你觉得 单池、双池、四池 哪个更适合？或者有其他想法？

  Jackpot 层级结构

  ┌───────┬──────────┬──────────┬──────────┐
  │ 层级  │ 初始金额 │ 触发概率 │ 分配比例 │
  ├───────┼──────────┼──────────┼──────────┤
  │ Mini  │ 100      │ 高       │ 30%      │
  ├───────┼──────────┼──────────┼──────────┤
  │ Minor │ 500      │ 中高     │ 25%      │
  ├───────┼──────────┼──────────┼──────────┤
  │ Major │ 2,000    │ 中       │ 25%      │
  ├───────┼──────────┼──────────┼──────────┤
  │ Grand │ 10,000   │ 低       │ 20%      │
  └───────┴──────────┴──────────┴──────────┘

  2. 触发机制

  方案A：符号触发（推荐）
  - 5 个 Wild 符号连线 → 进入 Jackpot 轮盘
  - 轮盘随机决定获得哪个层级的 Jackpot

  方案B：随机触发
  - 每次 Spin 有极小概率（如 1/10000）直接触发
  - 触发后进入 Jackpot 游戏

  3. 累积机制

  每笔下注的 1% 进入 Jackpot 池
  ├── 30% → Mini 池
  ├── 25% → Minor 池
  ├── 25% → Major 池
  └── 20% → Grand 池

  4. 数据库设计

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
  CREATE TABLE jackpot_pool (
      id BIGINT PRIMARY KEY AUTO_INCREMENT,
      game_id VARCHAR(32),
      pool_type VARCHAR(16), -- mini/minor/major/grand
      current_amount DECIMAL(18,2),
      last_win_time DATETIME,
      win_count BIGINT DEFAULT 0
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

  5. Redis 结构

  # 实时奖池金额
  jackpot:{game_id}:{pool_type}:amount → DECIMAL

  # 奖池状态信息
  jackpot:{game_id}:info → HASH {
      mini_amount: "1234.56",
      minor_amount: "5678.90",
      major_amount: "12345.67",
      grand_amount: "67890.12",
      last_update: "2026-06-01T21:00:00Z"
  }