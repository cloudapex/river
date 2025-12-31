package hapibase

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/cloudapex/river/hapi"
	"github.com/cloudapex/river/mqrpc"
	"github.com/cloudapex/river/tools/aes"
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
func (h *HttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req, err := RequestToProto(r)
	if err != nil {
		er := hapi.InternalServerError("httpgateway", "request parse failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(er.Error()))
		return
	}

	// 解密请求数据
	if err = h.decryptRequest(req); err != nil {
		er := hapi.InternalServerError("httpgateway", "decrypt request failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(er.Error()))
		return
	}

	service, err := h.Opts.Route(r)
	if err != nil {
		er := hapi.InternalServerError("httpgateway", "service route failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(er.Error()))
		return
	}
	rsp := &hapi.Response{}
	if err = h.Opts.RpcHandle(service, req, rsp); err != nil {
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

	// 加密应答数据
	if err = h.encryptResponse(req, rsp); err != nil {
		er := hapi.InternalServerError("httpgateway", "encrypt response failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(er.Error()))
		return
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

// internal RPCHandler
func (h *HttpHandler) callRpcService(service *hapi.Service, req *hapi.Request, rsp *hapi.Response) error {
	return mqrpc.MsgPack(rsp, mqrpc.RpcResult(service.Server.GetRPC().Call(context.TODO(), service.Topic, req)))
}

// decryptRequest ...
func (h *HttpHandler) decryptRequest(req *hapi.Request) error {
	debugKey := ""
	if pair, ok := req.Header[hapi.HTTP_HEAD_KEY_DEBUG_KEY]; ok && len(pair.Values) > 0 {
		debugKey = pair.Values[0]
	}

	if h.Opts.EncryptKey == "" || debugKey == h.Opts.DebugKey {
		return nil
	}

	// 解密 Body 字段
	if req.Body != "" {
		// 先进行 base64 解码
		decodedBody, err := base64.StdEncoding.DecodeString(req.Body)
		if err != nil {
			return fmt.Errorf("base64 decode request body failed: %v", err)
		}

		decryptedBody, err := aes.AES_ECB_Decrypt(decodedBody, []byte(h.Opts.EncryptKey))
		if err != nil {
			return fmt.Errorf("decrypt request body failed: %v", err)
		}
		req.Body = string(decryptedBody)
	}

	return nil
}

// encryptResponse ...
func (h *HttpHandler) encryptResponse(req *hapi.Request, rsp *hapi.Response) error {
	debugKey := ""
	if pair, ok := req.Header[hapi.HTTP_HEAD_KEY_DEBUG_KEY]; ok && len(pair.Values) > 0 {
		debugKey = pair.Values[0]
	}

	if h.Opts.EncryptKey == "" || debugKey == h.Opts.DebugKey {
		return nil
	}

	// 加密 Body 字段
	if rsp.Body != "" {
		encryptedBody, err := aes.AES_ECB_Encrypt([]byte(rsp.Body), []byte(h.Opts.EncryptKey))
		if err != nil {
			return fmt.Errorf("encrypt response body failed: %v", err)
		}
		rsp.Body = string(encryptedBody)
	}

	return nil
}
