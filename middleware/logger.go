package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-pay/limiter"
)

// Limiter gin middleware limiter
// if rl is nil, default Bucket = 1000, Rate = 1000
func Limiter(appName string, rl *limiter.RateLimiter) gin.HandlerFunc {
	if rl == nil {
		rl = limiter.NewLimiter(nil)
	}
	limitKey := appName
	return func(c *gin.Context) {
		if limitKey == "" {
			limitKey = strings.Split(c.Request.RequestURI, "?")[0][1:]
		}
		// log.Warning("key:", path[1:])
		l := rl.LimiterGroup.Get(limitKey)
		if !l.Allow() {
			c.JSON(http.StatusOK, struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			}{
				Code:    503,
				Message: "服务器忙，请稍后重试...",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
