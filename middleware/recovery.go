package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-pay/ecode"
	"github.com/go-pay/web/metadata"
	"github.com/go-pay/xlog"
)

type RecoveryInfo struct {
	Time        string `json:"time"`
	RequestURI  string `json:"request_uri"`
	Body        string `json:"body"`
	RequestInfo string `json:"request_info"`
	Err         any    `json:"error"`
	Stack       string `json:"stack"`
}

// Recovery gin middleware recovery
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			rawReq, body []byte
		)
		body, _ = metadata.RequestBody(c.Request)
		if c.Request != nil {
			rawReq, _ = httputil.DumpRequest(c.Request, true)
		}
		defer func() {
			if err := recover(); err != nil {
				const size = 64 << 10
				stack := make([]byte, size)
				stack = stack[:runtime.Stack(stack, false)]
				bs, _ := json.Marshal(RecoveryInfo{
					Time:        time.Now().Format("2006-01-02 15:04:05.000"),
					RequestURI:  c.Request.Host + c.Request.RequestURI,
					Body:        string(body),
					RequestInfo: string(rawReq),
					Err:         err,
					Stack:       string(stack),
				})
				xlog.Errorf("[GinPanic] %s", string(bs))
				_ = c.AbortWithError(http.StatusInternalServerError, ecode.ServerErr)
			}
		}()
		c.Next()
	}
}
