// Package network tcp网络控制器
package network

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// ConnSet tcp连接管理器
type ConnSet map[net.Conn]struct{}

// TCPConn tcp连接
type TCPConn struct {
	io.Reader //Read(p []byte) (n int, err error)
	io.Writer //Write(p []byte) (n int, err error)
	sync.Mutex
	bufLocks  chan bool //当有写入一次数据设置一次
	buffer    bytes.Buffer
	conn      net.Conn
	closeFlag bool
}

func newTCPConn(conn net.Conn) *TCPConn {
	tcpConn := new(TCPConn)
	tcpConn.conn = conn

	return tcpConn
}

// Close 关闭tcp连接
func (tcpConn *TCPConn) Close() error {
	tcpConn.Lock()
	defer tcpConn.Unlock()
	if tcpConn.closeFlag {
		return nil
	}

	tcpConn.closeFlag = true
	return tcpConn.conn.Close()
}

// Write b must not be modified by the others goroutines
func (tcpConn *TCPConn) Write(b []byte) (n int, err error) {
	tcpConn.Lock()
	defer tcpConn.Unlock()
	if tcpConn.closeFlag || b == nil {
		return
	}

	return tcpConn.conn.Write(b)
}

// Read read data
func (tcpConn *TCPConn) Read(b []byte) (int, error) {
	return tcpConn.conn.Read(b)
}

// tcp not support ReadMessage
func (tcpConn *TCPConn) ReadMessage() (messageType int, p []byte, err error) {
	return 0, nil, fmt.Errorf("not impl")
}

// LocalAddr 本地socket端口地址
func (tcpConn *TCPConn) LocalAddr() net.Addr {
	return tcpConn.conn.LocalAddr()
}

// RemoteAddr 远程socket端口地址
func (tcpConn *TCPConn) RemoteAddr() net.Addr {
	return tcpConn.conn.RemoteAddr()
}

// SetDeadline A zero value for t means I/O operations will not time out.
func (tcpConn *TCPConn) SetDeadline(t time.Time) error {
	return tcpConn.conn.SetDeadline(t)
}

// SetReadDeadline sets the deadline for future Read calls.
// A zero value for t means Read will not time out.
func (tcpConn *TCPConn) SetReadDeadline(t time.Time) error {
	return tcpConn.conn.SetReadDeadline(t)
}

// SetWriteDeadline sets the deadline for future Write calls.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (tcpConn *TCPConn) SetWriteDeadline(t time.Time) error {
	return tcpConn.conn.SetWriteDeadline(t)
}
