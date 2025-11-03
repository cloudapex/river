package tools

import (
	"encoding/base64"
	"encoding/json"

	"github.com/vmihailenco/msgpack/v5"
)

// MsgPack → JSON
func MsgPackToJSON(msgpackData []byte) (string, error) {
	var data map[string]interface{}
	if err := msgpack.Unmarshal(msgpackData, &data); err != nil {
		return "", err
	}

	// 处理二进制数据
	processBinaryData(data)

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// JSON → MsgPack
func JSONToMsgPack(jsonStr string) ([]byte, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, err
	}

	// 处理 base64 字符串转回二进制
	processBase64Data(data)

	return msgpack.Marshal(data)
}

// 处理二进制字段（MsgPack → JSON 时）
func processBinaryData(data map[string]interface{}) {
	for k, v := range data {
		switch val := v.(type) {
		case []byte:
			data[k] = base64.StdEncoding.EncodeToString(val)
		case map[string]interface{}:
			processBinaryData(val)
		case []interface{}:
			for _, item := range val {
				if subMap, ok := item.(map[string]interface{}); ok {
					processBinaryData(subMap)
				}
			}
		}
	}
}

// 处理 base64 字段（JSON → MsgPack 时）
func processBase64Data(data map[string]interface{}) {
	for k, v := range data {
		switch val := v.(type) {
		case string:
			// 尝试解码 base64
			if decoded, err := base64.StdEncoding.DecodeString(val); err == nil {
				data[k] = decoded
			}
		case map[string]interface{}:
			processBase64Data(val)
		}
	}
}
