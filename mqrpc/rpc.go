// Package mqrpc rpc接口定义
package mqrpc

import (
	"context"
	"reflect"

	"github.com/cloudapex/river/mqrpc/core"
)

// MethodInfo 方法信息
type MethodInfo struct {
	Function  reflect.Value
	FuncType  reflect.Type
	InType    []reflect.Type
	Goroutine bool
}

// CallInfo RPC的请求信息
type CallInfo struct {
	RPCInfo  *core.RPCInfo
	Result   *core.ResultInfo
	Props    map[string]any
	ExecTime int64
	Agent    IMQServer //代理者  AMQPServer / LocalServer 都继承 Callback(callinfo CallInfo)(error) 方法
}

// IMQServer 代理者
type IMQServer interface {
	Callback(callinfo *CallInfo) error
}

// IRPCListener 事件监听器
type IRPCListener interface {
	// 当未找到请求的Method时会触发该方法(返回的MethodInfo会被执行,为nil时必须返回error)
	OnMethodNotFound(fn string) (*MethodInfo, error)

	// 会对method执行之前做一些前置处理，如：检查当前玩家是否已登录，打印统计日志等(当error不为nil时将直接返回改错误信息而不会再执行后续调用)
	OnBeforeHandle(fn string, callInfo *CallInfo) error

	// 当方法执行超时时会触发该回调
	OnTimeOut(fn string, Expired int64)

	// 当方法执行异常时会触发该回调
	OnError(fn string, callInfo *CallInfo, err error)

	// 当方法执行完成时会触发该回调(时间单位为 Nano 纳秒  1000000纳秒等于1毫秒)
	OnComplete(fn string, callInfo *CallInfo, result *core.ResultInfo, execTime int64)
}

// IGoroutineControl 服务协程数量控制
type IGoroutineControl interface {
	Wait() error
	Finish()
}

// IRPCServer 服务定义
type IRPCServer interface {
	Addr() string
	SetListener(listener IRPCListener) // 设置监听器
	SetGoroutineControl(control IGoroutineControl)
	GetExecuting() int64
	Register(id string, f any)   // 注册RPC方法,f第一个参数必须为context.Context(单线程)
	RegisterGO(id string, f any) // 注册RPC方法,f第一个参数必须为context.Context(多线程)
	Done() (err error)
}

// IRPCClient 客户端定义
type IRPCClient interface {
	Done() (err error)
	Call(ctx context.Context, _func string, params ...any) (any, error)
	CallArgs(ctx context.Context, _func string, argTypes []string, args [][]byte) (any, error) // ctx参数必须装进args中
	CallNR(ctx context.Context, _func string, params ...any) (err error)
	CallNRArgs(ctx context.Context, _func string, argTypes []string, args [][]byte) (err error) // ctx参数必须装进args中
}

// IMarshaler is a simple encoding interface used for the broker/transport
// where headers are not supported by the underlying implementation.
type IMarshaler interface {
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
	String() string
}

// IRPCSerialize 自定义参数序列化接口
type IRPCSerialize interface {
	/**
	序列化 结构体-->[]byte
	param 需要序列化的参数值
	@return ptype 当能够序列化这个值,并且正确解析为[]byte时 返回改值正确的类型,否则返回 ""即可
	@return p 解析成功得到的数据, 如果无法解析该类型,或者解析失败 返回nil即可
	@return err 无法解析该类型,或者解析失败 返回错误信息
	*/
	Serialize(param any) (ptype string, p []byte, err error)
	/**
	反序列化 []byte-->结构体
	ptype 参数类型 与Serialize函数中ptype 对应
	b   参数的字节流
	@return param 解析成功得到的数据结构
	@return err 无法解析该类型,或者解析失败 返回错误信息
	*/
	Deserialize(ptype string, b []byte) (param any, err error)
	/**
	返回这个接口能够处理的所有类型
	*/
	GetTypes() []string
}

// ParamOption ParamOption
type ParamOption func() []any

// Param 请求参数包装器
func Param(params ...any) ParamOption {
	return func() []any {
		return params
	}
}
