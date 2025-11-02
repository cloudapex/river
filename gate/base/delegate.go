// Package basegate handler
package gatebase

import (
	"context"
	"fmt"

	"sync"

	"github.com/cloudapex/river/gate"
	"github.com/cloudapex/river/log"
	"github.com/cloudapex/river/tools"
	"github.com/pkg/errors"
)

// NewDelegate NewDelegate
func NewDelegate(gate gate.IGate) *Delegate {
	return &Delegate{
		gate: gate,
	}
}

// Delegate Delegate(gate.IAgentLearner,gate.IDelegater)
type Delegate struct {
	gate     gate.IGate
	sessions sync.Map //连接列表
	lock     sync.RWMutex
	agentNum int // session size
}

// OnDestroy
func (this *Delegate) OnDestroy() {
	this.sessions.Range(func(key, value any) bool {
		value.(gate.IClientAgent).Close()
		this.sessions.Delete(key)
		return true
	})
}

// GetAgent
func (this *Delegate) GetAgent(sessionId string) (gate.IClientAgent, error) {
	agent, ok := this.sessions.Load(sessionId)
	if !ok || agent == nil {
		return nil, errors.New("No Sesssion found")
	}
	return agent.(gate.IClientAgent), nil
}

// GetAgentNum
func (this *Delegate) GetAgentNum() int {
	num := 0
	this.lock.RLock()
	num = this.agentNum
	this.lock.RUnlock()
	return num
}

// ========== IAgentLearner

// 当连接建立(握手成功)
func (this *Delegate) Connect(a gate.IClientAgent) {
	defer func() {
		if err := tools.Catch(recover()); err != nil {
			log.Error("gateHandler Connect(agent) panic:%v", err)
		}
	}()
	if a.GetSession() != nil {
		this.sessions.Store(a.GetSession().GetSessionID(), a)
		// 已经建连成功的才计算
		if a.IsShaked() { // 握手
			this.lock.Lock()
			this.agentNum++
			this.lock.Unlock()
		}
	}
	// 客户端连接和断开的监听器
	if this.gate.GetSessionLearner() != nil {
		go func() {
			this.gate.GetSessionLearner().OnConnect(a.GetSession())
		}()
	}
}

// 当连接关闭(客户端主动关闭或者异常断开)
func (this *Delegate) DisConnect(a gate.IClientAgent) {
	defer func() {
		if err := tools.Catch(recover()); err != nil {
			log.Error("handler DisConnect panic:%v", err)
		}
		if a.GetSession() != nil {
			this.sessions.Delete(a.GetSession().GetSessionID())
			// 已经建联成功的才计算
			if a.IsShaked() { // 握手
				this.lock.Lock()
				this.agentNum--
				this.lock.Unlock()
			}
		}
	}()
	// 客户端连接和断开的监听器
	if this.gate.GetSessionLearner() != nil {
		if a.GetSession() != nil {
			// 没有session的就不回调了
			this.gate.GetSessionLearner().OnDisConnect(a.GetSession())
		}
	}
}

// ========== Session相关 RPC方法回调

// Load the latest session
func (this *Delegate) OnRpcLoad(ctx context.Context, sessionId string) (gate.ISession, error) {
	agent, ok := this.sessions.Load(sessionId)
	if !ok || agent == nil {
		return nil, fmt.Errorf("No Sesssion found")
	}
	return agent.(gate.IClientAgent).GetSession(), nil
}

// Bind the session with the the userId.
func (this *Delegate) OnRpcBind(ctx context.Context, sessionId string, userId string) (gate.ISession, error) {
	agent, ok := this.sessions.Load(sessionId)
	if !ok || agent == nil {
		return nil, fmt.Errorf("No Sesssion found")
	}

	agent.(gate.IClientAgent).GetSession().SetUserID(userId)

	storager := this.gate.GetStorageHandler()
	if storager != nil && agent.(gate.IClientAgent).GetSession().GetUserID() != "" {
		// 可以持久化
		data, err := storager.Query(userId)
		if err == nil && data != nil {
			// 有已持久化的数据,可能是上一次连接保存的
			impSession, err := NewSession(data)
			if err == nil {
				// 合并两个map 并且以 agent.(Agent).GetSession().Settings 已有的优先
				agent.(gate.IClientAgent).GetSession().SetSettings(impSession.CloneSettings())
			} else {
				// 解析持久化数据失败
				log.Warning("Sesssion Resolve fail %s", err.Error())
			}
		}
		//数据持久化
		_ = storager.Storage(agent.(gate.IClientAgent).GetSession())
	}

	return agent.(gate.IClientAgent).GetSession(), nil
}

// UnBind the session with the the userId.
func (this *Delegate) OnRpcUnBind(ctx context.Context, sessionId string) (gate.ISession, error) {
	agent, ok := this.sessions.Load(sessionId)
	if !ok || agent == nil {
		return nil, fmt.Errorf("No Sesssion found")
	}
	agent.(gate.IClientAgent).GetSession().SetUserID("")
	return agent.(gate.IClientAgent).GetSession(), nil
}

// Push the session with the the userId.
func (this *Delegate) OnRpcPush(ctx context.Context, sessionId string, settings map[string]string) (gate.ISession, error) {
	agent, ok := this.sessions.Load(sessionId)
	if !ok || agent == nil {
		return nil, fmt.Errorf("No Sesssion found")
	}
	// 覆盖当前map对应的key-value
	for key, value := range settings {
		_ = agent.(gate.IClientAgent).GetSession().Set(key, value)
	}

	storager := this.gate.GetStorageHandler()
	if storager != nil && agent.(gate.IClientAgent).GetSession().GetUserID() != "" {
		err := storager.Storage(agent.(gate.IClientAgent).GetSession())
		if err != nil {
			log.Warning("gate session storage failure : %v", err)
		}
	}

	return agent.(gate.IClientAgent).GetSession(), nil
}

// Set values (one or many) for the session.
func (this *Delegate) OnRpcSet(ctx context.Context, sessionId string, key string, value string) (gate.ISession, error) {
	agent, ok := this.sessions.Load(sessionId)
	if !ok || agent == nil {
		return nil, fmt.Errorf("No Sesssion found")
	}

	_ = agent.(gate.IClientAgent).GetSession().Set(key, value)

	storager := this.gate.GetStorageHandler()
	if storager != nil && agent.(gate.IClientAgent).GetSession().GetUserID() != "" {
		err := storager.Storage(agent.(gate.IClientAgent).GetSession())
		if err != nil {
			log.Error("gate session storage failure : %v", err)
		}
	}

	return agent.(gate.IClientAgent).GetSession(), nil
}

// Del value from the session.
func (this *Delegate) OnRpcDel(ctx context.Context, sessionId string, key string) (gate.ISession, error) {
	agent, ok := this.sessions.Load(sessionId)
	if !ok || agent == nil {
		return nil, fmt.Errorf("No Sesssion found")
	}

	_ = agent.(gate.IClientAgent).GetSession().Del(key)

	storager := this.gate.GetStorageHandler()
	if storager != nil && agent.(gate.IClientAgent).GetSession().GetUserID() != "" {
		err := storager.Storage(agent.(gate.IClientAgent).GetSession())
		if err != nil {
			log.Error("gate session storage failure :%v", err)
		}
	}

	return agent.(gate.IClientAgent).GetSession(), nil
}

// Send message to the session.
func (this *Delegate) OnRpcSend(ctx context.Context, sessionId string, topic string, body []byte) (bool, error) {
	agent, ok := this.sessions.Load(sessionId)
	if !ok || agent == nil {
		return false, fmt.Errorf("No Sesssion found")
	}
	// 组装一个pack{topic,data}
	err := agent.(gate.IClientAgent).SendPack(&gate.Pack{Topic: topic, Body: body})
	if err != nil {
		return false, err
	}
	return true, nil
}

// check connect is normal for the session
func (this *Delegate) OnRpcConnected(ctx context.Context, sessionId string) (bool, error) {
	agent, ok := this.sessions.Load(sessionId)
	if !ok || agent == nil {
		return false, fmt.Errorf("No Sesssion found")
	}
	return agent.(gate.IClientAgent).IsClosed(), nil
}

// Proactively close the connection of session
func (this *Delegate) OnRpcClose(ctx context.Context, sessionId string) (bool, error) {
	agent, ok := this.sessions.Load(sessionId)
	if !ok || agent == nil {
		return false, fmt.Errorf("No Sesssion found")
	}
	agent.(gate.IClientAgent).Close()
	return true, nil
}

// ========== Global的 RPC方法回调

// broadcast message to all session of the gate
func (this *Delegate) OnRpcBroadcast(ctx context.Context, topic string, body []byte) (int64, error) {
	var count int64 = 0
	this.sessions.Range(func(key, agent any) bool {
		e := agent.(gate.IClientAgent).SendPack(&gate.Pack{Topic: topic, Body: body})
		if e != nil {
			log.Warning("IAgent.SendPack error:", e.Error())
		} else {
			count++
		}
		return true
	})
	return count, nil
}
