// Package httpgateway 网关配置
package httpgate

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/cloudapex/river/app"
	"github.com/cloudapex/river/registry"
	"github.com/cloudapex/river/selector"
)

// Service represents an API service
type Service struct {
	// hander
	Hander string
	// node
	SrvSession app.IServerSession
}

// DefaultRoute 默认路由规则
var DefaultRoute = func(r *http.Request) (*Service, error) {
	if r.URL.Path == "" {
		return nil, errors.New("path is nil")
	}
	handers := strings.Split(r.URL.Path, "/")
	if len(handers) < 2 {
		return nil, errors.New("path is not /[server]/path")
	}
	server := handers[1]
	if server == "" {
		return nil, errors.New("server is nil")
	}
	session, err := app.App().GetRouteServer(server,
		selector.WithStrategy(func(services []*registry.Service) selector.Next {
			var nodes []*registry.Node

			// Filter the nodes for datacenter
			for _, service := range services {
				for _, node := range service.Nodes {
					nodes = append(nodes, node)
				}
			}

			var mtx sync.Mutex
			//log.Info("services[0] $v",services[0].Nodes[0])
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

// Route 路由器定义
type Route func(r *http.Request) (*Service, error)

// Option 配置
type Option func(*Options)

// Options 网关配置项
type Options struct {
	TimeOut time.Duration
	Route   Route
}

// NewOptions 创建配置
func NewOptions(opts ...Option) Options {
	opt := Options{
		Route:   DefaultRoute,
		TimeOut: app.App().Options().RPCExpired,
	}

	for _, o := range opts {
		o(&opt)
	}

	return opt
}

// SetRoute 设置路由器
func SetRoute(s Route) Option {
	return func(o *Options) {
		o.Route = s
	}
}

// TimeOut 设置网关超时时间
func TimeOut(s time.Duration) Option {
	return func(o *Options) {
		o.TimeOut = s
	}
}
