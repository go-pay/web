package web

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-pay/limiter"
	"github.com/go-pay/web/middleware"
	"github.com/go-pay/xlog"
	"github.com/go-pay/xtime"
)

type GinEngine struct {
	server           *http.Server
	Gin              *gin.Engine
	timeout          time.Duration
	wg               sync.WaitGroup
	addrPort         string
	IgnoreReleaseLog bool
	hookMaps         map[hookType][]func(c context.Context)
}

func InitGin(c *Config) *GinEngine {
	if c == nil {
		c = &Config{Addr: ":2233"}
	}
	g := gin.New()
	engine := &GinEngine{Gin: g, wg: sync.WaitGroup{}, addrPort: c.Addr, hookMaps: make(map[hookType][]func(c context.Context))}

	if c.ReadTimeout == 0 {
		c.ReadTimeout = xtime.Duration(60 * time.Second)
	}
	if c.WriteTimeout == 0 {
		c.WriteTimeout = xtime.Duration(60 * time.Second)
	}
	engine.timeout = time.Duration(c.ReadTimeout)
	engine.server = &http.Server{
		Addr:         engine.addrPort,
		Handler:      g,
		ReadTimeout:  time.Duration(c.ReadTimeout),
		WriteTimeout: time.Duration(c.WriteTimeout),
	}
	g.Use(engine.Logger(false), middleware.Recovery())
	if c.Limiter != nil && c.Limiter.Rate != 0 {
		g.Use(middleware.Limiter("", limiter.NewLimiter(c.Limiter)))
	}
	if !c.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	return engine
}

// 添加 GinServer 服务关闭时的钩子函数
func (g *GinEngine) AddShutdownHook(hooks ...HookFunc) *GinEngine {
	for _, fn := range hooks {
		if fn != nil {
			g.hookMaps[_HookShutdown] = append(g.hookMaps[_HookShutdown], fn)
		}
	}
	return g
}

// 添加 GinServer 进程退出时钩子函数
func (g *GinEngine) AddExitHook(hooks ...HookFunc) *GinEngine {
	for _, fn := range hooks {
		if fn != nil {
			g.hookMaps[_HookExit] = append(g.hookMaps[_HookExit], fn)
		}
	}
	return g
}

func (g *GinEngine) Start() {
	// monitoring signal
	go g.goNotifySignal()

	// start gin http server
	xlog.Warnf("Listening and serving HTTP on %s", g.addrPort)
	if err := g.server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			panic(fmt.Sprintf("server.ListenAndServe(), error(%+v).", err))
		}
		xlog.Warn("http: Server closed")
	}
	xlog.Color(xlog.Green).Warnf("wait for process working finished")
	// wait for process finished
	g.wg.Wait()
	xlog.Color(xlog.Green).Warnf("process exit")
}

// 监听信号
func (g *GinEngine) goNotifySignal() {
	g.wg.Add(1)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	for {
		si := <-ch
		switch si {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			xlog.Color(xlog.Yellow).Warnf("get a signal %s, stop the process", si.String())
			// close gin http server
			g.Close()
			ctx, cancelFunc := context.WithTimeout(context.Background(), g.timeout)
			// call before close hooks
			go func() {
				if a := recover(); a != nil {
					xlog.Errorf("panic: %v", a)
				}
				for _, fn := range g.hookMaps[_HookShutdown] {
					fn(ctx)
				}
			}()
			// wait for program finish processing
			xlog.Warnf("waiting for the process to finish %v", g.timeout)
			time.Sleep(g.timeout)
			cancelFunc()
			// call after close hooks
			for _, fn := range g.hookMaps[_HookExit] {
				fn(context.Background())
			}
			// notify process exit
			g.wg.Done()
			runtime.Gosched()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}

func (g *GinEngine) Close() {
	if g.server != nil {
		// disable keep-alives on existing connections
		g.server.SetKeepAlivesEnabled(false)
		_ = g.server.Shutdown(context.Background())
	}
}

// Logger
func (g *GinEngine) Logger(ignoreRelease bool) gin.HandlerFunc {
	g.IgnoreReleaseLog = ignoreRelease
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
		if gin.Mode() == gin.ReleaseMode && g.IgnoreReleaseLog {
			return
		}

		// End time
		end := time.Now()
		fmt.Fprintf(os.Stdout, "[GIN] %s | %3d | %13v | %15s | %-7s %#v\n%s", end.Format("2006/01/02 - 15:04:05"), c.Writer.Status(), end.Sub(start), c.ClientIP(), c.Request.Method, path, c.Errors.ByType(gin.ErrorTypePrivate).String())
	}
}
