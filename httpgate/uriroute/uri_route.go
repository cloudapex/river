package uriroute

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/cloudapex/river/app"
	"github.com/cloudapex/river/gate"
	"github.com/cloudapex/river/log"
	"github.com/cloudapex/river/mqrpc"
	"github.com/pkg/errors"
)

// FSelector 服务节点选择函数，可以自定义服务筛选规则
// 如不指定,默认使用 Scheme作为moduleType,Hostname作为服务节点nodeId
// 如随机到服务节点Hostname可以用modulus,cache,random等通用规则
// 例如:
// im://modulus/remove_feeds_member?msg_id=1002
type FSelector func(session gate.ISession, topic string, u *url.URL) (s app.IServerSession, err error)

// FDataParsing 指定数据解析函数
// 返回值如bean！=nil err==nil则会向后端模块传入 func(session,bean)(result, error)
// 否则使用json或[]byte传参
type FDataParsing func(topic string, u *url.URL, msg []byte) (bean any, err error)

// Option Option
type Option func(*URIRoute)

// NewURIRoute NewURIRoute
func NewURIRoute(module app.IRPCModule, opts ...Option) *URIRoute {
	route := &URIRoute{
		module:      module,
		CallTimeOut: app.App().Options().RPCExpired,
	}
	for _, o := range opts {
		o(route)
	}
	return route
}

// Selector Selector
func Selector(t FSelector) Option {
	return func(o *URIRoute) {
		o.Selector = t
	}
}

// DataParsing DataParsing
func DataParsing(t FDataParsing) Option {
	return func(o *URIRoute) {
		o.DataParsing = t
	}
}

// CallTimeOut CallTimeOut
func CallTimeOut(t time.Duration) Option {
	return func(o *URIRoute) {
		o.CallTimeOut = t
	}
}

// URIRoute URIRoute
type URIRoute struct {
	module      app.IRPCModule
	Selector    FSelector
	DataParsing FDataParsing
	CallTimeOut time.Duration
}

// OnRoute OnRoute
func (u *URIRoute) OnRoute(session gate.ISession, topic string, msg []byte) (bool, any, error) {
	needreturn := true
	uu, err := url.Parse(topic)
	if err != nil {
		return needreturn, nil, errors.Errorf("topic is not uri %v", err.Error())
	}
	var argTypes []string = nil
	var argDatas [][]byte = nil

	_func := uu.Path
	m, err := url.ParseQuery(uu.RawQuery)
	if err != nil {
		return needreturn, nil, errors.Errorf("parse query error %v", err.Error())
	}
	if _, ok := m["msg_id"]; !ok {
		needreturn = false
	}
	argTypes = make([]string, 2)
	argDatas = make([][]byte, 2)
	//session.SetTopic(topic)
	var serverSession app.IServerSession
	if u.Selector != nil {
		ss, err := u.Selector(session, topic, uu)
		if err != nil {
			return needreturn, nil, err
		}
		serverSession = ss
	} else {
		moduleType := uu.Scheme
		if uu.Hostname() == "modulus" {
			//取模
		} else if uu.Hostname() == "cache" {
			//缓存
		} else if uu.Hostname() == "random" {
			//随机
		} else {
			//其他规则就是 module://[user:pass@]nodeId/path
			moduleType = fmt.Sprintf("%v@%v", moduleType, uu.Hostname())
		}
		ss, err := u.module.GetRouteServer(moduleType)
		if err != nil {
			return needreturn, nil, errors.Errorf("Service(type:%s) not found", moduleType)
		}
		serverSession = ss
	}

	if u.DataParsing != nil {
		bean, err := u.DataParsing(topic, uu, msg)
		if err == nil && bean != nil {
			if needreturn {
				ctx, cancel := context.WithTimeout(context.TODO(), u.CallTimeOut)
				defer cancel()
				result, e := serverSession.Call(ctx, _func, session, bean)
				if e != nil {
					return needreturn, result, e
				}
				return needreturn, result, nil
			}

			e := serverSession.CallNR(context.TODO(), _func, session, bean)
			if e != nil {
				log.Warning("Gate rpc", e.Error())
				return needreturn, nil, e
			}

			return needreturn, nil, nil
		}
	}

	// 默认参数
	if len(msg) > 0 && msg[0] == '{' && msg[len(msg)-1] == '}' {
		//尝试解析为json为map
		var obj any // var obj map[string]any
		err := json.Unmarshal(msg, &obj)
		if err != nil {
			return needreturn, nil, errors.Errorf("The JSON format is incorrect %v", err)
		}
		argTypes[1] = mqrpc.JSMAP
		argDatas[1] = msg
	} else {
		argTypes[1] = mqrpc.BYTES
		argDatas[1] = msg
	}
	s := session.Clone()
	//s.SetTopic(topic)

	ctx := context.Background()
	ctx = mqrpc.ContextWithValue(ctx, gate.CONTEXT_TRANSKEY_SESSION, s)
	ctx = mqrpc.ContextWithValue(ctx, log.CONTEXT_TRANSKEY_TRACE, session.GetTraceSpan())
	argTypes[0], argDatas[0], err = mqrpc.ArgToData(ctx)
	if err != nil {
		return needreturn, nil, err
	}
	if needreturn {
		ctx, cancel := context.WithTimeout(ctx, u.CallTimeOut)
		defer cancel()
		result, e := serverSession.CallArgs(ctx, _func, argTypes, argDatas)
		if e != nil {
			return needreturn, result, e
		}
		return needreturn, result, nil
	}

	e := serverSession.CallNRArgs(ctx, _func, argTypes, argDatas)
	if e != nil {
		log.Warning("Gate rpc", e.Error())
		return needreturn, nil, e
	}

	return needreturn, nil, nil
}
