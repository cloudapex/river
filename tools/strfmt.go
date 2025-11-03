package tools

import (
	"fmt"
	"regexp"
	"strings"
)

// github.com/fatih/structs 是一个用于处理结构体的反射操作的包

var (
	regFields = regexp.MustCompile(`\{(\w+)\}`)
	regField  = regexp.MustCompile(`[\{\}]`)
)

// Sprintf 字符串格式,{name}可以替换为map中的value内容
// eg 你的名字是{name}   extra=map[string]string{"name","river"}
func Sprintf(format string, extra map[string]any) string {
	fields := regFields.FindAllString(format, -1)
	ret := format
	for _, fieldName := range fields {
		field := regField.ReplaceAllString(fieldName, "")
		if v, ok := extra[field]; !ok {
		} else {
			ret = strings.Replace(ret, fieldName, fmt.Sprintf("%v", v), 1)
		}
	}
	return ret
}
