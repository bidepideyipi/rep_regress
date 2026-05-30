package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/platform-games/gateway/config"
	"github.com/platform-games/gateway/internal/handlers"
	"go.uber.org/zap"
)

func Setup(cfg *config.Config, logger *zap.Logger) *gin.Engine {
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(LoggerMiddleware(logger))

	authHandler := handlers.NewAuthHandler(logger)
	userHandler := handlers.NewUserHandler(logger)
	gameHandler := handlers.NewGameHandler(logger)
	financeHandler := handlers.NewFinanceHandler(logger)

	v1 := r.Group("/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/game-access", authHandler.GameAccess)
			auth.POST("/verify-token", authHandler.VerifyToken)
		}

		user := v1.Group("/user")
		{
			user.GET("/info", userHandler.GetUserInfo)
			user.PUT("/info", userHandler.UpdateUserInfo)
		}

		game := v1.Group("/game")
		{
			game.POST("/action", gameHandler.GameAction)
			game.POST("/sync", gameHandler.GameSync)
		}

		finance := v1.Group("/finance")
		{
			finance.GET("/balance", financeHandler.GetBalance)
			finance.GET("/transactions", financeHandler.GetTransactions)
		}
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return r
}

func LoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		ip := c.ClientIP()

		logger.Info("HTTP Request",
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", ip),
			zap.Int("status", status),
			zap.Duration("latency", latency),
		)
	}
}
