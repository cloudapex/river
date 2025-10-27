package rpcbase

import "github.com/cloudapex/river/mqrpc/core"

type ClinetCallInfo struct {
	correlation_id string
	timeout        int64 //超时
	call           chan *core.ResultInfo
}
