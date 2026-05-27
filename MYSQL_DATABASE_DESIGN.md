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
| CHECK | chk_status | status | 状态值检查约束 |
| CHECK | chk_financial_permission | financial_permission | 财务权限检查约束 |

---

### 2.2 游戏配置表 (game_config)

#### 2.2.1 表基本信息

| 项目 | 内容 |
|------|------|
| 表名 | game_config |
| 中文名 | 游戏配置表 |
| 用途 | 存储游戏配置信息，包括符号定义、卷轴权重、赔付表等 |
| 存储引擎 | InnoDB |
| 字符集 | utf8mb4 |

#### 2.2.2 字段设计

| 字段名 | 类型 | 长度 | 允许NULL | 默认值 | 主键 | 说明 |
|--------|------|------|----------|--------|------|------|
| game_id | VARCHAR | 32 | NO | - | YES | 游戏ID |
| game_name | VARCHAR | 64 | NO | - | NO | 游戏名称 |
| game_type | VARCHAR | 32 | NO | - | NO | 游戏类型（slot_3x3、slot_5x3等） |
| symbol_config | JSON | - | NO | - | NO | 符号定义配置 |
| reel_weights | JSON | - | NO | - | NO | 卷轴权重配置 |
| pay_table | JSON | - | NO | - | NO | 赔付表配置 |
| min_bet | DECIMAL | 18,2 | NO | 0.10 | NO | 最小下注金额 |
| max_bet | DECIMAL | 18,2 | NO | 1000.00 | NO | 最大下注金额 |
| min_lines | INT | 11 | NO | 1 | NO | 最小下注线数 |
| max_lines | INT | 11 | NO | 20 | NO | 最大下注线数 |
| rtp_target | DECIMAL | 5,2 | NO | 95.00 | NO | 目标RTP值 |
| volatility_level | VARCHAR | 16 | NO | medium | NO | 波动性等级（low/medium/high） |
| special_features | JSON | - | YES | NULL | NO | 特殊功能配置 |
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

---

### 2.3 用户信息表 (user_info)

#### 2.3.1 表基本信息

| 项目 | 内容 |
|------|------|
| 表名 | user_info |
| 中文名 | 用户信息表 |
| 用途 | 存储用户基本信息和账户余额 |
| 存储引擎 | InnoDB |
| 字符集 | utf8mb4 |

#### 2.3.2 字段设计

| 字段名 | 类型 | 长度 | 允许NULL | 默认值 | 主键 | 说明 |
|--------|------|------|----------|--------|------|------|
| user_id | VARCHAR | 32 | NO | - | YES | 用户ID |
| integrator_id | VARCHAR | 32 | NO | - | NO | 所属集成商ID |
| user_name | VARCHAR | 64 | YES | NULL | NO | 用户名称 |
| nickname | VARCHAR | 64 | YES | NULL | NO | 用户昵称 |
| email | VARCHAR | 64 | YES | NULL | NO | 用户邮箱 |
| phone | VARCHAR | 20 | YES | NULL | NO | 用户电话 |
| balance | DECIMAL | 18,2 | NO | 0.00 | NO | 账户余额 |
| total_bet | DECIMAL | 18,2 | NO | 0.00 | NO | 总下注金额 |
| total_win | DECIMAL | 18,2 | NO | 0.00 | NO | 总赔付金额 |
| total_games | BIGINT | 20 | NO | 0 | NO | 总游戏次数 |
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
| idx_status | 普通索引 | status | BTREE | 状态查询索引 |
| idx_email | 普通索引 | email | BTREE | 邮箱查询索引 |
| idx_phone | 普通索引 | phone | BTREE | 电话查询索引 |
| idx_create_time | 普通索引 | create_time | BTREE | 创建时间查询索引 |

#### 2.3.4 约束条件

| 约束类型 | 约束名称 | 约束字段 | 约束说明 |
|----------|----------|----------|----------|
| PRIMARY KEY | pk_user_info | user_id | 主键约束 |
| FOREIGN KEY | fk_integrator_id | integrator_id | 外键约束关联integrator_config表 |
| CHECK | chk_status | status | 状态值检查约束 |
| CHECK | chk_balance | balance | 余额非负检查约束 |

---

### 2.4 游戏记录表 (game_records)

#### 2.4.1 表基本信息

| 项目 | 内容 |
|------|------|
| 表名 | game_records |
| 中文名 | 游戏记录表（迁移到ClickHouse） |
| 用途 | **注意：详细游戏日志已迁移到ClickHouse，本表仅存储核心业务信息，支持快速查询和关联** |
| **迁移说明** | 详细游戏日志数据已迁移至ClickHouse的game_log_detail表，MySQL仅保留核心字段 |

#### 2.4.2 字段设计

| 字段名 | 类型 | 长度 | 允许NULL | 默认值 | 主键 | 说明 |
|--------|------|------|----------|--------|------|------|


#### 2.4.3 字段说明变更

| 字段 | 原设计 | 当前设计 | 变更原因 |
|------|--------|----------|----------|

#### 2.4.4 数据同步说明

- **实时同步**：游戏结果同时写入MySQL和ClickHouse
- **ClickHouse职责**：存储详细游戏日志，支持大数据分析和RTP统计
- **数据关联**：通过代码实现跨库关联查询，例如查询用户游戏记录时，同时从ClickHouse查询详细日志

#### 2.4.5 索引设计

| 索引名称 | 索引类型 | 索引字段 | 索引方法 | 说明 |
|----------|----------|----------|----------|------|

#### 2.4.6 约束条件

| 约束类型 | 约束名称 | 约束字段 | 约束说明 |
|----------|----------|----------|----------|

---

### 2.5 交易记录表 (transaction_records)

#### 2.5.1 表基本信息

| 项目 | 内容 |
|------|------|
| 表名 | transaction_records |
| 中文名 | 交易记录表 |
| 用途 | 存储用户资金变动记录，支持账户余额追溯 |
| 存储引擎 | InnoDB |
| 字符集 | utf8mb4 |

#### 2.5.2 字段设计

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

#### 2.5.3 索引设计

| 索引名称 | 索引类型 | 索引字段 | 索引方法 | 说明 |
|----------|----------|----------|----------|------|
| PRIMARY | 主键索引 | id | - | 主键索引 |
| uk_transaction_id | 唯一索引 | transaction_id | BTREE | 交易ID唯一索引 |
| idx_user_type | 组合索引 | user_id, transaction_type | BTREE | 用户交易类型查询索引 |
| idx_game_session | 普通索引 | game_session_id | BTREE | 游戏会话关联查询索引 |
| idx_create_time | 普通索引 | create_time | BTREE | 创建时间查询索引 |
| idx_status | 普通索引 | status | BTREE | 交易状态查询索引 |

#### 2.5.4 约束条件

| 约束类型 | 约束名称 | 约束字段 | 约束说明 |
|----------|----------|----------|----------|
| PRIMARY KEY | pk_transaction_records | id | 主键约束 |
| UNIQUE KEY | uk_transaction_id | transaction_id | 交易ID唯一约束 |
| FOREIGN KEY | fk_user_id | user_id | 外键约束关联user_info表 |
| FOREIGN KEY | fk_integrator_id | integrator_id | 外键约束关联integrator_config表 |
| CHECK | chk_status | status | 状态值检查约束 |

---

### 2.6 用户会话表 (user_session)

#### 2.6.1 表基本信息

| 项目 | 内容 |
|------|------|
| 表名 | user_session |
| 中文名 | 用户会话表 |
| 用途 | 存储用户游戏会话信息，支持会话管理和状态跟踪 |
| 存储引擎 | InnoDB |
| 字符集 | utf8mb4 |

#### 2.6.2 字段设计

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

#### 2.6.3 索引设计

| 索引名称 | 索引类型 | 索引字段 | 索引方法 | 说明 |
|----------|----------|----------|----------|------|
| PRIMARY | 主键索引 | session_id | - | 主键索引 |
| idx_user_id | 普通索引 | user_id | BTREE | 用户ID查询索引 |
| idx_integrator_id | 普通索引 | integrator_id | BTREE | 集成商ID查询索引 |
| idx_access_token | 普通索引 | access_token | BTREE | 访问令牌查询索引 |
| idx_expire_time | 普通索引 | expire_time | BTREE | 过期时间查询索引 |
| idx_status | 普通索引 | status | BTREE | 状态查询索引 |

#### 2.6.4 约束条件

| 约束类型 | 约束名称 | 约束字段 | 约束说明 |
|----------|----------|----------|----------|
| PRIMARY KEY | pk_user_session | session_id | 主键约束 |
| FOREIGN KEY | fk_user_id | user_id | 外键约束关联user_info表 |
| FOREIGN KEY | fk_integrator_id | integrator_id | 外键约束关联integrator_config表 |
| CHECK | chk_status | status | 状态值检查约束 |

---

### 2.7 游戏配置版本表 (game_config_version)

#### 2.7.1 表基本信息

| 项目 | 内容 |
|------|------|
| 表名 | game_config_version |
| 中文名 | 游戏配置版本表 |
| 用途 | 存储游戏配置的历史版本，支持配置回滚和审计 |
| 存储引擎 | InnoDB |
| 字符集 | utf8mb4 |

#### 2.7.2 字段设计

| 字段名 | 类型 | 长度 | 允许NULL | 默认值 | 主键 | 说明 |
|--------|------|------|----------|--------|------|------|
| id | BIGINT | 20 | NO | AUTO_INCREMENT | YES | 自增主键 |
| game_id | VARCHAR | 32 | NO | - | NO | 游戏ID |
| version | INT | 11 | NO | - | NO | 版本号 |
| config_snapshot | JSON | - | NO | - | NO | 配置快照JSON |
| change_reason | VARCHAR | 256 | YES | NULL | NO | 变更原因 |
| operator | VARCHAR | 32 | YES | NULL | NO | 操作人员 |
| create_time | DATETIME | - | NO | CURRENT_TIMESTAMP | NO | 创建时间 |

#### 2.7.3 索引设计

| 索引名称 | 索引类型 | 索引字段 | 索引方法 | 说明 |
|----------|----------|----------|----------|------|
| PRIMARY | 主键索引 | id | - | 主键索引 |
| uk_game_version | 唯一索引 | game_id, version | BTREE | 游戏ID和版本号唯一索引 |
| idx_game_id | 普通索引 | game_id | BTREE | 游戏ID查询索引 |
| idx_create_time | 普通索引 | create_time | BTREE | 创建时间查询索引 |

#### 2.7.4 约束条件

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
| game_records | 按月分表 | create_time | 按需 | 支持历史数据归档和查询优化 |
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
| v1.0.0 | 2025-01-01 | 初始版本发布 | - |