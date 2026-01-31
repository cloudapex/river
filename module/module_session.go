// Package basemodule 服务节点实例定义
package module

import (
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
	rpc  mqrpc.IRPCClient
}

func (this *moduleServerSession) GetID() string {
	return this.node.Id
}

func (this *moduleServerSession) GetName() string {
	return this.name
}
func (this *moduleServerSession) GetRPC() mqrpc.IRPCClient {
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
	return nil
}
