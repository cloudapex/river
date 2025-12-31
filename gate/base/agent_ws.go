package gatebase

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"

	"github.com/cloudapex/river/gate"
	"github.com/cloudapex/river/tools/aes"
)

func NewWSClientAgent(h gate.FunRecvPackHandler) gate.IClientAgent {
	return &WSClientAgent{agentBase{recvHandler: h}}
}

type WSClientAgent struct {
	agentBase
}

// 读取数据并解码出Pack
func (this *WSClientAgent) OnReadDecodingPack() (*gate.Pack, error) {
	_, datas, err := this.conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	if len(datas) == 0 {
		return nil, nil
	}
	// 1 读取pack长度(2字节)
	if len(datas) < gate.PACK_HEAD_TOTAL_LEN_SIZE {
		return nil, fmt.Errorf("package len tool small")
	}
	// 解出pack长度 pkgLen(仅作验证)
	pkgLen := binary.LittleEndian.Uint16(datas[0:gate.PACK_HEAD_TOTAL_LEN_SIZE])
	if pkgLen != uint16(len(datas)) {
		return nil, fmt.Errorf("package len notmatch headLen")
	}
	// 2 读取body体
	bodyData := datas[gate.PACK_HEAD_TOTAL_LEN_SIZE:]
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
	// 3 从body中读取msgid长度(2个字节)
	topicLen := binary.LittleEndian.Uint16(bodyData[0:gate.PACK_HEAD_MSG_ID_LEN_SIZE])

	// 4 取得 msg_id(string) 和 msg data
	return &gate.Pack{
		Topic: string(bodyData[gate.PACK_HEAD_MSG_ID_LEN_SIZE : gate.PACK_HEAD_MSG_ID_LEN_SIZE+topicLen]),
		Body:  bodyData[gate.PACK_HEAD_MSG_ID_LEN_SIZE+topicLen:],
	}, nil
}
