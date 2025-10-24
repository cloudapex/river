// Copyright 2014 mqant Author. All Rights Reserved.
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

// Package network 网络代理器
package network

import (
	"net"
)

// Conn 网络代理接口
type Conn interface {
	// Read 和 Write 方法处理的数据必须是一个完整的数据包
	net.Conn
	ReadMessage() (messageType int, p []byte, err error) // 只有ws_conn有实现

	Destroy()
	doDestroy()
}

// Agent 代理
type Agent interface {
	Run() error
	OnClose() error
}
