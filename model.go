package web

import (
	"context"

	"github.com/go-pay/limiter"
	"github.com/go-pay/xtime"
)

const (
	_HookShutdown hookType = "shutdown"
	_HookExit     hookType = "exit"
)

type hookType string

type HookFunc func(c context.Context)

type Config struct {
	Addr         string          `json:"addr" yaml:"addr" toml:"addr"`                            // addr, default :2233
	ReadTimeout  xtime.Duration  `json:"read_timeout" yaml:"read_timeout" toml:"read_timeout"`    // read_timeout, default 60s
	WriteTimeout xtime.Duration  `json:"write_timeout" yaml:"write_timeout" toml:"write_timeout"` // write_timeout, default 60s
	Debug        bool            `json:"debug" yaml:"debug" toml:"debug"`                         // is show log
	Limiter      *limiter.Config `json:"limiter" yaml:"limiter" toml:"limiter"`                   // interface limit
}

type CommonRsp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type HttpRsp[V any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    V      `json:"data,omitempty"`
}
