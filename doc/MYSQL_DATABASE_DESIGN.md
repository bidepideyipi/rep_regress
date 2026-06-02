# MySQL数据库设计文档

## 文档说明
本文档描述游戏服务模块的MySQL数据库设计，包括表结构、字段说明、索引设计、约束条件等。

**重要说明**：
- 本数据库主要用于存储核心业务数据和实时查询
- 详细游戏日志数据已迁移至ClickHouse存储
- MySQL和ClickHouse通过分布式事务保证数据一致性
- 详细的数据分工请参考《ClickHouse数据库设计文档》第8节

---

## 1. 数据库基础信息

| 项目 | 内容 |
|------|------|
| 数据库名称 | slot_game_service |
| 数据库版本 | MySQL 8.0+ |
| 字符集 | utf8mb4 |
| 排序规则 | utf8mb4_unicode_ci |
| 存储引擎 | InnoDB |
| 时区 | UTC |

---

## 2. 表结构设计

### 2.0 表结构总览

| 表编号 | 表名 | 中文名 | 用途说明 |
|--------|------|------|----------|
| 2.1 | integrator_config | 集成商配置表 | 存储集成商的配置信息和权限设置 |
| 2.2 | game_config | 游戏配置表 | 存储游戏基础配置信息（详细配置在Nacos） |
| 2.3 | user_info | 用户信息表 | 存储用户基本信息、登录信息和账户余额（独立行级锁） |
| 2.4 | user_finance_stats | 用户财务统计表 | 存储用户游戏统计信息（不含余额，减少锁竞争） |
| 2.5 | transaction_records | 交易记录表 | 存储用户资金变动记录 |
| 2.6 | user_session | 用户会话表 | 存储用户游戏会话信息 |
| 2.7 | jackpot_config | Jackpot配置表 | 存储Jackpot基础配置信息 |
| 2.8 | jackpot_pool | Jackpot实时池表 | 存储Jackpot奖池实时金额（game_id="0"表示全局共享池） |

**表数量统计**：核心业务表8个，其中用户相关表2个（user_info含余额，user_finance_stats含统计），Jackpot相关表2个<br>
**说明**：
- 游戏详细配置（符号、卷轴、权重、赔付表等）已在Nacos配置中心管理
- Jackpot中奖记录写入ClickHouse，通过聚合查询统计

---

### 2.1 集成商配置表 (integrator_config)

#### 2.1.1 表基本信息

| 项目 | 内容 |
|------|------|
| 表名 | integrator_config |
| 中文名 | 集成商配置表 |
| 用途 | 存储集成商的配置信息，包括API密钥、权限设置等 |
| 存储引擎 | InnoDB |
| 字符集 | utf8mb4 |

#### 2.1.2 字段设计

| 字段名 | 类型 | 长度 | 允许NULL | 默认值 | 主键 | 说明 |
|--------|------|------|----------|--------|------|------|
| integrator_id | VARCHAR | 32 | NO | - | YES | 集成商ID |
| integrator_name | VARCHAR | 64 | NO | - | NO | 集成商名称 |
| company_name | VARCHAR | 128 | YES | NULL | NO | 公司名称 |
| contact_person | VARCHAR | 32 | YES | NULL | NO | 联系人 |
| contact_email | VARCHAR | 64 | YES | NULL | NO | 联系邮箱 |
| contact_phone | VARCHAR | 20 | YES | NULL | NO | 联系电话 |
| is_platform_self | TINYINT | 1 | NO | 0 | NO | 是否平台自营（0-否，1-是） |
| api_key | VARCHAR | 64 | NO | - | NO | API密钥 |
| api_secret | VARCHAR | 128 | NO | - | NO | API密钥（加密存储） |
| game_access_permitted | JSON | - | YES | NULL | NO | 游戏访问权限列表 |
| financial_permission | TINYINT | 1 | NO | 1 | NO | 财务权限（0-否，1-是） |
| status | TINYINT | 1 | NO | 1 | NO | 状态（0-禁用，1-启用） |
| remark | TEXT | - | YES | NULL | NO | 备注 |
| create_time | DATETIME | - | NO | CURRENT_TIMESTAMP | NO | 创建时间 |
| update_time | DATETIME | - | NO | CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | NO | 更新时间 |

#### 2.1.3 索引设计

| 索引名称 | 索引类型 | 索引字段 | 索引方法 | 说明 |
|----------|----------|----------|----------|------|
| PRIMARY | 主键索引 | integrator_id | - | 主键索引 |
| uk_api_key | 唯一索引 | api_key | BTREE | API密钥唯一索引 |
| idx_status | 普通索引 | status | BTREE | 状态查询索引 |
| idx_company | 普通索引 | company_name | BTREE | 公司名称查询索引 |

#### 2.1.4 约束条件

| 约束类型 | 约束名称 | 约束字段 | 约束说明 |
|----------|----------|----------|----------|
| PRIMARY KEY | pk_integrator_config | integrator_id | 主键约束 |
| UNIQUE KEY | uk_api_key | api_key | API密钥唯一约束 |
| CHECK | chk_status | status | 状态值检查约束(0-1) |
| CHECK | chk_financial_permission | financial_permission | 财务权限检查约束(0-1) |
| CHECK | chk_is_platform_self | is_platform_self | 是否自营检查约束(0-1) |

---

### 2.2 游戏配置表 (game_config)

#### 2.2.1 表基本信息

| 项目 | 内容 |
|------|------|
| 表名 | game_config |
| 中文名 | 游戏配置表 |
| 用途 | 存储游戏基础配置信息，详细配置（符号、卷轴、权重等）在Nacos管理 |
| 存储引擎 | InnoDB |
| 字符集 | utf8mb4 |

#### 2.2.2 字段设计

| 字段名 | 类型 | 长度 | 允许NULL | 默认值 | 主键 | 说明 |
|--------|------|------|----------|--------|------|------|
| game_id | VARCHAR | 32 | NO | - | YES | 游戏ID |
| game_name | VARCHAR | 64 | NO | - | NO | 游戏名称（英文） |
| game_type | VARCHAR | 32 | NO | - | NO | 游戏类型（slot_3x3、slot_5x3等） |
| reel_count | INT | 11 | NO | 3 | NO | 卷轴数量 |
| symbol_count | INT | 11 | NO | 10 | NO | 符号数量 |
| min_bet | DECIMAL | 18,2 | NO | 0.10 | NO | 最小下注金额 |
| max_bet | DECIMAL | 18,2 | NO | 1000.00 | NO | 最大下注金额 |
| min_lines | INT | 11 | NO | 1 | NO | 最小下注线数 |
| max_lines | INT | 11 | NO | 20 | NO | 最大下注线数 |
| rtp_target | DECIMAL | 5,2 | NO | 95.00 | NO | 目标RTP值 |
| volatility_level | VARCHAR | 16 | NO | medium | NO | 波动性等级（low/medium/high） |
| special_features | JSON | - | YES | NULL | NO | 特殊功能配置（保留JSON格式） |
| nacos_config_id | VARCHAR | 64 | NO | - | NO | Nacos配置ID（关联详细配置） |
| status | TINYINT | 1 | NO | 1 | NO | 状态（0-禁用，1-启用） |
| version | INT | 11 | NO | 1 | NO | 配置版本号 |
| remark | TEXT | - | YES | NULL | NO | 备注 |
| create_time | DATETIME | - | NO | CURRENT_TIMESTAMP | NO | 创建时间 |
| update_time | DATETIME | - | NO | CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | NO | 更新时间 |

#### 2.2.3 索引设计

| 索引名称 | 索引类型 | 索引字段 | 索引方法 | 说明 |
|----------|----------|----------|----------|------|
| PRIMARY | 主键索引 | game_id | - | 主键索引 |
| idx_status | 普通索引 | status | BTREE | 状态查询索引 |
| idx_game_type | 普通索引 | game_type | BTREE | 游戏类型查询索引 |
| idx_nacos_config | 普通索引 | nacos_config_id | BTREE | Nacos配置ID查询索引 |
| idx_version | 普通索引 | version | BTREE | 版本号查询索引 |

#### 2.2.4 约束条件

| 约束类型 | 约束名称 | 约束字段 | 约束说明 |
|----------|----------|----------|----------|
| PRIMARY KEY | pk_game_config | game_id | 主键约束 |
| CHECK | chk_status | status | 状态值检查约束 |
| CHECK | chk_min_bet | min_bet | 最小下注金额检查 |
| CHECK | chk_max_bet | max_bet | 最大下注金额检查 |
| CHECK | chk_reel_count | reel_count | 卷轴数量检查(1-10) |
| CHECK | chk_symbol_count | symbol_count | 符号数量检查(1-20) |

---

### 2.3 用户信息表 (user_info)

#### 2.3.1 表基本信息

| 项目 | 内容 |
|------|------|
| 表名 | user_info |
| 中文名 | 用户信息表 |
| 用途 | 存储用户基本信息、登录信息和账户余额（余额字段独立行级锁，优化并发性能） |
| 存储引擎 | InnoDB |
| 字符集 | utf8mb4 |

#### 2.3.2 字段设计

| 字段名 | 类型 | 长度 | 允许NULL | 默认值 | 主键 | 说明 |
|--------|------|------|----------|--------|------|------|
| user_id | VARCHAR | 32 | NO | - | YES | 用户ID |
| integrator_id | VARCHAR | 32 | NO | - | NO | 所属集成商ID（必须关联集成商） |
| user_name | VARCHAR | 64 | YES | NULL | NO | 用户名称 |
| nickname | VARCHAR | 64 | YES | NULL | NO | 用户昵称 |
| email | VARCHAR | 64 | YES | NULL | NO | 用户邮箱 |
| phone | VARCHAR | 20 | YES | NULL | NO | 用户电话 |
| password_hash | VARCHAR | 128 | YES | NULL | NO | 密码哈希值（加密存储） |
| rtp_tolerance_threshold | DECIMAL | 5,2 | NO | 0.00 | NO | RTP容忍阈值（百分比，0表示完全容忍） |
| balance | DECIMAL | 18,2 | NO | 0.00 | NO | 账户余额（独立行级锁） |
| vip_level | TINYINT | 2 | NO | 0 | NO | VIP等级 |
| status | TINYINT | 1 | NO | 1 | NO | 状态（0-禁用，1-启用，2-冻结） |
| last_login_time | DATETIME | - | YES | NULL | NO | 最后登录时间 |
| last_login_ip | VARCHAR | 64 | YES | NULL | NO | 最后登录IP |
| create_time | DATETIME | - | NO | CURRENT_TIMESTAMP | NO | 创建时间 |
| update_time | DATETIME | - | NO | CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | NO | 更新时间 |

#### 2.3.3 索引设计

| 索引名称 | 索引类型 | 索引字段 | 索引方法 | 说明 |
|----------|----------|----------|----------|------|
| PRIMARY | 主键索引 | user_id | - | 主键索引 |
| idx_integrator | 普通索引 | integrator_id | BTREE | 集成商ID查询索引 |
| idx_email | 普通索引 | email | BTREE | 邮箱查询索引 |
| idx_phone | 普通索引 | phone | BTREE | 电话查询索引 |
| idx_status | 普通索引 | status | BTREE | 状态查询索引 |
| idx_create_time | 普通索引 | create_time | BTREE | 创建时间查询索引 |
| idx_rtp_tolerance | 普通索引 | rtp_tolerance_threshold | BTREE | RTP容忍阈值查询索引 |
| idx_balance | 普通索引 | balance | BTREE | 余额查询索引 |

#### 2.3.4 约束条件

| 约束类型 | 约束名称 | 约束字段 | 约束说明 |
|----------|----------|----------|----------|
| PRIMARY KEY | pk_user_info | user_id | 主键约束 |
| FOREIGN KEY | fk_integrator_id | integrator_id | 外键约束关联integrator_config表 |
| CHECK | chk_status | status | 状态值检查约束 |
| CHECK | chk_rtp_tolerance | rtp_tolerance_threshold | RTP容忍阈值检查约束(0-100) |
| CHECK | chk_balance | balance | 余额非负检查约束 |

---

### 2.3.5 用户财务统计表 (user_finance_stats)

#### 2.3.5.1 表基本信息

| 项目 | 内容 |
|------|------|
| 表名 | user_finance_stats |
| 中文名 | 用户财务统计表 |
| 用途 | 存储用户游戏统计数据（余额信息已移至user_info表，独立行级锁） |
| 存储引擎 | InnoDB |
| 字符集 | utf8mb4 |

#### 2.3.5.2 字段设计

| 字段名 | 类型 | 长度 | 允许NULL | 默认值 | 主键 | 说明 |
|--------|------|------|----------|--------|------|------|
| id | BIGINT | 20 | NO | AUTO_INCREMENT | YES | 自增主键 |
| user_id | VARCHAR | 32 | NO | - | NO | 关联用户ID |
| total_bet | DECIMAL | 18,2 | NO | 0.00 | NO | 总下注金额 |
| total_win | DECIMAL | 18,2 | NO | 0.00 | NO | 总赔付金额 |
| total_games | BIGINT | 20 | NO | 0 | NO | 总游戏次数 |
| net_result | DECIMAL | 18,2 | NO | 0.00 | NO | 净收益（总赔付-总下注） |
| rtp_value | DECIMAL | 5,2 | NO | 0.00 | NO | 用户RTP值（总赔付/总下注*100） |
| max_single_win | DECIMAL | 18,2 | NO | 0.00 | NO | 最大单次赔付 |
| update_time | DATETIME | - | NO | CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | NO | 更新时间 |

#### 2.3.5.3 索引设计

| 索引名称 | 索引类型 | 索引字段 | 索引方法 | 说明 |
|----------|----------|----------|----------|------|
| PRIMARY | 主键索引 | id | - | 主键索引 |
| UNIQUE KEY | uk_user_id | user_id | BTREE | 用户ID唯一索引 |
| idx_total_bet | 普通索引 | total_bet | BTREE | 总下注查询索引 |

#### 2.3.5.4 约束条件

| 约束类型 | 约束名称 | 约束字段 | 约束说明 |
|----------|----------|----------|----------|
| PRIMARY KEY | pk_user_finance_stats | id | 主键约束 |
| UNIQUE KEY | uk_user_finance | user_id | 用户财务记录唯一约束 |
| FOREIGN KEY | fk_user_id | user_id | 外键约束关联user_info表 |
| CHECK | chk_balance | balance | 余额非负检查约束 |

---

### 2.4 交易记录表 (transaction_records)

#### 2.4.1 表基本信息

| 项目 | 内容 |
|------|------|
| 表名 | transaction_records |
| 中文名 | 交易记录表 |
| 用途 | 存储用户资金变动记录，支持账户余额追溯 |
| 存储引擎 | InnoDB |
| 字符集 | utf8mb4 |

#### 2.4.2 字段设计

| 字段名 | 类型 | 长度 | 允许NULL | 默认值 | 主键 | 说明 |
|--------|------|------|----------|--------|------|------|
| id | BIGINT | 20 | NO | AUTO_INCREMENT | YES | 自增主键 |
| transaction_id | VARCHAR | 32 | NO | - | NO | 交易ID |
| user_id | VARCHAR | 32 | NO | - | NO | 用户ID |
| integrator_id | VARCHAR | 32 | NO | - | NO | 集成商ID |
| transaction_type | VARCHAR | 16 | NO | - | NO | 交易类型（bet/win/refund等） |
| amount | DECIMAL | 18,2 | NO | - | NO | 变动金额 |
| balance_before | DECIMAL | 18,2 | NO | - | NO | 交易前余额 |
| balance_after | DECIMAL | 18,2 | NO | - | NO | 交易后余额 |
| game_session_id | VARCHAR | 32 | YES | NULL | NO | 关联游戏会话ID |
| reference_id | VARCHAR | 32 | YES | NULL | NO | 关联参考ID |
| status | TINYINT | 1 | NO | 1 | NO | 交易状态（0-失败，1-成功，2-处理中） |
| error_message | VARCHAR | 256 | YES | NULL | NO | 错误信息 |
| create_time | DATETIME | - | NO | CURRENT_TIMESTAMP | NO | 创建时间 |

#### 2.4.3 索引设计

| 索引名称 | 索引类型 | 索引字段 | 索引方法 | 说明 |
|----------|----------|----------|----------|------|
| PRIMARY | 主键索引 | id | - | 主键索引 |
| uk_transaction_id | 唯一索引 | transaction_id | BTREE | 交易ID唯一索引 |
| idx_user_type | 组合索引 | user_id, transaction_type | BTREE | 用户交易类型查询索引 |
| idx_game_session | 普通索引 | game_session_id | BTREE | 游戏会话关联查询索引 |
| idx_create_time | 普通索引 | create_time | BTREE | 创建时间查询索引 |
| idx_status | 普通索引 | status | BTREE | 交易状态查询索引 |

#### 2.4.4 约束条件

| 约束类型 | 约束名称 | 约束字段 | 约束说明 |
|----------|----------|----------|----------|
| PRIMARY KEY | pk_transaction_records | id | 主键约束 |
| UNIQUE KEY | uk_transaction_id | transaction_id | 交易ID唯一约束 |
| FOREIGN KEY | fk_user_id | user_id | 外键约束关联user_info表 |
| FOREIGN KEY | fk_integrator_id | integrator_id | 外键约束关联integrator_config表 |
| CHECK | chk_status | status | 状态值检查约束 |

---

### 2.5 用户会话表 (user_session)

#### 2.5.1 表基本信息

| 项目 | 内容 |
|------|------|
| 表名 | user_session |
| 中文名 | 用户会话表 |
| 用途 | 存储用户游戏会话信息，支持会话管理和状态跟踪 |
| 存储引擎 | InnoDB |
| 字符集 | utf8mb4 |

#### 2.5.2 字段设计

| 字段名 | 类型 | 长度 | 允许NULL | 默认值 | 主键 | 说明 |
|--------|------|------|----------|--------|------|------|
| session_id | VARCHAR | 32 | NO | - | YES | 会话ID |
| user_id | VARCHAR | 32 | NO | - | NO | 用户ID |
| integrator_id | VARCHAR | 32 | NO | - | NO | 集成商ID |
| access_token | VARCHAR | 256 | NO | - | NO | 访问令牌 |
| game_url | VARCHAR | 512 | NO | - | NO | 加密游戏地址 |
| expire_time | DATETIME | - | NO | - | NO | 过期时间 |
| current_game_id | VARCHAR | 32 | YES | NULL | NO | 当前游戏ID |
| total_bet_amount | DECIMAL | 18,2 | NO | 0.00 | NO | 会话总下注金额 |
| total_win_amount | DECIMAL | 18,2 | NO | 0.00 | NO | 会话总赔付金额 |
| session_duration | INT | 11 | NO | 0 | NO | 会话持续时长（秒） |
| status | TINYINT | 1 | NO | 1 | NO | 会话状态（0-结束，1-活跃，2-过期） |
| client_ip | VARCHAR | 64 | YES | NULL | NO | 客户端IP地址 |
| user_agent | VARCHAR | 512 | YES | NULL | NO | 用户代理信息 |
| create_time | DATETIME | - | NO | CURRENT_TIMESTAMP | NO | 创建时间 |
| update_time | DATETIME | - | NO | CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | NO | 更新时间 |

#### 2.5.3 索引设计

| 索引名称 | 索引类型 | 索引字段 | 索引方法 | 说明 |
|----------|----------|----------|----------|------|
| PRIMARY | 主键索引 | session_id | - | 主键索引 |
| idx_user_id | 普通索引 | user_id | BTREE | 用户ID查询索引 |
| idx_integrator_id | 普通索引 | integrator_id | BTREE | 集成商ID查询索引 |
| idx_access_token | 普通索引 | access_token | BTREE | 访问令牌查询索引 |
| idx_expire_time | 普通索引 | expire_time | BTREE | 过期时间查询索引 |
| idx_status | 普通索引 | status | BTREE | 状态查询索引 |

#### 2.5.4 约束条件

| 约束类型 | 约束名称 | 约束字段 | 约束说明 |
|----------|----------|----------|----------|
| PRIMARY KEY | pk_user_session | session_id | 主键约束 |
| FOREIGN KEY | fk_user_id | user_id | 外键约束关联user_info表 |
| FOREIGN KEY | fk_integrator_id | integrator_id | 外键约束关联integrator_config表 |
| CHECK | chk_status | status | 状态值检查约束 |

---

### 2.6 游戏配置版本表 (game_config_version)

#### 2.6.1 表基本信息

| 项目 | 内容 |
|------|------|
| 表名 | game_config_version |
| 中文名 | 游戏配置版本表 |
| 用途 | 存储游戏配置的历史版本，支持配置回滚和审计 |
| 存储引擎 | InnoDB |
| 字符集 | utf8mb4 |

#### 2.6.2 字段设计

| 字段名 | 类型 | 长度 | 允许NULL | 默认值 | 主键 | 说明 |
|--------|------|------|----------|--------|------|------|
| id | BIGINT | 20 | NO | AUTO_INCREMENT | YES | 自增主键 |
| game_id | VARCHAR | 32 | NO | - | NO | 游戏ID |
| version | INT | 11 | NO | - | NO | 版本号 |
| config_snapshot | JSON | - | NO | - | NO | 配置快照JSON |
| change_reason | VARCHAR | 256 | YES | NULL | NO | 变更原因 |
| operator | VARCHAR | 32 | YES | NULL | NO | 操作人员 |
| create_time | DATETIME | - | NO | CURRENT_TIMESTAMP | NO | 创建时间 |

#### 2.6.3 索引设计

| 索引名称 | 索引类型 | 索引字段 | 索引方法 | 说明 |
|----------|----------|----------|----------|------|
| PRIMARY | 主键索引 | id | - | 主键索引 |
| uk_game_version | 唯一索引 | game_id, version | BTREE | 游戏ID和版本号唯一索引 |
| idx_game_id | 普通索引 | game_id | BTREE | 游戏ID查询索引 |
| idx_create_time | 普通索引 | create_time | BTREE | 创建时间查询索引 |

#### 2.6.4 约束条件

| 约束类型 | 约束名称 | 约束字段 | 约束说明 |
|----------|----------|----------|----------|
| PRIMARY KEY | pk_game_config_version | id | 主键约束 |
| UNIQUE KEY | uk_game_version | game_id, version | 游戏ID版本号唯一约束 |
| FOREIGN KEY | fk_game_id | game_id | 外键约束关联game_config表 |

---

### 2.7 Jackpot配置表 (jackpot_config)

#### 2.7.1 表基本信息

| 项目 | 内容 |
|------|------|
| 表名 | jackpot_config |
| 中文名 | Jackpot配置表 |
| 用途 | 存储Jackpot基础配置信息，包括注入比例、种子金额等 |
| 存储引擎 | InnoDB |
| 字符集 | utf8mb4 |

#### 2.7.2 字段设计

| 字段名 | 类型 | 长度 | 允许NULL | 默认值 | 主键 | 说明 |
|--------|------|------|----------|--------|------|------|
| game_id | VARCHAR | 32 | NO | - | YES | 游戏ID（"0"表示全局配置） |
| enabled | TINYINT | 1 | NO | 1 | NO | 是否启用（0-否，1-是） |
| contribution_rate | DECIMAL | 5,4 | NO | 0.0100 | NO | 注入比例（1%） |
| mini_seed | DECIMAL | 18,2 | NO | 100.00 | NO | Mini池种子金额 |
| minor_seed | DECIMAL | 18,2 | NO | 500.00 | NO | Minor池种子金额 |
| major_seed | DECIMAL | 18,2 | NO | 2000.00 | NO | Major池种子金额 |
| grand_seed | DECIMAL | 18,2 | NO | 10000.00 | NO | Grand池种子金额 |
| mini_ratio | DECIMAL | 4,3 | NO | 0.300 | NO | Mini池分配比例（30%） |
| minor_ratio | DECIMAL | 4,3 | NO | 0.250 | NO | Minor池分配比例（25%） |
| major_ratio | DECIMAL | 4,3 | NO | 0.250 | NO | Major池分配比例（25%） |
| grand_ratio | DECIMAL | 4,3 | NO | 0.200 | NO | Grand池分配比例（20%） |
| create_time | DATETIME | - | NO | CURRENT_TIMESTAMP | NO | 创建时间 |
| update_time | DATETIME | - | NO | CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | NO | 更新时间 |

#### 2.7.3 索引设计

| 索引名称 | 索引类型 | 索引字段 | 索引方法 | 说明 |
|----------|----------|----------|----------|------|
| PRIMARY | 主键索引 | game_id | - | 主键索引 |
| idx_enabled | 普通索引 | enabled | BTREE | 启用状态查询索引 |

#### 2.7.4 约束条件

| 约束类型 | 约束名称 | 约束字段 | 约束说明 |
|----------|----------|----------|----------|
| PRIMARY KEY | pk_jackpot_config | game_id | 主键约束 |
| FOREIGN KEY | fk_jackpot_game | game_id | 外键约束关联game_config表（game_id="0"除外） |
| CHECK | chk_enabled | enabled | 启用状态检查约束 |
| CHECK | chk_ratios | mini_ratio+minor_ratio+major_ratio+grand_ratio | 分配比例总和为1 |

---

### 2.8 Jackpot实时池表 (jackpot_pool)

#### 2.8.1 表基本信息

| 项目 | 内容 |
|------|------|
| 表名 | jackpot_pool |
| 中文名 | Jackpot实时池表 |
| 用途 | 存储Jackpot奖池实时金额，Redis为主，DB为备份 |
| 存储引擎 | InnoDB |
| 字符集 | utf8mb4 |

#### 2.8.2 字段设计

| 字段名 | 类型 | 长度 | 允许NULL | 默认值 | 主键 | 说明 |
|--------|------|------|----------|--------|------|------|
| game_id | VARCHAR | 32 | NO | - | YES | 游戏ID（"0"表示全局共享池） |
| pool_type | VARCHAR | 16 | NO | - | YES | 池子类型（mini/minor/major/grand） |
| current_amount | DECIMAL | 18,2 | NO | 0.00 | NO | 当前金额 |
| last_win_time | DATETIME | - | YES | NULL | NO | 最后中奖时间 |
| win_count | BIGINT | 20 | NO | 0 | NO | 中奖次数 |
| create_time | DATETIME | - | NO | CURRENT_TIMESTAMP | NO | 创建时间 |
| update_time | DATETIME | - | NO | CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | NO | 更新时间 |

#### 2.8.3 索引设计

| 索引名称 | 索引类型 | 索引字段 | 索引方法 | 说明 |
|----------|----------|----------|----------|------|
| PRIMARY | 主键索引 | game_id, pool_type | - | 联合主键索引 |
| idx_pool_type | 普通索引 | pool_type | BTREE | 池子类型查询索引 |
| idx_last_win_time | 普通索引 | last_win_time | BTREE | 最后中奖时间查询索引 |

#### 2.8.4 约束条件

| 约束类型 | 约束名称 | 约束字段 | 约束说明 |
|----------|----------|----------|----------|
| PRIMARY KEY | pk_jackpot_pool | game_id, pool_type | 联合主键约束 |
| CHECK | chk_pool_type | pool_type | 池子类型检查约束 |
| CHECK | chk_amount | current_amount | 金额非负检查约束 |

#### 2.8.5 设计说明

**混合池模式规则：**
- **全局池 (game_id="0")**：仅 Grand 池，所有游戏共享
- **游戏独立池**：当 game_id 在 jackpot_config 中存在且 enabled=1 时，创建该游戏的 Mini/Minor/Major 池

**判断逻辑：**
```
if jackpot_config 中存在 game_id 且 enabled=1:
    该游戏使用独立池 (mini/minor/major)
else:
    该游戏仅使用全局池 (game_id="0" 的 grand)
```

**Redis Key设计：**
```
jackpot:{game_id}:{pool_type}:amount → DECIMAL

# 示例：
# jackpot:0:grand:amount     # 全局Grand池
# jackpot:game_a:mini:amount  # 游戏A独立Mini池（如果配置且enabled=1）
```

---

## 3. 数据库性能优化

### 3.1 分表策略

| 表名 | 分表策略 | 分表字段 | 分表数量 | 说明 |
|------|----------|----------|----------|------|
| transaction_records | 按月分表 | create_time | 按需 | 支持交易记录归档 |

### 3.2 读写分离

| 数据源 | 类型 | 权限 | 用途 |
|--------|------|------|------|
| master | 写入库 | 读/写 | 处理所有写操作 |
| slave1 | 读取库 | 只读 | 处理查询操作 |
| slave2 | 读取库 | 只读 | 处理查询操作 |

### 3.3 缓存策略

| 数据类型 | 缓存位置 | 缓存时长 | 更新策略 |
|----------|----------|----------|----------|
| 游戏配置 | Redis | 1小时 | 主动更新+过期更新 |
| 集成商信息 | Redis | 30分钟 | 主动更新+过期更新 |
| 用户信息 | Redis | 10分钟 | 主动更新+过期更新 |
| 热门游戏统计 | Redis | 5分钟 | 实时更新 |
| Jackpot奖池 | Redis | 永久 | 实时更新，DB异步备份 |

---

## 4. 数据库安全配置

### 4.1 权限管理

| 用户名 | 权限范围 | 操作权限 | 说明 |
|--------|----------|----------|------|
| game_app_user | 业务库 | SELECT, INSERT, UPDATE, DELETE | 应用程序访问权限 |
| game_read_user | 业务库 | SELECT | 只读访问权限 |
| game_admin_user | 业务库 | ALL | 管理员权限 |

### 4.2 数据加密

| 数据类型 | 加密方式 | 加密字段 |
|----------|----------|----------|
| API密钥 | AES-256 | api_secret |
| 用户余额 | 应用层加密 | balance |
| 敏感配置 | AES-256 | 游戏配置中的敏感信息 |

### 4.3 备份策略

| 备份类型 | 备份频率 | 保留周期 | 存储位置 |
|----------|----------|----------|----------|
| 全量备份 | 每日一次 | 30天 | 云存储 |
| 增量备份 | 每小时一次 | 7天 | 云存储 |
| 日志备份 | 每日一次 | 90天 | 云存储 |

---

## 5. 数据库监控指标

### 5.1 性能监控

| 监控指标 | 阈值 | 告警级别 | 处理建议 |
|----------|------|----------|----------|
| QPS | > 10000 | 严重 | 扩容或优化查询 |
| 响应时间 | > 500ms | 严重 | 检查慢查询和索引 |
| 连接数 | > 800 | 警告 | 调整连接池配置 |
| CPU使用率 | > 80% | 警告 | 检查负载情况 |
| 磁盘使用率 | > 85% | 严重 | 清理数据或扩容 |

### 5.2 数据一致性监控

| 监控指标 | 检查频率 | 异常处理 |
|----------|----------|----------|
| 主从同步延迟 | 每分钟 | 延迟>1秒时告警 |
| 数据一致性校验 | 每日 | 发现不一致时立即修复 |
| 事务回滚监控 | 实时 | 回滚率>0.1%时告警 |

---

## 6. 版本记录

| 版本号 | 日期 | 修改内容 | 修改人 |
|--------|------|----------|--------|
| v2.3.0 | 2025-01-06 | 1. 删除jackpot_win_record表，中奖记录写入ClickHouse<br>2. 表数量从9个减少到8个 | - |
| v2.2.0 | 2025-01-05 | 1. 删除游戏详细配置相关表（symbol_config、symbol_multiplier、symbol_special_property、reel_config、reel_symbol_weight、pay_table_config）<br>2. 详细配置已迁移至Nacos配置中心管理<br>3. game_config表新增nacos_config_id字段关联Nacos配置<br>4. 表数量从15个减少到9个 | - |
| v2.1.0 | 2025-01-04 | 1. 新增Jackpot相关表：jackpot_config、jackpot_pool<br>2. 采用混合池模式：Grand池全局共享（game_id="0"），Mini/Minor/Major各游戏独立<br>3. 表数量从12个增加到15个<br>4. 新增Jackpot缓存策略说明 | - |
| v2.0.0 | 2025-01-03 | 1. 游戏配置表结构规范化重构<br>2. 将JSON字段拆分为独立表：symbol_config、symbol_multiplier、symbol_special_property、reel_config、reel_symbol_weight、pay_table_config<br>3. 新增5个关联表，减少数据冗余，提高数据一致性<br>4. 新增详细的表关系图和数据关联示例<br>5. 新增规范化设计优势说明和配置迁移建议 | - |
| v1.1.0 | 2025-01-02 | 1. 删除game_records表，所有游戏记录迁移至ClickHouse<br>2. 新增symbol_config JSON字段详细结构说明<br>3. 新增Nacos配置中心架构设计建议<br>4. 更新表编号结构，表数量从7个减少到6个 | - |
| v1.0.0 | 2025-01-01 | 初始版本发布 | - |