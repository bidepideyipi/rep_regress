package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"platform-games/slot-game/config"
	"platform-games/slot-game/consumer"
	"platform-games/slot-game/controllers"
	"platform-games/slot-game/models"
	"platform-games/slot-game/nacos"
	"platform-games/slot-game/rocketmq"
	"platform-games/slot-game/routers"

	"github.com/gin-gonic/gin"
)

const (
	// Nacos中ClickHouse配置
	ClickHouseConfigDataID = "app-config"
	ClickHouseConfigGroup  = "DEFAULT_GROUP"
)

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	flag.Parse()

	// 加载应用配置
	appConfig, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 设置Gin模式
	gin.SetMode(appConfig.Server.Mode)

	// 初始化Nacos配置管理器
	configManager, err := nacos.NewConfigManager(
		appConfig.Nacos.ServerAddr,
		appConfig.Nacos.GameConfig.GameID,
		appConfig.Nacos.Namespace,
		appConfig.Nacos.Group,
		appConfig.Nacos.GameConfig.DataID,
	)
	if err != nil {
		log.Fatalf("创建Nacos配置管理器失败: %v", err)
	}
	defer configManager.Close()

	// 加载游戏配置
	gameConfig, err := configManager.LoadConfig()
	if err != nil {
		log.Printf("从Nacos加载配置失败，使用默认配置: %v", err)
		log.Fatal("游戏配置加载失败")
	}

	// 初始化游戏控制器
	gameController := controllers.NewGameController(configManager)
	gameController.InitializeGame(gameConfig)

	// 从Nacos加载应用配置（用于RocketMQ和ClickHouse）
	nacosAppConfigContent, err := configManager.GetRawConfig(ClickHouseConfigDataID, ClickHouseConfigGroup)
	if err != nil {
		log.Printf("从Nacos加载应用配置失败: %v", err)
	}

	var appConfigForConsumer *consumer.AppConfig
	if nacosAppConfigContent != "" {
		var cfg consumer.AppConfig
		if err := json.Unmarshal([]byte(nacosAppConfigContent), &cfg); err != nil {
			log.Printf("解析Nacos应用配置失败: %v", err)
		} else {
			appConfigForConsumer = &cfg
		}
	}

	if appConfigForConsumer == nil {
		log.Printf("使用默认应用配置")
		appConfigForConsumer = &consumer.AppConfig{
			RocketMQ: consumer.RocketMQConfig{
				NameServers: []string{"127.0.0.1:9876"},
				Producer: consumer.ProducerConfig{
					GroupName: "slot_game_producer_group",
					Topic:     "game_log_topic",
				},
				Consumer: consumer.ConsumerConfig{
					GroupName: "slot_game_consumer_group",
					Topic:     "game_log_topic",
					BatchSize: 100,
				},
			},
			ClickHouse: consumer.ClickHouseConfig{
				Host:     "127.0.0.1",
				Port:     9000,
				Username: "default",
				Password: "",
				Database: "rtp_analytics",
			},
		}
	}

	// 初始化游戏消费者
	gameConsumer, err := consumer.NewGameConsumer(appConfigForConsumer)
	if err != nil {
		log.Printf("初始化游戏消费者失败: %v", err)
	} else {
		if err := gameConsumer.Start(); err != nil {
			log.Printf("启动游戏消费者失败: %v", err)
		}
		defer func() {
			if gameConsumer != nil {
				log.Println("正在关闭游戏消费者...")
				if err := gameConsumer.Stop(); err != nil {
					log.Printf("关闭游戏消费者失败: %v", err)
				}
			}
		}()
	}

	// 初始化RocketMQ生产者
	mqProducer, err := rocketmq.NewProducer()
	if err != nil {
		log.Printf("初始化RocketMQ生产者失败: %v", err)
		// 非致命错误，继续启动服务
	} else {
		gameController.SetMQProducer(mqProducer)
		defer func() {
			log.Println("正在关闭RocketMQ生产者...")
			if err := mqProducer.Shutdown(); err != nil {
				log.Printf("关闭RocketMQ生产者失败: %v", err)
			}
		}()
		log.Printf("RocketMQ生产者初始化成功")
	}

	// 监听配置变化
	err = configManager.WatchConfig(func(newConfig *models.GameConfig) {
		log.Println("检测到配置变更，重新初始化游戏...")
		gameController.InitializeGame(newConfig)
	})
	if err != nil {
		log.Printf("监听配置变化失败: %v", err)
	}

	// 启动自动刷新配置（如果启用）
	if gameConfig.CacheSettings.EnableLocalCache && gameConfig.CacheSettings.RefreshIntervalSeconds > 0 {
		go func() {
			interval := time.Duration(gameConfig.CacheSettings.RefreshIntervalSeconds) * time.Second
			configManager.StartAutoRefresh(interval)
		}()
	}

	// 设置路由
	router := routers.SetupRouter(gameController)

	// 配置HTTP服务器
	server := &http.Server{
		Addr:         config.GetServerAddr(),
		Handler:      router,
		ReadTimeout:  time.Duration(appConfig.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(appConfig.Server.WriteTimeout) * time.Second,
	}

	// 启动服务器
	go func() {
		log.Printf("服务器启动在端口 %d", appConfig.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(appConfig.Server.ShutdownTimeout)*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("服务器强制关闭: %v", err)
	}

	log.Println("服务器已关闭")
}

// getLocalIP 获取本机IP地址
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}
