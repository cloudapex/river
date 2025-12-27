package hapibase

import (
	"context"
	"net/http"

	"github.com/cloudapex/river/hapi"
	"github.com/cloudapex/river/mqrpc"
)

// NewHandler 创建常规http handler
func NewHandler(opts hapi.Options) *HttpHandler {
	return &HttpHandler{Opts: opts}
}

// HttpHandler 网关handler
type HttpHandler struct {
	Opts hapi.Options
}

// API handler is the default handler which takes api.Request and returns api.Response
func (a *HttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req, err := RequestToProto(r)
	if err != nil {
		er := hapi.InternalServerError("httpgateway", err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(er.Error()))
		return
	}
	service, err := a.Opts.Route(r)
	if err != nil {
		er := hapi.InternalServerError("httpgateway", err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(er.Error()))
		return
	}
	rsp := &hapi.Response{}
	if err = a.Opts.RpcHandle(service, req, rsp); err != nil {
		w.Header().Set("Content-Type", "application/json")
		ce := hapi.ParseError(err.Error())
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

	for _, header := range rsp.Header {
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
func (a *HttpHandler) callRpcService(service *hapi.Service, req *hapi.Request, rsp *hapi.Response) error {
	return mqrpc.MsgPack(rsp, mqrpc.RpcResult(service.SrvSession.Call(context.TODO(), service.Topic, req)))
}
