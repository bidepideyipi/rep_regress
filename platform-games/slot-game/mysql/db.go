package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// DBConfig MySQL数据库配置
type DBConfig struct {
	Host         string
	Port         int
	Username     string
	Password     string
	Database     string
	Charset      string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  time.Duration
}

// DB 数据库连接封装
type DB struct {
	*sql.DB
	config *DBConfig
	mu     sync.RWMutex
}

var (
	globalDB *DB
	once     sync.Once
)

// NewDB 创建数据库连接
func NewDB(config *DBConfig) (*DB, error) {
	if config == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	// 设置默认值
	if config.Charset == "" {
		config.Charset = "utf8mb4"
	}
	if config.MaxOpenConns == 0 {
		config.MaxOpenConns = 100
	}
	if config.MaxIdleConns == 0 {
		config.MaxIdleConns = 20
	}
	if config.MaxLifetime == 0 {
		config.MaxLifetime = time.Hour
	}

	// 构建DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true&loc=Local",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
		config.Charset,
	)

	// 创建连接
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("打开数据库连接失败: %w", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.MaxLifetime)

	log.Printf("MySQL连接成功: %s:%d/%s (用户: %s)",
		config.Host, config.Port, config.Database, config.Username)

	return &DB{
		DB:     db,
		config: config,
	}, nil
}

// InitDB 初始化全局数据库连接
func InitDB(config *DBConfig) error {
	var initErr error
	once.Do(func() {
		db, err := NewDB(config)
		if err != nil {
			initErr = err
			return
		}
		globalDB = db
	})
	return initErr
}

// GetDB 获取全局数据库连接
func GetDB() *DB {
	return globalDB
}

// Close 关闭数据库连接
func (db *DB) Close() error {
	if db != nil && db.DB != nil {
		return db.DB.Close()
	}
	return nil
}

// CloseGlobal 关闭全局数据库连接
func CloseGlobal() error {
	if globalDB != nil {
		return globalDB.Close()
	}
	return nil
}

// BeginTx 开始事务
func (db *DB) BeginTx() (*sql.Tx, error) {
	return db.DB.Begin()
}

// IsHealthy 检查数据库健康状态
func (db *DB) IsHealthy() bool {
	if db == nil || db.DB == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return db.PingContext(ctx) == nil
}

// Stats 获取连接池状态
func (db *DB) Stats() sql.DBStats {
	return db.DB.Stats()
}

// GetConfig 获取配置
func (db *DB) GetConfig() *DBConfig {
	return db.config
}
