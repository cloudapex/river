package gatebase

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"sync"

	"github.com/cloudapex/river/gate"
	"github.com/cloudapex/river/tools/aes"
)

func NewTCPClientAgent(h gate.FunRecvPackHandler) gate.IClientAgent {
	return &TCPClientAgent{
		agentBase: agentBase{recvHandler: h},
		pkgLenDataPool: &sync.Pool{
			New: func() interface{} {
				return make([]byte, gate.PACK_HEAD_TOTAL_LEN_SIZE)
			},
		},
		bodyDataPool: &sync.Pool{
			New: func() interface{} {
				// 初始化为512KB，这是一个合理的中间值
				return make([]byte, gate.PACK_BODY_DEFAULT_SIZE_IN_POOL)
			},
		},
	}
}

type TCPClientAgent struct {
	agentBase
	// 每个连接实例的缓冲池
	pkgLenDataPool *sync.Pool
	bodyDataPool   *sync.Pool
}

// 读取数据并解码出Pack
func (this *TCPClientAgent) OnReadDecodingPack() (*gate.Pack, error) {
	// 从缓冲池获取包长度数据缓冲区
	pkgLenData := this.pkgLenDataPool.Get().([]byte)
	defer this.pkgLenDataPool.Put(pkgLenData)

	// 1 读取pack长度(2字节)
	_, err := io.ReadFull(this.r, pkgLenData)
	if err != nil {
		return nil, err
	}
	// 1.1 解出pack长度 pkgLen
	pkgLen := binary.LittleEndian.Uint16(pkgLenData)

	// 1.2 计算需要的包体大小
	bodyLen := int(pkgLen - gate.PACK_HEAD_TOTAL_LEN_SIZE)

	var bodyData []byte
	var needPutBack bool

	// 判断是否需要使用缓冲池
	if bodyLen <= gate.PACK_BODY_DEFAULT_SIZE_IN_POOL {
		// 从缓冲池获取包体数据缓冲区
		buf := this.bodyDataPool.Get().([]byte)
		// 调整切片长度以匹配实际需要的大小
		bodyData = buf[:bodyLen]
		needPutBack = true
	} else {
		// 如果包体太大，直接创建新的缓冲区（不使用缓冲池）
		bodyData = make([]byte, bodyLen)
		needPutBack = false
	}

	// 确保缓冲区在函数结束时正确处理
	if needPutBack {
		defer this.bodyDataPool.Put(bodyData[:gate.PACK_BODY_DEFAULT_SIZE_IN_POOL]) // 归还完整大小的缓冲区
	}

	// 2 读取body体
	_, err = io.ReadFull(this.r, bodyData)
	if err != nil {
		return nil, err
	}
	if this.gate.Options().EncryptKey != "" {
		cbc, err := aes.AES_ECB_Decrypt(bodyData, []byte(this.gate.Options().EncryptKey))
		if err != nil {
			return nil, fmt.Errorf("decrypt cbc, err:%v", err)
		}
		b64Data, err := base64.StdEncoding.DecodeString(string(cbc))
		if err != nil {
			return nil, fmt.Errorf("decrypt base64, err:%v", err)
		}
		bodyData = b64Data
	}
	// 检查解密后的数据长度是否足够
	if len(bodyData) < gate.PACK_HEAD_MSG_ID_LEN_SIZE {
		return nil, fmt.Errorf("package len too small after decrypt")
	}
	// 3 从body中读取msgid长度(2个字节)
	topicLen := binary.LittleEndian.Uint16(bodyData[0:gate.PACK_HEAD_MSG_ID_LEN_SIZE])

	// 检查数据长度是否足够包含topic和body
	if len(bodyData) < int(gate.PACK_HEAD_MSG_ID_LEN_SIZE+topicLen) {
		return nil, fmt.Errorf("package len not enough for topic and body")
	}
	// 4 取得 string版msg_id 和 msg data
	return &gate.Pack{
		Topic: string(bodyData[gate.PACK_HEAD_MSG_ID_LEN_SIZE : gate.PACK_HEAD_MSG_ID_LEN_SIZE+topicLen]),
		Body:  bodyData[gate.PACK_HEAD_MSG_ID_LEN_SIZE+topicLen:],
	}, nil
}
