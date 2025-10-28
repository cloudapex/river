package httpgate

import (
	"time"

	"github.com/cloudapex/river/app"
	"github.com/cloudapex/river/module/server"
)

// Option 配置
type Option func(*Options)

// Options 网关配置项
type Options struct {
	TimeOut time.Duration
	Route   Route

	Opts []server.Option // 用来控制Module属性的
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

// ServerOpts ServerOpts
func ServerOpts(s []server.Option) Option {
	return func(o *Options) {
		o.Opts = s
	}
}
