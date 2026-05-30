package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/platform-games/gateway/pkg/response"
	"golang.org/x/time/rate"
)

type RateLimiter struct {
	limiters sync.Map
	rate     rate.Limit
	burst    int
}

type limiterInfo struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func NewRateLimiter(rps int, burst int) *RateLimiter {
	return &RateLimiter{
		rate:  rate.Limit(rps),
		burst: burst,
	}
}

func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	if li, ok := rl.limiters.Load(key); ok {
		info := li.(*limiterInfo)
		info.lastSeen = time.Now()
		return info.limiter
	}

	limiter := rate.NewLimiter(rl.rate, rl.burst)
	rl.limiters.Store(key, &limiterInfo{
		limiter:  limiter,
		lastSeen: time.Now(),
	})

	return limiter
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	go rl.cleanup()

	return func(c *gin.Context) {
		key := c.ClientIP()

		if !rl.getLimiter(key).Allow() {
			response.TooManyRequests(c, "Rate limit exceeded")
			c.Abort()
			return
		}

		c.Next()
	}
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.limiters.Range(func(key, value interface{}) bool {
			info := value.(*limiterInfo)
			if time.Since(info.lastSeen) > 5*time.Minute {
				rl.limiters.Delete(key)
			}
			return true
		})
	}
}

func (rl *RateLimiter) MerchantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		merchantID := c.GetHeader("X-Merchant-ID")
		if merchantID == "" {
			c.Next()
			return
		}

		key := "merchant:" + merchantID

		if !rl.getLimiter(key).Allow() {
			response.TooManyRequests(c, "Merchant rate limit exceeded")
			c.Abort()
			return
		}

		c.Next()
	}
}
