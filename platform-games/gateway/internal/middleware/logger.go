package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type LoggerMiddleware struct {
	logger *zap.Logger
}

func NewLoggerMiddleware(logger *zap.Logger) *LoggerMiddleware {
	return &LoggerMiddleware{logger: logger}
}

func (m *LoggerMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		ip := c.ClientIP()
		userAgent := c.Request.UserAgent()
		requestID := c.Writer.Header().Get("X-Request-ID")

		fields := []zap.Field{
			zap.String("request_id", requestID),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", ip),
			zap.String("user_agent", userAgent),
			zap.Int("status", status),
			zap.Duration("latency", latency),
		}

		if len(c.Errors) > 0 {
			fields = append(fields, zap.Any("errors", c.Errors.String()))
		}

		m.logger.Info("HTTP Request", fields...)
	}
}
