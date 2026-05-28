package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"platform-games/slot-game/config"
	"platform-games/slot-game/controllers"
	"platform-games/slot-game/models"
	"platform-games/slot-game/nacos"
	"platform-games/slot-game/routers"
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
		// 这里可以加载默认配置或退出
		log.Fatal("游戏配置加载失败")
	}

	// 初始化游戏控制器
	gameController := controllers.NewGameController(configManager)
	gameController.InitializeGame(gameConfig)

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