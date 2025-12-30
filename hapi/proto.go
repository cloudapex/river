package hapi

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudapex/river/mqrpc"
)

const (
	RPC_CONTEXT_KEY_HEADER = "rtx_header" // 定义需要RPC传输session的ContextKey
)

// 注册RPC_CONTEXT_KEY_HEADER
func init() {
	mqrpc.RegTranslatableCtxKey(RPC_CONTEXT_KEY_HEADER, func() mqrpc.IMarshaler {
		return HttpHead(nil)
	})
}

// header's value
type Pair struct {
	Key    string   `msgpack:"key" json:"key"`
	Values []string `msgpack:"values,omitempty" json:"values,omitempty"`
}

// A HTTP request as RPC, Forward by the api handler
type Request struct {
	Method string           `msgpack:"method" json:"method"`
	Path   string           `msgpack:"path" json:"path"`
	Header map[string]*Pair `msgpack:"header,omitempty" json:"header,omitempty"`
	Get    map[string]*Pair `msgpack:"get,omitempty" json:"get,omitempty"`
	Post   map[string]*Pair `msgpack:"post,omitempty" json:"post,omitempty"`
	Body   string           `msgpack:"body,omitempty" json:"body,omitempty"` // raw request body; if not application/x-www-form-urlencoded
	Url    string           `msgpack:"url,omitempty" json:"url,omitempty"`
}

// A HTTP response as RPC, Expected response for the api handler
type Response struct {
	StatusCode int32            `msgpack:"status_code" json:"status_code"`
	Header     map[string]*Pair `msgpack:"header,omitempty" json:"header,omitempty"`
	Body       string           `msgpack:"body,omitempty" json:"body,omitempty"`
}

// A HTTP event as RPC, Forwarded by the event handler
type Event struct {
	Name      string           `msgpack:"name" json:"name"`
	Id        string           `msgpack:"id" json:"id"`
	Timestamp int64            `msgpack:"timestamp" json:"timestamp"`
	Header    map[string]*Pair `msgpack:"header,omitempty" json:"header,omitempty"`
	Data      string           `msgpack:"data,omitempty" json:"data,omitempty"`
}

// 包装一下 Request.Header
func HttpHead(head map[string]*Pair) mqrpc.IMarshaler {
	if head != nil {
		return &httpHead{head: head}
	}
	return &httpHead{head: map[string]*Pair{}}
}

type httpHead struct {
	head map[string]*Pair
}

func (t *httpHead) Marshal() ([]byte, error) {
	bytes, err := json.Marshal(t.head)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}
func (t *httpHead) Unmarshal(bytes []byte) error {
	return json.Unmarshal(bytes, &t.head)
}
func (t *httpHead) String() string {
	return fmt.Sprintf("%v", t.head)
}

// get Header from context
func ContextValHeader(ctx context.Context) map[string]*Pair {
	out := map[string]*Pair{}
	val, ok := ctx.Value(RPC_CONTEXT_KEY_HEADER).(*httpHead)
	if !ok {
		return out
	}
	return val.head
}
