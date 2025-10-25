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

// LoadConfig 加载本地配置
func LoadConfig(Path string) {
	fmt.Println("app configuration path :", Path)

	// Read config
	if err := readFileInto(Path); err != nil {
		panic(err)
	}
}

// Config 配置结构体
type Config struct {
	Log      map[string]interface{} // 不用定制
	BI       map[string]interface{} // 不用定制
	RpcLog   bool
	Module   map[string][]*ModuleSettings
	Nats     Nats
	Settings map[string]interface{}
}

// ModuleSettings 模块配置
type ModuleSettings struct {
	ID         string `json:"ID"` // 节点id(指@符号后面的值)
	Host       string // 没啥用
	ProcessEnv string
	Settings   map[string]interface{}
}

// Nats nats配置
type Nats struct {
	Addr          string
	MaxReconnects int
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
