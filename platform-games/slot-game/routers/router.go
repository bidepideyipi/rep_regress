package routers

import (
	"platform-games/slot-game/controllers"
	"github.com/gin-gonic/gin"
)

// SetupRouter 设置路由
func SetupRouter(gameController *controllers.GameController) *gin.Engine {
	router := gin.Default()

	// 健康检查
	router.GET("/health", gameController.HealthCheck)
	router.GET("/health/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "pong"})
	})

	// API路由组
	api := router.Group("/api")
	{
		// 游戏相关路由
		game := api.Group("/game")
		{
			game.POST("/spin", gameController.Spin)           // 执行旋转
			game.GET("/config", gameController.GetConfig)      // 获取游戏配置
			game.POST("/config/refresh", gameController.RefreshConfig) // 刷新配置
		}
	}

	return router
}