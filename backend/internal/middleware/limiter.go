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
// 注意：需要用户维度限流时，把 Auth 中间件挂在 Limiter 之前。
func Limiter(fillPerSec float64, burst int) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := fmt.Sprintf("ip:%s", c.ClientIP())
		if uid, role := CurrentUser(c); uid > 0 && role >= 0 {
			key = fmt.Sprintf("uid:%d", uid)
		}
		if !getLimiter(key, fillPerSec, burst).Allow() {
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

// getLimiter 取/建限流器；顺带清理 10 分钟未活动的桶
func getLimiter(key string, fillPerSec float64, burst int) *rate.Limiter {
	limiterMu.Lock()
	defer limiterMu.Unlock()
	if e, ok := limiterMap[key]; ok {
		e.lastSeen = time.Now()
		return e.lim
	}
	lim := rate.NewLimiter(rate.Limit(fillPerSec), burst)
	limiterMap[key] = &limiterEntry{lim: lim, lastSeen: time.Now()}
	return lim
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
