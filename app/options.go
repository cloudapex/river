package app

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/cloudapex/river/registry"
	"github.com/cloudapex/river/selector"
	"github.com/cloudapex/river/selector/cache"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/nats-io/nats.go"
)

// 启动参数(命令行优先级比环境变量高)
type startUpArgs struct {
	WordDir    string `env:"wd" env-default:""`
	ProcessEnv string `env:"env" env-default:"dev"`
	ConsulAddr string `env:"consul" env-default:"127.0.0.1:8500"` // configKey: config/{env}/server
	LogPath    string `env:"log" env-default:""`
	BiPath     string `env:"bi" env-default:""`
	PProfAddr  string `env:"pprof" env-default:""`
}

// APP选项
func NewOptions(opts ...Option) Options {

	// default value
	opt := Options{
		Version:          "1.0.0",
		Selector:         cache.NewSelector(), // 这两个
		RegisterInterval: time.Second * time.Duration(10),
		RegisterTTL:      time.Second * time.Duration(20),
		KillWaitTTL:      time.Second * time.Duration(60),
		RPCExpired:       time.Second * time.Duration(10),
		RPCMaxCoroutine:  0, //不限制
		Debug:            true,
		Parse:            true,
		LogFileName: func(logdir, prefix, processID, suffix string) string {
			return fmt.Sprintf("%s/%v%s%s", logdir, prefix, processID, suffix)
		},
		BIFileName: func(logdir, prefix, processID, suffix string) string {
			return fmt.Sprintf("%s/%v%s%s", logdir, prefix, processID, suffix)
		},
	}

	for _, o := range opts {
		o(&opt)
	}

	// 解析启动参数
	var startArgs = startUpArgs{}
	if opt.Parse {
		// 其次使用环境变量
		if err := cleanenv.ReadEnv(&startArgs); err != nil {
			panic(err)
		}
		// 优先使用命令行参数
		flag.StringVar(&startArgs.WordDir, "wd", startArgs.WordDir, "Server work directory")
		flag.StringVar(&startArgs.ProcessEnv, "env", startArgs.ProcessEnv, "Server ProcessEnv")
		flag.StringVar(&startArgs.ConsulAddr, "consul", startArgs.ConsulAddr, "Consul server addr")
		flag.StringVar(&startArgs.LogPath, "log", startArgs.LogPath, "Log file directory")
		flag.StringVar(&startArgs.BiPath, "bi", startArgs.BiPath, "bi file directory")
		flag.StringVar(&startArgs.PProfAddr, "pprof", startArgs.PProfAddr, "listen pprof addr")
		flag.Parse()
	}

	// 工作目录(优先使用代码中的)
	if opt.WorkDir == "" {
		opt.WorkDir = startArgs.WordDir
	}

	// 设置进程分组环境(优先使用代码中的)
	if opt.ProcessEnv == "" {
		opt.ProcessEnv = startArgs.ProcessEnv
		if opt.ProcessEnv == "" {
			opt.ProcessEnv = "dev"
		}
	}

	// 最终检查设置工作目录
	ApplicationDir := ""
	if opt.WorkDir != "" {
		_, err := os.Open(opt.WorkDir)
		if err != nil {
			panic(err)
		}
		os.Chdir(opt.WorkDir)
		ApplicationDir, err = os.Getwd()
	} else {
		var err error
		ApplicationDir, err = os.Getwd()
		if err != nil {
			file, _ := exec.LookPath(os.Args[0])
			ApplicationPath, _ := filepath.Abs(file)
			ApplicationDir, _ = filepath.Split(ApplicationPath)
		}

	}
	opt.WorkDir = ApplicationDir

	// nats addr
	if len(opt.ConsulAddr) == 0 {
		opt.ConsulAddr = append(opt.ConsulAddr, startArgs.ConsulAddr)
	}

	// configkey
	if opt.ConfigKey == "" {
		opt.ConfigKey = fmt.Sprintf("config/%v/server", startArgs.ProcessEnv)
	}

	// 创建日志文件
	defaultLogPath := fmt.Sprintf("%s/bin/logs", ApplicationDir)
	defaultBIPath := fmt.Sprintf("%s/bin/bi", ApplicationDir)

	if opt.LogDir == "" { // 优先使用代码中的
		opt.LogDir = defaultLogPath
		if startArgs.LogPath != "" {
			opt.LogDir = startArgs.LogPath
		}
	}
	if opt.BIDir == "" { // 优先使用代码中的
		opt.BIDir = defaultBIPath
		if startArgs.BiPath != "" {
			opt.BIDir = startArgs.BiPath
		}
	}

	if _, err := os.Stat(opt.LogDir); os.IsNotExist(err) {
		if err := os.Mkdir(opt.LogDir, os.ModePerm); err != nil {
			fmt.Println(err)
		}
	}
	if _, err := os.Stat(opt.BIDir); os.IsNotExist(err) {
		if err := os.Mkdir(opt.BIDir, os.ModePerm); err != nil {
			fmt.Println(err)
		}
	}

	// pprof
	opt.PProfAddr = startArgs.PProfAddr
	if opt.PProfAddr != "" {
		go func() {
			err := http.ListenAndServe(opt.PProfAddr, nil)
			fmt.Printf("StartPProf err:%v \n", err)
		}()
	}

	return opt
}

// Option 应用级别配置项
type Option func(*Options)

// Options 应用级别配置
type Options struct {
	Version     string
	Debug       bool
	Parse       bool // 是否由框架解析启动环境变量,默认为true
	WorkDir     string
	ProcessEnv  string   // 进程分组ID(development)
	ConfigKey   string   // for consule
	ConsulAddr  []string // for consule
	LogDir      string
	BIDir       string
	PProfAddr   string
	KillWaitTTL time.Duration // 服务关闭超时强杀(60s)

	Nats             *nats.Conn
	Registry         registry.Registry // 注册服务发现(registry.DefaultRegistry)
	Selector         selector.Selector // 节点选择器(在Registry基础上)(cache.NewSelector())
	RegisterInterval time.Duration     // 服务注册发现续约频率(10s)
	RegisterTTL      time.Duration     // 服务注册发现续约生命周期(20s)

	ClientRPChandler   ClientRPCHandler   // 配置RPC调用方监控器(nil)
	ServerRPCHandler   ServerRPCHandler   // 配置RPC服务方监控器(nil)
	RpcCompleteHandler RpcCompleteHandler // 配置RPC执行结果监控器(nil)
	RPCExpired         time.Duration      // RPC调用超时(10s)
	RPCMaxCoroutine    int                // 默认0(不限制) 没用

	// 自定义日志文件名字(主要作用方便k8s映射日志不会被冲突，建议使用k8s pod实现)
	LogFileName FileNameHandler // 日志文件名称(默认):fmt.Sprintf("%s/%v%s%s", logdir, prefix, processID, suffix)
	// 自定义BI日志名字
	BIFileName FileNameHandler //  BI文件名称(默认):fmt.Sprintf("%s/%v%s%s", logdir, prefix, processID, suffix)
}

// Version 应用版本
func Version(v string) Option {
	return func(o *Options) {
		o.Version = v
	}
}

// Debug 只有是在调试模式下才会在控制台打印日志, 非调试模式下只在日志文件中输出日志
func Debug(t bool) Option {
	return func(o *Options) {
		o.Debug = t
	}
}

// WorkDir 进程工作目录
func WorkDir(v string) Option {
	return func(o *Options) {
		o.WorkDir = v
	}
}

// Configure 配置key
func ConfigKey(v string) Option {
	return func(o *Options) {
		o.ConfigKey = v
	}
}

// Configure consule 地址
func ConsulAddr(v ...string) Option {
	return func(o *Options) {
		o.ConsulAddr = append(o.ConsulAddr, v...)
	}
}

// LogDir 日志存储路径
func LogDir(v string) Option {
	return func(o *Options) {
		o.LogDir = v
	}
}

// ProcessID 进程分组ID
func ProcessID(v string) Option {
	return func(o *Options) {
		o.ProcessEnv = v
	}
}

// BILogDir  BI日志路径
func BILogDir(v string) Option {
	return func(o *Options) {
		o.BIDir = v
	}
}

// Nats  nats配置
func Nats(nc *nats.Conn) Option {
	return func(o *Options) {
		o.Nats = nc
	}
}

// Registry sets the registry for the service
// and the underlying components
func Registry(r registry.Registry) Option {
	return func(o *Options) {
		o.Registry = r
		o.Selector.Apply(selector.Registry(r))
	}
}

// Selector 路由选择器
func Selector(r selector.Selector) Option {
	return func(o *Options) {
		o.Selector = r
	}
}

// RegisterTTL specifies the TTL to use when registering the service
func RegisterTTL(t time.Duration) Option {
	return func(o *Options) {
		o.RegisterTTL = t
	}
}

// RegisterInterval specifies the interval on which to re-register
func RegisterInterval(t time.Duration) Option {
	return func(o *Options) {
		o.RegisterInterval = t
	}
}

// KillWaitTTL specifies the interval on which to re-register
func KillWaitTTL(t time.Duration) Option {
	return func(o *Options) {
		o.KillWaitTTL = t
	}
}

// SetClientRPChandler 配置调用者监控器
func SetClientRPChandler(t ClientRPCHandler) Option {
	return func(o *Options) {
		o.ClientRPChandler = t
	}
}

// SetServerRPCHandler 配置服务方监控器
func SetServerRPCHandler(t ServerRPCHandler) Option {
	return func(o *Options) {
		o.ServerRPCHandler = t
	}
}

// SetServerRPCCompleteHandler 服务RPC执行结果监控器
func SetRpcCompleteHandler(t RpcCompleteHandler) Option {
	return func(o *Options) {
		o.RpcCompleteHandler = t
	}
}

// Parse mqant框架是否解析环境参数
func Parse(t bool) Option {
	return func(o *Options) {
		o.Parse = t
	}
}

// RPC超时时间
func RPCExpired(t time.Duration) Option {
	return func(o *Options) {
		o.RPCExpired = t
	}
}

// 单个节点RPC同时并发协程数
func RPCMaxCoroutine(t int) Option {
	return func(o *Options) {
		o.RPCMaxCoroutine = t
	}
}

// WithLogFile 日志文件名称
func WithLogFile(name FileNameHandler) Option {
	return func(o *Options) {
		o.LogFileName = name
	}
}

// WithBIFile Bi日志名称
func WithBIFile(name FileNameHandler) Option {
	return func(o *Options) {
		o.BIFileName = name
	}
}
