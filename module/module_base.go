// Package basemodule BaseModule定义
package module

import (
	"context"
	"fmt"
	"os"

	"github.com/cloudapex/river/app"
	"github.com/cloudapex/river/conf"
	"github.com/cloudapex/river/log"
	"github.com/cloudapex/river/module/server"
	"github.com/cloudapex/river/module/service"
	"github.com/cloudapex/river/mqrpc"
	"github.com/cloudapex/river/mqrpc/core"
	"github.com/cloudapex/river/selector"
	"github.com/cloudapex/river/tools"
	"github.com/pkg/errors"
)

// ModuleBase 默认的RPCModule实现
type ModuleBase struct {
	//context.Context

	Impl     app.IRPCModule
	settings *conf.ModuleSettings

	serviceStoped chan bool
	exit          context.CancelFunc

	service  service.Service // 内含server
	listener mqrpc.RPCListener
}

// Init 模块初始化(由派生类调用)
func (this *ModuleBase) Init(impl app.IRPCModule, settings *conf.ModuleSettings, opt ...server.Option) {
	// 初始化模块
	this.Impl = impl
	this.settings = settings

	// 创建一个供远程调用的RPCService
	opts := server.Options{
		Metadata: map[string]string{},
	}
	for _, o := range opt {
		o(&opts)
	}
	if opts.Registry == nil {
		opt = append(opt, server.Registry(app.App().Registrar()))
	}

	if opts.RegisterInterval == 0 {
		opt = append(opt, server.RegisterInterval(app.App().Options().RegisterInterval))
	}

	if opts.RegisterTTL == 0 {
		opt = append(opt, server.RegisterTTL(app.App().Options().RegisterTTL))
	}

	if len(opts.Name) == 0 {
		opt = append(opt, server.Name(this.Impl.GetType()))
	}

	if len(opts.ID) == 0 {
		if settings.ID != "" {
			opt = append(opt, server.ID(settings.ID))
		} else {
			opt = append(opt, server.ID(tools.GenerateID().String()))
		}
	}

	if len(opts.Version) == 0 {
		opt = append(opt, server.Version(this.Impl.Version()))
	}

	server := server.NewServer(opt...) // opts.Address = nats_server.addr
	err := server.OnInit(this.Impl, settings)
	if err != nil {
		log.Warning("server OnInit fail id(%s) error(%s)", this.GetServerID(), err)
	}
	hostname, _ := os.Hostname()
	server.Options().Metadata["hostname"] = hostname
	server.Options().Metadata["pid"] = fmt.Sprintf("%v", os.Getpid())
	ctx, cancel := context.WithCancel(context.Background())
	this.exit = cancel
	this.serviceStoped = make(chan bool)
	this.service = service.NewService(
		service.Server(server),
		service.RegisterInterval(app.App().Options().RegisterInterval),
		service.Context(ctx),
	)

	go func() {
		err := this.service.Run()
		if err != nil {
			log.Warning("service run fail id(%s) error(%s)", this.GetServerID(), err)
		}
		close(this.serviceStoped)
	}()
	this.GetServer().SetListener(this)
}

// OnInit 当模块初始化时调用
func (this *ModuleBase) OnInit(settings *conf.ModuleSettings) {
	// 所有初始化逻辑都放到Init中, 重载OnInit不可调用基类!
	panic("ModuleBase: OnInit() must be implemented")
}

// OnDestroy 当模块注销时调用
func (this *ModuleBase) OnDestroy() {
	this.exit()

	select {
	case <-this.serviceStoped: // 等待注册中心注销完成
	}
	_ = this.GetServer().OnDestroy() //一定别忘了关闭RPC
}

// SetListener  mqrpc.RPCListener
func (this *ModuleBase) SetListener(listener mqrpc.RPCListener) {
	this.listener = listener
}

// GetImpl 获取子类
func (this *ModuleBase) GetImpl() app.IRPCModule {
	return this.Impl
}

// GetServer server.Server
func (this *ModuleBase) GetServer() server.Server {
	return this.service.Server()
}

// GetServerID 节点ID
func (this *ModuleBase) GetServerID() string {
	//很关键,需要与配置文件中的Module配置对应
	if this.service != nil && this.service.Server() != nil {
		return this.service.Server().ID()
	}
	return "no server"
}

// GetModuleSettings  获取Config.Module[typ].Settings
func (this *ModuleBase) GetModuleSettings() *conf.ModuleSettings {
	return this.settings
}

// OnConfChanged 当配置变更时调用(目前没用)
func (this *ModuleBase) OnConfChanged(settings *conf.ModuleSettings) {}

// OnAppConfigurationLoaded 当应用配置加载完成时调用
func (this *ModuleBase) OnAppConfigurationLoaded() {
	// 当App初始化时调用，这个接口不管这个模块是否在这个进程运行都会调用
}

// GetRouteServer 获取服务实例(通过服务ID|服务类型,可设置选择器过滤)
func (this *ModuleBase) GetRouteServer(service string, opts ...selector.SelectOption) (s app.IServerSession, err error) {
	return app.App().GetRouteServer(service, opts...)
}

// GetServerByID 通过服务ID(moduleType@id)获取服务实例
func (this *ModuleBase) GetServerByID(serverID string) (app.IServerSession, error) {
	return app.App().GetServerByID(serverID)
}

// GetServersByType 通过服务类型(moduleType)获取服务实例列表
func (this *ModuleBase) GetServersByType(serviceName string) []app.IServerSession {
	return app.App().GetServersByType(serviceName)
}

// GetServerBySelector 通过服务类型(moduleType)获取服务实例(可设置选择器)
func (this *ModuleBase) GetServerBySelector(serviceName string, opts ...selector.SelectOption) (app.IServerSession, error) {
	return app.App().GetServerBySelector(serviceName, opts...)
}

// Call  RPC调用(需要等待结果)
func (this *ModuleBase) Call(ctx context.Context, moduleType, _func string, params mqrpc.ParamOption, opts ...selector.SelectOption) (any, error) {
	return app.App().Call(ctx, moduleType, _func, params, opts...)
}

// CallNR  RPC调用(需要等待结果)
func (this *ModuleBase) CallNR(ctx context.Context, moduleType, _func string, params ...any) (err error) {
	return app.App().CallNR(ctx, moduleType, _func, params...)
}

// CallBroadcast RPC调用(群发,无需等待结果)
func (this *ModuleBase) CallBroadcast(ctx context.Context, moduleType, _func string, params ...any) {
	app.App().CallBroadcast(ctx, moduleType, _func, params...)
}

// ================= RPCListener[监听事件]

// NoFoundFunction  当hander未找到时调用
func (this *ModuleBase) NoFoundFunction(fn string) (*mqrpc.FunctionInfo, error) {
	if this.listener != nil {
		return this.listener.NoFoundFunction(fn)
	}
	return nil, errors.Errorf("Remote function(%s) not found", fn)
}

// BeforeHandle  hander执行前调用
func (this *ModuleBase) BeforeHandle(fn string, callInfo *mqrpc.CallInfo) error {
	if this.listener != nil {
		return this.listener.BeforeHandle(fn, callInfo)
	}
	return nil
}

// OnTimeOut  hander执行超时调用
func (this *ModuleBase) OnTimeOut(fn string, Expired int64) {
	if this.listener != nil {
		this.listener.OnTimeOut(fn, Expired)
	}
}

// OnError  hander执行错误调用
func (this *ModuleBase) OnError(fn string, callInfo *mqrpc.CallInfo, err error) {
	if this.listener != nil {
		this.listener.OnError(fn, callInfo, err)
	}
}

// OnComplete hander成功执行完成时调用
// fn 		方法名
// params		参数
// result		执行结果
// exec_time 	方法执行时间 单位为 Nano 纳秒  1000000纳秒等于1毫秒
func (this *ModuleBase) OnComplete(fn string, callInfo *mqrpc.CallInfo, result *core.ResultInfo, execTime int64) {
	if this.listener != nil {
		this.listener.OnComplete(fn, callInfo, result, execTime)
	}
}
