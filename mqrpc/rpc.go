// Package mqrpc rpc接口定义
package mqrpc

import (
	"context"
	"reflect"

	"github.com/cloudapex/river/log"
	rpcpb "github.com/cloudapex/river/mqrpc/pb"
)

type ContextTransKey string

var registedContextTransfer = map[ContextTransKey]func() Marshaler{
	// 默认注册log.TraceSpanImp
	ContextTransTrace: func() Marshaler { return &log.TraceSpanImp{} },
}

// 把需要通过Context进行RPC传输的KV注册进来
func RegistContextTransValue(key ContextTransKey, makeFun func() Marshaler) {
	registedContextTransfer[key] = makeFun
}

// 定义需要RPC传输TraceSpan的ContextKey
var ContextTransTrace = ContextTransKey("ContextTransTrace")

// ContextTransTrace快捷WithValue方法
func ContextWithTrace(ctx context.Context, trace log.TraceSpan) context.Context {
	return context.WithValue(ctx, ContextTransTrace, trace)
}

// FunctionInfo handler接口信息
type FunctionInfo struct {
	Function  reflect.Value
	FuncType  reflect.Type
	InType    []reflect.Type
	Goroutine bool
}

// MQServer 代理者
type MQServer interface {
	Callback(callinfo *CallInfo) error
}

// CallInfo RPC的请求信息
type CallInfo struct {
	RPCInfo  *rpcpb.RPCInfo
	Result   *rpcpb.ResultInfo
	Props    map[string]interface{}
	ExecTime int64
	Agent    MQServer //代理者  AMQPServer / LocalServer 都继承 Callback(callinfo CallInfo)(error) 方法
}

// RPCListener 事件监听器
type RPCListener interface {
	/**
	NoFoundFunction 当未找到请求的handler时会触发该方法
	*FunctionInfo  选择可执行的handler
	return error
	*/
	NoFoundFunction(fn string) (*FunctionInfo, error)
	/**
	BeforeHandle会对请求做一些前置处理，如：检查当前玩家是否已登录，打印统计日志等。
	@session  可能为nil
	return error  当error不为nil时将直接返回改错误信息而不会再执行后续调用
	*/
	BeforeHandle(fn string, callInfo *CallInfo) error
	OnTimeOut(fn string, Expired int64)
	OnError(fn string, callInfo *CallInfo, err error)
	/**
	fn 		方法名
	params		参数
	result		执行结果
	exec_time 	方法执行时间 单位为 Nano 纳秒  1000000纳秒等于1毫秒
	*/
	OnComplete(fn string, callInfo *CallInfo, result *rpcpb.ResultInfo, execTime int64)
}

// GoroutineControl 服务协程数量控制
type GoroutineControl interface {
	Wait() error
	Finish()
}

// RPCServer 服务定义
type RPCServer interface {
	Addr() string
	SetListener(listener RPCListener) // 设置监听器
	SetGoroutineControl(control GoroutineControl)
	GetExecuting() int64
	Register(id string, f interface{})   // 注册RPC方法,f第一个参数必须为context.Context(单线程)
	RegisterGO(id string, f interface{}) // 注册RPC方法,f第一个参数必须为context.Context(多线程)
	Done() (err error)
}

// RPCClient 客户端定义
type RPCClient interface {
	Done() (err error)
	Call(ctx context.Context, _func string, params ...interface{}) (interface{}, error)
	CallArgs(ctx context.Context, _func string, argsType []string, args [][]byte) (interface{}, error) // ctx参数必须装进args中
	CallNR(ctx context.Context, _func string, params ...interface{}) (err error)
	CallNRArgs(ctx context.Context, _func string, argsType []string, args [][]byte) (err error) // ctx参数必须装进args中
}

// Marshaler is a simple encoding interface used for the broker/transport
// where headers are not supported by the underlying implementation.
type Marshaler interface {
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
	String() string
}

// RPCSerialize 自定义参数序列化接口
type RPCSerialize interface {
	/**
	序列化 结构体-->[]byte
	param 需要序列化的参数值
	@return ptype 当能够序列化这个值,并且正确解析为[]byte时 返回改值正确的类型,否则返回 ""即可
	@return p 解析成功得到的数据, 如果无法解析该类型,或者解析失败 返回nil即可
	@return err 无法解析该类型,或者解析失败 返回错误信息
	*/
	Serialize(param interface{}) (ptype string, p []byte, err error)
	/**
	反序列化 []byte-->结构体
	ptype 参数类型 与Serialize函数中ptype 对应
	b   参数的字节流
	@return param 解析成功得到的数据结构
	@return err 无法解析该类型,或者解析失败 返回错误信息
	*/
	Deserialize(ptype string, b []byte) (param interface{}, err error)
	/**
	返回这个接口能够处理的所有类型
	*/
	GetTypes() []string
}

// ParamOption ParamOption
type ParamOption func() []interface{}

// Param 请求参数包装器
func Param(params ...interface{}) ParamOption {
	return func() []interface{} {
		return params
	}
}
