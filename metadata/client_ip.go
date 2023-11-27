package metadata

import (
	"net"
	"net/http"
	"strings"
)

func ClientIP(r *http.Request, rHeader http.Header) string {
	cIp := rHeader.Get("X-Forwarded-For")
	cIp = strings.TrimSpace(strings.Split(cIp, ",")[0])
	if cIp == "" {
		cIp = strings.TrimSpace(rHeader.Get("X-Real-Ip"))
	}
	if cIp != "" {
		return cIp
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}
	return ""
}
