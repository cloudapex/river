// Package network tcp服务器
package network

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cloudapex/river/log"
)

// TCPServer tcp服务器
type TCPServer struct {
	Addr       string
	TLS        bool //是否支持tls
	CertFile   string
	KeyFile    string
	MaxConnNum int
	// mutexConns   sync.Mutex
	NewConnAgent func(*TCPConn) Agent
	ln           net.Listener
	wgLn         sync.WaitGroup
	wgConns      sync.WaitGroup
}

// Start 开始tcp监听
func (server *TCPServer) Start() {
	server.init()
	log.Info("TCP Listen :%s", server.Addr)
	go server.run()
}

func (server *TCPServer) init() {
	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		log.Error("%v", err)
		panic(fmt.Sprintf("TCPServer.Start.Listen err:%v", err))
	}

	if server.NewConnAgent == nil {
		log.Error("NewConnAgent must not be nil")
		panic(fmt.Sprintf("TCPServer.NewConnAgent must not be nil"))
	}
	if server.TLS {
		tlsConf := new(tls.Config)
		tlsConf.Certificates = make([]tls.Certificate, 1)
		tlsConf.Certificates[0], err = tls.LoadX509KeyPair(server.CertFile, server.KeyFile)
		if err == nil {
			ln = tls.NewListener(ln, tlsConf)
			log.Info("TCP Listen TLS load success")
		} else {
			log.Warning("tcp_server tls :%v", err)
		}
	}

	server.ln = ln
}
func (server *TCPServer) run() {
	server.wgLn.Add(1)
	defer server.wgLn.Done()

	var connNum int32
	var tempDelay time.Duration
	for {
		conn, err := server.ln.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				log.Info("accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return
		}
		tempDelay = 0
		current := atomic.LoadInt32(&connNum)
		if server.MaxConnNum > 0 && int(current) >= server.MaxConnNum {
			log.Warning("TCP Server reach max connection number:%d, current:%d", server.MaxConnNum, current)
			conn.Close()
			continue
		}
		atomic.AddInt32(&connNum, 1)

		tcpConn := newTCPConn(conn)
		agent := server.NewConnAgent(tcpConn)
		server.wgConns.Add(1)
		go func() {
			defer func() {
				atomic.AddInt32(&connNum, -1)
				server.wgConns.Done()
			}()
			agent.Run()

			// cleanup
			tcpConn.Close()
			agent.OnClose()
		}()
	}
}

// Close 关闭TCP监听
func (server *TCPServer) Close() {
	server.ln.Close()
	server.wgLn.Wait()
	server.wgConns.Wait()
}
