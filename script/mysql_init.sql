-- MySQL数据库初始化脚本
-- 数据库名称：slot_game_service
-- 版本：v2.0.0
-- 说明：游戏服务模块MySQL数据库表初始化脚本（规范化表结构版本）
-- 变更记录：
--   v2.0.0: 数据库结构规范化重构，将游戏配置JSON字段拆分为多个关联表
--   v1.0.1: 删除game_records表，所有游戏记录数据存储在ClickHouse中
--   v1.0.0: 初始版本

-- 创建数据库
CREATE DATABASE IF NOT EXISTS slot_game_service DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE slot_game_service;

-- 禁用外键检查，方便重建表结构
SET FOREIGN_KEY_CHECKS = 0;

-- 设置时区
SET time_zone = '+00:00';

-- ============================================
-- 2.1 集成商配置表 (integrator_config)
-- ============================================
DROP TABLE IF EXISTS integrator_config;

CREATE TABLE integrator_config (
    integrator_id VARCHAR(32) NOT NULL COMMENT '集成商ID',
    integrator_name VARCHAR(64) NOT NULL COMMENT '集成商名称',
    company_name VARCHAR(128) DEFAULT NULL COMMENT '公司名称',
    contact_person VARCHAR(32) DEFAULT NULL COMMENT '联系人',
    contact_email VARCHAR(64) DEFAULT NULL COMMENT '联系邮箱',
    contact_phone VARCHAR(20) DEFAULT NULL COMMENT '联系电话',
    is_platform_self TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否平台自营（0-否，1-是）',
    api_key VARCHAR(64) NOT NULL COMMENT 'API密钥',
    api_secret VARCHAR(128) NOT NULL COMMENT 'API密钥（加密存储）',
    game_access_permitted JSON DEFAULT NULL COMMENT '游戏访问权限列表',
    financial_permission TINYINT(1) NOT NULL DEFAULT 1 COMMENT '财务权限（0-否，1-是）',
    status TINYINT(1) NOT NULL DEFAULT 1 COMMENT '状态（0-禁用，1-启用）',
    remark TEXT DEFAULT NULL COMMENT '备注',
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    PRIMARY KEY (integrator_id),
    UNIQUE KEY uk_api_key (api_key),
    KEY idx_status (status),
    KEY idx_company (company_name),
    CONSTRAINT chk_status CHECK (status IN (0, 1)),
    CONSTRAINT chk_financial_permission CHECK (financial_permission IN (0, 1))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='集成商配置表';

-- ============================================
-- 2.2.0 游戏配置主表 (game_config) - 基础配置
-- ============================================
DROP TABLE IF EXISTS game_config;

CREATE TABLE game_config (
    game_id VARCHAR(32) NOT NULL COMMENT '游戏ID',
    game_name VARCHAR(64) NOT NULL COMMENT '游戏名称（英文）',
    game_type VARCHAR(32) NOT NULL COMMENT '游戏类型（slot_3x3、slot_5x3等）',
    reel_count INT(11) NOT NULL DEFAULT 3 COMMENT '卷轴数量',
    symbol_count INT(11) NOT NULL DEFAULT 10 COMMENT '符号数量',
    min_bet DECIMAL(18,2) NOT NULL DEFAULT 0.10 COMMENT '最小下注金额',
    max_bet DECIMAL(18,2) NOT NULL DEFAULT 1000.00 COMMENT '最大下注金额',
    min_lines INT(11) NOT NULL DEFAULT 1 COMMENT '最小下注线数',
    max_lines INT(11) NOT NULL DEFAULT 20 COMMENT '最大下注线数',
    rtp_target DECIMAL(5,2) NOT NULL DEFAULT 95.00 COMMENT '目标RTP值',
    volatility_level VARCHAR(16) NOT NULL DEFAULT 'medium' COMMENT '波动性等级（low/medium/high）',
    special_features JSON DEFAULT NULL COMMENT '特殊功能配置（保留JSON格式）',
    status TINYINT(1) NOT NULL DEFAULT 1 COMMENT '状态（0-禁用，1-启用）',
    version INT(11) NOT NULL DEFAULT 1 COMMENT '配置版本号',
    remark TEXT DEFAULT NULL COMMENT '备注',
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    PRIMARY KEY (game_id),
    KEY idx_status (status),
    KEY idx_game_type (game_type),
    KEY idx_version (version),
    CONSTRAINT chk_game_status CHECK (status IN (0, 1)),
    CONSTRAINT chk_min_bet CHECK (min_bet > 0),
    CONSTRAINT chk_max_bet CHECK (max_bet >= min_bet),
    CONSTRAINT chk_reel_count CHECK (reel_count BETWEEN 1 AND 10),
    CONSTRAINT chk_symbol_count CHECK (symbol_count BETWEEN 1 AND 20)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='游戏配置主表';

-- ============================================
-- 2.2.1 符号配置表 (symbol_config)
-- ============================================
DROP TABLE IF EXISTS symbol_config;

CREATE TABLE symbol_config (
    id BIGINT(20) NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    game_id VARCHAR(32) NOT NULL COMMENT '关联游戏ID',
    symbol_id VARCHAR(32) NOT NULL COMMENT '符号ID（英文，如cherry）',
    symbol_name VARCHAR(64) NOT NULL COMMENT '符号名称（英文，如cherry）',
    symbol_type VARCHAR(16) NOT NULL DEFAULT 'normal' COMMENT '符号类型：normal、wild、scatter',
    description VARCHAR(256) DEFAULT NULL COMMENT '符号描述',
    is_active TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否启用（0-否，1-是）',
    sort_order INT(11) NOT NULL DEFAULT 0 COMMENT '排序顺序',
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    PRIMARY KEY (id),
    UNIQUE KEY uk_game_symbol (game_id, symbol_id),
    KEY idx_symbol_type (symbol_type),
    KEY idx_sort_order (sort_order),
    CONSTRAINT fk_symbol_game_id FOREIGN KEY (game_id) REFERENCES game_config (game_id) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT chk_symbol_type CHECK (symbol_type IN ('normal', 'wild', 'scatter')),
    CONSTRAINT chk_is_active CHECK (is_active IN (0, 1))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='符号配置表';

-- ============================================
-- 2.2.2 符号赔付倍数表 (symbol_multiplier)
-- ============================================
DROP TABLE IF EXISTS symbol_multiplier;

CREATE TABLE symbol_multiplier (
    id BIGINT(20) NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    symbol_config_id BIGINT(20) NOT NULL COMMENT '关联符号配置ID',
    match_count INT(11) NOT NULL COMMENT '连击数（如2、3、4、5）',
    multiplier DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '赔付倍数',
    is_bet_line TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否需要匹配下注线（0-否，1-是）',
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    
    PRIMARY KEY (id),
    UNIQUE KEY uk_symbol_match (symbol_config_id, match_count),
    KEY idx_multiplier (multiplier),
    CONSTRAINT fk_symbol_config_id FOREIGN KEY (symbol_config_id) REFERENCES symbol_config (id) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT chk_match_count CHECK (match_count BETWEEN 1 AND 10),
    CONSTRAINT chk_multiplier CHECK (multiplier > 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='符号赔付倍数表';

-- ============================================
-- 2.2.3 符号特殊属性表 (symbol_special_property)
-- ============================================
DROP TABLE IF EXISTS symbol_special_property;

CREATE TABLE symbol_special_property (
    id BIGINT(20) NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    symbol_config_id BIGINT(20) NOT NULL COMMENT '关联符号配置ID',
    property_name VARCHAR(32) NOT NULL COMMENT '属性名称：substitute、free_spins、any_position等',
    property_value VARCHAR(256) NOT NULL COMMENT '属性值（JSON字符串或具体值）',
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    
    PRIMARY KEY (id),
    UNIQUE KEY uk_symbol_property (symbol_config_id, property_name),
    KEY idx_property_name (property_name),
    CONSTRAINT fk_symbol_special_property_config_id FOREIGN KEY (symbol_config_id) REFERENCES symbol_config (id) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='符号特殊属性表';

-- ============================================
-- 2.2.4 卷轴配置表 (reel_config)
-- ============================================
DROP TABLE IF EXISTS reel_config;

CREATE TABLE reel_config (
    id BIGINT(20) NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    game_id VARCHAR(32) NOT NULL COMMENT '关联游戏ID',
    reel_index INT(11) NOT NULL COMMENT '卷轴索引（1,2,3,4,5）',
    reel_name VARCHAR(32) NOT NULL COMMENT '卷轴名称（如reel1, reel2）',
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    
    PRIMARY KEY (id),
    UNIQUE KEY uk_game_reel (game_id, reel_index),
    CONSTRAINT fk_reel_game_id FOREIGN KEY (game_id) REFERENCES game_config (game_id) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT chk_reel_index CHECK (reel_index BETWEEN 1 AND 10)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='卷轴配置表';

-- ============================================
-- 2.2.5 卷轴符号权重表 (reel_symbol_weight)
-- ============================================
DROP TABLE IF EXISTS reel_symbol_weight;

CREATE TABLE reel_symbol_weight (
    id BIGINT(20) NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    reel_config_id BIGINT(20) NOT NULL COMMENT '关联卷轴配置ID',
    symbol_config_id BIGINT(20) NOT NULL COMMENT '关联符号配置ID',
    weight INT(11) NOT NULL DEFAULT 0 COMMENT '符号权重（权重越大出现概率越高）',
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    
    PRIMARY KEY (id),
    UNIQUE KEY uk_reel_symbol (reel_config_id, symbol_config_id),
    KEY idx_weight (weight),
    CONSTRAINT fk_reel_config_id FOREIGN KEY (reel_config_id) REFERENCES reel_config (id) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT fk_reel_symbol_config_id FOREIGN KEY (symbol_config_id) REFERENCES symbol_config (id) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT chk_weight CHECK (weight >= 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='卷轴符号权重表';

-- ============================================
-- 2.2.6 赔付表配置表 (pay_table_config)
-- ============================================
DROP TABLE IF EXISTS pay_table_config;

CREATE TABLE pay_table_config (
    id BIGINT(20) NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    game_id VARCHAR(32) NOT NULL COMMENT '关联游戏ID',
    pay_line_count INT(11) NOT NULL DEFAULT 20 COMMENT '赔付线数量',
    pay_line_pattern JSON DEFAULT NULL COMMENT '赔付线模式配置',
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    PRIMARY KEY (id),
    UNIQUE KEY uk_game_paytable (game_id),
    CONSTRAINT fk_pay_table_game_id FOREIGN KEY (game_id) REFERENCES game_config (game_id) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT chk_pay_line_count CHECK (pay_line_count BETWEEN 1 AND 100)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='赔付表配置表';

-- ============================================
-- 2.3 用户信息表 (user_info)
-- ============================================
DROP TABLE IF EXISTS user_info;

CREATE TABLE user_info (
    user_id VARCHAR(32) NOT NULL COMMENT '用户ID',
    integrator_id VARCHAR(32) NOT NULL COMMENT '所属集成商ID（必须关联集成商）',
    user_name VARCHAR(64) DEFAULT NULL COMMENT '用户名称',
    nickname VARCHAR(64) DEFAULT NULL COMMENT '用户昵称',
    email VARCHAR(64) DEFAULT NULL COMMENT '用户邮箱',
    phone VARCHAR(20) DEFAULT NULL COMMENT '用户电话',
    password_hash VARCHAR(128) DEFAULT NULL COMMENT '密码哈希值（加密存储）',
    rtp_tolerance_threshold DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT 'RTP容忍阈值（百分比，0表示完全容忍）',
    balance DECIMAL(18,2) NOT NULL DEFAULT 0.00 COMMENT '账户余额（独立行级锁）',
    vip_level TINYINT(2) NOT NULL DEFAULT 0 COMMENT 'VIP等级',
    status TINYINT(1) NOT NULL DEFAULT 1 COMMENT '状态（0-禁用，1-启用，2-冻结）',
    last_login_time DATETIME DEFAULT NULL COMMENT '最后登录时间',
    last_login_ip VARCHAR(64) DEFAULT NULL COMMENT '最后登录IP',
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    PRIMARY KEY (user_id),
    KEY idx_integrator (integrator_id),
    KEY idx_email (email),
    KEY idx_phone (phone),
    KEY idx_status (status),
    KEY idx_create_time (create_time),
    KEY idx_rtp_tolerance (rtp_tolerance_threshold),
    KEY idx_balance (balance),
    CONSTRAINT fk_user_integrator FOREIGN KEY (integrator_id) REFERENCES integrator_config (integrator_id) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT chk_user_status CHECK (status IN (0, 1, 2)),
    CONSTRAINT chk_rtp_tolerance CHECK (rtp_tolerance_threshold >= 0 AND rtp_tolerance_threshold <= 100),
    CONSTRAINT chk_balance CHECK (balance >= 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户信息表';

-- ============================================
-- 2.3.5 用户财务统计表 (user_finance_stats)
-- ============================================
DROP TABLE IF EXISTS user_finance_stats;

CREATE TABLE user_finance_stats (
    id BIGINT(20) NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    user_id VARCHAR(32) NOT NULL COMMENT '关联用户ID',
    total_bet DECIMAL(18,2) NOT NULL DEFAULT 0.00 COMMENT '总下注金额',
    total_win DECIMAL(18,2) NOT NULL DEFAULT 0.00 COMMENT '总赔付金额',
    total_games BIGINT(20) NOT NULL DEFAULT 0 COMMENT '总游戏次数',
    net_result DECIMAL(18,2) NOT NULL DEFAULT 0.00 COMMENT '净收益（总赔付-总下注）',
    rtp_value DECIMAL(5,2) NOT NULL DEFAULT 0.00 COMMENT '用户RTP值（总赔付/总下注*100）',
    max_single_win DECIMAL(18,2) NOT NULL DEFAULT 0.00 COMMENT '最大单次赔付',
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    PRIMARY KEY (id),
    UNIQUE KEY uk_user_id (user_id),
    KEY idx_total_bet (total_bet),
    CONSTRAINT fk_user_finance_user_id FOREIGN KEY (user_id) REFERENCES user_info (user_id) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户财务统计表';

-- ============================================
-- 2.4 交易记录表 (transaction_records)
-- ============================================
DROP TABLE IF EXISTS transaction_records;

CREATE TABLE transaction_records (
    id BIGINT(20) NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    transaction_id VARCHAR(32) NOT NULL COMMENT '交易ID',
    user_id VARCHAR(32) NOT NULL COMMENT '用户ID',
    integrator_id VARCHAR(32) NOT NULL COMMENT '集成商ID',
    transaction_type VARCHAR(16) NOT NULL COMMENT '交易类型（bet/win/refund等）',
    amount DECIMAL(18,2) NOT NULL COMMENT '变动金额',
    balance_before DECIMAL(18,2) NOT NULL COMMENT '交易前余额',
    balance_after DECIMAL(18,2) NOT NULL COMMENT '交易后余额',
    game_session_id VARCHAR(32) DEFAULT NULL COMMENT '关联游戏会话ID',
    reference_id VARCHAR(32) DEFAULT NULL COMMENT '关联参考ID',
    status TINYINT(1) NOT NULL DEFAULT 1 COMMENT '交易状态（0-失败，1-成功，2-处理中）',
    error_message VARCHAR(256) DEFAULT NULL COMMENT '错误信息',
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    
    PRIMARY KEY (id),
    UNIQUE KEY uk_transaction_id (transaction_id),
    KEY idx_user_type (user_id, transaction_type),
    KEY idx_game_session (game_session_id),
    KEY idx_create_time (create_time),
    KEY idx_status (status),
    CONSTRAINT fk_transaction_user FOREIGN KEY (user_id) REFERENCES user_info (user_id) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT fk_transaction_integrator FOREIGN KEY (integrator_id) REFERENCES integrator_config (integrator_id) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT chk_transaction_status CHECK (status IN (0, 1, 2))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='交易记录表';

-- ============================================
-- 2.5 用户会话表 (user_session)
-- ============================================
DROP TABLE IF EXISTS user_session;

CREATE TABLE user_session (
    session_id VARCHAR(32) NOT NULL COMMENT '会话ID',
    user_id VARCHAR(32) NOT NULL COMMENT '用户ID',
    integrator_id VARCHAR(32) NOT NULL COMMENT '集成商ID',
    access_token VARCHAR(256) NOT NULL COMMENT '访问令牌',
    game_url VARCHAR(512) NOT NULL COMMENT '加密游戏地址',
    expire_time DATETIME NOT NULL COMMENT '过期时间',
    current_game_id VARCHAR(32) DEFAULT NULL COMMENT '当前游戏ID',
    total_bet_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00 COMMENT '会话总下注金额',
    total_win_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00 COMMENT '会话总赔付金额',
    session_duration INT(11) NOT NULL DEFAULT 0 COMMENT '会话持续时长（秒）',
    status TINYINT(1) NOT NULL DEFAULT 1 COMMENT '会话状态（0-结束，1-活跃，2-过期）',
    client_ip VARCHAR(64) DEFAULT NULL COMMENT '客户端IP地址',
    user_agent VARCHAR(512) DEFAULT NULL COMMENT '用户代理信息',
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    PRIMARY KEY (session_id),
    KEY idx_user_id (user_id),
    KEY idx_integrator_id (integrator_id),
    KEY idx_access_token (access_token),
    KEY idx_expire_time (expire_time),
    KEY idx_status (status),
    CONSTRAINT fk_session_user FOREIGN KEY (user_id) REFERENCES user_info (user_id) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT fk_session_integrator FOREIGN KEY (integrator_id) REFERENCES integrator_config (integrator_id) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT chk_session_status CHECK (status IN (0, 1, 2))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户会话表';

-- ============================================
-- 2.6 游戏配置版本表 (game_config_version)
-- ============================================
DROP TABLE IF EXISTS game_config_version;

CREATE TABLE game_config_version (
    id BIGINT(20) NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    game_id VARCHAR(32) NOT NULL COMMENT '游戏ID',
    version INT(11) NOT NULL COMMENT '版本号',
    config_snapshot JSON NOT NULL COMMENT '配置快照JSON',
    change_reason VARCHAR(256) DEFAULT NULL COMMENT '变更原因',
    operator VARCHAR(32) DEFAULT NULL COMMENT '操作人员',
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    
    PRIMARY KEY (id),
    UNIQUE KEY uk_game_version (game_id, version),
    KEY idx_game_id (game_id),
    KEY idx_create_time (create_time),
    CONSTRAINT fk_version_game FOREIGN KEY (game_id) REFERENCES game_config (game_id) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='游戏配置版本表';

-- ============================================
-- 创建索引优化查询性能
-- ============================================

-- 为交易记录表添加复合索引优化查询
ALTER TABLE transaction_records ADD INDEX idx_user_time_type (user_id, create_time, transaction_type);

-- 为用户会话表添加复合索引优化活跃会话查询
ALTER TABLE user_session ADD INDEX idx_user_status_time (user_id, status, create_time);

-- ============================================
-- 插入初始化数据
-- ============================================

-- 插入示例集成商数据
INSERT INTO integrator_config (integrator_id, integrator_name, company_name, contact_person, contact_email, contact_phone, is_platform_self, api_key, api_secret, status) VALUES
('platform_self_001', 'Platform Self', 'Platform Self Operations', 'Admin Team', 'admin@platform.com', '400-800-1234', 1, 'platform_api_key_001', HEX(AES_ENCRYPT('platform_secret_001', 'encryption_key')), 1),
('test_integrator_001', 'Test Integrator', 'Test Company', 'John Doe', 'test@example.com', '123-456-7890', 0, 'test_api_key_001', HEX(AES_ENCRYPT('test_secret_001', 'encryption_key')), 1);

-- ============================================
-- 插入示例游戏配置数据 - Lucky Fruits 3x3 slot
-- ============================================

-- 插入游戏配置主表数据
INSERT INTO game_config (game_id, game_name, game_type, reel_count, symbol_count, min_bet, max_bet, min_lines, max_lines, rtp_target, volatility_level, status) VALUES
('game_001', 'Lucky Fruits', 'slot_3x3', 3, 10, 0.10, 1000.00, 1, 20, 95.00, 'medium', 1);

-- 插入符号配置数据
INSERT INTO symbol_config (game_id, symbol_id, symbol_name, symbol_type, sort_order, is_active) VALUES
('game_001', 'cherry', 'cherry', 'normal', 1, 1),
('game_001', 'lemon', 'lemon', 'normal', 2, 1),
('game_001', 'orange', 'orange', 'normal', 3, 1),
('game_001', 'plum', 'plum', 'normal', 4, 1),
('game_001', 'grape', 'grape', 'normal', 5, 1),
('game_001', 'watermelon', 'watermelon', 'normal', 6, 1),
('game_001', 'bell', 'bell', 'normal', 7, 1),
('game_001', 'seven', 'seven', 'normal', 8, 1),
('game_001', 'wild', 'wild', 'wild', 9, 1),
('game_001', 'scatter', 'scatter', 'scatter', 10, 1);

-- 插入符号赔付倍数数据
INSERT INTO symbol_multiplier (symbol_config_id, match_count, multiplier, is_bet_line) VALUES
(1, 2, 3, 0), (1, 3, 10, 1),
(2, 2, 5, 0), (2, 3, 15, 1),
(3, 2, 8, 0), (3, 3, 20, 1),
(4, 2, 10, 0), (4, 3, 25, 1),
(5, 2, 12, 0), (5, 3, 30, 1),
(6, 2, 15, 0), (6, 3, 40, 1),
(7, 2, 20, 0), (7, 3, 50, 1),
(8, 2, 40, 0), (8, 3, 100, 1),
(9, 2, 80, 0), (9, 3, 200, 1),
(10, 3, 50, 0);

-- 插入符号特殊属性数据
INSERT INTO symbol_special_property (symbol_config_id, property_name, property_value) VALUES
(10, 'free_spins', '10'),
(10, 'any_position', 'true'),
(9, 'substitute', 'normal');

-- 插入卷轴配置数据
INSERT INTO reel_config (game_id, reel_index, reel_name) VALUES
('game_001', 1, 'reel1'),
('game_001', 2, 'reel2'),
('game_001', 3, 'reel3');

-- 插入卷轴符号权重数据
-- Reel 1 权重配置
INSERT INTO reel_symbol_weight (reel_config_id, symbol_config_id, weight) VALUES
(1, 1, 25), (1, 2, 20), (1, 3, 18), (1, 4, 15), (1, 5, 12),
(1, 6, 8), (1, 7, 6), (1, 8, 4), (1, 9, 2), (1, 10, 1);

-- Reel 2 权重配置
INSERT INTO reel_symbol_weight (reel_config_id, symbol_config_id, weight) VALUES
(2, 1, 25), (2, 2, 20), (2, 3, 18), (2, 4, 15), (2, 5, 12),
(2, 6, 8), (2, 7, 6), (2, 8, 4), (2, 9, 2), (2, 10, 1);

-- Reel 3 权重配置
INSERT INTO reel_symbol_weight (reel_config_id, symbol_config_id, weight) VALUES
(3, 1, 25), (3, 2, 20), (3, 3, 18), (3, 4, 15), (3, 5, 12),
(3, 6, 8), (3, 7, 6), (3, 8, 4), (3, 9, 2), (3, 10, 1);

-- 插入赔付表配置数据
INSERT INTO pay_table_config (game_id, pay_line_count, pay_line_pattern) VALUES
('game_001', 20, '[[0,0],[0,1],[0,2]]');

-- 插入示例用户数据
INSERT INTO user_info (user_id, integrator_id, user_name, nickname, email, password_hash, rtp_tolerance_threshold, balance, status) VALUES
('test_user_001', 'test_integrator_001', 'Test User', 'Player001', 'testuser@example.com', NULL, 0.00, 1000.00, 1);

-- 插入更多用户测试数据（不同RTP容忍阈值）
INSERT INTO user_info (user_id, integrator_id, user_name, nickname, email, password_hash, rtp_tolerance_threshold, balance, status) VALUES
('platform_user_001', 'platform_self_001', 'Platform User One', 'VIPPlayer1', 'platform1@example.com', NULL, 0.00, 5000.00, 1),
('strict_user_001', 'test_integrator_001', 'Strict User', 'StrictPlayer', 'strict@example.com', NULL, 5.00, 2000.00, 1),
('normal_user_001', 'test_integrator_001', 'Normal User', 'NormalPlayer', 'normal@example.com', NULL, 10.00, 1500.00, 1);

-- 插入用户财务统计数据
INSERT INTO user_finance_stats (user_id, total_bet, total_win, total_games, net_result, rtp_value, max_single_win) VALUES
('test_user_001', 0.00, 0.00, 0, 0.00, 0.00, 0.00),
('platform_user_001', 0.00, 0.00, 0, 0.00, 0.00, 0.00),
('strict_user_001', 0.00, 0.00, 0, 0.00, 0.00, 0.00),
('normal_user_001', 0.00, 0.00, 0, 0.00, 0.00, 0.00);

-- ============================================
-- 创建视图优化查询
-- ============================================

-- 创建用户总览视图
CREATE OR REPLACE VIEW user_overview AS
SELECT 
    u.user_id,
    u.integrator_id,
    i.integrator_name,
    u.user_name,
    u.nickname,
    u.email,
    u.rtp_tolerance_threshold,
    u.balance,
    f.total_bet,
    f.total_win,
    f.rtp_value as user_rtp,
    f.total_games,
    f.net_result,
    f.max_single_win,
    u.vip_level,
    u.status,
    u.last_login_time,
    u.create_time
FROM user_info u
LEFT JOIN integrator_config i ON u.integrator_id = i.integrator_id
LEFT JOIN user_finance_stats f ON u.user_id = f.user_id;

-- ============================================
-- 创建存储过程
-- ============================================

DELIMITER //

-- 更新用户余额和统计数据的存储过程
DROP PROCEDURE IF EXISTS update_user_after_game //

CREATE PROCEDURE update_user_after_game(
    IN p_user_id VARCHAR(32),
    IN p_bet_amount DECIMAL(18,2),
    IN p_win_amount DECIMAL(18,2),
    OUT p_new_balance DECIMAL(18,2)
)
BEGIN
    DECLARE v_balance DECIMAL(18,2);
    DECLARE v_total_bet DECIMAL(18,2);
    DECLARE v_total_win DECIMAL(18,2);
    DECLARE v_total_games BIGINT(20);
    DECLARE v_max_single_win DECIMAL(18,2);
    
    -- 从user_info表获取当前用户余额（独立行级锁，优化并发性能）
    SELECT balance INTO v_balance
    FROM user_info 
    WHERE user_id = p_user_id
    FOR UPDATE;
    
    -- 从user_finance_stats表获取当前用户财务统计数据
    SELECT total_bet, total_win, total_games, max_single_win 
    INTO v_total_bet, v_total_win, v_total_games, v_max_single_win
    FROM user_finance_stats 
    WHERE user_id = p_user_id;
    
    -- 计算新的最大单次赔付
    IF p_win_amount > v_max_single_win THEN
        SET v_max_single_win = p_win_amount;
    END IF;
    
    -- 更新用户余额（在user_info表中）
    UPDATE user_info 
    SET balance = balance + (p_win_amount - p_bet_amount)
    WHERE user_id = p_user_id;
    
    -- 更新用户财务统计数据（在user_finance_stats表中）
    UPDATE user_finance_stats 
    SET 
        total_bet = v_total_bet + p_bet_amount,
        total_win = v_total_win + p_win_amount,
        total_games = v_total_games + 1,
        net_result = (v_total_win + p_win_amount) - (v_total_bet + p_bet_amount),
        rtp_value = CASE WHEN (v_total_bet + p_bet_amount) > 0 
                        THEN ((v_total_win + p_win_amount) / (v_total_bet + p_bet_amount)) * 100 
                        ELSE 0 END,
        max_single_win = v_max_single_win
    WHERE user_id = p_user_id;
    
    -- 获取更新后的余额
    SELECT balance INTO p_new_balance FROM user_info WHERE user_id = p_user_id;
    
END //

-- 游戏配置版本管理的存储过程
DROP PROCEDURE IF EXISTS create_game_config_version //

CREATE PROCEDURE create_game_config_version(
    IN p_game_id VARCHAR(32),
    IN p_version INT,
    IN p_change_reason VARCHAR(256),
    IN p_operator VARCHAR(32)
)
BEGIN
    INSERT INTO game_config_version (game_id, version, config_snapshot, change_reason, operator)
    SELECT 
        p_game_id,
        p_version,
        JSON_OBJECT(
            'game_name', game_name,
            'game_type', game_type,
            'reel_count', reel_count,
            'symbol_count', symbol_count,
            'min_bet', min_bet,
            'max_bet', max_bet,
            'min_lines', min_lines,
            'max_lines', max_lines,
            'rtp_target', rtp_target,
            'volatility_level', volatility_level,
            'special_features', special_features,
            'symbols', (
                SELECT JSON_ARRAYAGG(
                    JSON_OBJECT(
                        'symbol_id', s.symbol_id,
                        'symbol_name', s.symbol_name,
                        'symbol_type', s.symbol_type,
                        'multipliers', (
                            SELECT JSON_ARRAYAGG(
                                JSON_OBJECT(
                                    'match_count', m.match_count,
                                    'multiplier', m.multiplier,
                                    'is_bet_line', m.is_bet_line
                                )
                            )
                            FROM symbol_multiplier m
                            WHERE m.symbol_config_id = s.id
                        ),
                        'properties', (
                            SELECT JSON_ARRAYAGG(
                                JSON_OBJECT(
                                    'property_name', sp.property_name,
                                    'property_value', sp.property_value
                                )
                            )
                            FROM symbol_special_property sp
                            WHERE sp.symbol_config_id = s.id
                        )
                    )
                )
                FROM symbol_config s
                WHERE s.game_id = p_game_id
            ),
            'reels', (
                SELECT JSON_ARRAYAGG(
                    JSON_OBJECT(
                        'reel_index', r.reel_index,
                        'reel_name', r.reel_name,
                        'weights', (
                            SELECT JSON_ARRAYAGG(
                                JSON_OBJECT(
                                    'symbol_id', s2.symbol_id,
                                    'weight', rsw.weight
                                )
                            )
                            FROM reel_symbol_weight rsw
                            JOIN symbol_config s2 ON rsw.symbol_config_id = s2.id
                            WHERE rsw.reel_config_id = r.id
                        )
                    )
                )
                FROM reel_config r
                WHERE r.game_id = p_game_id
            ),
            'pay_table', (
                SELECT JSON_OBJECT(
                    'pay_line_count', pt.pay_line_count,
                    'pay_line_pattern', pt.pay_line_pattern
                )
                FROM pay_table_config pt
                WHERE pt.game_id = p_game_id
            )
        ),
        p_change_reason,
        p_operator
    FROM game_config 
    WHERE game_id = p_game_id;
    
END //

DELIMITER ;

-- ============================================
-- 创建触发器
-- ============================================

-- 游戏配置更新时自动创建版本记录
DELIMITER //

DROP TRIGGER IF EXISTS game_config_version_trigger //

CREATE TRIGGER game_config_version_trigger
AFTER UPDATE ON game_config
FOR EACH ROW
BEGIN
    IF OLD.version != NEW.version THEN
        INSERT INTO game_config_version (game_id, version, config_snapshot, change_reason, operator)
        VALUES (
            NEW.game_id,
            NEW.version,
            JSON_OBJECT(
                'game_name', NEW.game_name,
                'game_type', NEW.game_type,
                'reel_count', NEW.reel_count,
                'symbol_count', NEW.symbol_count,
                'min_bet', NEW.min_bet,
                'max_bet', NEW.max_bet,
                'min_lines', NEW.min_lines,
                'max_lines', NEW.max_lines,
                'rtp_target', NEW.rtp_target,
                'volatility_level', NEW.volatility_level,
                'special_features', NEW.special_features,
                'note', '详细配置请查看相关关联表：symbol_config, symbol_multiplier, symbol_special_property, reel_config, reel_symbol_weight, pay_table_config'
            ),
            '配置更新',
            CURRENT_USER()
        );
    END IF;
END //
DELIMITER ;

-- 重新启用外键检查
SET FOREIGN_KEY_CHECKS = 1;

-- ============================================
-- 完成脚本执行
-- ============================================
SELECT 'MySQL数据库初始化脚本执行完成！' AS status;
SELECT '数据库：slot_game_service' AS database_info;
SELECT '表数量：12个（规范化表结构：1个游戏配置主表 + 6个关联表 + 5个业务表）' AS table_count;
SELECT '视图数量：1个' AS view_count;
SELECT '存储过程数量：2个' AS procedure_count;
SELECT '触发器数量：1个' AS trigger_count;
SELECT '数据库版本：v2.0.2（修复存储过程和触发器重复创建问题）' AS version_info;