package middleware

import (
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"acking/pkg/response"
)

// Limiter 基于 IP（未登录）或用户 ID（已登录）的令牌桶限流。
// fillPerSec: 每秒补充令牌数；burst: 桶容量。
// 注意：需要用户维度限流时，把 Auth/OptionalAuth 中间件挂在 Limiter 之前。
func Limiter(fillPerSec float64, burst int) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := fmt.Sprintf("ip:%s", c.ClientIP())
		if uid, role := CurrentUser(c); uid > 0 && role >= 0 {
			key = fmt.Sprintf("uid:%d", uid)
		}
		if !allow(key, fillPerSec, burst) {
			response.FailCode(c, response.CodeRateLimited)
			c.Abort()
			return
		}
		c.Next()
	}
}

// LimiterDual 已登录与未登录使用不同桶参数（如文件上传：登录宽松按 UID，未登录严格按 IP）。
func LimiterDual(authRate float64, authBurst int, guestRate float64, guestBurst int) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := fmt.Sprintf("ip:%s", c.ClientIP())
		fill, burst := guestRate, guestBurst
		if uid, role := CurrentUser(c); uid > 0 && role >= 0 {
			key = fmt.Sprintf("uid:%d", uid)
			fill, burst = authRate, authBurst
		}
		if !allow(key, fill, burst) {
			response.FailCode(c, response.CodeRateLimited)
			c.Abort()
			return
		}
		c.Next()
	}
}

type limiterEntry struct {
	lim      *rate.Limiter
	lastSeen time.Time
}

var (
	limiterMu  sync.Mutex
	limiterMap = make(map[string]*limiterEntry)
)

// allow 取/建令牌桶并尝试消费一枚令牌。
// 缓存键带上限流参数：同一身份访问不同接口时各自的桶独立，
// 避免"先到的接口决定桶参数"（如浏览接口的宽松桶被上传接口继承）。
func allow(key string, fillPerSec float64, burst int) bool {
	limiterMu.Lock()
	defer limiterMu.Unlock()
	cacheKey := fmt.Sprintf("%s|%v/%d", key, fillPerSec, burst)
	e, ok := limiterMap[cacheKey]
	if !ok {
		e = &limiterEntry{lim: rate.NewLimiter(rate.Limit(fillPerSec), burst)}
		limiterMap[cacheKey] = e
	}
	e.lastSeen = time.Now()
	return e.lim.Allow()
}

func init() {
	go func() {
		for range time.Tick(5 * time.Minute) {
			limiterMu.Lock()
			for k, e := range limiterMap {
				if time.Since(e.lastSeen) > 10*time.Minute {
					delete(limiterMap, k)
				}
			}
			limiterMu.Unlock()
		}
	}()
}
