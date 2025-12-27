package app

import (
	"context"

	"github.com/cloudapex/river/conf"
	"github.com/cloudapex/river/mqrpc"
	"github.com/cloudapex/river/mqrpc/core"
	"github.com/cloudapex/river/registry"
	"github.com/cloudapex/river/selector"
	"github.com/nats-io/nats.go"
)

// IApp 应用定义
type IApp interface {
	OnInit() error
	OnDestroy() error

	Run(mods ...IModule) error

	// Config 获取启动配置
	Config() conf.Config

	// Options 获取应用配置
	Options() Options
	// Transporter 获取消息传输对象
	Transporter() *nats.Conn
	// Registrar 获取服务注册对象
	Registrar() registry.Registry
	// WorkDir 获取进程工作目录
	WorkDir() string
	// GetProcessEnv 获取应用进程分组ID(dev,test,...)
	GetProcessEnv() string

	// UpdateOptions 允许再次更新应用配置(before app.Run)
	UpdateOptions(opts ...Option) error
	// 设置服务路由器(动态转换service名称)
	SetServiceRoute(fn func(route string) string) error

	// 获取服务实例(通过服务ID|服务类型,可设置selector.WithFilter和selector.WithStrategy)
	GetRouteServer(service string, opts ...selector.SelectOption) (IModuleServerSession, error)
	// 获取服务实例(通过服务ID(moduleType@id))
	GetServerByID(serverID string) (IModuleServerSession, error)
	// 获取服务实例(通过服务类型(moduleType),可设置可设置selector.WithFilter和selector.WithStrategy)
	GetServerBySelector(serviceName string, opts ...selector.SelectOption) (IModuleServerSession, error)
	// 获取多个服务实例(通过服务类型(moduleType))
	GetServersByType(serviceName string) []IModuleServerSession

	// Call RPC调用(需要等待结果)
	Call(ctx context.Context, moduleServer, _func string, param mqrpc.ParamOption, opts ...selector.SelectOption) (any, error)
	// Call RPC调用(无需等待结果)
	CallNR(ctx context.Context, moduleServer, _func string, params ...any) error
	// Call RPC调用(群发,无需等待结果)
	CallBroadcast(ctx context.Context, moduleType, _func string, params ...any)

	// 回调(hook)
	OnConfigurationLoaded(func()) error                           // 设置应用启动配置初始化完成后回调
	OnModuleInited(func(module IModule)) error                    // 设置每个模块初始化完成后回调
	GetModuleInited() func(module IModule)                        // 获取每个模块初始化完成后回调函数
	OnStartup(func()) error                                       // 设置应用启动完成后回调
	OnServiceBreak(_func func(moduleName, serverId string)) error // 设置当模块服务断开删除时回调
}

// IModule 基本模块定义
type IModule interface {
	GetType() string // 模块类型
	Version() string // 模块版本

	Run(closeSig chan bool)

	OnInit(settings *conf.ModuleSettings) // 所有初始化逻辑都放到Init中, 重载OnInit不可调用基类!(由Init层层调用base.Init)即可
	OnDestroy()
	OnAppConfigurationLoaded()                   // 当App初始化时调用，这个接口不管这个模块是否在这个进程运行都会调用
	OnConfChanged(settings *conf.ModuleSettings) // 为以后动态服务发现做准备(目前没用)
}

// IRPCModule RPC模块定义
type IRPCModule interface {
	IModule
	//context.Context

	// 模块服务ID
	GetServerID() string
	GetModuleSettings() (settings *conf.ModuleSettings)

	// 注册RPC方法(f的第一个参数必须是context.Context,返回参数(最多两个)最后一个必须是error)
	Register(msg string, f interface{})   // 同步
	RegisterGO(msg string, f interface{}) // 并发

	// 获取服务实例(通过服务ID|服务类型,可设置选择器过滤)
	GetRouteServer(service string, opts ...selector.SelectOption) (IModuleServerSession, error) //获取经过筛选过的服务
	// 通过服务ID(moduleType@id)获取服务实例
	GetServerByID(serverID string) (IModuleServerSession, error)
	// 通过服务类型(moduleType)获取服务实例列表
	GetServersByType(serviceName string) []IModuleServerSession
	// 通过服务类型(moduleType)获取服务实例(可设置选择器)
	GetServerBySelector(serviceName string, opts ...selector.SelectOption) (IModuleServerSession, error)

	// RPC方法
	Call(ctx context.Context, moduleServer, _func string, params mqrpc.ParamOption, opts ...selector.SelectOption) (any, error)
	CallNR(ctx context.Context, moduleServer, _func string, params ...any) error
	CallBroadcast(ctx context.Context, moduleType, _func string, params ...any)
}

// IModuleServerSession Module服务会话代理
type IModuleServerSession interface {
	// 服务ID
	GetID() string
	// 服务名称(moduleType)
	GetName() string
	// RPC客户端
	GetRPC() mqrpc.IRPCClient

	// 服务节点信息
	GetNode() *registry.Node
	SetNode(node *registry.Node) (err error)
}

// FileNameHandler 自定义日志文件名字
type FileNameHandler func(logdir, prefix, processID, suffix string) string

// ClientRPCHandler 调用方RPC监控
type ClientRPCHandler func(server registry.Node, rpcinfo *core.RPCInfo, result any, err error, exec_time int64)

// ServerRPCHandler 服务方RPC监控
type ServerRPCHandler func(module IModule, callInfo *mqrpc.CallInfo)

// ServerRPCHandler 服务方RPC完成监控
// type RpcCompleteHandler func(module IModule, callInfo *mqrpc.CallInfo, input []any, out []any, execTime time.Duration)
