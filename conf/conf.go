package conf

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Conf 全局配置结构体
var Conf = Config{}

// Config 配置结构体
type Config struct {
	RpcLog   bool                         `json:"rpc_log"`
	Module   map[string][]*ModuleSettings `json:"module"`
	Nats     Nats                         `json:"nats"`
	Settings map[string]any               `json:"settings"`
	Log      map[string]any               `json:"log"` // 不用定制
	BI       map[string]any               `json:"bi"`  // 不用定制
}

// ModuleSettings 模块配置
type ModuleSettings struct {
	ID         string         `json:"id"`   // 节点id(指@符号后面的值)
	Host       string         `json:"host"` // 没啥用
	ProcessEnv string         `json:"env"`
	Settings   map[string]any `json:"settings"`
}

// Nats nats配置
type Nats struct {
	Addr          string
	MaxReconnects int
}

// --------------- 本地配置

// LoadConfig 加载本地配置
func LoadConfig(path string) {
	fmt.Println("app configuration path :", path)

	// Read config
	if err := readFileInto(path); err != nil {
		panic(err)
	}
}
func readFileInto(path string) error {
	var data []byte
	buf := new(bytes.Buffer)
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	r := bufio.NewReader(f)
	for {
		line, err := r.ReadSlice('\n')
		if err != nil {
			if len(line) > 0 {
				buf.Write(line)
			}
			break
		}
		if !strings.HasPrefix(strings.TrimLeft(string(line), "\t "), "//") {
			buf.Write(line)
		}
	}
	data = buf.Bytes()
	//fmt.Print(string(data))
	return json.Unmarshal(data, &Conf)
}
