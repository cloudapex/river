package gatebase

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cloudapex/river/app"
	"github.com/cloudapex/river/conf"
	"github.com/cloudapex/river/gate"
	"github.com/cloudapex/river/module"
	"github.com/cloudapex/river/network"
)

var _ app.IRPCModule = &GateBase{}

type GateBase struct {
	module.ModuleBase

	opts gate.Options

	delegater    gate.IDelegater                       // ession管理接口
	agentCreater func(netTyp string) gate.IClientAgent // 创建客户端连接代理接口
	shakeHandle  func(r *http.Request) error           // 建立连接时鉴权(ws)

	storager        gate.StorageHandler     // Session持久化接口
	router          gate.RouteHandler       // 路由控制接口
	sessionLearner  gate.ISessionLearner    // 客户端连接和断开的监听器(业务使用)
	agentLearner    gate.IAgentLearner      // 客户端连接和断开的监听器(内部使用)
	recvPackHandler gate.FunRecvPackHandler // 接收数据包处理接口
	sendMessageHook gate.FunSendMessageHook // 发送消息时的钩子回调
}

func (this *GateBase) Init(subclass app.IRPCModule, settings *conf.ModuleSettings, opts ...gate.Option) {
	this.ModuleBase.Init(subclass, settings, this.opts.Opts...) // 这是必须的

	// 使用settings的配置覆盖opts
	this.opts = gate.NewOptions(opts...)
	if WSAddr, ok := settings.Settings["WSAddr"]; ok {
		this.opts.WsAddr = WSAddr.(string)
	}
	if TCPAddr, ok := settings.Settings["TCPAddr"]; ok {
		this.opts.TcpAddr = TCPAddr.(string)
	}

	if tls, ok := settings.Settings["TLS"]; ok {
		this.opts.TLS = tls.(bool)
	}
	if CertFile, ok := settings.Settings["CertFile"]; ok {
		this.opts.CertFile = CertFile.(string)
	}
	if KeyFile, ok := settings.Settings["KeyFile"]; ok {
		this.opts.KeyFile = KeyFile.(string)
	}

	if EncryptKey, ok := settings.Settings["EncryptKey"]; ok {
		this.opts.EncryptKey = EncryptKey.(string)
	}

	// for member
	delegate := NewDelegate(this)
	this.delegater = delegate
	this.agentLearner = delegate
	this.agentCreater = this.defaultClientAgentCreater
	this.recvPackHandler = this.defaultRecvPackHandler

	// for session
	this.GetServer().RegisterGO("Load", delegate.OnRpcLoad)
	this.GetServer().RegisterGO("Bind", delegate.OnRpcBind)
	this.GetServer().RegisterGO("UnBind", delegate.OnRpcUnBind)
	this.GetServer().RegisterGO("Push", delegate.OnRpcPush)
	this.GetServer().RegisterGO("Set", delegate.OnRpcSet)
	this.GetServer().RegisterGO("Del", delegate.OnRpcDel)
	this.GetServer().RegisterGO("Send", delegate.OnRpcSend)
	this.GetServer().RegisterGO("Connected", delegate.OnRpcConnected)
	this.GetServer().RegisterGO("Close", delegate.OnRpcClose)
	// for global
	this.GetServer().RegisterGO("Broadcast", delegate.OnRpcBroadcast)
}
func (this *GateBase) GetType() string { return "Gate" }

func (this *GateBase) Version() string { return "1.0.0" }

func (this *GateBase) Options() gate.Options { return this.opts }

func (this *GateBase) OnInit(settings *conf.ModuleSettings) {
	// 所有初始化逻辑都放到Init中, 重载OnInit不可调用基类OnInit!
	panic("GateBase: OnInit() must be implemented")
}
func (this *GateBase) OnDestroy() {
	this.ModuleBase.OnDestroy()
}
func (this *GateBase) OnAppConfigurationLoaded() {
	this.ModuleBase.OnAppConfigurationLoaded()
}
func (this *GateBase) OnConfChanged(settings *conf.ModuleSettings) {
	this.ModuleBase.OnConfChanged(settings)
}

func (this *GateBase) Run(closeSig chan bool) {
	// for wss
	var wsServer *network.WSServer
	if this.opts.WsAddr != "" {
		wsServer = new(network.WSServer)
		wsServer.Addr = this.opts.WsAddr
		wsServer.HTTPTimeout = 12 * time.Second
		wsServer.TLS = this.opts.TLS
		wsServer.CertFile = this.opts.CertFile
		wsServer.KeyFile = this.opts.KeyFile
		wsServer.ShakeFunc = this.shakeHandle
		wsServer.MaxMsgLen = uint32(this.opts.MaxPackSize)
		wsServer.NewAgent = func(conn *network.WSConn) network.Client {
			agent := this.agentCreater("ws")
			agent.Init(agent, this, conn)
			return agent
		}
	}
	// for tcp
	var tcpServer *network.TCPServer
	if this.opts.TcpAddr != "" {
		tcpServer = new(network.TCPServer)
		tcpServer.Addr = this.opts.TcpAddr
		tcpServer.TLS = this.opts.TLS
		tcpServer.CertFile = this.opts.CertFile
		tcpServer.KeyFile = this.opts.KeyFile
		tcpServer.NewAgent = func(conn *network.TCPConn) network.Client {
			agent := this.agentCreater("tcp")
			agent.Init(agent, this, conn)
			return agent
		}
	}

	if wsServer != nil {
		wsServer.Start()
	}
	if tcpServer != nil {
		tcpServer.Start()
	}
	<-closeSig
	if this.delegater != nil {
		this.delegater.OnDestroy()
	}
	if wsServer != nil {
		wsServer.Close()
	}
	if tcpServer != nil {
		tcpServer.Close()
	}
}

// --------------- AgentCreater

// SetAgentCreater 设置创建客户端Agent的函数
func (this *GateBase) SetAgentCreater(cfunc func(netTyp string) gate.IClientAgent) {
	this.agentCreater = cfunc
}

// defaultClientAgentCreater 默认的创建客户端连接Agent的方法
func (this *GateBase) defaultClientAgentCreater(netTyp string) gate.IClientAgent {
	switch netTyp {
	case "ws":
		return NewWSClientAgent(this.recvPackHandler)
	case "tcp":
		return NewTCPClientAgent(this.recvPackHandler)
	}
	return NewWSClientAgent(this.recvPackHandler) // default use ws
}

// --------------- Delegater

// SetDelegater 设置代理处理器
func (this *GateBase) SetDelegater(handler gate.IDelegater) error {
	this.delegater = handler
	return nil
}

// GetDelegater 获取代理处理器
func (this *GateBase) GetDelegater() gate.IDelegater { return this.delegater }

// --------------- ShakeHandler

// SetShakeHandler 设置建立连接时鉴权器(ws)
func (this *GateBase) SetShakeHandler(handler func(r *http.Request) error) error {
	this.shakeHandle = handler
	return nil
}

// GetShakeHandler 获取建立连接时鉴权器(ws)
func (this *GateBase) GetShakeHandler() func(r *http.Request) error { return this.shakeHandle }

// --------------- StorageHandler

// SetStorageHandler 设置Session信息持久化接口
func (this *GateBase) SetStorageHandler(storager gate.StorageHandler) error {
	this.storager = storager
	return nil
}

// GetStorageHandler 获取Session信息持久化接口
func (this *GateBase) GetStorageHandler() (storager gate.StorageHandler) { return this.storager }

// --------------- RouteHandler

// SetRouteHandler 设置路由接口
func (this *GateBase) SetRouteHandler(router gate.RouteHandler) error {
	this.router = router
	return nil
}

// GetRouteHandler 获取路由接口
func (this *GateBase) GetRouteHandler() gate.RouteHandler { return this.router }

// --------------- ISessionLearner

// SetSessionLearner 设置客户端连接和断开的监听器
func (this *GateBase) SetSessionLearner(learner gate.ISessionLearner) error {
	this.sessionLearner = learner
	return nil
}

// GetSessionLearner 获取客户端连接和断开的监听器
func (this *GateBase) GetSessionLearner() gate.ISessionLearner { return this.sessionLearner }

// --------------- IAgentLearner(内部用)

// SetAgentLearner 设置客户端连接和断开的监听器(内部用)
func (this *GateBase) SetAgentLearner(learner gate.IAgentLearner) error {
	this.agentLearner = learner
	return nil
}

// SetAgentLearner 获取客户端连接和断开的监听器(内部用)
func (this *GateBase) GetAgentLearner() gate.IAgentLearner { return this.agentLearner }

// --------------- FunRecvPackHandler

// SetRecvPackHandler 设置接收数据包处理接口
func (this *GateBase) SetRecvPackHandler(handler gate.FunRecvPackHandler) error {
	this.recvPackHandler = handler
	return nil
}

// defaultRecvPackHandler 默认接收数据包处理接口
func (this *GateBase) defaultRecvPackHandler(session gate.ISession, pack *gate.Pack) error {
	// 默认是通过topic解析出路由规则
	topic := strings.FieldsFunc(pack.Topic, func(r rune) bool {
		return r == '/' || r == '_'
	})
	if len(topic) < 2 {
		return fmt.Errorf("pack.Topic resolving faild with:%v", pack.Topic)
	}

	moduleTyp, msgId := topic[0], topic[1]

	// 优先在已绑定的Module中提供服务
	serverId, _ := session.Get(moduleTyp)
	if serverId != "" {
		return app.App().CallNR(session.GenRPCContext(), serverId, gate.RPC_CLIENT_MSG, msgId, pack.Body)
	}

	// 然后按照默认路由规则随机取得Module服务
	server, err := app.App().GetRouteServer(moduleTyp)
	if err != nil {
		return fmt.Errorf("Service(moduleType:%s) not found", moduleTyp)
	}

	return server.CallNR(session.GenRPCContext(), gate.RPC_CLIENT_MSG, msgId, pack.Body)
}

// --------------- FunSendMessageHook

// SetsendMessageHook 设置发送消息时的钩子回调
func (this *GateBase) SetSendMessageHook(hook gate.FunSendMessageHook) error {
	this.sendMessageHook = hook
	return nil
}

// GetSendMessageHook 获取发送消息时的钩子回调
func (this *GateBase) GetSendMessageHook() gate.FunSendMessageHook { return this.sendMessageHook }
