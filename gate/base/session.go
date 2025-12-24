// Package basegate gate.Session
package gatebase

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/cloudapex/river/app"
	"github.com/cloudapex/river/gate"
	"github.com/cloudapex/river/log"
	"github.com/cloudapex/river/mqrpc"
	"github.com/cloudapex/river/tools"
	"github.com/vmihailenco/msgpack/v5"
)

func init() {
	mqrpc.RegTransContextKey(gate.CONTEXT_TRANSKEY_SESSION, func() mqrpc.Marshaler {
		s, _ := NewSessionByMap(map[string]any{})
		return s
	})
}

// NewSession data必须要有效
func NewSession(data []byte) (gate.ISession, error) {
	agent := &sessionAgent{
		lock: new(sync.RWMutex),
	}
	err := agent.initByDat(data)
	if err != nil {
		return nil, err
	}
	if agent.session.Settings == nil {
		agent.session.Settings = make(map[string]string)
	}
	return agent, nil
}

// NewSessionByMap data可以为空则构造一个空的Session
func NewSessionByMap(data map[string]any) (gate.ISession, error) {
	agent := &sessionAgent{
		session: new(SessionImp),
		lock:    new(sync.RWMutex),
	}
	err := agent.initByMap(data)
	if err != nil {
		return nil, err
	}
	if agent.session.Settings == nil {
		agent.session.Settings = make(map[string]string)
	}
	return agent, nil
}

type sessionAgent struct {
	session  *SessionImp
	lock     *sync.RWMutex // for session.UserId and session.Setting
	userdata any
	// guestJudger func(session gate.ISession) bool
}

func (s *sessionAgent) initByDat(data []byte) error {
	se := &SessionImp{}
	err := msgpack.Unmarshal(data, se)
	if err != nil {
		return err
	}
	s.session = se
	return nil
}
func (s *sessionAgent) initByMap(datas map[string]any) error {
	if uid := datas["UserId"]; uid != nil {
		s.session.UserId = uid.(string)
	}
	if ip := datas["IP"]; ip != nil {
		s.session.IP = ip.(string)
	}
	if network := datas["Network"]; network != nil {
		s.session.Network = network.(string)
	}
	if sessionId := datas["SessionId"]; sessionId != nil {
		s.session.SessionId = sessionId.(string)
	}
	if serverId := datas["ServerId"]; serverId != nil {
		s.session.ServerId = serverId.(string)
	}
	if settings := datas["Settings"]; settings != nil {
		s.lock.Lock()
		s.session.Settings = settings.(map[string]string)
		s.lock.Unlock()
	}
	return nil
}

func (s *sessionAgent) GetIP() string        { return s.session.IP }
func (s *sessionAgent) GetNetwork() string   { return s.session.Network }
func (s *sessionAgent) GetSessionID() string { return s.session.SessionId }
func (s *sessionAgent) GetServerID() string  { return s.session.ServerId }
func (s *sessionAgent) GetUserData() any     { return s.userdata }

func (s *sessionAgent) GetUserID() string {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.session.UserId
}
func (s *sessionAgent) GetUserIDUint() uint64 {
	uid, err := strconv.ParseUint(s.GetUserID(), 10, 64)
	if err != nil {
		return 0
	}
	return uid
}
func (s *sessionAgent) SetUserID(userId string) {
	s.lock.Lock()
	s.session.UserId = userId
	s.lock.Unlock()
}
func (s *sessionAgent) Get(key string) (string, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if s.session.Settings == nil {
		return "", false
	}
	if result, ok := s.session.Settings[key]; ok {
		return result, ok
	} else {
		return "", false
	}
}
func (s *sessionAgent) Set(key, value string) error {
	s.lock.Lock()
	s.session.Settings[key] = value
	s.lock.Unlock()
	return nil
}
func (s *sessionAgent) Del(key string) error {
	s.lock.Lock()
	delete(s.session.Settings, key)
	s.lock.Unlock()
	return nil
}
func (s *sessionAgent) SetSettings(settings map[string]string) {
	s.lock.Lock()
	s.session.Settings = settings
	s.lock.Unlock()
}

// 合并两个map 并且以 s.Settings 已有的优先
func (s *sessionAgent) ImportSettings(settings map[string]string) error {
	s.lock.Lock()
	if s.session.Settings == nil {
		s.session.Settings = settings
	} else {
		for k, v := range settings {
			if _, ok := s.session.Settings[k]; ok {
				//不用替换
			} else {
				s.session.Settings[k] = v
			}
		}
	}
	s.lock.Unlock()
	return nil
}

// SettingsRange 安全遍历Settings
func (s *sessionAgent) SettingsRange(f func(k, v string) bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.session.Settings == nil {
		return
	}
	for k, v := range s.session.Settings {
		c := f(k, v)
		if c == false {
			return
		}
	}
}

// update 更新为新的session数据(只更新Settings数据)
func (s *sessionAgent) update(session gate.ISession) error {
	userId := session.GetUserID()

	iP := session.GetIP()

	network := session.GetNetwork()

	sessionId := session.GetSessionID()

	serverId := session.GetServerID()

	settings := map[string]string{}
	session.SettingsRange(func(k, v string) bool {
		settings[k] = v
		return true
	})

	s.lock.Lock()
	s.session.UserId = userId
	s.session.IP = iP
	s.session.Network = network
	s.session.SessionId = sessionId
	s.session.ServerId = serverId
	s.session.Settings = settings
	s.lock.Unlock()
	return nil
}

// Clone 每次rpc调用都拷贝一份新的Session进行传输
func (s *sessionAgent) Clone() gate.ISession {
	s.lock.Lock()
	defer s.lock.Unlock()
	tmp := map[string]string{}
	for k, v := range s.session.Settings {
		tmp[k] = v
	}
	agent := &sessionAgent{
		userdata: s.userdata,
		lock:     new(sync.RWMutex),
		session: &SessionImp{
			IP:        s.session.IP,
			Network:   s.session.Network,
			UserId:    s.session.UserId,
			SessionId: s.session.SessionId,
			ServerId:  s.session.ServerId,
			TraceId:   s.session.TraceId,
			SpanId:    tools.GenerateID().String(),
			Settings:  tmp,
		},
	}
	return agent
}

// CloneSettings 只Clone Settings
func (s *sessionAgent) CloneSettings() map[string]string {
	tmp := map[string]string{}

	s.lock.Lock()
	defer s.lock.Unlock()
	for k, v := range s.session.Settings {
		tmp[k] = v
	}
	return tmp
}

// IsGuest 是否是访客(未登录), 默认判断规则为(userId=="")
func (s *sessionAgent) IsGuest() bool {
	if s.GetUserID() == "" {
		return true
	}
	return false
}

// GenRPCContext 生成RPC方法需要的context
func (s *sessionAgent) GenRPCContext() context.Context {
	ctx := context.Background()
	mqrpc.ContextWithValue(ctx, gate.CONTEXT_TRANSKEY_SESSION, s)
	mqrpc.ContextWithValue(ctx, log.CONTEXT_TRANSKEY_TRACE, s.GetTraceSpan())
	return ctx
}

// ========== TraceLog 部分
func (s *sessionAgent) GenTraceSpan() {
	s.session.TraceId = tools.GenerateID().String()
	s.session.SpanId = tools.GenerateID().String()
}
func (s *sessionAgent) GetTraceSpan() log.TraceSpan {
	return log.CreateTrace(s.session.TraceId, s.session.SpanId)
}

// ========== Session RPC方法封装

// update local Session(从Gate拉取最新数据)
func (s *sessionAgent) ToUpdate() error {
	server, err := app.Default().GetServerByID(s.session.ServerId)
	if err != nil {
		return fmt.Errorf("Gate not found serverId(%s), err:%v", s.session.ServerId, err)
	}
	result, err := server.Call(context.TODO(), "Load", s.session.SessionId)
	if err != nil {
		return fmt.Errorf("Call Gate serverId(%v) 'Load' err:%v", s.session.ServerId, err)
	}
	if result != nil { // 重新更新当前Session
		s.update(result.(gate.ISession))
	}
	return nil
}

// Bind the session with the the userId.
func (s *sessionAgent) ToBind(userId string) error {
	if app.Default() == nil {
		return fmt.Errorf("app.App is nil")
	}
	server, err := app.Default().GetServerByID(s.session.ServerId)
	if err != nil {
		return fmt.Errorf("Gate not found serverId(%s), err:%v", s.session.ServerId, err)
	}
	result, err := server.Call(context.TODO(), "Bind", s.session.SessionId, userId)
	if err != nil {
		return fmt.Errorf("Call Gate serverId(%v) 'Bind' err:%v", s.session.ServerId, err)
	}
	if result != nil { // 绑定成功,重新更新当前Session
		s.update(result.(gate.ISession))
	}
	return nil
}

// UnBind the session with the the userId.
func (s *sessionAgent) ToUnBind() error {
	if app.Default() == nil {
		return fmt.Errorf("app.App is nil")
	}
	server, err := app.Default().GetServerByID(s.session.ServerId)
	if err != nil {
		return fmt.Errorf("Gate not found serverId(%s), err:%v", s.session.ServerId, err)
	}
	result, err := server.Call(context.TODO(), "UnBind", s.session.SessionId)
	if err != nil {
		return fmt.Errorf("Call Gate serverId(%v) 'UnBind' err:%v", s.session.ServerId, err)
	}
	if result != nil { // 绑定成功,重新更新当前Session
		s.update(result.(gate.ISession))
	}
	return nil
}

// Set values (one) for the session.
func (s *sessionAgent) ToSet(key string, value string) error {
	if app.Default() == nil {
		return fmt.Errorf("app.App is nil")
	}
	server, err := app.Default().GetServerByID(s.session.ServerId)
	if err != nil {
		return fmt.Errorf("Gate not found serverId(%s), err:%v", s.session.ServerId, err)
	}

	result, err := server.Call(context.TODO(), "Set", s.session.SessionId, s.session.SessionId, key, value)
	if err != nil {
		return fmt.Errorf("Call Gate serverId(%v) 'Set' err:%v", s.session.ServerId, err)
	}
	if result != nil { // 绑定成功,重新更新当前Session
		s.update(result.(gate.ISession))
	}
	return nil
}

// Set values (many) for the session(直接用参数Push).
func (s *sessionAgent) ToSetBatch(settings map[string]string) error {
	if app.Default() == nil {
		return fmt.Errorf("app.App is nil")
	}
	server, err := app.Default().GetServerByID(s.session.ServerId)
	if err != nil {
		return fmt.Errorf("Gate not found serverId(%s), err:%v", s.session.ServerId, err)
	}
	result, err := server.Call(context.TODO(), "Push", s.session.SessionId, settings)
	if err != nil {
		return fmt.Errorf("Call Gate serverId(%v) 'Push' err:%v", s.session.ServerId, err)
	}
	if result != nil { // 绑定成功,重新更新当前Session
		s.update(result.(gate.ISession))
	}
	return nil
}

// Push all Settings values for the session(拿自己的Settings去Push).
func (s *sessionAgent) ToPush() error {
	if app.Default() == nil {
		return fmt.Errorf("app.App is nil")
	}
	server, err := app.Default().GetServerByID(s.session.ServerId)
	if err != nil {
		return fmt.Errorf("Gate not found serverId(%s), err:%v", s.session.ServerId, err)
	}
	tmp := map[string]string{}
	s.lock.Lock()
	for k, v := range s.session.Settings {
		tmp[k] = v
	}
	s.lock.Unlock()
	result, err := server.Call(context.TODO(), "Push", s.session.SessionId, tmp)
	if err != nil {
		return fmt.Errorf("Call Gate serverId(%v) 'Push' err:%v", s.session.ServerId, err)
	}
	if result != nil { // 绑定成功,重新更新当前Session
		s.update(result.(gate.ISession))
	}
	return nil
}

// Remove value from the session.
func (s *sessionAgent) ToDel(key string) error {
	if app.Default() == nil {
		return fmt.Errorf("app.App is nil")
	}
	server, err := app.Default().GetServerByID(s.session.ServerId)
	if err != nil {
		return fmt.Errorf("Gate not found serverId(%s), err:%v", s.session.ServerId, err)
	}
	result, err := server.Call(context.TODO(), "Del", s.session.SessionId, key)
	if err != nil {
		return fmt.Errorf("Call Gate serverId(%v) 'Remove' err:%v", s.session.ServerId, err)
	}
	if result != nil { // 绑定成功,重新更新当前Session
		s.update(result.(gate.ISession))
	}
	return nil
}

// Send message to the session.
func (s *sessionAgent) ToSend(topic string, body []byte) error {
	if app.Default() == nil {
		return fmt.Errorf("app.App is nil")
	}
	server, err := app.Default().GetServerByID(s.session.ServerId)
	if err != nil {
		return fmt.Errorf("Gate not found serverId(%s), err:%v", s.session.ServerId, err)
	}
	return server.CallNR(context.TODO(), "Send", s.session.SessionId, topic, body)
}

// the session is connect status
func (s *sessionAgent) ToConnected() (bool, error) {
	if app.Default() == nil {
		return false, fmt.Errorf("app.App is nil")
	}
	server, err := app.Default().GetServerByID(s.session.ServerId)
	if err != nil {
		return false, fmt.Errorf("Gate not found serverId(%s), err:%v", s.session.ServerId, err)
	}
	result, err := server.Call(context.TODO(), "Connected", s.session.SessionId)
	return result.(bool), err
}

// Close the session connect
func (s *sessionAgent) ToClose() error {
	if app.Default() == nil {
		return fmt.Errorf("app.App is nil")
	}
	server, err := app.Default().GetServerByID(s.session.ServerId)
	if err != nil {
		return fmt.Errorf("Gate not found serverId(%s), err:%v", s.session.ServerId, err)
	}
	return server.CallNR(context.TODO(), "Close", s.session.SessionId)
}

// ========== mqrpc.Marshaler 接口

func (s *sessionAgent) Marshal() ([]byte, error) {
	s.lock.RLock()
	data, err := msgpack.Marshal(s.session)
	s.lock.RUnlock()
	if err != nil {
		return nil, err
	}
	return data, nil
}
func (s *sessionAgent) Unmarshal(data []byte) error {
	se := &SessionImp{}
	err := msgpack.Unmarshal(data, se)
	if err != nil {
		return err
	}

	s.session = se
	if s.session.Settings == nil {
		s.lock.Lock()
		s.session.Settings = make(map[string]string)
		s.lock.Unlock()
	}
	return nil
}
func (s *sessionAgent) String() string {
	return "gate.Session"
}
