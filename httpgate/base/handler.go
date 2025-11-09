package httpgatebase

import (
	"context"
	"net/http"

	"github.com/cloudapex/river/httpgate"
	"github.com/cloudapex/river/mqrpc"
)

// NewHandler 创建网关
func NewHandler(opts httpgate.Options) http.Handler {
	h := &HttpHandler{Opts: opts}
	if opts.RpcHandle == nil {
		opts.RpcHandle = h.handleRpc
	}
	return h
}

// HttpHandler 网关handler
type HttpHandler struct {
	Opts httpgate.Options
}

// API handler is the default handler which takes api.Request and returns api.Response
func (a *HttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req, err := RequestToProto(r)
	if err != nil {
		er := httpgate.InternalServerError("httpgateway", err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(er.Error()))
		return
	}
	service, err := a.Opts.Route(r)
	if err != nil {
		er := httpgate.InternalServerError("httpgateway", err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(er.Error()))
		return
	}
	rsp := &httpgate.Response{}
	if err = a.Opts.RpcHandle(service, req, rsp); err != nil {
		w.Header().Set("Content-Type", "application/json")
		ce := httpgate.ParseError(err.Error())
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
func (a *HttpHandler) handleRpc(service *httpgate.Service, req *httpgate.Request, rsp *httpgate.Response) error {
	ctx, cancel := context.WithTimeout(context.TODO(), a.Opts.TimeOut)
	defer cancel()
	return mqrpc.MsgPack(rsp, mqrpc.RpcResult(service.SrvSession.Call(ctx, service.Hander, req)))
}
