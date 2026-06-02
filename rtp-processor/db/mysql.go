package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rtp-processor/config"
)

// MySQLManager MySQL连接管理器
type MySQLManager struct {
	db *sql.DB
}

// NewMySQLManager 创建MySQL管理器
func NewMySQLManager(cfg *config.Config) (*MySQLManager, error) {
	if cfg.MySQL.Host == "" || cfg.MySQL.Port == 0 || cfg.MySQL.Database == "" {
		return nil, fmt.Errorf("MySQL配置不完整")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local",
		cfg.MySQL.Username,
		cfg.MySQL.Password,
		cfg.MySQL.Host,
		cfg.MySQL.Port,
		cfg.MySQL.Database)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("连接MySQL失败: %v", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("MySQL Ping失败: %v", err)
	}

	log.Printf("[MySQL] 连接成功: %s:%d/%s", cfg.MySQL.Host, cfg.MySQL.Port, cfg.MySQL.Database)

	return &MySQLManager{db: db}, nil
}

// GetDB 获取数据库连接
func (m *MySQLManager) GetDB() *sql.DB {
	return m.db
}

// Close 关闭连接
func (m *MySQLManager) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}
