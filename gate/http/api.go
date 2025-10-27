// Package httpgateway provides an http-rpc handler which provides the entire http request over rpc
package httpgate

import (
	"context"
	"net/http"

	"github.com/cloudapex/river/app"
	httpgateapi "github.com/cloudapex/river/gate/http/api"
	"github.com/cloudapex/river/gate/http/errors"
	go_api "github.com/cloudapex/river/gate/http/proto"
	"github.com/cloudapex/river/mqrpc"
)

// APIHandler 网关handler
type APIHandler struct {
	Opts Options
	App  app.IApp
}

// API handler is the default handler which takes api.Request and returns api.Response
func (a *APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	request, err := httpgateapi.RequestToProto(r)
	if err != nil {
		er := errors.InternalServerError("httpgateway", err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(er.Error()))
		return
	}
	server, err := a.Opts.Route(r)
	if err != nil {
		er := errors.InternalServerError("httpgateway", err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(er.Error()))
		return
	}
	rsp := &go_api.Response{}
	ctx, _ := context.WithTimeout(context.TODO(), a.Opts.TimeOut)
	if err = mqrpc.MsgPack(rsp, mqrpc.RpcResult(server.SrvSession.Call(ctx, server.Hander, request))); err != nil {
		w.Header().Set("Content-Type", "application/json")
		ce := errors.Parse(err.Error())
		switch ce.Code {
		case 0:
			w.WriteHeader(500)
		default:
			w.WriteHeader(int(ce.Code))
		}
		_, err = w.Write([]byte(ce.Error()))
		return
	} else if rsp.StatusCode == 0 {
		rsp.StatusCode = http.StatusOK
	}

	for _, header := range rsp.GetHeader() {
		for _, val := range header.Values {
			w.Header().Add(header.Key, val)
		}
	}

	if len(w.Header().Get("Content-Type")) == 0 {
		w.Header().Set("Content-Type", "application/json")
	}

	w.WriteHeader(int(rsp.StatusCode))
	w.Write([]byte(rsp.Body))
}

// NewHandler 创建网关
func NewHandler(opts ...Option) http.Handler {
	options := NewOptions(opts...)
	return &APIHandler{
		Opts: options,
	}
}
