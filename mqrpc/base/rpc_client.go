package rpcbase

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cloudapex/river/app"
	"github.com/cloudapex/river/log"
	"github.com/cloudapex/river/mqrpc"
	"github.com/cloudapex/river/mqrpc/core"
	"github.com/cloudapex/river/tools/uuid"
)

type RPCClient struct {
	nats_client *NatsClient
}

func NewRPCClient(session app.IServerSession) (mqrpc.RPCClient, error) {
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
	// 检测是否含有log.TraceSpan
	span, ok := ctx.Value(log.CONTEXT_TRANSKEY_TRACE).(log.TraceSpan)
	if !ok {
		_ctx = mqrpc.ContextWithValue(_ctx, log.CONTEXT_TRANSKEY_TRACE, log.CreateRootTrace())
	} else {
		_ctx = mqrpc.ContextWithValue(_ctx, log.CONTEXT_TRANSKEY_TRACE, span.ExtractSpan())
	}

	params = append([]any{_ctx}, params...)
	for k, arg := range params {
		var err error = nil
		argTypes[k], argDatas[k], err = mqrpc.ArgToData(arg)
		if err != nil {
			return nil, fmt.Errorf("args[%d] error %s", k, err.Error())
		}
	}
	start := time.Now()
	r, err := c.CallArgs(ctx, _func, argTypes, argDatas)
	if app.Default().Config().RpcLog {
		span, _ := ctx.Value(log.CONTEXT_TRANSKEY_TRACE).(log.TraceSpan)
		log.TInfo(span, "rpc Call ServerId = %v Func = %v Elapsed = %v Result = %v ERROR = %v", c.nats_client.session.GetID(), _func, time.Since(start), r, err)
	}
	return r, err
}
func (c *RPCClient) CallArgs(ctx context.Context, _func string, argTypes []string, argDatas [][]byte) (any, error) {
	var err error
	var result any

	caller, _ := os.Hostname()
	if ctx != nil {
		cr, ok := ctx.Value("caller").(string)
		if ok {
			caller = cr
		}
	}

	start := time.Now()
	var correlation_id = uuid.Rand().Hex()
	rpcInfo := &core.RPCInfo{
		Fn:       _func,
		Reply:    true,
		Expired:  (start.UTC().Add(app.Default().Options().RPCExpired).UnixNano()) / 1000000,
		Cid:      correlation_id,
		Args:     argDatas,
		ArgsType: argTypes,
		Caller:   caller,
		Hostname: caller,
	}
	defer func() {
		//异常日志都应该打印
		if app.Default().Options().ClientRPChandler != nil {
			exec_time := time.Since(start).Nanoseconds()
			app.Default().Options().ClientRPChandler(*c.nats_client.session.GetNode(), rpcInfo, result, err, exec_time)
		}
	}()
	callInfo := &mqrpc.CallInfo{
		RPCInfo: rpcInfo,
	}
	callback := make(chan *core.ResultInfo, 1)
	//优先使用本地rpc
	//if c.local_client != nil {
	//	err = c.local_client.Call(*callInfo, callback)
	//} else
	err = c.nats_client.Call(callInfo, callback)
	if err != nil {
		return nil, err
	}

	// 没有设置超时的话使用默认超时
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, app.Default().Options().RPCExpired)
	}
	defer func() {
		if cancel != nil {
			cancel()
		}
	}()

	select {
	case resultInfo, ok := <-callback:
		if !ok {
			return nil, fmt.Errorf("client closed")
		}
		result, err = mqrpc.DataToArg(resultInfo.ResultType, resultInfo.Result)
		if err != nil {
			return nil, err
		}

		return result, fmt.Errorf(resultInfo.Error)
	case <-ctx.Done():
		_ = c.nats_client.Delete(rpcInfo.Cid)
		c.close_callback_chan(callback)
		return nil, fmt.Errorf("deadline exceeded")
		//case <-time.After(time.Second * time.Duration(app.App().GetSettings().rpc.RPCExpired)):
		//	close(callback)
		//	c.nats_client.Delete(rpcInfo.Cid)
		//	return nil, "deadline exceeded"
	}
}

func (c *RPCClient) CallNR(ctx context.Context, _func string, params ...any) (err error) {
	_ctx := ctx
	var argTypes []string = make([]string, len(params)+1)
	var argDatas [][]byte = make([][]byte, len(params)+1)
	// 检测是否含有log.TraceSpan
	span, ok := ctx.Value(log.CONTEXT_TRANSKEY_TRACE).(log.TraceSpan)
	if !ok {
		_ctx = mqrpc.ContextWithValue(_ctx, log.CONTEXT_TRANSKEY_TRACE, log.CreateRootTrace())
	} else {
		_ctx = mqrpc.ContextWithValue(_ctx, log.CONTEXT_TRANSKEY_TRACE, span.ExtractSpan())
	}
	params = append([]any{_ctx}, params...)
	for k, arg := range params {
		argTypes[k], argDatas[k], err = mqrpc.ArgToData(arg)
		if err != nil {
			return fmt.Errorf("args[%d] error %s", k, err.Error())
		}
	}
	start := time.Now()
	err = c.CallNRArgs(ctx, _func, argTypes, argDatas)
	if app.Default().Config().RpcLog {
		span, _ := ctx.Value(log.CONTEXT_TRANSKEY_TRACE).(log.TraceSpan)
		log.TInfo(span, "rpc CallNR ServerId = %v Func = %v Elapsed = %v ERROR = %v", c.nats_client.session.GetID(), _func, time.Since(start), err)
	}
	return err
}
func (c *RPCClient) CallNRArgs(ctx context.Context, _func string, argTypes []string, argDatas [][]byte) (err error) {
	caller, _ := os.Hostname()
	if ctx != nil {
		cr, ok := ctx.Value("caller").(string)
		if ok {
			caller = cr
		}
	}

	var correlation_id = uuid.Rand().Hex()
	rpcInfo := &core.RPCInfo{
		Fn:       _func,
		Reply:    false,
		Expired:  (time.Now().UTC().Add(app.Default().Options().RPCExpired).UnixNano()) / 1000000,
		Cid:      correlation_id,
		Args:     argDatas,
		ArgsType: argTypes,
		Caller:   caller,
		Hostname: caller,
	}
	callInfo := &mqrpc.CallInfo{
		RPCInfo: rpcInfo,
	}
	//优先使用本地rpc
	//if c.local_client != nil {
	//	err = c.local_client.CallNR(*callInfo)
	//} else
	return c.nats_client.CallNR(callInfo)
}

func (c *RPCClient) close_callback_chan(ch chan *core.ResultInfo) {
	defer func() {
		if recover() != nil {
			// close(ch) panic occur
		}
	}()

	close(ch) // panic if ch is closed
}
