package proto

type Pair struct {
	Key    string   `msgpack:"key" json:"key"`
	Values []string `msgpack:"values,omitempty" json:"values,omitempty"`
}

// A HTTP request as RPC
// Forward by the api handler
type Request struct {
	Method string           `msgpack:"method" json:"method"`
	Path   string           `msgpack:"path" json:"path"`
	Header map[string]*Pair `msgpack:"header,omitempty" json:"header,omitempty"`
	Get    map[string]*Pair `msgpack:"get,omitempty" json:"get,omitempty"`
	Post   map[string]*Pair `msgpack:"post,omitempty" json:"post,omitempty"`
	Body   string           `msgpack:"body,omitempty" json:"body,omitempty"` // raw request body; if not application/x-www-form-urlencoded
	Url    string           `msgpack:"url,omitempty" json:"url,omitempty"`
}

// A HTTP response as RPC
// Expected response for the api handler
type Response struct {
	StatusCode int32            `msgpack:"status_code" json:"status_code"`
	Header     map[string]*Pair `msgpack:"header,omitempty" json:"header,omitempty"`
	Body       string           `msgpack:"body,omitempty" json:"body,omitempty"`
}

// A HTTP event as RPC
// Forwarded by the event handler
type Event struct {
	Name      string           `msgpack:"name" json:"name"`
	Id        string           `msgpack:"id" json:"id"`
	Timestamp int64            `msgpack:"timestamp" json:"timestamp"`
	Header    map[string]*Pair `msgpack:"header,omitempty" json:"header,omitempty"`
	Data      string           `msgpack:"data,omitempty" json:"data,omitempty"`
}
