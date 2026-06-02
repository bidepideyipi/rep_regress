-- MySQL数据库初始化脚本
-- 数据库名称：slot_game_service
-- 版本：v2.2.0
-- 说明：游戏服务模块MySQL数据库表初始化脚本
-- 变更记录：
--   v2.2.0: 删除游戏详细配置相关表，配置已迁移至Nacos；新增nacos_config_id字段
--   v2.1.0: 新增Jackpot相关表（jackpot_config、jackpot_pool、jackpot_win_record），采用混合池模式
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
-- 2.2 游戏配置表 (game_config)
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
    nacos_config_id VARCHAR(64) NOT NULL COMMENT 'Nacos配置ID（关联详细配置）',
    status TINYINT(1) NOT NULL DEFAULT 1 COMMENT '状态（0-禁用，1-启用）',
    version INT(11) NOT NULL DEFAULT 1 COMMENT '配置版本号',
    remark TEXT DEFAULT NULL COMMENT '备注',
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    PRIMARY KEY (game_id),
    KEY idx_status (status),
    KEY idx_game_type (game_type),
    KEY idx_nacos_config (nacos_config_id),
    KEY idx_version (version),
    CONSTRAINT chk_game_status CHECK (status IN (0, 1)),
    CONSTRAINT chk_min_bet CHECK (min_bet > 0),
    CONSTRAINT chk_max_bet CHECK (max_bet >= min_bet),
    CONSTRAINT chk_reel_count CHECK (reel_count BETWEEN 1 AND 10),
    CONSTRAINT chk_symbol_count CHECK (symbol_count BETWEEN 1 AND 20)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='游戏配置表（详细配置在Nacos）';

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
-- 2.7 Jackpot配置表 (jackpot_config)
-- ============================================
DROP TABLE IF EXISTS jackpot_config;

CREATE TABLE jackpot_config (
    game_id VARCHAR(32) NOT NULL COMMENT '游戏ID（"0"表示全局配置）',
    enabled TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否启用（0-否，1-是）',
    contribution_rate DECIMAL(5,4) NOT NULL DEFAULT 0.0100 COMMENT '注入比例（1%）',
    mini_seed DECIMAL(18,2) NOT NULL DEFAULT 100.00 COMMENT 'Mini池种子金额',
    minor_seed DECIMAL(18,2) NOT NULL DEFAULT 500.00 COMMENT 'Minor池种子金额',
    major_seed DECIMAL(18,2) NOT NULL DEFAULT 2000.00 COMMENT 'Major池种子金额',
    grand_seed DECIMAL(18,2) NOT NULL DEFAULT 10000.00 COMMENT 'Grand池种子金额',
    mini_ratio DECIMAL(4,3) NOT NULL DEFAULT 0.300 COMMENT 'Mini池分配比例（30%）',
    minor_ratio DECIMAL(4,3) NOT NULL DEFAULT 0.250 COMMENT 'Minor池分配比例（25%）',
    major_ratio DECIMAL(4,3) NOT NULL DEFAULT 0.250 COMMENT 'Major池分配比例（25%）',
    grand_ratio DECIMAL(4,3) NOT NULL DEFAULT 0.200 COMMENT 'Grand池分配比例（20%）',
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    PRIMARY KEY (game_id),
    KEY idx_enabled (enabled),
    CONSTRAINT chk_jackpot_enabled CHECK (enabled IN (0, 1)),
    CONSTRAINT chk_jackpot_ratios CHECK (mini_ratio + minor_ratio + major_ratio + grand_ratio = 1.0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Jackpot配置表';

-- ============================================
-- 2.8 Jackpot实时池表 (jackpot_pool)
-- ============================================
DROP TABLE IF EXISTS jackpot_pool;

CREATE TABLE jackpot_pool (
    game_id VARCHAR(32) NOT NULL COMMENT '游戏ID（"0"表示全局共享池）',
    pool_type VARCHAR(16) NOT NULL COMMENT '池子类型（mini/minor/major/grand）',
    current_amount DECIMAL(18,2) NOT NULL DEFAULT 0.00 COMMENT '当前金额',
    last_win_time DATETIME DEFAULT NULL COMMENT '最后中奖时间',
    win_count BIGINT(20) NOT NULL DEFAULT 0 COMMENT '中奖次数',
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    PRIMARY KEY (game_id, pool_type),
    KEY idx_pool_type (pool_type),
    KEY idx_last_win_time (last_win_time),
    CONSTRAINT chk_jackpot_pool_type CHECK (pool_type IN ('mini', 'minor', 'major', 'grand')),
    CONSTRAINT chk_jackpot_amount CHECK (current_amount >= 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Jackpot实时池表';

-- ============================================
-- 2.9 Jackpot中奖记录表 (jackpot_win_record)
-- ============================================
DROP TABLE IF EXISTS jackpot_win_record;

CREATE TABLE jackpot_win_record (
    transaction_id VARCHAR(64) NOT NULL COMMENT '交易ID',
    integrator_id VARCHAR(32) NOT NULL COMMENT '集成商ID',
    user_id VARCHAR(32) NOT NULL COMMENT '用户ID',
    game_id VARCHAR(32) NOT NULL COMMENT '游戏ID',
    pool_type VARCHAR(16) NOT NULL COMMENT '中奖池子类型',
    win_amount DECIMAL(18,2) NOT NULL COMMENT '中奖金额',
    pool_amount_before DECIMAL(18,2) NOT NULL COMMENT '中奖前池金额',
    pool_amount_after DECIMAL(18,2) NOT NULL COMMENT '中奖后池金额（重置为种子金额）',
    session_id VARCHAR(64) DEFAULT NULL COMMENT '游戏会话ID',
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',

    PRIMARY KEY (transaction_id),
    KEY idx_user_id (user_id),
    KEY idx_game_id (game_id),
    KEY idx_pool_type (pool_type),
    KEY idx_create_time (create_time),
    CONSTRAINT fk_jackpot_integrator FOREIGN KEY (integrator_id) REFERENCES integrator_config (integrator_id) ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT chk_jackpot_win_pool_type CHECK (pool_type IN ('mini', 'minor', 'major', 'grand')),
    CONSTRAINT chk_jackpot_win_amount CHECK (win_amount > 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Jackpot中奖记录表';

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

-- 插入游戏配置数据（详细配置在Nacos）
INSERT INTO game_config (game_id, game_name, game_type, reel_count, symbol_count, min_bet, max_bet, min_lines, max_lines, rtp_target, volatility_level, nacos_config_id, status) VALUES
('game_001', 'Lucky Fruits', 'slot_3x3', 3, 10, 0.10, 1000.00, 1, 20, 95.00, 'medium', 'nacos_game_config_001', 1);

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
-- 插入Jackpot初始化数据
-- ============================================

-- 插入全局Jackpot配置（game_id="0"表示全局配置）
INSERT INTO jackpot_config (game_id, enabled, contribution_rate, mini_seed, minor_seed, major_seed, grand_seed, mini_ratio, minor_ratio, major_ratio, grand_ratio) VALUES
('0', 1, 0.0100, 100.00, 500.00, 2000.00, 10000.00, 0.300, 0.250, 0.250, 0.200);

-- 插入游戏001的Jackpot配置
INSERT INTO jackpot_config (game_id, enabled, contribution_rate, mini_seed, minor_seed, major_seed, grand_seed, mini_ratio, minor_ratio, major_ratio, grand_ratio) VALUES
('game_001', 1, 0.0100, 100.00, 500.00, 2000.00, 10000.00, 0.300, 0.250, 0.250, 0.200);

-- 插入全局共享Grand池
INSERT INTO jackpot_pool (game_id, pool_type, current_amount, last_win_time, win_count) VALUES
('0', 'grand', 10000.00, NULL, 0);

-- 插入游戏001的Mini/Minor/Major池
INSERT INTO jackpot_pool (game_id, pool_type, current_amount, last_win_time, win_count) VALUES
('game_001', 'mini', 100.00, NULL, 0),
('game_001', 'minor', 500.00, NULL, 0),
('game_001', 'major', 2000.00, NULL, 0);

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
            'nacos_config_id', nacos_config_id,
            'note', '详细配置（符号、卷轴、权重等）请查看Nacos配置中心'
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
                'nacos_config_id', NEW.nacos_config_id,
                'note', '详细配置（符号、卷轴、权重等）请查看Nacos配置中心'
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
SELECT '表数量：9个（游戏详细配置在Nacos管理）' AS table_count;
SELECT '视图数量：1个' AS view_count;
SELECT '存储过程数量：2个' AS procedure_count;
SELECT '触发器数量：1个' AS trigger_count;
SELECT '数据库版本：v2.2.0（删除游戏详细配置表，配置迁移至Nacos）' AS version_info;