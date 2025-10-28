package httpgate

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"

	"github.com/cloudapex/river/app"
	"github.com/cloudapex/river/registry"
	"github.com/cloudapex/river/selector"
)

// Service represents an API service
type Service struct {
	// URL.Path
	Hander string
	// node
	SrvSession app.IServerSession
}

// Route 路由器定义
type Route func(r *http.Request) (*Service, error)

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
		return nil, errors.New(err.Error())
	}
	return &Service{SrvSession: session, Hander: r.URL.Path}, err
}
