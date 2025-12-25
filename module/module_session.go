// Package basemodule 服务节点实例定义
package module

import (
	"context"
	"sync"

	"github.com/cloudapex/river/app"
	"github.com/cloudapex/river/mqrpc"
	rpcbase "github.com/cloudapex/river/mqrpc/base"
	"github.com/cloudapex/river/registry"
)

// NewModuleServerSession 创建一个节点实例的访问Session(rpcClient)
func NewModuleServerSession(name string, node *registry.Node) (app.IModuleServerSession, error) {
	session := &moduleServerSession{
		name: name,
		node: node,
	}
	rpc, err := rpcbase.NewRPCClient(session)
	if err != nil {
		return nil, err
	}
	session.rpc = rpc
	return session, err
}

type moduleServerSession struct {
	mu   sync.RWMutex
	node *registry.Node
	name string
	rpc  mqrpc.RPCClient
}

func (this *moduleServerSession) GetID() string {
	return this.node.Id
}

func (this *moduleServerSession) GetName() string {
	return this.name
}
func (this *moduleServerSession) GetRPC() mqrpc.RPCClient {
	return this.rpc
}

func (this *moduleServerSession) GetNode() *registry.Node {
	this.mu.RLock()
	defer this.mu.RUnlock()
	return this.node
}

func (this *moduleServerSession) SetNode(node *registry.Node) (err error) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.node = node
	return
}

// 消息请求 需要回复
func (this *moduleServerSession) Call(ctx context.Context, _func string, params ...any) (any, error) {
	return this.rpc.Call(ctx, _func, params...)
}

// 消息请求 不需要回复
func (this *moduleServerSession) CallNR(ctx context.Context, _func string, params ...any) (err error) {
	return this.rpc.CallNR(ctx, _func, params...)
}

// 消息请求 需要回复
func (this *moduleServerSession) CallArgs(ctx context.Context, _func string, argTypes []string, argDatas [][]byte) (any, error) {
	return this.rpc.CallArgs(ctx, _func, argTypes, argDatas)
}

// 消息请求 不需要回复
func (this *moduleServerSession) CallNRArgs(ctx context.Context, _func string, argTypes []string, argDatas [][]byte) (err error) {
	return this.rpc.CallNRArgs(ctx, _func, argTypes, argDatas)
}
