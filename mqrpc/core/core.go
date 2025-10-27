package core

type RPCInfo struct {
	Cid      string   `msgpack:"cid" json:"cid"`                               // 调用ID
	Fn       string   `msgpack:"fn" json:"fn"`                                 // 函数名
	ReplyTo  string   `msgpack:"reply_to,omitempty" json:"reply_to,omitempty"` // 回复地址
	Track    string   `msgpack:"track,omitempty" json:"track,omitempty"`       // 跟踪信息
	Expired  int64    `msgpack:"expired" json:"expired"`                       // 过期时间
	Reply    bool     `msgpack:"reply" json:"reply"`                           // 是否需要回复
	ArgsType []string `msgpack:"args_type" json:"args_type"`                   // 参数类型列表
	Args     [][]byte `msgpack:"args" json:"args"`                             // 参数数据
	Caller   string   `msgpack:"caller,omitempty" json:"caller,omitempty"`     // 调用者
	Hostname string   `msgpack:"hostname,omitempty" json:"hostname,omitempty"` // 主机名
}

type ResultInfo struct {
	Cid        string `msgpack:"cid" json:"cid"`                                     // 调用ID
	Error      string `msgpack:"error,omitempty" json:"error,omitempty"`             // 错误信息
	ResultType string `msgpack:"result_type,omitempty" json:"result_type,omitempty"` // 结果类型
	Result     []byte `msgpack:"result,omitempty" json:"result,omitempty"`           // 结果数据
}
