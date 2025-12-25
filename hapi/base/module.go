package hapibase

import (
	"context"
	"net/http"
	"time"

	"github.com/cloudapex/river/app"
	"github.com/cloudapex/river/conf"
	"github.com/cloudapex/river/hapi"
	"github.com/cloudapex/river/log"
	"github.com/cloudapex/river/module"
	"github.com/gin-gonic/gin"
)

var _ app.IRPCModule = &HApiBase{}

type HApiBase struct {
	module.ModuleBase

	opts hapi.Options

	router *gin.Engine
}

func (this *HApiBase) Init(subclass app.IRPCModule, settings *conf.ModuleSettings, opts ...hapi.Option) {
	this.ModuleBase.Init(subclass, settings, this.opts.Opts...) // 这是必须的

	// 使用settings的配置覆盖opts
	this.opts = hapi.NewOptions(opts...)
	if WSAddr, ok := settings.Settings["Addr"]; ok {
		this.opts.Addr = WSAddr.(string)
	}

	if tls, ok := settings.Settings["TLS"]; ok {
		this.opts.TLS = tls.(bool)
	}

	if CertFile, ok := settings.Settings["CertFile"]; ok {
		this.opts.CertFile = CertFile.(string)
	}

	if KeyFile, ok := settings.Settings["KeyFile"]; ok {
		this.opts.KeyFile = KeyFile.(string)
	}

	if TimeOut, ok := settings.Settings["TimeOut"]; ok {
		this.opts.TimeOut = time.Duration(TimeOut.(int)) * time.Second
	}

	if ReadTimeout, ok := settings.Settings["ReadTimeout"]; ok {
		this.opts.ReadTimeout = time.Duration(ReadTimeout.(int)) * time.Second
	}

	if WriteTimeout, ok := settings.Settings["WriteTimeout"]; ok {
		this.opts.WriteTimeout = time.Duration(WriteTimeout.(int)) * time.Second
	}

	if IdleTimeout, ok := settings.Settings["IdleTimeout"]; ok {
		this.opts.IdleTimeout = time.Duration(IdleTimeout.(int)) * time.Second
	}

	if MaxHeaderBytes, ok := settings.Settings["MaxHeaderBytes"]; ok {
		this.opts.MaxHeaderBytes = MaxHeaderBytes.(int)
	}

	// 创建路由
	gin.SetMode(gin.ReleaseMode)
	this.router = gin.New()
	this.router.Use(gin.Logger())
	this.router.Use(gin.Recovery())

	// 创建处理器
	_handler := NewHandler(this.opts)
	if this.opts.RpcHandle == nil {
		this.opts.RpcHandle = _handler.callRpcService
	}
	this.router.NoRoute(func(ctx *gin.Context) {
		// 将 gin.Context 转换为标准 http.ResponseWriter 和 *http.Request
		_handler.ServeHTTP(ctx.Writer, ctx.Request)
	})
}
func (this *HApiBase) GetType() string {
	// 很关键,需要与配置文件中的Module配置对应
	return "hapi"
}
func (this *HApiBase) Version() string {
	// 可以在监控时了解代码版本
	return "1.0.0"
}
func (this *HApiBase) Options() hapi.Options { return this.opts }

func (this *HApiBase) OnInit(settings *conf.ModuleSettings) {
	// 所有初始化逻辑都放到Init中, 重载OnInit不可调用基类OnInit!
	panic("HApiBase: OnInit() must be implemented")
}
func (this *HApiBase) OnDestroy() {
	this.ModuleBase.OnDestroy() // 一定别忘了继承
}
func (this *HApiBase) RouterGroup() *gin.RouterGroup { return &this.router.RouterGroup }

func (this *HApiBase) Run(closeSig chan bool) {
	srv := this.startHttpServer()

	<-closeSig

	if err := srv.Shutdown(context.Background()); err != nil {
		log.Error("Shutdown() error: %s", err)
	}
}

// ---------------

func (this *HApiBase) startHttpServer() *http.Server {
	srv := &http.Server{
		Addr:           this.opts.Addr,
		Handler:        this.router,
		ReadTimeout:    this.opts.ReadTimeout,
		WriteTimeout:   this.opts.WriteTimeout,
		IdleTimeout:    this.opts.IdleTimeout,
		MaxHeaderBytes: this.opts.MaxHeaderBytes,
	}

	go func() {
		var err error
		if this.opts.TLS {
			// TLS配置存在，使用HTTPS
			log.Info("Starting HTTPS server on %s with cert %s and key %s", this.opts.Addr, this.opts.CertFile, this.opts.KeyFile)
			err = srv.ListenAndServeTLS(this.opts.CertFile, this.opts.KeyFile)
		} else {
			// 没有TLS配置，使用HTTP
			log.Info("Starting HTTP server on %s", this.opts.Addr)
			err = srv.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			// cannot panic, because this probably is an intentional close
			log.Error("ListenAndServe() error: %s", err)
		}
	}()
	// returning reference so caller can call Shutdown()
	return srv
}
