package gatebase

import (
	"bufio"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/cloudapex/river/app"
	"github.com/cloudapex/river/gate"
	"github.com/cloudapex/river/log"
	"github.com/cloudapex/river/network"
	"github.com/cloudapex/river/tools"
	"github.com/cloudapex/river/tools/aes"
)

type agentBase struct {
	Impl gate.IAgent

	gate         gate.IGate
	session      gate.ISession
	conn         network.Conn
	r            *bufio.Reader
	w            *bufio.Writer
	ch           chan int        // 控制模块可同时开启的最大协程数(暂时没用)
	sendPackChan chan *gate.Pack // 需要发送的消息缓存
	isClosed     int32
	isShaked     int32
	recvNum      int64
	sendNum      int64
	connTime     time.Time
	lastError    error
}

func (this *agentBase) Init(impl gate.IAgent, gt gate.IGate, conn network.Conn) error {
	this.Impl = impl
	this.ch = make(chan int, gt.Options().ConcurrentTasks)
	this.conn = conn
	this.gate = gt
	this.r = bufio.NewReaderSize(conn, gt.Options().BufSize)
	this.w = bufio.NewWriterSize(conn, gt.Options().BufSize)

	this.isClosed = 0
	this.isShaked = 0
	this.recvNum = 0
	this.sendNum = 0
	this.sendPackChan = make(chan *gate.Pack, gt.Options().SendPackBuffNum)
	return nil
}
func (this *agentBase) Close() {
	go func() { // 关闭连接部分情况下会阻塞超时，因此放协程去处理
		if this.conn != nil {
			this.conn.Close()
		}
	}()
}
func (this *agentBase) OnClose() error {
	atomic.StoreInt32(&this.isClosed, 1)
	close(this.sendPackChan)
	this.gate.GetAgentLearner().DisConnect(this) //发送连接断开的事件
	return nil
}
func (this *agentBase) Destroy() { // 没用
	if this.conn != nil {
		this.conn.Destroy()
	}
}
func (this *agentBase) Run() (err error) {
	defer func() {
		if err := tools.Catch(recover()); err != nil {
			log.Error("agent.recvLoop() panic:%v", err)
		}
		this.Close()
	}()

	addr := this.conn.RemoteAddr()
	this.session, err = NewSessionByMap(map[string]any{
		"IP":        addr.String(),
		"Network":   addr.Network(),
		"SessionId": tools.GenerateID().String(),
		"ServerId":  this.gate.GetServerID(),
		"Settings":  make(map[string]string),
	})

	this.session.UpdTraceSpan() // 代码跟踪
	this.connTime = time.Now()
	this.isShaked = 1
	this.gate.GetAgentLearner().Connect(this) //发送连接成功的事件

	log.Info("gate create agent sessionId:%s, current gate agents num:%d", this.session.GetSessionID(), this.gate.GetDelegater().GetAgentNum())

	go this.sendLoop()     // 发送数据线程
	return this.recvLoop() // 接收数据线程
}

// ========== 属性方法

// ConnTime 建立连接的时间
func (this *agentBase) ConnTime() time.Time { return this.connTime }

// IsClosed 是否关闭了
func (this *agentBase) IsClosed() bool { return atomic.LoadInt32(&this.isClosed) == 1 }

// IsShaked 连接就绪(握手/认证...)
func (this *agentBase) IsShaked() bool { return this.isShaked == 1 }

// RecvNum 接收消息的数量
func (this *agentBase) RecvNum() int64 { return atomic.LoadInt64(&this.recvNum) }

// SendNum 发送消息的数量
func (this *agentBase) SendNum() int64 { return atomic.LoadInt64(&this.sendNum) }

// GetSession 管理的ClientSession
func (this *agentBase) GetSession() gate.ISession { return this.session }

// ========== 处理发送
func (this *agentBase) sendLoop() {
	defer func() {
		if err := tools.Catch(recover()); err != nil {
			log.Error("agent.sendLoop() panic:%v")
		}
		this.Close()
	}()

	for pack := range this.sendPackChan {
		atomic.AddInt64(&this.sendNum, 1)
		sendData := this.Impl.OnWriteEncodingPack(pack)
		this.conn.SetWriteDeadline(time.Now().Add(30 * time.Second))
		if _, err := this.conn.Write(sendData); err != nil {
			this.lastError = err
			log.Error("sendLoop, userId:%v sessionId:%v topic:%v dataLen:%v, err:%v", this.session.GetUserID(), this.session.GetSessionID(), pack.Topic, len(sendData), err)
		} else {
			log.Debug("sendLoop, userId:%v sessionId:%v topic:%v dataLen:%v ok.", this.session.GetUserID(), this.session.GetSessionID(), pack.Topic, len(sendData))
		}
	}
}

// SendPack 提供发送数据包的方法
func (this *agentBase) SendPack(pack *gate.Pack) error {
	if this.IsClosed() {
		return nil
	}

	if hook := this.gate.GetSendMessageHook(); hook != nil {
		bb, err := hook(this.GetSession(), pack.Topic, pack.Body)
		if err != nil {
			return err
		}
		pack.Body = bb
	}
	select {
	case this.sendPackChan <- pack:
		return nil
	default:
		return fmt.Errorf("too many unsent messages")
	}
}

// ========== 处理接收
func (this *agentBase) recvLoop() error {
	heartOverTime := this.gate.Options().HeartOverTimer
	for {
		nowTime := time.Now()
		if heartOverTime > 0 {
			_ = this.conn.SetReadDeadline(nowTime.Add(heartOverTime))
		}
		// this.recvWait()
		pack, err := this.Impl.OnReadDecodingPack()
		if err != nil {
			if heartOverTime > 0 && time.Since(nowTime) >= (heartOverTime) {
				log.Error("recvLoop heartOverTime, userId:%v sessionId:%v", this.session.GetSessionID(), this.session.GetUserID())
			} else {
				log.Error("recvLoop heartOverTime, userId:%v sessionId:%v err:%s", this.session.GetSessionID(), this.session.GetUserID(), err.Error())
			}
			this.lastError = err
			return err
		}

		if pack == nil {
			continue
		}
		atomic.AddInt64(&this.recvNum, 1)
		if route := this.gate.GetRouteHandler(); route != nil {
			done, err := route.OnRoute(this.GetSession(), pack.Topic, pack.Body)
			if err != nil {
				this.lastError = err
				return err
			}
			if done {
				continue
			}
		}
		if err := this.OnHandRecvPack(pack); err != nil {
			this.lastError = err
			return err
		}
		log.Debug("recvLoop, userId:%v sessionId:%v topic:%v dataLen:%v ok.", this.session.GetUserID(), this.session.GetSessionID(), pack.Topic, len(pack.Body))
	}
}
func (this *agentBase) recvWait() error {
	// 如果ch满了则会处于阻塞，从而达到限制最大协程的功能
	select {
	case this.ch <- 1:
	}
	return nil
}

func (this *agentBase) recvFinish() {
	// 完成则从ch推出数据
	select {
	case <-this.ch:
	default:
	}
}

// 自行实现如何处理收到的数据包
func (this *agentBase) OnHandRecvPack(pack *gate.Pack) error {
	// 处理保活(默认不处理保活,留给上层处理)

	// 默认是通过topic解析出路由规则
	topic := strings.Split(pack.Topic, "/")
	if len(topic) < 2 {
		return fmt.Errorf("pack.Topic resolving faild with:%v", pack.Topic)
	}
	moduleTyp, msgId := topic[0], topic[1]

	// 优先在已绑定的Module中提供服务
	serverId, _ := this.session.Get(moduleTyp)
	if serverId != "" {
		if server, _ := app.App().GetServerByID(serverId); server != nil {
			_, err := server.Call(this.session.GenRPCContext(), gate.RPC_CLIENT_MSG, msgId, pack.Body)
			return err
		}
	}

	// 然后按照默认路由规则随机取得Module服务
	server, err := app.App().GetRouteServer(moduleTyp)
	if err != nil {
		return fmt.Errorf("Service(moduleType:%s) not found", moduleTyp)
	}

	_, err = server.Call(this.session.GenRPCContext(), gate.RPC_CLIENT_MSG, msgId, pack.Body)
	return err
}

// 获取最后发生的错误
func (this *agentBase) GetError() error {
	if !this.IsClosed() {
		return nil
	}
	return this.lastError
}

// ========== Pack编码默认实现

// OnWriteEncodingPack 处理编码Pack后的数据用于发送
func (this *agentBase) OnWriteEncodingPack(pack *gate.Pack) []byte {
	// [普通不加密]
	// headLen := gate.PACK_HEAD_TOTAL_LEN_SIZE + gate.PACK_HEAD_MSG_ID_LEN_SIZE
	// totalLen := headLen + idLen + len(pack.Body)
	// sendData := make([]byte, headLen, totalLen)
	// binary.LittleEndian.PutUint16(sendData, uint16(totalLen))                              // for PACK_HEAD_TOTAL_LEN_SIZE
	// binary.LittleEndian.PutUint16(sendData[gate.PACK_HEAD_TOTAL_LEN_SIZE:], uint16(idLen)) // for PACK_HEAD_MSG_ID_LEN_SIZE
	// sendData = append(sendData, []byte(pack.Topic)...)
	// sendData = append(sendData, pack.Body...)
	// [end]

	idLen := len(pack.Topic)

	// 需要加密的数据: PACK_HEAD_MSG_ID_LEN_SIZE + msgId + msgData
	bodyLen := gate.PACK_HEAD_MSG_ID_LEN_SIZE + idLen + len(pack.Body)
	bodyData := make([]byte, bodyLen)
	binary.LittleEndian.PutUint16(bodyData, uint16(idLen))
	copy(bodyData[gate.PACK_HEAD_MSG_ID_LEN_SIZE:], []byte(pack.Topic))
	copy(bodyData[gate.PACK_HEAD_MSG_ID_LEN_SIZE+idLen:], []byte(pack.Body))

	// 处理加密: 先base加密 + 再ecb加密
	if this.gate.Options().EncryptKey != "" {
		b64Data := base64.StdEncoding.EncodeToString(bodyData)
		encryptedData, err := aes.AES_ECB_Encrypt([]byte(b64Data), []byte(this.gate.Options().EncryptKey))
		if err != nil {
			bodyData = []byte(err.Error())
			log.Error("AES_ECB_Encrypt err:%v", err)
		}
		bodyData = encryptedData
	}

	// 发送数据总长度: PACK_HEAD_TOTAL_LEN_SIZE + len(bodyData)
	totalLen := gate.PACK_HEAD_TOTAL_LEN_SIZE + len(bodyData)
	sendData := make([]byte, totalLen)
	binary.LittleEndian.PutUint16(sendData, uint16(totalLen))
	copy(sendData[gate.PACK_HEAD_TOTAL_LEN_SIZE:], bodyData)
	return sendData
}

// OnReadDecodingPack 从连接中读取数据并解码出Pack
func (this *agentBase) OnReadDecodingPack() (*gate.Pack, error) {
	panic("agentBase: OnReadDecodingPack() must be implemented")
}
