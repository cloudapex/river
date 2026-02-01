package hapi

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"

	"github.com/cloudapex/river/app"
	"github.com/cloudapex/river/mqrpc"
	"github.com/cloudapex/river/registry"
	"github.com/cloudapex/river/selector"
)

// Service represents an API service
type Service struct {
	// topic
	Topic string // msg_id
	// module server
	Server app.IModuleServerSession
}

// --------------- 路由器

// Router 路由器定义
type Router func(r *http.Request) (*Service, error)

// DefaultRoute 默认路由规则
var DefaultRoute = func(r *http.Request) (*Service, error) {
	if r.URL.Path == "" {
		return nil, errors.New("path is nil")
	}
	handers := strings.Split(r.URL.Path, "/")
	if len(handers) < 2 {
		return nil, errors.New("path is not /[server]/path")
	}
	service := handers[1]
	if service == "" {
		return nil, errors.New("module server is nil")
	}
	session, err := app.App().GetRouteServer(service,
		selector.WithStrategy(func(services []*registry.Service) selector.Next {
			var nodes []*registry.Node

			// Filter the nodes for datacenter
			for _, service := range services {
				for _, node := range service.Nodes {
					nodes = append(nodes, node)
				}
			}

			var mtx sync.Mutex
			return func() (*registry.Node, error) {
				mtx.Lock()
				defer mtx.Unlock()
				if len(nodes) == 0 {
					return nil, fmt.Errorf("no node")
				}
				index := rand.Intn(int(len(nodes)))
				return nodes[index], nil
			}
		}),
	)
	if err != nil {
		return nil, err
	}
	return &Service{Server: session, Topic: r.URL.Path}, err
}

// --------------- 转发器

// Transfer 转发器定义（Transfer）
type Transfer func(service *Service, req *Request, rsp *Response) error

// DefaultTransfe 默认转发规则
var DefaultTransfe = func(service *Service, req *Request, rsp *Response) error {
	return mqrpc.MsgPack(rsp, mqrpc.RpcResult(service.Server.GetRPC().Call(context.TODO(), service.Topic, req)))
}
