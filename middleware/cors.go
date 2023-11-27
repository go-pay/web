package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS gin middleware cors
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin") // 请求头部
		if origin == "" {
			origin = c.Request.Host
		}
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			// 允许跨域返回的Header
			c.Header("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, Content-Length, X-CSRF-Token, Token, Session, Origin, Host, Connection, Accept-Encoding, Accept-Language, X-Requested-With")
			// 允许的方法
			c.Header("Access-Control-Allow-Methods", "POST, PUT ,GET, OPTIONS, DELETE, HEAD, TRACE, UPDATE")
			// 允许客户端解析的Header
			c.Header("Access-Control-Expose-Headers", "Authorization, Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
			// 缓存时间
			c.Header("Access-Control-Max-Age", "172800")
			// 允许客户端传递校验信息，cookie
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Request.Header.Del("Origin")
		c.Next()
	}
}
