# River

River是一款基于Golang语言的简洁、高效、高性能的分布式微服务框架。其灵活的架构也适用于即时通讯、物联网及其他分布式应用领域。

[![Go Report Card](https://goreportcard.com/badge/github.com/cloudapex/river)](https://goreportcard.com/report/github.com/cloudapex/river)
[![GoDoc](https://godoc.org/github.com/cloudapex/river?status.svg)](https://godoc.org/github.com/cloudapex/river)
[![Release](https://img.shields.io/github/release/cloudapex/river.svg?style=flat-square)](https://github.com/cloudapex/river/releases)

## 目录

- [版本](#版本)
- [特性](#特性)
- [架构设计](#架构设计)
  - [核心组件](#核心组件)
- [安装](#安装)
  - [环境要求](#环境要求)
  - [获取代码](#获取代码)
  - [依赖管理](#依赖管理)
- [快速开始](#快速开始)
  - [启动依赖服务](#启动依赖服务)
  - [配置文件](#配置文件)
  - [创建应用](#创建应用)
  - [创建业务模块](#创建业务模块)
  - [网关模块配置](#网关模块配置)
  - [运行应用](#运行应用)
- [使用示例](#使用示例)
  - [RPC调用](#rpc调用)
  - [网关消息处理](#网关消息处理)
  - [模块间通信](#模块间通信)
- [配置详解](#配置详解)
  - [网关配置参数](#网关配置参数)
  - [超时配置说明](#超时配置说明)
- [内置模块](#内置模块)
- [技术栈](#技术栈)
- [性能优化](#性能优化)
- [贡献](#贡献)
- [许可证](#许可证)
- [联系方式](#联系方式)
- [相关项目](#相关项目)

## 版本

当前版本: v1.2.3

## 特性

- **分布式架构**：支持高并发、高实时性，适用于游戏、即时通讯、物联网场景
- **无回调编程模型**：基于Goroutine实现，开发过程全程做到无callback回调，代码可读性更高
- **微服务支持**：完整的微服务框架，支持分布式服务注册发现
- **多协议支持**：多种网关支持, 网关层支持HTTP、TCP、WebSocket协议及自定义粘包协议
- **灵活的RPC通信**：使用NATS作为RPC通信通道，提供高效的消息传递分发机制
- **服务治理**：使用Consul实现服务注册与发现，支持服务监控和管理
- **模块化设计**：核心服务模块管理，支持灵活扩展
- **高效数据序列化**：只需使用MsgPack进行数据编码,让数据传输更简单更干净
- **连接池优化**：针对高频网络操作进行缓冲区复用优化
- **安全特性**：支持TLS加密和数据包加密

## 架构设计

River采用分层架构设计，主要包括以下几个核心组件：

```
┌─────────────────────────────────────────────────────────────┐
│                        Client Layer                         │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Gateway Layer (TCP/WebSocket) │  HTTP Gateway Layer        │
│                                                             │
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

1. **长连接网关层(Gate)**
   - 支持TCP和WebSocket协议
   - 支持自定义粘包协议
   - 提供客户端连接管理和消息路由
   - 支持TLS加密和数据包加密
   - 心跳超时控制（默认60秒）
   - 可配置最大包大小、发送缓冲区等参数

2. **短连接网关层(HAPI)**
   - 提供HTTP/HTTPS API服务
   - 支持RESTful风格路由
   - 支持TLS加密
   - 可配置超时参数（读超时、写超时、空闲超时）
   - 使用Gin框架处理HTTP请求

3. **应用层(App)**
   - 提供应用实例创建和管理
   - 支持模块注册和运行
   - 实现服务发现和RPC调用
   - 集成Consul服务注册与发现
   - 支持配置动态加载

4. **模块系统(Module)**
   - 支持自定义业务模块
   - 提供定时器模块等基础模块（基于时间轮算法）
   - 支持模块间RPC通信
   - 模块生命周期管理

5. **RPC通信(MQRPC)**
   - 基于NATS的消息队列实现
   - 支持同步和异步调用
   - 支持广播调用
   - 提供服务注册和发现机制

6. **服务注册与发现(Registry)**
   - 基于Consul实现
   - 支持服务监控和健康检查
   - 服务自动注册与注销

7. **工具集(Tools)**
   - AES加密/解密
   - ID生成
   - Base62编码
   - ID,IP工具
   - 环形Queue,安全Map等实用工具

## 安装

### 环境要求

- Go版本 >= 1.25.0
- NATS消息队列服务
- Consul服务注册与发现服务
- 支持Linux、Windows、macOS等操作系统

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

主要依赖：
- **NATS** (`github.com/nats-io/nats.go`) - RPC通信
- **Consul** (`github.com/hashicorp/consul/api`) - 服务注册与发现
- **WebSocket** (`github.com/gorilla/websocket`) - WebSocket支持
- **MsgPack** (`github.com/vmihailenco/msgpack/v5`) - 高效数据序列化
- **cleanenv** (`github.com/ilyakaznacheev/cleanenv`) - 配置解析
- **Gin** (`github.com/gin-gonic/gin`) - HTTP网关路由
- **assert** (`github.com/stretchr/testify/assert`) - 测试断言
-	**uuid** (`github.com/google/uuid`) - UUID生成

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
  "RpcLog": true,
  "Module": {
    "gate": [
      {
        "ID": "gate-1",
        "ProcessEnv": "dev",
        "Settings": {
          "TCPAddr": ":8091",
          "WsAddr": ":8092"
        }
      }
    ],
    "hapi": [
      {
        "ID": "hapi-1",
        "ProcessEnv": "dev",
        "Settings": {
          "Addr": ":8088"
        }
      }
    ]
  },
  "Nats": {
    "Addr": "127.0.0.1:4222",
    "MaxReconnects": 1000
  },
  "BI": {
    "file": {
      "prefix": "",
      "suffix": ".log"
    }
  },
  "Log": {
    "file": {
      "prefix": "",
      "suffix": ".log"
    }
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

// 注册模块到应用
func main() {
  app := river.CreateApp(/* ... */)
  gameModule := &GameModule{}
  app.Run(gameModule)
}
```

### 5. 网关模块配置

```go
// TCP/WebSocket网关配置
import "github.com/cloudapex/river/gate"

opts := gate.NewOptions(
  gate.WsAddr(":3654"),
  gate.TcpAddr(":3653"),
  gate.TLS(false),
  gate.HeartOverTimer(60*time.Second),
  gate.MaxPackSize(65535),
  gate.SendPackBuffNum(100),
)

// HTTP网关配置
import "github.com/cloudapex/river/hapi"

httpOpts := hapi.NewOptions(
  hapi.Addr(":8090"),
  hapi.TLS(false),
  hapi.ReadTimeout(5*time.Second),
  hapi.WriteTimeout(10*time.Second),
  hapi.IdleTimeout(60*time.Second),
)
```

### 6. 运行应用

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

// 关闭会话
session.ToClose()
```

### 模块间通信

```go
// 在模块中获取其他模块实例
server, err := app.GetRouteServer("game@server1")
if err != nil {
  // 处理错误
}

// 调用远程方法
result, err := server.Call(ctx, "Method", "param1", "param2")
```

## 配置详解

### 网关配置参数

**TCP/WebSocket网关(gate)**:
- `WsAddr`: WebSocket监听地址
- `TcpAddr`: TCP监听地址
- `TLS`: 是否启用TLS
- `CertFile`: TLS证书文件路径
- `KeyFile`: TLS私钥文件路径
- `HeartOverTimer`: 心跳超时时间（默认60秒）
- `MaxPackSize`: 单个协议包最大数据量（默认65535字节）
- `SendPackBuffSize`: 发送消息缓冲队列大小（默认100）

**HTTP网关(hapi)**:
- `Addr`: HTTP监听地址
- `TLS`: 是否启用HTTPS
- `CertFile`: HTTPS证书文件路径
- `KeyFile`: HTTPS私钥文件路径
- `ReadTimeout`: 读取超时时间（默认5秒）
- `WriteTimeout`: 写入超时时间（默认10秒）
- `IdleTimeout`: 空闲超时时间（默认60秒）
- `MaxHeaderBytes`: 最大HTTP头部字节数（默认4KB）

### 超时配置说明

River框架在多个层面提供了超时控制：

1. **连接超时**：WebSocket服务器使用`HTTPTimeout`参数控制HTTP层面的读写超时（默认10秒，网关中配置为12秒）
2. **心跳超时**：TCP/WebSocket网关使用`HeartOverTimer`参数控制心跳超时（默认60秒）
3. **HTTP超时**：HTTP网关提供读、写、空闲超时配置
4. **RPC超时**：通过`TimeOut`参数控制RPC调用超时

## 内置模块

River提供了多个内置模块：

- **Timer模块**：提供定时器功能，基于时间轮算法实现（精度10ms，36个槽位，单圈360ms）
- **Gate模块**：提供TCP/WebSocket网关服务，支持自定义协议
- **HTTP模块**：提供HTTP/HTTPS API服务，支持RESTful路由

### 模块配置示例

```json
{
  "Module": {
    "Timer": [
      {
        "id": "timer-1",
        "env": "dev"
      }
    ],
    "Gate": [
      {
        "id": "gate-1",
        "env": "dev",
        "settings": {
          "TCPAddr": ":3653",
          "WsAddr": ":3654",
          "TLS": false,
          "HeartOverTimer": "60s",
          "MaxPackSize": 65535
        }
      }
    ],
    "hapi": [
      {
        "id": "hapi-1",
        "env": "dev",
        "settings": {
          "Addr": ":8090",
          "TLS": false,
          "ReadTimeout": "5s",
          "WriteTimeout": "10s",
          "IdleTimeout": "60s"
        }
      }
    ]
  }
}
```

## 技术栈

- **语言**：Golang 1.25.0+
- **RPC通信**：[NATS](https://nats.io/) 
- **服务注册发现**：[Consul](https://www.consul.io/)
- **网络协议**：TCP, WebSocket, HTTP/HTTPS
- **序列化**：[MsgPack](https://msgpack.org/)
- **Web框架**：[Gin](https://gin-gonic.com/)
- **配置解析**：[cleanenv](https://github.com/ilyakaznacheev/cleanenv)
- **日志系统**：基于Beego日志组件封装
- **加密算法**：AES ECB/CBC模式
- **工具库**：UUID, Base62, IP工具等

## 性能优化

River针对高并发场景进行了多项优化：

- 基于Goroutine的并发模型，避免回调地狱
- 高效的消息序列化和反序列化
- 连接复用和池化技术
- 内存预分配和对象复用
- 缓冲区池化（sync.Pool）减少GC压力
- 零拷贝技术优化数据传输
- 时间轮算法实现高效定时器

## 贡献

欢迎提交Issue和Pull Request来帮助改进River。

## 许可证

River基于Apache License 2.0许可证开源。

## 联系方式

如有问题，请提交Issue或联系项目维护者。

## 相关项目

- [NATS](https://nats.io/) - 高性能消息队列
- [Consul](https://www.consul.io/) - 服务发现和配置
- [Gorilla WebSocket](https://github.com/gorilla/websocket) - WebSocket实现
- [MsgPack](https://msgpack.org/) - 高效二进制序列化
- [Gin](https://gin-gonic.com/) - Web框架