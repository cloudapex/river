package hapi

import (
	"time"

	"github.com/cloudapex/river/module/server"
)

// 配置Setting键常量定义
const (
	// 服务器配置
	SettingKeyAddr = "addr" // 监听地址

	// tls
	SettingKeyTLS      = "tls"           // 是否启用TLS
	SettingKeyCertFile = "tls_cert_file" // 证书文件路径
	SettingKeyKeyFile  = "tls_key_file"  // 私钥文件路径

	SettingKeyReadTimeout    = "read_timeout"     // 读取超时（秒）
	SettingKeyWriteTimeout   = "write_timeout"    // 写入超时（秒）
	SettingKeyIdleTimeout    = "idle_timeout"     // 空闲超时（秒）
	SettingKeyMaxHeaderBytes = "max_header_bytes" // 最大头部字节数

	// 安全配置
	SettingKeyDebugKey   = "debug_key"   // 调试密钥
	SettingKeyEncryptKey = "encrypt_key" // 加密密钥
)

// Option 配置
type Option func(*Options)

// Options 网关配置项
type Options struct {
	Addr           string   // Settings["Addr"]
	Route          Router   // 控制如何选择rpc服务
	Transfer       Transfer // 控制如何处理api请求
	TLS            bool
	CertFile       string
	KeyFile        string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
	MaxHeaderBytes int
	DebugKey       string // 调试用(可不用加密调试)(Settings["DebugKey"])
	EncryptKey     string // 消息包加密key(Settings["EncryptKey"])(must 16, 24 or 32 bytes)

	Opts []server.Option // 用来控制Module属性的
}

// NewOptions 创建配置
func NewOptions(opts ...Option) Options {
	opt := Options{
		Addr:           ":8090",
		Route:          DefaultRoute,
		Transfer:       DefaultTransfe,
		TLS:            false,
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
func Transfers(t Transfer) Option {
	return func(o *Options) {
		o.Transfer = t
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

// DebugKey 调试时的key(可不用加密)
func DebugKey(key string) Option {
	return func(o *Options) {
		o.DebugKey = key
	}
}

// EncryptKey 设置消息包加密Key
func EncryptKey(key string) Option {
	return func(o *Options) {
		o.EncryptKey = key
	}
}
