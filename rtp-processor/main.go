package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/rtp-processor/config"
	"github.com/rtp-processor/consumer"
	"github.com/rtp-processor/db"
	"github.com/rtp-processor/processor"
	"github.com/sirupsen/logrus"
)

var (
	nacosAddr = flag.String("nacos_addr", "127.0.0.1", "Nacos server address")
	namespace = flag.String("namespace", "public", "Nacos namespace")
	group     = flag.String("group", "DEFAULT_GROUP", "Nacos group")
	dataId    = flag.String("data_id", "app-config", "Nacos data id")
)

/**
 * @brief 主函数
 * @note 加载配置，连接 ClickHouse，启动批处理服务和消费者，等待信号量，优雅停机
 */
func main() {
	logrus.SetLevel(logrus.WarnLevel)

	flag.Parse()
	// 加载配置
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 调试：输出 ClickHouse 配置
	passwordSet := "已设置"
	if cfg.ClickHouse.Password == "" {
		passwordSet = "未设置"
	}
	log.Printf("ClickHouse 配置: Host=%s, Port=%d, Database=%s, Username=%s, Password=%s",
		cfg.ClickHouse.Host, cfg.ClickHouse.Port, cfg.ClickHouse.Database, cfg.ClickHouse.Username, passwordSet)

	// 连接 ClickHouse
	chWriter, err := db.NewClickHouseWriter(cfg)
	if err != nil {
		log.Fatalf("连接 ClickHouse 失败: %v", err)
	}
	defer chWriter.Close()

	dbConn := chWriter.GetConn()
	log.Printf("ClickHouse 连接成功")

	// 启动批处理服务
	svc := processor.NewBatchService(dbConn, cfg)

	log.Printf("RTP 批处理服务启动")
	log.Printf("用户聚合间隔: %v, 游戏聚合间隔: %v, 告警间隔: %v", cfg.GetAggregateUserInterval(), cfg.GetAggregateGameInterval(), cfg.GetAlertInterval())

	svc.Start()

	// 启动 RocketMQ Push 消费者
	pushConsumer, err := consumer.NewPushConsumer(cfg)
	if err != nil {
		log.Printf("创建Push消费者失败: %v", err)
	} else {
		if err := pushConsumer.Start(); err != nil {
			log.Printf("启动Push消费者失败: %v", err)
		} else {
			log.Printf("RocketMQ Push消费者启动成功")
		}
	}

	gracefullShutdown(svc, pushConsumer)
}

/**
 * @brief 优雅停机
 * @param svc 批处理服务
 * @param pushConsumer Push消费者（可能为nil）
 */
func gracefullShutdown(svc *processor.BatchService, pushConsumer *consumer.PushConsumer) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭服务...")

	// 先停止消费者
	if pushConsumer != nil {
		log.Println("正在关闭消费者...")
		if err := pushConsumer.Stop(); err != nil {
			log.Printf("关闭消费者失败: %v", err)
		}
	}

	// 再停止批处理服务
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
