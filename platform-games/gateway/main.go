package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/platform-games/gateway/config"
	"github.com/platform-games/gateway/internal/router"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	cfg, err := config.Load("./config/config.yaml")
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	logger, err := initLogger(cfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to init logger: %v", err))
	}
	defer logger.Sync()

	logger.Info("Starting gateway service",
		zap.String("host", cfg.Server.Host),
		zap.Int("port", cfg.Server.Port),
		zap.String("mode", cfg.Server.Mode),
	)

	ginEngine := router.Setup(cfg, logger)

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      ginEngine,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		logger.Info("Server started", zap.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}

func initLogger(cfg *config.Config) (*zap.Logger, error) {
	var logger *zap.Logger
	var err error

	if cfg.Log.Format == "json" {
		logger, err = zap.NewProduction()
		if err != nil {
			return nil, err
		}
	} else {
		logger, err = zap.NewDevelopment()
		if err != nil {
			return nil, err
		}
	}

	if cfg.Log.Level != "" {
		level, err := zap.ParseAtomicLevel(cfg.Log.Level)
		if err == nil {
			config := zap.NewProductionConfig()
			config.Level = level
			config.Development = (cfg.Log.Format != "json")
			config.EncoderConfig.TimeKey = "timestamp"
			config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
			logger, err = config.Build(zap.AddCaller())
			if err != nil {
				return nil, err
			}
		}
	}

	return logger, nil
}
