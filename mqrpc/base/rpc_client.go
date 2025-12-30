package rpcbase

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/cloudapex/river/app"
	"github.com/cloudapex/river/log"
	"github.com/cloudapex/river/mqrpc"
	"github.com/cloudapex/river/mqrpc/core"
	"github.com/google/uuid"
)

type RPCClient struct {
	nats_client *NatsClient
}

func NewRPCClient(session app.IModuleServerSession) (mqrpc.IRPCClient, error) {
	rpc_client := new(RPCClient)
	nats_client, err := NewNatsClient(session)
	if err != nil {
		log.Error("Dial: %s", err)
		return nil, err
	}
	rpc_client.nats_client = nats_client
	return rpc_client, nil
}

func (c *RPCClient) Done() (err error) {
	if c.nats_client != nil {
		err = c.nats_client.Done()
	}
	return
}

func (c *RPCClient) Call(ctx context.Context, _func string, params ...any) (any, error) {
	_ctx := ctx
	var argTypes []string = make([]string, len(params)+1)
	var argDatas [][]byte = make([][]byte, len(params)+1)

	// 检测添加log.TraceSpan到ctx
	span, ok := ctx.Value(log.RPC_CONTEXT_KEY_TRACE).(log.TraceSpan)
	if !ok {
		_ctx = mqrpc.ContextWithValue(_ctx, log.RPC_CONTEXT_KEY_TRACE, log.CreateRootTrace())
	} else {
		_ctx = mqrpc.ContextWithValue(_ctx, log.RPC_CONTEXT_KEY_TRACE, span.ExtractSpan())
	}

	// 重新组装参数(ctx放到首位)
	params = append([]any{_ctx}, params...)
	for k, arg := range params {
		var err error = nil
		argTypes[k], argDatas[k], err = mqrpc.ArgToData(arg)
		if err != nil {
			return nil, fmt.Errorf("args[%d] error %s", k, err.Error())
		}
	}

	// CallArgs
	return c.CallArgs(ctx, _func, argTypes, argDatas)
}
func (c *RPCClient) CallArgs(ctx context.Context, _func string, argTypes []string, argDatas [][]byte) (any, error) {
	var err error
	var result any
	var result_info = core.ResultInfo{ResultType: "unknown", Result: nil}

	caller, _ := os.Hostname()
	if ctx != nil {
		cr, ok := ctx.Value("caller").(string)
		if ok {
			caller = cr
		}
	}

	start := time.Now()
	rpcInfo := &core.RPCInfo{
		Fn:       _func,
		Reply:    true,
		Expired:  (start.UTC().Add(app.App().Options().RPCExpired).UnixNano()) / 1000000,
		Cid:      uuid.New().String(),
		Args:     argDatas,
		ArgsType: argTypes,
		Caller:   caller,
		Hostname: caller,
	}

	defer func() { // 全局监控(调用方)
		if app.App().Config().RpcLog || err != nil { // 打印调用日志
			span, _ := ctx.Value(log.RPC_CONTEXT_KEY_TRACE).(log.TraceSpan)
			log.TInfo(span, "rpc Call ServerId = %v, Func = %v, Elapsed = %v, Result = <%s-len:%d>, Error = %v",
				c.nats_client.session.GetID(), _func, time.Since(start), result_info.ResultType, len(result_info.Result), err)
		}
		if handle := app.App().Options().ClientRPCHandler; handle != nil {
			handle(*c.nats_client.session.GetNode(), rpcInfo, result, err, time.Since(start).Nanoseconds())
		}
	}()

	// call
	callInfo := &mqrpc.CallInfo{
		RPCInfo: rpcInfo,
	}
	callback := make(chan *core.ResultInfo, 1)
	err = c.nats_client.Call(callInfo, callback)
	if err != nil {
		return nil, err
	}

	// 没有设置超时的话使用默认超时
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, app.App().Options().RPCExpired)
		defer cancel()
	}

	select {
	case resultInfo, ok := <-callback: // 结果
		if !ok {
			return nil, fmt.Errorf("client closed")
		}
		result_info = *resultInfo
		result, err = mqrpc.DataToArg(resultInfo.ResultType, resultInfo.Result)
		if err != nil {
			return nil, err
		}
		if resultInfo.Error == "" {
			return result, nil
		}
		return result, errors.New(resultInfo.Error)

	case <-ctx.Done(): // 超时
		_ = c.nats_client.Delete(rpcInfo.Cid)
		c.close_callback_chan(callback)
		return nil, fmt.Errorf("deadline exceeded")
	}
}

func (c *RPCClient) CallNR(ctx context.Context, _func string, params ...any) error {
	_ctx := ctx
	var argTypes []string = make([]string, len(params)+1)
	var argDatas [][]byte = make([][]byte, len(params)+1)

	// 检测添加log.TraceSpan到ctx
	span, ok := ctx.Value(log.RPC_CONTEXT_KEY_TRACE).(log.TraceSpan)
	if !ok {
		_ctx = mqrpc.ContextWithValue(_ctx, log.RPC_CONTEXT_KEY_TRACE, log.CreateRootTrace())
	} else {
		_ctx = mqrpc.ContextWithValue(_ctx, log.RPC_CONTEXT_KEY_TRACE, span.ExtractSpan())
	}

	// 重新组装参数(ctx放到首位)
	params = append([]any{_ctx}, params...)
	for k, arg := range params {
		var err error = nil
		argTypes[k], argDatas[k], err = mqrpc.ArgToData(arg)
		if err != nil {
			return fmt.Errorf("args[%d] error %s", k, err.Error())
		}
	}

	// CallNRArgs
	return c.CallNRArgs(ctx, _func, argTypes, argDatas)
}
func (c *RPCClient) CallNRArgs(ctx context.Context, _func string, argTypes []string, argDatas [][]byte) error {
	var err error
	caller, _ := os.Hostname()
	if ctx != nil {
		cr, ok := ctx.Value("caller").(string)
		if ok {
			caller = cr
		}
	}

	rpcInfo := &core.RPCInfo{
		Fn:       _func,
		Reply:    false,
		Expired:  (time.Now().UTC().Add(app.App().Options().RPCExpired).UnixNano()) / 1000000,
		Cid:      uuid.New().String(),
		Args:     argDatas,
		ArgsType: argTypes,
		Caller:   caller,
		Hostname: caller,
	}
	callInfo := &mqrpc.CallInfo{
		RPCInfo: rpcInfo,
	}

	defer func() { // 全局监控(调用方)
		if app.App().Config().RpcLog || err != nil { // 打印调用日志
			span, _ := ctx.Value(log.RPC_CONTEXT_KEY_TRACE).(log.TraceSpan)
			log.TInfo(span, "rpc CallNR ServerId = %v, Func = %v, Elapsed = %v, Result = <%T-val:%v>, Error = %v",
				c.nats_client.session.GetID(), _func, 0, nil, false, err)
		}
		if handle := app.App().Options().ClientRPCHandler; handle != nil {
			handle(*c.nats_client.session.GetNode(), rpcInfo, nil, err, 0)
		}
	}()
	err = c.nats_client.CallNR(callInfo)
	return err
}

func (c *RPCClient) close_callback_chan(ch chan *core.ResultInfo) {
	defer func() {
		if recover() != nil {
			// close(ch) panic occur
		}
	}()

	close(ch) // panic if ch is closed
}
