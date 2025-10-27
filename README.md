# River

River是一款基于Golang语言的简洁、高效、高性能的分布式微服务游戏服务器框架。研发的初衷是要实现一款能支持高并发、高性能、高实时性的游戏服务器框架，也希望River未来能够做即时通讯和物联网方面的应用。

[![Go Report Card](https://goreportcard.com/badge/github.com/cloudapex/river)](https://goreportcard.com/report/github.com/cloudapex/river)
[![GoDoc](https://godoc.org/github.com/cloudapex/river?status.svg)](https://godoc.org/github.com/cloudapex/river)
[![Release](https://img.shields.io/github/release/cloudapex/river.svg?style=flat-square)](https://github.com/cloudapex/river/releases)

## 特性

- **高性能分布式架构**：支持高并发、高实时性，适用于游戏、即时通讯、物联网场景
- **无回调编程模型**：基于Goroutine实现，开发过程全程做到无callback回调，代码可读性更高
- **微服务支持**：完整的微服务框架，支持分布式服务注册发现
- **多协议支持**：网关层支持MQTT协议及自定义粘包协议，兼容多平台客户端（iOS、Android、PC、WebSocket）
- **灵活的RPC通信**：使用NATS作为RPC通信通道，提供高效的消息传递机制
- **服务治理**：使用Consul实现服务注册与发现，支持服务监控和管理
- **模块化设计**：核心服务模块管理，支持灵活扩展
- **高效数据序列化**：使用MsgPack进行高效数据编码

## 架构设计

River采用分层架构设计，主要包括以下几个核心组件：

```
┌─────────────────────────────────────────────────────────────┐
│                        Client Layer                         │
├─────────────────────────────────────────────────────────────┤
│                     Gateway Layer (MQTT)                    │
├─────────────────────────────────────────────────────────────┤
│                    Application Layer                        │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────────────┐ │
│  │ Module1 │  │ Module2 │  │ Module3 │  │ Custom Modules  │ │
│  └─────────┘  └─────────┘  └─────────┘  └─────────────────┘ │
├─────────────────────────────────────────────────────────────┤
│                     Service Layer                           │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │ NATS Broker │  │ RPC Server  │  │ Registry (Consul)   │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### 核心组件

1. **网关层(Gate)**
   - 支持MQTT协议和WebSocket
   - 支持自定义粘包协议
   - 提供客户端连接管理和消息路由

2. **应用层(App)**
   - 提供应用实例创建和管理
   - 支持模块注册和运行
   - 实现服务发现和RPC调用

3. **模块系统(Module)**
   - 支持自定义业务模块
   - 提供定时器模块等基础模块
   - 支持模块间RPC通信

4. **RPC通信(MQRPC)**
   - 基于NATS的消息队列实现
   - 支持同步和异步调用
   - 提供服务注册和发现机制

5. **服务注册与发现(Registry)**
   - 基于Consul实现
   - 支持服务监控和健康检查

## 安装

### 环境要求

- Go版本 >= 1.25.0
- NATS消息队列服务
- Consul服务注册与发现服务

### 获取代码

```bash
git clone https://github.com/cloudapex/river.git
cd river
```

### 依赖管理

River使用Go Modules进行依赖管理：

```bash
go mod tidy
```

## 快速开始

### 1. 启动依赖服务

首先确保NATS和Consul服务已启动：

```bash
# 启动NATS服务
docker run -d --name nats -p 4222:4222 nats:latest

# 启动Consul服务
docker run -d --name consul -p 8500:8500 consul:latest
```

### 2. 配置文件

创建配置文件`config.json`：

```json
{
  "Nats": {
    "Addr": "127.0.0.1:4222",
    "MaxReconnects": 100
  },
  "Log": {
    "LogPath": "./logs",
    "LogLevel": "debug"
  }
}
```

### 3. 创建应用

```go
package main

import (
  "github.com/cloudapex/river"
  "github.com/cloudapex/river/app"
)

func main() {
  // 创建应用实例
  app := river.CreateApp(
    app.ConsulAddr("127.0.0.1:8500"),
    app.ConfigKey("/river/config"),
  )
  
  // 运行应用
  app.Run()
}
```

### 4. 创建业务模块

```go
package main

import (
  "context"
  
  "github.com/cloudapex/river/app"
  "github.com/cloudapex/river/conf"
)

type GameModule struct {
  app.ModuleBase
}

func (m *GameModule) GetType() string {
  return "game"
}

func (m *GameModule) Version() string {
  return "1.0.0"
}

func (m *GameModule) OnInit(settings *conf.ModuleSettings) {
  // 模块初始化逻辑
}

func (m *GameModule) Run(closeSig chan bool) {
  // 模块运行逻辑
  <-closeSig
}

func (m *GameModule) OnDestroy() {
  // 模块销毁逻辑
}
```

### 5. 运行应用

```bash
go run main.go
```

## 使用示例

### RPC调用

```go
// 同步调用
result, err := app.Call(context.Background(), "game@server1", "Hello", 
  func() []any { return []any{"world"} })

// 异步调用
err := app.CallNR(context.Background(), "game", "Notify", "message")

// 广播调用
app.CallBroadcast(context.Background(), "game", "Broadcast", "notice")
```

### 网关消息处理

```go
// 发送消息给客户端
session.ToSend("topic", []byte("message"))

// 绑定用户ID
session.ToBind("user123")

// 设置会话属性
session.ToSet("key", "value")
```

## 文档

[在线文档](https://cloudapex.github.io/river/)

## 模块

River提供了多个内置模块：

- **Timer模块**：提供定时器功能
- **Gate模块**：提供网关服务
- **HTTP模块**：提供HTTP API服务

## 技术栈

- **语言**：Golang 1.25.0+
- **RPC通信**：[NATS](https://nats.io/) 
- **服务注册发现**：[Consul](https://www.consul.io/)
- **网络协议**：MQTT, WebSocket
- **序列化**：[MsgPack](https://msgpack.org/)
- **配置解析**：[cleanenv](https://github.com/ilyakaznacheev/cleanenv)
- **日志系统**：基于Beego日志组件封装

## 性能

River针对游戏服务器场景进行了优化：

- 基于Goroutine的并发模型，避免回调地狱
- 高效的消息序列化和反序列化
- 连接复用和池化技术
- 内存预分配和对象复用

## 贡献

欢迎提交Issue和Pull Request来帮助改进River。

## 许可证

River基于Apache License 2.0许可证开源。

## 联系方式

如有问题，请提交Issue或联系项目维护者。