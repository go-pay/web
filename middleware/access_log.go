package middleware

import (
	"bytes"
	"encoding/json"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-pay/web/metadata"
)

type CommonRsp struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

type OutputLog struct {
	// common
	AppName string `json:"app_name"`
	CostMs  int64  `json:"cost_ms"`
	Ts      int64  `json:"ts"`

	// request
	ClientIP  string `json:"client_ip"`
	Method    string `json:"method"`
	Schema    string `json:"schema"`
	Host      string `json:"host"`
	Path      string `json:"path"`
	ReqHeader string `json:"req_header"`
	ReqBody   string `json:"req_body"`

	// response
	StatusCode int    `json:"status_code"`
	ResHeader  string `json:"res_header"`
	ResCode    int    `json:"res_code"`
	ResMsg     string `json:"res_msg"`
	ResBody    string `json:"res_body"`
}

var (
	defaultHeaderKey = []string{"Content-Type", "Content-Length", "Accept", "Origin", "Host", "Connection",
		"Accept-Encoding", "Accept-Language", "User-Agent", "Referer", "Cookie", "Authorization",
		"X-Real-IP", "X-Forwarded-For", "X-Forwarded-Proto", "X-Forwarded-Host", "X-Forwarded-Port",
		"X-Forwarded-Server", "X-Forwarded-For-Original"}
	defaultHeaderKeyMap = map[string]int{
		"Content-Type":             1,
		"Content-Length":           1,
		"Accept":                   1,
		"Origin":                   1,
		"Host":                     1,
		"Connection":               1,
		"Accept-Encoding":          1,
		"Accept-Language":          1,
		"User-Agent":               1,
		"Referer":                  1,
		"Cookie":                   1,
		"Authorization":            1,
		"X-Real-IP":                1,
		"X-Forwarded-For":          1,
		"X-Forwarded-Proto":        1,
		"X-Forwarded-Host":         1,
		"X-Forwarded-Port":         1,
		"X-Forwarded-Server":       1,
		"X-Forwarded-For-Original": 1,
	}
)

// AccessLog middleware for request and response body
func AccessLog(appName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			st        = time.Now()
			rHost     = c.Request.Host
			rUri      = c.Request.RequestURI
			rMethod   = c.Request.Method
			rHeader   = c.Request.Header
			rClientIP = metadata.ClientIP(c.Request, rHeader)
			reqHead   = map[string]string{}
			resHead   = map[string]string{}
			schema    = "http"
		)
		if c.Request.TLS != nil {
			schema = "https"
		}
		reqBs, err := metadata.RequestBody(c.Request)
		if err != nil {
			return
		}
		writer := responseWriter{ResponseWriter: c.Writer, resBs: &bytes.Buffer{}}
		c.Writer = writer
		defer func() {
			if len(defaultHeaderKey) != 0 {
				for _, v := range defaultHeaderKey {
					h := rHeader.Get(v)
					if h == "" && defaultHeaderKeyMap[v] == 1 {
						continue
					}
					reqHead[v] = h
				}
			}

			resHeader := c.Writer.Header()
			for k, v := range resHeader {
				if len(v) == 0 {
					continue
				}
				resHead[k] = v[0]
			}

			rbs := writer.resBs.Bytes()
			rsp := &CommonRsp{}
			_ = json.Unmarshal(rbs, rsp)

			output := &OutputLog{
				AppName:    appName,
				ClientIP:   rClientIP,
				CostMs:     time.Since(st).Milliseconds(),
				Host:       rHost,
				Method:     rMethod,
				Path:       rUri,
				ReqHeader:  marshalString(reqHead),
				ReqBody:    string(reqBs),
				ResHeader:  marshalString(resHead),
				ResCode:    rsp.Code,
				ResMsg:     rsp.Message,
				ResBody:    marshalString(rsp),
				Schema:     schema,
				StatusCode: c.Writer.Status(),
				Ts:         st.Unix(),
			}
			log.Printf("access_log: %s\n\n", marshalString(output))
		}()
		c.Next()
	}
}

func SetAccessLogHeader(headers []string) {
	defaultHeaderKey = make([]string, len(headers))
	copy(defaultHeaderKey, headers)
}

func AddAccessLogHeader(headers []string) {
	defaultHeaderKey = append(defaultHeaderKey, headers...)
}

func marshalString(v any) string {
	bs, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(bs)
}

func marshalBytes(v any) []byte {
	bs, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return bs
}

// 自定义一个结构体，实现 gin.ResponseWriter interface
type responseWriter struct {
	gin.ResponseWriter
	resBs *bytes.Buffer
}

// 重写 Write([]byte) (int, error) 方法
func (w responseWriter) Write(b []byte) (int, error) {
	// 向一个bytes.buffer中写一份数据来为获取body使用
	w.resBs.Write(b)
	// 完成gin.Context.Writer.Write()原有功能
	return w.ResponseWriter.Write(b)
}
