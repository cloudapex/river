package rpcbase

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/cloudapex/river/app"
	"github.com/cloudapex/river/log"
	"github.com/cloudapex/river/mqrpc"
	"github.com/cloudapex/river/mqrpc/core"
	"github.com/nats-io/nats.go"
	"github.com/vmihailenco/msgpack/v5"
)

type NatsServer struct {
	call_chan chan mqrpc.CallInfo
	addr      string
	server    *RPCServer
	done      chan bool
	stopeds   chan bool
	subs      *nats.Subscription
	isClose   bool
}

func setAddrs(addrs []string) []string {
	var cAddrs []string
	for _, addr := range addrs {
		if len(addr) == 0 {
			continue
		}
		if !strings.HasPrefix(addr, "nats://") {
			addr = "nats://" + addr
		}
		cAddrs = append(cAddrs, addr)
	}
	if len(cAddrs) == 0 {
		cAddrs = []string{nats.DefaultURL}
	}
	return cAddrs
}

func NewNatsServer(s *RPCServer) (*NatsServer, error) {
	server := new(NatsServer)
	server.server = s
	server.done = make(chan bool)
	server.stopeds = make(chan bool)
	server.isClose = false
	server.addr = nats.NewInbox()
	go func() {
		server.on_request_handle()
		safeClose(server.stopeds)
	}()
	return server, nil
}
func (s *NatsServer) Addr() string {
	return s.addr
}

func safeClose(ch chan bool) {
	defer func() {
		if recover() != nil {
			// close(ch) panic occur
		}
	}()

	close(ch) // panic if ch is closed
}

/*
*
注销消息队列
*/
func (s *NatsServer) Shutdown() (err error) {
	safeClose(s.done)
	s.isClose = true
	select {
	case <-s.stopeds:
		//等待nats注销完成
	}
	return
}

func (s *NatsServer) Callback(callinfo *mqrpc.CallInfo) error {
	body, err := s.MarshalResult(callinfo.Result)
	if err != nil {
		return err
	}
	reply_to := callinfo.Props["reply_to"].(string)
	return app.Default().Transporter().Publish(reply_to, body)
}

/*
*
接收请求信息
*/
func (s *NatsServer) on_request_handle() (err error) {
	defer func() {
		if r := recover(); r != nil {
			var rn = ""
			switch r.(type) {

			case string:
				rn = r.(string)
			case error:
				rn = r.(error).Error()
			}
			buf := make([]byte, 1024)
			l := runtime.Stack(buf, false)
			errstr := string(buf[:l])
			log.Error("%s\n ----Stack----\n%s", rn, errstr)
			fmt.Println(errstr)
		}
	}()
	s.subs, err = app.Default().Transporter().SubscribeSync(s.addr)
	if err != nil {
		return err
	}

	go func() {
		select {
		case <-s.done:
			//服务关闭
		}
		s.subs.Unsubscribe()
	}()

	for !s.isClose {
		m, err := s.subs.NextMsg(time.Minute)
		if err != nil && err == nats.ErrTimeout {
			//fmt.Println(err.Error())
			//log.Warning("NatsServer error with '%v'",err)
			if !s.subs.IsValid() {
				//订阅已关闭，需要重新订阅
				s.subs, err = app.Default().Transporter().SubscribeSync(s.addr)
				if err != nil {
					log.Error("NatsServer SubscribeSync[1] error with '%v'", err)
					continue
				}
			}
			continue
		} else if err != nil {
			log.Warning("NatsServer error with '%v'", err)
			if !s.subs.IsValid() {
				//订阅已关闭，需要重新订阅
				s.subs, err = app.Default().Transporter().SubscribeSync(s.addr)
				if err != nil {
					log.Error("NatsServer SubscribeSync[2] error with '%v'", err)
					continue
				}
			}
			continue
		}

		rpcInfo, err := s.Unmarshal(m.Data)
		if err == nil {
			callInfo := &mqrpc.CallInfo{
				RPCInfo: rpcInfo,
			}
			callInfo.Props = map[string]any{
				"reply_to": rpcInfo.ReplyTo,
			}

			callInfo.Agent = s //设置代理为NatsServer

			s.server.Call(callInfo)
		} else {
			fmt.Println("error ", err)
		}
	}
	return nil
}

func (s *NatsServer) Unmarshal(data []byte) (*core.RPCInfo, error) {
	//fmt.Println(msg)
	//保存解码后的数据，Value可以为任意数据类型
	var rpcInfo core.RPCInfo
	err := msgpack.Unmarshal(data, &rpcInfo)
	if err != nil {
		return nil, err
	} else {
		return &rpcInfo, err
	}
}

// goroutine safe
func (s *NatsServer) MarshalResult(resultInfo *core.ResultInfo) ([]byte, error) {
	//log.Error("",map2)
	b, err := msgpack.Marshal(resultInfo)
	return b, err
}
