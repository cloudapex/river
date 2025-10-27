// Package network 网络代理器
package network

import (
	"net"
)

// Conn 网络代理接口
type Conn interface {
	// Read 和 Write 方法处理的数据必须是一个完整的数据包
	net.Conn
	ReadMessage() (messageType int, p []byte, err error)
}

// Agent 代理
type Agent interface {
	Run() error
	OnClose() error
}
