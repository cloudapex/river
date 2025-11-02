package network

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cloudapex/river/log"
	"github.com/gorilla/websocket"
)

// WSHandler websocket 处理器
type WSHandler struct {
	newConnAgent func(*WSConn) Client
}

func (handler *WSHandler) work(conn *websocket.Conn, r *http.Request) {
	wsConn := newWSConn(conn, r)
	agent := handler.newConnAgent(wsConn)
	agent.Run()

	// cleanup
	wsConn.Close()
	agent.OnClose()
}

// WSServer websocket服务器
type WSServer struct {
	Addr        string
	TLS         bool //是否支持tls
	CertFile    string
	KeyFile     string
	MaxConnNum  int
	MaxMsgLen   uint32
	HTTPTimeout time.Duration
	NewAgent    func(*WSConn) Client
	ln          net.Listener
	handler     *WSHandler
	ShakeFunc   func(r *http.Request) error
	wgConns     sync.WaitGroup
}

// Start 开启监听websocket端口
func (server *WSServer) Start() {
	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		log.Error("%v", err)
		panic(fmt.Sprintf("WSServer.Start.Listen err:%v", err))
	}

	if server.HTTPTimeout <= 0 {
		server.HTTPTimeout = 10 * time.Second
		log.Warning("invalid HTTPTimeout, reset to %v", server.HTTPTimeout)
	}
	if server.NewAgent == nil {
		log.Error("NewConnAgent must not be nil")
		panic(fmt.Sprintf("TCPServer.NewConnAgent must not be nil"))
	}
	if server.TLS {
		tlsConf := new(tls.Config)
		tlsConf.Certificates = make([]tls.Certificate, 1)
		tlsConf.Certificates[0], err = tls.LoadX509KeyPair(server.CertFile, server.KeyFile)
		if err == nil {
			ln = tls.NewListener(ln, tlsConf)
			log.Info("WS Listen TLS load success")
		} else {
			log.Warning("ws_server tls :%v", err)
		}
	}
	server.ln = ln
	server.handler = &WSHandler{
		newConnAgent: server.NewAgent,
	}

	// upgrader connect
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024 * 5,
		WriteBufferSize: 1024 * 5,
		// 开启跨域
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	var connNum int32 // 连接计数器
	httpHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if server.ShakeFunc != nil {
			if err := server.ShakeFunc(r); err != nil {
				http.Error(w, "Handshake error", http.StatusBadRequest)
				return
			}
		}

		current := atomic.LoadInt32(&connNum)
		if server.MaxConnNum > 0 && int(current) >= server.MaxConnNum {
			log.Warning("WS Server reach max connection number:%d, current:%d", server.MaxConnNum, current)
			http.Error(w, "reach max connection num", http.StatusServiceUnavailable)
			return
		}

		// 使用 websocket.Upgrader 升级为 WebSocket 连接
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, fmt.Sprintf("clientAcceptor 提升为 websocket 失败, %v", err), http.StatusBadRequest)
			return
		}

		// 原子增加连接数
		atomic.AddInt32(&connNum, 1)
		server.wgConns.Add(1)

		// 设置 WebSocket 配置参数
		conn.SetReadLimit(int64(server.MaxMsgLen))
		conn.SetWriteDeadline(time.Now().Add(server.HTTPTimeout))
		conn.SetReadDeadline(time.Now().Add(server.HTTPTimeout))

		// 处理 WebSocket 连接
		go func() {
			defer func() {
				atomic.AddInt32(&connNum, -1)
				server.wgConns.Done()
			}()
			server.handler.work(conn, r)
		}()
	}

	httpServer := &http.Server{
		Addr:           server.Addr,
		Handler:        http.HandlerFunc(httpHandler),
		ReadTimeout:    server.HTTPTimeout,
		WriteTimeout:   server.HTTPTimeout,
		MaxHeaderBytes: 1024,
	}
	log.Info("WS Listen :%s", server.Addr)
	go httpServer.Serve(ln)
}

// Close 停止监听websocket端口
func (server *WSServer) Close() {
	server.ln.Close()

	server.wgConns.Wait()
}
