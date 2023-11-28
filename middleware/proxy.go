package middleware

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-pay/ecode"
)

var (
	httpCli = &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: defaultTransportDialContext(&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}),
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			DisableKeepAlives:     true,
			ForceAttemptHTTP2:     true,
		},
	}
)

type HttpRsp[V any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    V      `json:"data,omitempty"`
}

// GinProxy gin request proxy and get rsp
func GinProxy[Rsp any](c *gin.Context, method, host, uri string) (rspParam Rsp, err error) {
	var (
		//req     *http.Request
		//reader  *strings.Reader
		rMethod = c.Request.Method
		rHeader = c.Request.Header
		rUri    = c.Request.RequestURI
		//pa      = c.Request.Form.Encode()
		//rBody   = c.Request.Body
	)
	vo := reflect.ValueOf(rspParam)
	if vo.Kind() != reflect.Ptr {
		err = ecode.New(500, "", "rspParam must be point kind")
		return
	}
	if method != "" {
		rMethod = strings.ToUpper(method)
	}
	if uri != "" {
		rUri = uri
	}
	uri = host + rUri
	// Request
	req, e := http.NewRequestWithContext(c, rMethod, uri, c.Request.Body)
	if err != nil {
		err = e
		return
	}
	// Request Header
	req.Header = rHeader
	// Do
	resp, e := httpCli.Do(req)
	if e != nil {
		err = e
		return
	}
	defer resp.Body.Close()
	rspBytes, e := io.ReadAll(resp.Body)
	if e != nil {
		err = e
		return
	}
	if resp.StatusCode != 200 {
		err = ecode.New(resp.StatusCode, "", string(rspBytes))
		return
	}
	res := &HttpRsp[Rsp]{}
	if err = json.Unmarshal(rspBytes, res); err != nil {
		return
	}
	return res.Data, nil
}

// GinPureProxy gin request proxy
func GinPureProxy(c *gin.Context, host string) {
	var (
		w       = c.Writer
		r       = c.Request
		rMethod = r.Method
		rUri    = r.RequestURI
	)
	uri := host + rUri
	// Request
	req, err := http.NewRequestWithContext(c, rMethod, uri, c.Request.Body)
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}
	// Request Header
	req.Header = c.Request.Header
	// Do
	resp, err := httpCli.Do(req)
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	// Response Header
	for k, vs := range resp.Header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
