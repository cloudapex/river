// Package gate 网关配置
package gate

import (
	"time"

	"github.com/cloudapex/river/module/server"
)

// Option 网关配置项
type Option func(*Options)

// Options 网关配置项
type Options struct {
	WsAddr           string
	TcpAddr          string
	ConcurrentTasks  int // 单个连接允许的同时并发协程数,控制流量(20)(目前没用)
	BufSize          int // 连接数据缓存大小(2048)(只对TCP有用)
	MaxPackSize      int // 单个协议包数据最大值(uint16:65535)
	SendPackBuffSize int // 发送消息的缓冲队列(100)
	TLS              bool
	CertFile         string
	KeyFile          string
	EncryptKey       string // 消息包加密key
	//OverTime        time.Duration // 建立连接超时(10s)
	HeartOverTimer time.Duration // 心跳超时时间(本质是读取超时)(60s)

	Opts []server.Option // 用来控制module server属性的
}

// NewOptions 网关配置项
func NewOptions(opts ...Option) Options {
	opt := Options{
		Opts:             []server.Option{},
		ConcurrentTasks:  20,
		BufSize:          2048,
		MaxPackSize:      65535,
		SendPackBuffSize: 100,
		//OverTime:        time.Second * 10,
		HeartOverTimer: time.Second * 60,
		TLS:            false,
	}

	for _, o := range opts {
		o(&opt)
	}

	return opt
}

// ConcurrentTasks 设置单个连接允许的同时并发协程数(目前没用)
func ConcurrentTasks(s int) Option {
	return func(o *Options) {
		o.ConcurrentTasks = s
	}
}

// BufSize 单个连接网络数据缓存大小
func BufSize(s int) Option {
	return func(o *Options) {
		o.BufSize = s
	}
}

// MaxPackSize 单个协议包数据最大值
func MaxPackSize(s int) Option {
	return func(o *Options) {
		o.MaxPackSize = s
	}
}

// SendPackBuffNum 发送消息的缓冲队列数量
func SendPackBuffNum(n int) Option {
	return func(o *Options) {
		o.SendPackBuffSize = n
	}
}

// HeartOverTimer 心跳超时时间
func HeartOverTimer(s time.Duration) Option {
	return func(o *Options) {
		o.HeartOverTimer = s
	}
}

// OverTime 超时时间
// func OverTime(s time.Duration) Option {
// 	return func(o *Options) {
// 		o.OverTime = s
// 	}
// }

// TLS TLS
func TLS(s bool) Option {
	return func(o *Options) {
		o.TLS = s
	}
}

// TcpAddr tcp监听地址
// Deprecated: 因为命名规范问题函数将废弃,请用TCPAddr代替
func TcpAddr(s string) Option {
	return func(o *Options) {
		o.TcpAddr = s
	}
}

// WsAddr websocket监听端口
func WsAddr(s string) Option {
	return func(o *Options) {
		o.WsAddr = s
	}
}

// CertFile TLS 证书cert文件
func CertFile(s string) Option {
	return func(o *Options) {
		o.CertFile = s
	}
}

// KeyFile TLS 证书key文件
func KeyFile(s string) Option {
	return func(o *Options) {
		o.KeyFile = s
	}
}

// 消息包加密Key
func EncryptKey(s string) Option {
	return func(o *Options) {
		o.EncryptKey = s
	}
}

// ServerOpts ServerOpts
func ServerOpts(s []server.Option) Option {
	return func(o *Options) {
		o.Opts = s
	}
}
