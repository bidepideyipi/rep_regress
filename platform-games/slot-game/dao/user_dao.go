package dao

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"platform-games/slot-game/models"

	"github.com/google/uuid"
)

// UserDAO 用户数据访问对象
type UserDAO struct {
	db *sql.DB
}

// NewUserDAO 创建UserDAO
func NewUserDAO(db *sql.DB) *UserDAO {
	return &UserDAO{db: db}
}

// GetUserByIntegratorAndID 根据集成商ID和用户ID获取用户信息
func (dao *UserDAO) GetUserByIntegratorAndID(ctx context.Context, integratorID, userID string) (*models.UserInfo, error) {
	query := `
		SELECT user_id, integrator_id, user_name, nickname, email, phone,
		       rtp_tolerance_threshold, balance, vip_level, status,
		       last_login_time, last_login_ip, create_time, update_time
		FROM user_info
		WHERE integrator_id = ? AND user_id = ?
	`

	user := &models.UserInfo{}
	err := dao.db.QueryRowContext(ctx, query, integratorID, userID).Scan(
		&user.UserID, &user.IntegratorID, &user.UserName, &user.Nickname,
		&user.Email, &user.Phone, &user.RTPtoleranceThreshold, &user.Balance,
		&user.VIPLevel, &user.Status, &user.LastLoginTime, &user.LastLoginIP,
		&user.CreateTime, &user.UpdateTime,
	)

	if err == sql.ErrNoRows {
		return nil, nil // 用户不存在
	}
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	return user, nil
}

// GetUserForUpdate 根据集成商ID和用户ID获取用户信息（加行锁）
func (dao *UserDAO) GetUserForUpdate(ctx context.Context, integratorID, userID string) (*models.UserInfo, error) {
	query := `
		SELECT user_id, integrator_id, user_name, nickname, email, phone,
		       rtp_tolerance_threshold, balance, vip_level, status,
		       last_login_time, last_login_ip, create_time, update_time
		FROM user_info
		WHERE integrator_id = ? AND user_id = ?
		FOR UPDATE
	`

	user := &models.UserInfo{}
	err := dao.db.QueryRowContext(ctx, query, integratorID, userID).Scan(
		&user.UserID, &user.IntegratorID, &user.UserName, &user.Nickname,
		&user.Email, &user.Phone, &user.RTPtoleranceThreshold, &user.Balance,
		&user.VIPLevel, &user.Status, &user.LastLoginTime, &user.LastLoginIP,
		&user.CreateTime, &user.UpdateTime,
	)

	if err == sql.ErrNoRows {
		return nil, nil // 用户不存在
	}
	if err != nil {
		return nil, fmt.Errorf("查询用户失败(加锁): %w", err)
	}

	return user, nil
}

// UpdateBalance 更新用户余额
func (dao *UserDAO) UpdateBalance(ctx context.Context, userID string, newBalance float64) error {
	query := `UPDATE user_info SET balance = ?, update_time = NOW() WHERE user_id = ?`
	result, err := dao.db.ExecContext(ctx, query, newBalance, userID)
	if err != nil {
		return fmt.Errorf("更新余额失败: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("用户不存在: %s", userID)
	}

	return nil
}

// DeductBalance 扣减用户余额（使用select for update）
func (dao *UserDAO) DeductBalance(ctx context.Context, integratorID, userID string, amount float64) (*models.UserInfo, error) {
	tx, err := dao.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("开启事务失败: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 1. 查询用户并加锁
	user, err := dao.getUserForUpdateTx(ctx, tx, integratorID, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("用户不存在: integrator=%s, user=%s", integratorID, userID)
	}

	// 2. 检查用户状态
	if user.Status != models.UserStatusEnabled {
		return nil, fmt.Errorf("用户状态异常: status=%d", user.Status)
	}

	// 3. 检查余额是否足够
	if user.Balance < amount {
		return nil, fmt.Errorf("余额不足: balance=%.2f, amount=%.2f", user.Balance, amount)
	}

	// 4. 扣减余额
	newBalance := user.Balance - amount
	if err := dao.updateBalanceTx(ctx, tx, userID, newBalance); err != nil {
		return nil, err
	}

	// 5. 提交事务
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	// 更新用户信息中的余额
	user.Balance = newBalance

	log.Printf("[余额扣减] 成功: integrator=%s, user=%s, 扣减=%.2f, 剩余=%.2f",
		integratorID, userID, amount, newBalance)

	return user, nil
}

// AddBalance 增加用户余额（使用select for update）
func (dao *UserDAO) AddBalance(ctx context.Context, integratorID, userID string, amount float64) (*models.UserInfo, error) {
	tx, err := dao.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("开启事务失败: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 1. 查询用户并加锁
	user, err := dao.getUserForUpdateTx(ctx, tx, integratorID, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("用户不存在: integrator=%s, user=%s", integratorID, userID)
	}

	// 2. 检查用户状态
	if user.Status != models.UserStatusEnabled {
		return nil, fmt.Errorf("用户状态异常: status=%d", user.Status)
	}

	// 3. 增加余额
	newBalance := user.Balance + amount
	if err := dao.updateBalanceTx(ctx, tx, userID, newBalance); err != nil {
		return nil, err
	}

	// 4. 提交事务
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	// 更新用户信息中的余额
	user.Balance = newBalance

	log.Printf("[余额增加] 成功: integrator=%s, user=%s, 增加=%.2f, 余额=%.2f",
		integratorID, userID, amount, newBalance)

	return user, nil
}

// CreateTransaction 创建交易记录
func (dao *UserDAO) CreateTransaction(ctx context.Context, tx *sql.Tx, record *models.TransactionRecord) error {
	query := `
		INSERT INTO transaction_records
		(transaction_id, user_id, integrator_id, transaction_type, amount,
		 balance_before, balance_after, game_session_id, reference_id, status, error_message)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := tx.ExecContext(ctx, query,
		record.TransactionID, record.UserID, record.IntegratorID,
		record.TransactionType, record.Amount, record.BalanceBefore,
		record.BalanceAfter, record.GameSessionID, record.ReferenceID,
		record.Status, record.ErrorMessage,
	)
	if err != nil {
		return fmt.Errorf("创建交易记录失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	record.ID = id
	record.CreateTime = time.Now()

	return nil
}

// 内部方法：在事务中查询用户（加锁）
func (dao *UserDAO) getUserForUpdateTx(ctx context.Context, tx *sql.Tx, integratorID, userID string) (*models.UserInfo, error) {
	query := `
		SELECT user_id, integrator_id, user_name, nickname, email, phone,
		       rtp_tolerance_threshold, balance, vip_level, status,
		       last_login_time, last_login_ip, create_time, update_time
		FROM user_info
		WHERE integrator_id = ? AND user_id = ?
		FOR UPDATE
	`

	user := &models.UserInfo{}
	err := tx.QueryRowContext(ctx, query, integratorID, userID).Scan(
		&user.UserID, &user.IntegratorID, &user.UserName, &user.Nickname,
		&user.Email, &user.Phone, &user.RTPtoleranceThreshold, &user.Balance,
		&user.VIPLevel, &user.Status, &user.LastLoginTime, &user.LastLoginIP,
		&user.CreateTime, &user.UpdateTime,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询用户失败(加锁): %w", err)
	}

	return user, nil
}

// 内部方法：在事务中更新余额
func (dao *UserDAO) updateBalanceTx(ctx context.Context, tx *sql.Tx, userID string, newBalance float64) error {
	query := `UPDATE user_info SET balance = ?, update_time = NOW() WHERE user_id = ?`
	result, err := tx.ExecContext(ctx, query, newBalance, userID)
	if err != nil {
		return fmt.Errorf("更新余额失败: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("用户不存在: %s", userID)
	}

	return nil
}

// GenerateTransactionID 生成交易ID
func GenerateTransactionID() string {
	return fmt.Sprintf("TXN-%s", uuid.New().String())
}
