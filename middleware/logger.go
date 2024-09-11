package middleware

import (
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	ignoreTraceLog     bool
	ignoreTraceLogPath = map[string]bool{
		"/":            true,
		"/ping":        true,
		"/health":      true,
		"/healthCheck": true,
	}
)

// Logger
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start time
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()
		if raw != "" {
			path = path + "?" + raw
		}

		// ignore logger output
		if ignoreTraceLog {
			return
		}
		if ignoreTraceLogPath[path] {
			return
		}

		// End time
		end := time.Now()
		fmt.Fprintf(os.Stdout, "[GIN] %s | %3d | %13v | %15s | %-7s %#v\n%s", end.Format("2006/01/02 - 15:04:05"), c.Writer.Status(), end.Sub(start), c.ClientIP(), c.Request.Method, path, c.Errors.ByType(gin.ErrorTypePrivate).String())
	}
}

func SetIgnoreTraceLog(ignore bool) {
	ignoreTraceLog = ignore
}

func AddIgnoreTraceLogPath(path string, ignore bool) {
	if path != "" {
		ignoreTraceLogPath[path] = ignore
	}
}
