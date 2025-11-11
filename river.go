// Copyright 2017 river Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package river river
package river

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/cloudapex/river/app"
	"github.com/cloudapex/river/conf"
	"github.com/cloudapex/river/log"
	"github.com/cloudapex/river/module"
	"github.com/cloudapex/river/module/modules"
	"github.com/cloudapex/river/mqrpc"
	"github.com/cloudapex/river/registry"
	"github.com/cloudapex/river/registry/consul"
	"github.com/cloudapex/river/selector"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
)

// CreateApp 创建app实例
func CreateApp(opts ...app.Option) app.IApp {
	out := NewApp(append(opts, app.Version(version))...)
	if app.DefaultApp == nil {
		app.DefaultApp = out
	}
	return out
}

// NewApp 创建app
func NewApp(opts ...app.Option) app.IApp {
	out := &DefaultApp{
		opts:    app.NewOptions(opts...),
		manager: module.NewModuleManager(),
	}
	return out
}

// DefaultApp 默认应用
type DefaultApp struct {
	opts app.Options

	manager *module.ModuleManager

	serverList sync.Map

	//将一个RPC调用路由到新的路由上
	serviceRoute func(route string) string

	onConfigurationLoaded func()                              // 应用启动配置初始化完成后回调
	onModuleInited        func(module app.IModule)            // 每个模块初始化完成后回调
	onStartup             func()                              // 应用启动完成后回调
	onServiceDeleteds     []func(moduleName, serverId string) // 当模块服务断开删除时回调
}

// 初始化 consul
func (this *DefaultApp) initConsul() error {
	err := this.opts.Selector.Apply(selector.SetWatcher(this.Watcher))
	if err != nil {
		return err
	}

	if this.opts.Registry == nil {
		rs := consul.NewRegistry(func(options *registry.Options) {
			options.Addrs = this.opts.ConsulAddr
		})
		this.opts.Registry = rs
		err = this.opts.Selector.Apply(selector.Registry(rs))
		if err != nil {
			return err
		}
	}

	return nil
}

// 初始化 config
func (this *DefaultApp) initConfig() error {
	confData, err := this.Options().Registry.GetKV(this.Options().ConfigKey)
	if err != nil {
		return fmt.Errorf("无法从consul获取配置:%s, err:%v", this.Options().ConfigKey, err)
	}
	err = json.Unmarshal(confData, &conf.Conf)
	if err != nil {
		return fmt.Errorf("consul配置解析失败: err:%v, confData:%s", err.Error(), string(confData))
	}
	return nil
}

// 初始化 logs
func (this *DefaultApp) initLogs() error {
	log.Init(
		log.WithDebug(this.opts.Debug),
		log.WithProcessID(this.opts.ProcessEnv),
		log.WithBiDir(this.opts.BIDir),
		log.WithLogDir(this.opts.LogDir),
		log.WithLogFileName(this.opts.LogFileName),
		log.WithBiSetting(conf.Conf.BI),
		log.WithBIFileName(this.opts.BIFileName),
		log.WithLogSetting(conf.Conf.Log))
	return nil
}

// 初始化 nats
func (this *DefaultApp) initNats() error {
	if this.opts.Nats == nil {
		nc, err := nats.Connect(fmt.Sprintf("nats://%s", conf.Conf.Nats.Addr),
			nats.MaxReconnects(conf.Conf.Nats.MaxReconnects))
		if err != nil {
			return fmt.Errorf("initNats err:%v", err)
		}
		this.opts.Nats = nc
	}
	log.Info("nats addr:%s", conf.Conf.Nats.Addr)
	return nil
}

// Run 运行应用
func (this *DefaultApp) Run(mods ...app.IModule) error {
	var err error

	// init consul
	err = this.initConsul()
	if err != nil {
		return err
	}

	// init config
	err = this.initConfig()
	if err != nil {
		return err
	}

	// init log
	err = this.initLogs()
	if err != nil {
		return err
	}

	// callback
	if this.onConfigurationLoaded != nil {
		this.onConfigurationLoaded()
	}

	// init nats
	err = this.initNats()
	if err != nil {
		return err
	}

	// start modules
	log.Info("river %v starting...", this.opts.Version)

	// 1 RegisterRunMod
	this.manager.RegisterRunMod(modules.TimerModule()) // 先注册时间轮模块 每一个进程都默认运行

	// 2 Register
	for i := 0; i < len(mods); i++ {
		mods[i].OnAppConfigurationLoaded()
		this.manager.Register(mods[i])
	}
	this.OnInit() // 初始化modules之前回调(重载)

	// 2 init modules
	this.manager.Init(this.opts.ProcessEnv)

	// 3 startup callback
	if this.onStartup != nil {
		this.onStartup() // 初始化modules之后回调
	}
	log.Info("river %v started", this.opts.Version)

	// close
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	sig := <-c
	log.BiBeego().Flush()
	log.LogBeego().Flush()

	//如果一分钟都关不了则强制关闭
	timeout := time.NewTimer(this.opts.KillWaitTTL)
	wait := make(chan struct{})
	go func() {
		this.OnDestroy()
		wait <- struct{}{}
	}()
	select {
	case <-timeout.C:
		panic(fmt.Sprintf("river close timeout (signal: %v)", sig))
	case <-wait:
		log.Info("river closing down (signal: %v)", sig)
	}
	log.BiBeego().Close()
	log.LogBeego().Close()
	return nil
}

// Config 获取启动配置
func (this *DefaultApp) Config() conf.Config { return conf.Conf }

// Options 获取应用配置
func (this *DefaultApp) Options() app.Options { return this.opts }

// Transporter 获取消息传输对象
func (this *DefaultApp) Transporter() *nats.Conn { return this.opts.Nats }

// Registrar 获取服务注册对象
func (this *DefaultApp) Registrar() registry.Registry { return this.opts.Registry }

// WorkDir 获取进程工作目录
func (this *DefaultApp) WorkDir() string { return this.opts.WorkDir }

// GetProcessEnv 获取应用进程分组环境ID
func (this *DefaultApp) GetProcessEnv() string { return this.opts.ProcessEnv }

// UpdateOptions 允许再次更新应用配置(before this.Run)
func (this *DefaultApp) UpdateOptions(opts ...app.Option) error {
	for _, o := range opts {
		o(&this.opts)
	}
	return nil
}

// Watcher 监视服务节点注销(ServerSession删除掉)
func (this *DefaultApp) Watcher(node *registry.Node) {
	session, ok := this.serverList.Load(node.Id)
	if ok && session != nil {
		session.(app.IServerSession).GetRPC().Done()
		this.serverList.Delete(node.Id)
	}

	// 服务断开回调
	s := strings.Split(node.Id, "@")
	if len(s) < 2 {
		return
	}
	if len(this.onServiceDeleteds) != 0 {
		for _, f := range this.onServiceDeleteds {
			go f(s[0], node.Id)
		}
	}
}

// OnInit 初始化
func (this *DefaultApp) OnInit() error { return nil }

// OnDestroy 应用退出
func (this *DefaultApp) OnDestroy() error { this.manager.Destroy(); return nil }

// SetServiceRoute 设置服务路由器(动态转换service名称)
func (this *DefaultApp) SetServiceRoute(fn func(route string) string) error {
	this.serviceRoute = fn
	return nil
}

// GetRouteServer 获取服务实例(通过服务ID|服务类型,可设置可设置selector.WithFilter和selector.WithStrategy)
func (this *DefaultApp) GetRouteServer(service string, opts ...selector.SelectOption) (app.IServerSession, error) {
	if this.serviceRoute != nil { // 进行一次路由转换
		service = this.serviceRoute(service)
	}
	s := strings.Split(service, "@")
	if len(s) == 2 {
		serverID := service // s[0] + @ + s[1] = moduleType@moduleID
		moduleID := s[1]
		if moduleID != "" {
			return this.GetServerByID(serverID)
		}
	}
	moduleType := s[0]
	return this.GetServerBySelector(moduleType, opts...)
}

// GetServerByID 获取服务实例(通过服务ID(moduleType@id))
func (this *DefaultApp) GetServerByID(serverID string) (app.IServerSession, error) {
	session, ok := this.serverList.Load(serverID)
	if !ok {
		// s[0] + @ + s[1] = moduleType@moduleID
		s := strings.Split(serverID, "@")
		if len(s) != 2 {
			return nil, errors.Errorf("serverID is error %v", serverID)
		}
		moduleType := s[0]
		sessions := this.GetServersByType(moduleType)
		for _, s := range sessions {
			if s.GetNode().Id == serverID {
				return s, nil
			}
		}
	} else {
		return session.(app.IServerSession), nil
	}
	return nil, errors.Errorf("nofound %v", serverID)
}

// GetServerBySelector 获取服务实例(通过服务类型(moduleType),可设置可设置selector.WithFilter和selector.WithStrategy)
func (this *DefaultApp) GetServerBySelector(moduleType string, opts ...selector.SelectOption) (app.IServerSession, error) {
	next, err := this.opts.Selector.Select(moduleType, opts...)
	if err != nil {
		return nil, err
	}
	node, err := next()
	if err != nil {
		return nil, err
	}
	session, err := this.getServerSessionSafe(node, moduleType)
	if err != nil {
		return nil, err
	}
	return session.(app.IServerSession), nil
}

// GetServersByType 获取多个服务实例(通过服务类型(moduleType))
func (this *DefaultApp) GetServersByType(moduleType string) []app.IServerSession {
	sessions := make([]app.IServerSession, 0)
	services, err := this.opts.Selector.GetService(moduleType)
	if err != nil {
		log.Warning("GetServersByType %v", err)
		return sessions
	}
	for _, service := range services {
		for _, node := range service.Nodes {
			session, err := this.getServerSessionSafe(node, moduleType)
			if err != nil {
				log.Warning("getServerSessionSafe %v", err)
				continue
			}
			sessions = append(sessions, session.(app.IServerSession))
		}
	}
	return sessions
}

// getServerSessionSafe create and store serverSession safely
func (this *DefaultApp) getServerSessionSafe(node *registry.Node, moduleType string) (app.IServerSession, error) {
	session, ok := this.serverList.Load(node.Id)
	if ok {
		session.(app.IServerSession).SetNode(node)
		return session.(app.IServerSession), nil
	}
	// new
	s, err := module.NewModuleSession(moduleType, node)
	if err != nil {
		return nil, err
	}
	_session, _ := this.serverList.LoadOrStore(node.Id, s)
	_s := _session.(app.IServerSession)
	if s != _s { // 释放自己创建的那个
		go s.GetRPC().Done()
	}
	return s, nil
}

// Call RPC调用(需要等待结果)
func (this *DefaultApp) Call(ctx context.Context, moduleServer, _func string, param mqrpc.ParamOption, opts ...selector.SelectOption) (result any, err error) {
	server, err := this.GetRouteServer(moduleServer, opts...)
	if err != nil {
		return nil, err
	}
	return server.Call(ctx, _func, param()...)
}

// Call RPC调用(无需等待结果)
func (this *DefaultApp) CallNR(ctx context.Context, moduleServer, _func string, params ...any) (err error) {
	server, err := this.GetRouteServer(moduleServer)
	if err != nil {
		return
	}
	return server.CallNR(ctx, _func, params...)
}

// CallBroadcast RPC调用(群发,无需等待结果)
func (this *DefaultApp) CallBroadcast(ctx context.Context, moduleType, _func string, params ...any) {
	listSvr := this.GetServersByType(moduleType)
	for _, svr := range listSvr {
		svr.CallNR(ctx, _func, params...)
	}
}

// --------------- 回调(hook)

// OnConfigurationLoaded 设置应用启动配置初始化完成后回调
func (this *DefaultApp) OnConfigurationLoaded(_func func()) error {
	this.onConfigurationLoaded = _func
	return nil
}

// OnModuleInited 设置每个模块初始化完成后回调
func (this *DefaultApp) OnModuleInited(_func func(module app.IModule)) error {
	this.onModuleInited = _func
	return nil
}

// GetModuleInited 获取每个模块初始化完成后回调函数
func (this *DefaultApp) GetModuleInited() func(module app.IModule) {
	return this.onModuleInited
}

// OnStartup 设置应用启动完成后回调
func (this *DefaultApp) OnStartup(_func func()) error {
	this.onStartup = _func
	return nil
}

// OnServiceBreak 设置当模块服务断开删除时回调
func (this *DefaultApp) OnServiceBreak(_func func(moduleName, serverId string)) error {
	this.onServiceDeleteds = append(this.onServiceDeleteds, _func)
	return nil
}
