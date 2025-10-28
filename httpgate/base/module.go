package httpgatebase

import (
	"net/http"

	"github.com/cloudapex/river/app"
	"github.com/cloudapex/river/conf"
	"github.com/cloudapex/river/httpgate"
	"github.com/cloudapex/river/log"
	"github.com/cloudapex/river/module"
)

var _ app.IRPCModule = &HttpGateBase{}

type HttpGateBase struct {
	module.ModuleBase

	opts httpgate.Options
}

func (this *HttpGateBase) Init(subclass app.IRPCModule, settings *conf.ModuleSettings, opts ...httpgate.Option) {
	this.opts = httpgate.NewOptions(opts...)
	this.ModuleBase.Init(subclass, settings, this.opts.Opts...) // 这是必须的
}
func (this *HttpGateBase) GetType() string {
	// 很关键,需要与配置文件中的Module配置对应
	return "httpgate"
}
func (this *HttpGateBase) Version() string {
	// 可以在监控时了解代码版本
	return "1.0.0"
}
func (this *HttpGateBase) Run(closeSig chan bool) {
	srv := this.startHttpServer()

	<-closeSig

	if err := srv.Shutdown(nil); err != nil {
		log.Error("Shutdown() error: %s", err)
	}
}
func (this *HttpGateBase) OnDestroy() {
	// 一定别忘了继承
	this.ModuleBase.OnDestroy()
}

// ---------------

func (this *HttpGateBase) startHttpServer() *http.Server {
	srv := &http.Server{
		Addr:    ":8090",
		Handler: NewHandler(this.opts),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			// cannot panic, because this probably is an intentional close
			log.Error("ListenAndServe() error: %s", err)
		}
	}()
	// returning reference so caller can call Shutdown()
	return srv
}
