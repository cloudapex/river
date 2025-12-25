package hapi

import (
	"time"

	"github.com/cloudapex/river/app"
	"github.com/cloudapex/river/module/server"
)

// Option 配置
type Option func(*Options)

// Options 网关配置项
type Options struct {
	Addr           string
	Route          Router     // 控制如何选择rpc服务
	RpcHandle      RPCHandler // 控制如何处理api请求
	TLS            bool
	CertFile       string
	KeyFile        string
	TimeOut        time.Duration // rpc超时时间
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
	MaxHeaderBytes int

	Opts []server.Option // 用来控制Module属性的
}

// NewOptions 创建配置
func NewOptions(opts ...Option) Options {
	opt := Options{
		Addr:           ":8090",
		Route:          DefaultRoute,
		TLS:            false,
		TimeOut:        app.App().Options().RPCExpired,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 4 * 1024,
	}

	for _, o := range opts {
		o(&opt)
	}

	return opt
}

// Route 设置路由器
func Route(s Router) Option {
	return func(o *Options) {
		o.Route = s
	}
}

// RpcHandler 设置rpc处理器
func RpcHandler(h RPCHandler) Option {
	return func(o *Options) {
		o.RpcHandle = h
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

// Addr 设置监听地址
func Addr(addr string) Option {
	return func(o *Options) {
		o.Addr = addr
	}
}

// TLS 设置是否启用TLS
func TLS(enable bool) Option {
	return func(o *Options) {
		o.TLS = enable
	}
}

// CertFile 设置证书文件路径
func CertFile(certFile string) Option {
	return func(o *Options) {
		o.CertFile = certFile
	}
}

// KeyFile 设置私钥文件路径
func KeyFile(keyFile string) Option {
	return func(o *Options) {
		o.KeyFile = keyFile
	}
}

// ReadTimeout 设置读取超时
func ReadTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.ReadTimeout = timeout
	}
}

// WriteTimeout 设置写入超时
func WriteTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.WriteTimeout = timeout
	}
}

// IdleTimeout 设置空闲超时
func IdleTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.IdleTimeout = timeout
	}
}

// MaxHeaderBytes 设置最大头部字节数
func MaxHeaderBytes(bytes int) Option {
	return func(o *Options) {
		o.MaxHeaderBytes = bytes
	}
}
