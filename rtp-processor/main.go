package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/rtp-processor/config"
	"github.com/rtp-processor/processor"
)

var (
	nacosAddr = flag.String("nacos_addr", "127.0.0.1", "Nacos server address")
	namespace = flag.String("namespace", "public", "Nacos namespace")
	group     = flag.String("group", "DEFAULT_GROUP", "Nacos group")
	dataId    = flag.String("data_id", "app-config", "Nacos data id")
)

/**
 * @brief 主函数
 * @note 加载配置，连接 ClickHouse，启动批处理服务，等待信号量，优雅停机
 */
func main() {
	flag.Parse()

	// 加载配置
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 连接 ClickHouse
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{cfg.ClickHouseAddr()},
		Auth: clickhouse.Auth{
			Database: cfg.ClickHouse.Database,
			Username: cfg.ClickHouse.Username,
			Password: cfg.ClickHouse.Password,
		},
		DialTimeout: 10 * time.Second,
	})

	// 失败时退出
	if err != nil {
		log.Fatalf("连接 ClickHouse 失败: %v", err)
	}
	defer conn.Close()

	// 测试连接
	if err := conn.Ping(context.Background()); err != nil {
		log.Fatalf("Ping ClickHouse 失败: %v", err)
	}

	log.Printf("ClickHouse 连接成功: %s", cfg.ClickHouseAddr())

	// 启动批处理服务
	svc := processor.NewBatchService(conn, cfg)

	log.Printf("RTP 批处理服务启动")
	log.Printf("用户聚合间隔: %v, 游戏聚合间隔: %v, 告警间隔: %v", cfg.GetAggregateUserInterval(), cfg.GetAggregateGameInterval(), cfg.GetAlertInterval())

	svc.Start()

	gracefullShutdown(svc)
}

/**
 * @brief 优雅停机
 * @param svc 批处理服务
 */
func gracefullShutdown(svc *processor.BatchService) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭 RTP 批处理服务...")
	svc.Stop()
	log.Println("服务已关闭")
}

/**
 * @brief 加载配置
 * @return *config.Config 配置
 * @note 从 Nacos 加载配置
 */
func loadConfig() (*config.Config, error) {
	cfg, err := config.LoadConfigFromNacos(
		*nacosAddr,
		*namespace,
		*group,
		*dataId,
	)
	if err != nil {
		return nil, err
	}
	log.Printf("从 Nacos 加载配置成功")
	return cfg, nil
}
