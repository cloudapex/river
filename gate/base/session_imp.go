package gatebase

type SessionImp struct {
	IP        string            `msgpack:"ip,omitempty" json:"ip,omitempty"`
	Network   string            `msgpack:"network,omitempty" json:"network,omitempty"`
	UserId    string            `msgpack:"user_id,omitempty" json:"user_id,omitempty"`
	SessionId string            `msgpack:"session_id,omitempty" json:"session_id,omitempty"`
	ServerId  string            `msgpack:"server_id,omitempty" json:"server_id,omitempty"`
	TraceId   string            `msgpack:"trace_id,omitempty" json:"trace_id,omitempty"`
	SpanId    string            `msgpack:"span_id,omitempty" json:"span_id,omitempty"`
	Settings  map[string]string `msgpack:"settings,omitempty" json:"settings,omitempty"`
	Carrier   map[string]string `msgpack:"carrier,omitempty" json:"carrier,omitempty"`
	Topic     string            `msgpack:"topic,omitempty" json:"topic,omitempty"`
}
