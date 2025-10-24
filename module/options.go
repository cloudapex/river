package module

import (
	"time"

	"github.com/cloudapex/river/mqrpc"
	rpcpb "github.com/cloudapex/river/mqrpc/pb"
	"github.com/cloudapex/river/registry"
	"github.com/cloudapex/river/selector"
	"github.com/nats-io/nats.go"
)

// Option 应用级别配置项
type Option func(*Options)

// Options 应用级别配置
type Options struct {
	Version     string
	Debug       bool
	Parse       bool // 是否由框架解析启动环境变量,默认为true
	WorkDir     string
	ProcessEnv  string   // 进程分组ID(development)
	ConfigKey   string   // for consule
	ConsulAddr  []string // for consule
	LogDir      string
	BIDir       string
	PProfAddr   string
	KillWaitTTL time.Duration // 服务关闭超时强杀(60s)

	Nats             *nats.Conn
	Registry         registry.Registry // 注册服务发现(registry.DefaultRegistry)
	Selector         selector.Selector // 节点选择器(在Registry基础上)(cache.NewSelector())
	RegisterInterval time.Duration     // 服务注册发现续约频率(10s)
	RegisterTTL      time.Duration     // 服务注册发现续约生命周期(20s)

	ClientRPChandler   ClientRPCHandler   // 配置RPC调用方监控器(nil)
	ServerRPCHandler   ServerRPCHandler   // 配置RPC服务方监控器(nil)
	RpcCompleteHandler RpcCompleteHandler // 配置RPC执行结果监控器(nil)
	RPCExpired         time.Duration      // RPC调用超时(10s)
	RPCMaxCoroutine    int                // 默认0(不限制) 没用

	// 自定义日志文件名字(主要作用方便k8s映射日志不会被冲突，建议使用k8s pod实现)
	LogFileName FileNameHandler // 日志文件名称(默认):fmt.Sprintf("%s/%v%s%s", logdir, prefix, processID, suffix)
	// 自定义BI日志名字
	BIFileName FileNameHandler //  BI文件名称(默认):fmt.Sprintf("%s/%v%s%s", logdir, prefix, processID, suffix)
}

type FileNameHandler func(logdir, prefix, processID, suffix string) string

// ClientRPCHandler 调用方RPC监控
type ClientRPCHandler func(app IApp, server registry.Node, rpcinfo *rpcpb.RPCInfo, result interface{}, err error, exec_time int64)

// ServerRPCHandler 服务方RPC监控
type ServerRPCHandler func(app IApp, module IModule, callInfo *mqrpc.CallInfo)

// ServerRPCHandler 服务方RPC监控
type RpcCompleteHandler func(app IApp, module IModule, callInfo *mqrpc.CallInfo, input []interface{}, out []interface{}, execTime time.Duration)

// Version 应用版本
func Version(v string) Option {
	return func(o *Options) {
		o.Version = v
	}
}

// Debug 只有是在调试模式下才会在控制台打印日志, 非调试模式下只在日志文件中输出日志
func Debug(t bool) Option {
	return func(o *Options) {
		o.Debug = t
	}
}

// WorkDir 进程工作目录
func WorkDir(v string) Option {
	return func(o *Options) {
		o.WorkDir = v
	}
}

// Configure 配置key
func ConfigKey(v string) Option {
	return func(o *Options) {
		o.ConfigKey = v
	}
}

// Configure consule 地址
func ConsulAddr(v ...string) Option {
	return func(o *Options) {
		o.ConsulAddr = append(o.ConsulAddr, v...)
	}
}

// LogDir 日志存储路径
func LogDir(v string) Option {
	return func(o *Options) {
		o.LogDir = v
	}
}

// ProcessID 进程分组ID
func ProcessID(v string) Option {
	return func(o *Options) {
		o.ProcessEnv = v
	}
}

// BILogDir  BI日志路径
func BILogDir(v string) Option {
	return func(o *Options) {
		o.BIDir = v
	}
}

// Nats  nats配置
func Nats(nc *nats.Conn) Option {
	return func(o *Options) {
		o.Nats = nc
	}
}

// Registry sets the registry for the service
// and the underlying components
func Registry(r registry.Registry) Option {
	return func(o *Options) {
		o.Registry = r
		o.Selector.Init(selector.Registry(r))
	}
}

// Selector 路由选择器
func Selector(r selector.Selector) Option {
	return func(o *Options) {
		o.Selector = r
	}
}

// RegisterTTL specifies the TTL to use when registering the service
func RegisterTTL(t time.Duration) Option {
	return func(o *Options) {
		o.RegisterTTL = t
	}
}

// RegisterInterval specifies the interval on which to re-register
func RegisterInterval(t time.Duration) Option {
	return func(o *Options) {
		o.RegisterInterval = t
	}
}

// KillWaitTTL specifies the interval on which to re-register
func KillWaitTTL(t time.Duration) Option {
	return func(o *Options) {
		o.KillWaitTTL = t
	}
}

// SetClientRPChandler 配置调用者监控器
func SetClientRPChandler(t ClientRPCHandler) Option {
	return func(o *Options) {
		o.ClientRPChandler = t
	}
}

// SetServerRPCHandler 配置服务方监控器
func SetServerRPCHandler(t ServerRPCHandler) Option {
	return func(o *Options) {
		o.ServerRPCHandler = t
	}
}

// SetServerRPCCompleteHandler 服务RPC执行结果监控器
func SetRpcCompleteHandler(t RpcCompleteHandler) Option {
	return func(o *Options) {
		o.RpcCompleteHandler = t
	}
}

// Parse mqant框架是否解析环境参数
func Parse(t bool) Option {
	return func(o *Options) {
		o.Parse = t
	}
}

// RPC超时时间
func RPCExpired(t time.Duration) Option {
	return func(o *Options) {
		o.RPCExpired = t
	}
}

// 单个节点RPC同时并发协程数
func RPCMaxCoroutine(t int) Option {
	return func(o *Options) {
		o.RPCMaxCoroutine = t
	}
}

// WithLogFile 日志文件名称
func WithLogFile(name FileNameHandler) Option {
	return func(o *Options) {
		o.LogFileName = name
	}
}

// WithBIFile Bi日志名称
func WithBIFile(name FileNameHandler) Option {
	return func(o *Options) {
		o.BIFileName = name
	}
}
