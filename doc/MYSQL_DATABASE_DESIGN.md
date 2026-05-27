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
| 2.2.0 | game_config | 游戏配置主表 | 存储游戏基础配置信息 |
| 2.2.1 | symbol_config | 符号配置表 | 存储游戏符号的基本配置信息 |
| 2.2.2 | symbol_multiplier | 符号赔付倍数表 | 存储符号在不同连击数下的赔付倍数配置 |
| 2.2.3 | symbol_special_property | 符号特殊属性表 | 存储符号的特殊属性配置 |
| 2.2.4 | reel_config | 卷轴配置表 | 存储卷轴的基本信息 |
| 2.2.5 | reel_symbol_weight | 卷轴符号权重表 | 存储每个卷轴中各个符号的权重配置 |
| 2.2.6 | pay_table_config | 赔付表配置表 | 存储赔付规则的配置信息 |
| 2.3 | user_info | 用户信息表 | 存储用户基本信息、登录信息和账户余额（独立行级锁） |
| 2.3.5 | user_finance_stats | 用户财务统计表 | 存储用户游戏统计信息（不含余额，减少锁竞争） |
| 2.4 | transaction_records | 交易记录表 | 存储用户资金变动记录 |
| 2.5 | user_session | 用户会话表 | 存储用户游戏会话信息 |
| 2.6 | game_config_version | 游戏配置版本表 | 存储游戏配置的历史版本 |

**表数量统计**：核心业务表12个，其中游戏配置相关表7个（1个主表+6个关联表），用户相关表2个（user_info含余额，user_finance_stats含统计）

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

### 2.2 游戏配置相关表

#### 2.2.0 游戏配置主表 (game_config) - 基础配置

#### 2.2.1 game_config 表基本信息

| 项目 | 内容 |
|------|------|
| 表名 | game_config |
| 中文名 | 游戏配置表（基础信息） |
| 用途 | 存储游戏基础配置信息，详细配置通过关联表管理 |
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

#### 2.2.1 符号配置表 (symbol_config)

#### 2.2.1.1 表基本信息

| 项目 | 内容 |
|------|------|
| 表名 | symbol_config |
| 中文名 | 符号配置表 |
| 用途 | 存储游戏符号的基本配置信息 |
| 存储引擎 | InnoDB |
| 字符集 | utf8mb4 |

#### 2.2.1.2 字段设计

| 字段名 | 类型 | 长度 | 允许NULL | 默认值 | 主键 | 说明 |
|--------|------|------|----------|--------|------|------|
| id | BIGINT | 20 | NO | AUTO_INCREMENT | YES | 自增主键 |
| game_id | VARCHAR | 32 | NO | - | NO | 关联游戏ID |
| symbol_id | VARCHAR | 32 | NO | - | NO | 符号ID（英文，如cherry） |
| symbol_name | VARCHAR | 64 | NO | - | NO | 符号名称（英文，如cherry） |
| symbol_type | VARCHAR | 16 | NO | normal | NO | 符号类型：normal、wild、scatter |
| description | VARCHAR | 256 | YES | NULL | NO | 符号描述 |
| is_active | TINYINT | 1 | NO | 1 | NO | 是否启用（0-否，1-是） |
| sort_order | INT | 11 | NO | 0 | NO | 排序顺序 |
| create_time | DATETIME | - | NO | CURRENT_TIMESTAMP | NO | 创建时间 |
| update_time | DATETIME | - | NO | CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | NO | 更新时间 |

#### 2.2.1.3 索引设计

| 索引名称 | 索引类型 | 索引字段 | 索引方法 | 说明 |
|----------|----------|----------|----------|------|
| PRIMARY | 主键索引 | id | - | 主键索引 |
| uk_game_symbol | 唯一索引 | game_id, symbol_id | BTREE | 游戏和符号组合唯一索引 |
| idx_symbol_type | 普通索引 | symbol_type | BTREE | 符号类型查询索引 |
| idx_sort_order | 普通索引 | sort_order | BTREE | 排序查询索引 |

#### 2.2.1.4 约束条件

| 约束类型 | 约束名称 | 约束字段 | 约束说明 |
|----------|----------|----------|----------|
| PRIMARY KEY | pk_symbol_config | id | 主键约束 |
| UNIQUE KEY | uk_game_symbol | game_id, symbol_id | 游戏符号组合唯一约束 |
| FOREIGN KEY | fk_game_id | game_id | 外键约束关联game_config表 |
| CHECK | chk_symbol_type | symbol_type | 符号类型检查约束 |
| CHECK | chk_is_active | is_active | 启用状态检查约束 |

---

#### 2.2.2 符号赔付倍数表 (symbol_multiplier)

#### 2.2.2.1 表基本信息

| 项目 | 内容 |
|------|------|
| 表名 | symbol_multiplier |
| 中文名 | 符号赔付倍数表 |
| 用途 | 存储符号在不同连击数下的赔付倍数配置 |
| 存储引擎 | InnoDB |
| 字符集 | utf8mb4 |

#### 2.2.2.2 字段设计

| 字段名 | 类型 | 长度 | 允许NULL | 默认值 | 主键 | 说明 |
|--------|------|------|----------|--------|------|------|
| id | BIGINT | 20 | NO | AUTO_INCREMENT | YES | 自增主键 |
| symbol_config_id | BIGINT | 20 | NO | - | NO | 关联符号配置ID |
| match_count | INT | 11 | NO | - | NO | 连击数（如2、3、4、5） |
| multiplier | DECIMAL | 10,2 | NO | 0.00 | NO | 赔付倍数 |
| is_bet_line | TINYINT | 1 | NO | 1 | NO | 是否需要匹配下注线（0-否，1-是） |
| create_time | DATETIME | - | NO | CURRENT_TIMESTAMP | NO | 创建时间 |

#### 2.2.2.3 索引设计

| 索引名称 | 索引类型 | 索引字段 | 索引方法 | 说明 |
|----------|----------|----------|----------|------|
| PRIMARY | 主键索引 | id | - | 主键索引 |
| uk_symbol_match | 唯一索引 | symbol_config_id, match_count | BTREE | 符号和连击数组合唯一索引 |
| idx_multiplier | 普通索引 | multiplier | BTREE | 赔付倍数查询索引 |

#### 2.2.2.4 约束条件

| 约束类型 | 约束名称 | 约束字段 | 约束说明 |
|----------|----------|----------|----------|
| PRIMARY KEY | pk_symbol_multiplier | id | 主键约束 |
| UNIQUE KEY | uk_symbol_match | symbol_config_id, match_count | 符号连击数组合唯一约束 |
| FOREIGN KEY | fk_symbol_config_id | symbol_config_id | 外键约束关联symbol_config表 |
| CHECK | chk_match_count | match_count | 连击数检查约束(1-10) |
| CHECK | chk_multiplier | multiplier | 赔付倍数检查约束(>0) |

---

#### 2.2.3 符号特殊属性表 (symbol_special_property)

#### 2.2.3.1 表基本信息

| 项目 | 内容 |
|------|------|
| 表名 | symbol_special_property |
| 中文名 | 符号特殊属性表 |
| 用途 | 存储符号的特殊属性，如wild的可替代列表、scatter的免费旋转等 |
| 存储引擎 | InnoDB |
| 字符集 | utf8mb4 |

#### 2.2.3.2 字段设计

| 字段名 | 类型 | 长度 | 允许NULL | 默认值 | 主键 | 说明 |
|--------|------|------|----------|--------|------|------|
| id | BIGINT | 20 | NO | AUTO_INCREMENT | YES | 自增主键 |
| symbol_config_id | BIGINT | 20 | NO | - | NO | 关联符号配置ID |
| property_name | VARCHAR | 32 | NO | - | NO | 属性名称：substitute、free_spins、any_position等 |
| property_value | VARCHAR | 256 | NO | - | NO | 属性值（JSON字符串或具体值） |
| create_time | DATETIME | - | NO | CURRENT_TIMESTAMP | NO | 创建时间 |

#### 2.2.3.3 索引设计

| 索引名称 | 索引类型 | 索引字段 | 索引方法 | 说明 |
|----------|----------|----------|----------|------|
| PRIMARY | 主键索引 | id | - | 主键索引 |
| uk_symbol_property | 唯一索引 | symbol_config_id, property_name | BTREE | 符号属性组合唯一索引 |
| idx_property_name | 普通索引 | property_name | BTREE | 属性名称查询索引 |

#### 2.2.3.4 约束条件

| 约束类型 | 约束名称 | 约束字段 | 约束说明 |
|----------|----------|----------|----------|
| PRIMARY KEY | pk_symbol_special_property | id | 主键约束 |
| UNIQUE KEY | uk_symbol_property | symbol_config_id, property_name | 符号属性组合唯一约束 |
| FOREIGN KEY | fk_symbol_config_id | symbol_config_id | 外键约束关联symbol_config表 |

---

#### 2.2.4 卷轴配置表 (reel_config)

#### 2.2.4.1 表基本信息

| 项目 | 内容 |
|------|------|
| 表名 | reel_config |
| 中文名 | 卷轴配置表 |
| 用途 | 存储卷轴的基本信息 |
| 存储引擎 | InnoDB |
| 字符集 | utf8mb4 |

#### 2.2.4.2 字段设计

| 字段名 | 类型 | 长度 | 允许NULL | 默认值 | 主键 | 说明 |
|--------|------|------|----------|--------|------|------|
| id | BIGINT | 20 | NO | AUTO_INCREMENT | YES | 自增主键 |
| game_id | VARCHAR | 32 | NO | - | NO | 关联游戏ID |
| reel_index | INT | 11 | NO | - | NO | 卷轴索引（1,2,3,4,5） |
| reel_name | VARCHAR | 32 | NO | - | NO | 卷轴名称（如reel1, reel2） |
| create_time | DATETIME | - | NO | CURRENT_TIMESTAMP | NO | 创建时间 |

#### 2.2.4.3 索引设计

| 索引名称 | 索引类型 | 索引字段 | 索引方法 | 说明 |
|----------|----------|----------|----------|------|
| PRIMARY | 主键索引 | id | - | 主键索引 |
| uk_game_reel | 唯一索引 | game_id, reel_index | BTREE | 游戏和卷轴索引组合唯一索引 |

#### 2.2.4.4 约束条件

| 约束类型 | 约束名称 | 约束字段 | 约束说明 |
|----------|----------|----------|----------|
| PRIMARY KEY | pk_reel_config | id | 主键约束 |
| UNIQUE KEY | uk_game_reel | game_id, reel_index | 游戏卷轴组合唯一约束 |
| FOREIGN KEY | fk_game_id | game_id | 外键约束关联game_config表 |
| CHECK | chk_reel_index | reel_index | 卷轴索引检查约束(1-10) |

---

#### 2.2.5 卷轴符号权重表 (reel_symbol_weight)

#### 2.2.5.1 表基本信息

| 项目 | 内容 |
|------|------|
| 表名 | reel_symbol_weight |
| 中文名 | 卷轴符号权重表 |
| 用途 | 存储每个卷轴中各个符号的权重配置 |
| 存储引擎 | InnoDB |
| 字符集 | utf8mb4 |

#### 2.2.5.2 字段设计

| 字段名 | 类型 | 长度 | 允许NULL | 默认值 | 主键 | 说明 |
|--------|------|------|----------|--------|------|------|
| id | BIGINT | 20 | NO | AUTO_INCREMENT | YES | 自增主键 |
| reel_config_id | BIGINT | 20 | NO | - | NO | 关联卷轴配置ID |
| symbol_config_id | BIGINT | 20 | NO | - | NO | 关联符号配置ID |
| weight | INT | 11 | NO | 0 | NO | 符号权重（权重越大出现概率越高） |
| create_time | DATETIME | - | NO | CURRENT_TIMESTAMP | NO | 创建时间 |

#### 2.2.5.3 索引设计

| 索引名称 | 索引类型 | 索引字段 | 索引方法 | 说明 |
|----------|----------|----------|----------|------|
| PRIMARY | 主键索引 | id | - | 主键索引 |
| uk_reel_symbol | 唯一索引 | reel_config_id, symbol_config_id | BTREE | 卷轴和符号组合唯一索引 |
| idx_weight | 普通索引 | weight | BTREE | 权重查询索引 |

#### 2.2.5.4 约束条件

| 约束类型 | 约束名称 | 约束字段 | 约束说明 |
|----------|----------|----------|----------|
| PRIMARY KEY | pk_reel_symbol_weight | id | 主键约束 |
| UNIQUE KEY | uk_reel_symbol | reel_config_id, symbol_config_id | 卷轴符号组合唯一约束 |
| FOREIGN KEY | fk_reel_config_id | reel_config_id | 外键约束关联reel_config表 |
| FOREIGN KEY | fk_symbol_config_id | symbol_config_id | 外键约束关联symbol_config表 |
| CHECK | chk_weight | weight | 权重检查约束(>=0) |

---

#### 2.2.6 赔付表配置 (pay_table_config)

#### 2.2.6.1 表基本信息

| 项目 | 内容 |
|------|------|
| 表名 | pay_table_config |
| 中文名 | 赔付表配置表 |
| 用途 | 存储赔付规则的配置信息 |
| 存储引擎 | InnoDB |
| 字符集 | utf8mb4 |

#### 2.2.6.2 字段设计

| 字段名 | 类型 | 长度 | 允许NULL | 默认值 | 主键 | 说明 |
|--------|------|------|----------|--------|------|------|
| id | BIGINT | 20 | NO | AUTO_INCREMENT | YES | 自增主键 |
| game_id | VARCHAR | 32 | NO | - | NO | 关联游戏ID |
| pay_line_count | INT | 11 | NO | 20 | NO | 赔付线数量 |
| pay_line_pattern | JSON | - | YES | NULL | NO | 赔付线模式配置 |
| create_time | DATETIME | - | NO | CURRENT_TIMESTAMP | NO | 创建时间 |
| update_time | DATETIME | - | NO | CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | NO | 更新时间 |

#### 2.2.6.3 索引设计

| 索引名称 | 索引类型 | 索引字段 | 索引方法 | 说明 |
|----------|----------|----------|----------|------|
| PRIMARY | 主键索引 | id | - | 主键索引 |
| uk_game_paytable | 唯一索引 | game_id | BTREE | 游戏赔付表唯一索引 |

#### 2.2.6.4 约束条件

| 约束类型 | 约束名称 | 约束字段 | 约束说明 |
|----------|----------|----------|----------|
| PRIMARY KEY | pk_pay_table_config | id | 主键约束 |
| UNIQUE KEY | uk_game_paytable | game_id | 游戏赔付表唯一约束 |
| FOREIGN KEY | fk_game_id | game_id | 外键约束关联game_config表 |
| CHECK | chk_pay_line_count | pay_line_count | 赔付线数量检查约束(1-100) |

---

#### 2.2.7 游戏配置数据结构规范化说明

##### 规范化设计的优势

| 优势 | 说明 | 效果 |
|------|------|------|
| **减少数据冗余** | 符号配置可在多个游戏间复用 | 降低存储成本，提高数据一致性 |
| **稳定的数据结构** | 关系型表结构，约束明确 | 减少数据异常，提高数据质量 |
| **便于查询和维护** | 支持标准SQL查询，无需解析JSON | 简化应用逻辑，提高查询性能 |
| **版本控制精确** | 可以精确追踪单个符号或权重的变更 | 便于问题定位和配置回滚 |
| **数据完整性保证** | 外键约束确保引用完整性 | 防止孤儿数据，保证数据一致性 |

##### 表关系图

```
game_config (游戏配置主表)
├─ symbol_config (符号配置) - 1:N
│   ├─ symbol_multiplier (符号赔付倍数) - 1:N  
│   └─ symbol_special_property (符号特殊属性) - 1:N
├─ reel_config (卷轴配置) - 1:N
│   └─ reel_symbol_weight (卷轴符号权重) - 1:N
└─ pay_table_config (赔付表配置) - 1:1
```

##### 数据关联示例

**查询游戏所有符号及赔付倍数**：
```sql
SELECT 
    g.game_id,
    g.game_name,
    s.symbol_id,
    s.symbol_name,
    s.symbol_type,
    m.match_count,
    m.multiplier,
    m.is_bet_line
FROM game_config g
LEFT JOIN symbol_config s ON g.game_id = s.game_id
LEFT JOIN symbol_multiplier m ON s.id = m.symbol_config_id
WHERE g.game_id = 'game_001'
ORDER BY s.sort_order, m.match_count;
```

**查询特定卷轴的符号权重分布**：
```sql
SELECT 
    g.game_id,
    r.reel_index,
    r.reel_name,
    s.symbol_id,
    s.symbol_name,
    rsw.weight
FROM game_config g
JOIN reel_config r ON g.game_id = r.game_id
JOIN reel_symbol_weight rsw ON r.id = rsw.reel_config_id
JOIN symbol_config s ON rsw.symbol_config_id = s.id
WHERE g.game_id = 'game_001' AND r.reel_index = 1
ORDER BY rsw.weight DESC;
```

##### 配置迁移建议

| 迁移阶段 | 操作内容 | 注意事项 |
|----------|----------|----------|
| **第一阶段** | 创建新表结构，保留原有JSON字段 | 确保新表结构完整，测试数据导入 |
| **第二阶段** | 数据迁移，将JSON数据导入新表 | 保持数据一致性，处理异常数据 |
| **第三阶段** | 应用层适配，支持新旧两种配置读取方式 | 确保平滑过渡，无业务中断 |
| **第四阶段** | 验证新表数据正确性，删除旧JSON字段 | 充分测试，避免数据丢失 |

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
| v2.0.0 | 2025-01-03 | 1. 游戏配置表结构规范化重构<br>2. 将JSON字段拆分为独立表：symbol_config、symbol_multiplier、symbol_special_property、reel_config、reel_symbol_weight、pay_table_config<br>3. 新增5个关联表，减少数据冗余，提高数据一致性<br>4. 新增详细的表关系图和数据关联示例<br>5. 新增规范化设计优势说明和配置迁移建议 | - |
| v1.1.0 | 2025-01-02 | 1. 删除game_records表，所有游戏记录迁移至ClickHouse<br>2. 新增symbol_config JSON字段详细结构说明<br>3. 新增Nacos配置中心架构设计建议<br>4. 更新表编号结构，表数量从7个减少到6个 | - |
| v1.0.0 | 2025-01-01 | 初始版本发布 | - |