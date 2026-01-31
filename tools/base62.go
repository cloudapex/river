// Package tools 工具箱
package tools

import (
	"strings"
)

const (
	// base62 进制位
	base62 = 62

	// digits62 62进制码
	digits62 = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

// position34 62进制字符映射表
var position62 = [256]int8{}

func init() {
	// 初始化为-1表示未使用
	for i := range position62 {
		position62[i] = -1
	}
	// 初始化映射表
	for i, c := range digits62 {
		position62[byte(c)] = int8(i)
	}
}

// base62Map 62进制字符映射表
// var position34 = map[byte]int64{
// 	'0': 0, '1': 1, '2': 2, '3': 3, '4': 4, '5': 5, '6': 6, '7': 7, '8': 8, '9': 9,
// 	'a': 10, 'b': 11, 'c': 12, 'd': 13, 'e': 14, 'f': 15, 'g': 16, 'h': 17, 'i': 18,
// 	'j': 19, 'k': 20, 'l': 21, 'm': 22, 'n': 23, 'o': 24, 'p': 25, 'q': 26, 'r': 27,
// 	's': 28, 't': 29, 'u': 30, 'v': 31, 'w': 32, 'x': 33, 'y': 34, 'z': 35,
// 	'A': 36, 'B': 37, 'C': 38, 'D': 39, 'E': 40, 'F': 41, 'G': 42, 'H': 43, 'I': 44,
// 	'J': 45, 'K': 46, 'L': 47, 'M': 48, 'N': 49, 'O': 50, 'P': 51, 'Q': 52, 'R': 53,
// 	'S': 54, 'T': 55, 'U': 56, 'V': 57, 'W': 58, 'X': 59, 'Y': 60, 'Z': 61,
// }

// ToBase62 编码整数为base62字符串
func ToBase62(number int64) string {
	if number == 0 {
		return "0"
	}

	// 预分配足够的空间：int64最大值转换成62进制大约是11位
	var buf [12]byte // 稍微多分配一点
	i := len(buf)

	for number > 0 {
		i--
		buf[i] = digits62[number%base62]
		number /= base62
	}

	return string(buf[i:])
}

// FromBase62 解码base62字符串为整数(0表示非法字符或溢出)
func FromBase62(str string) int64 {
	str = strings.TrimSpace(str)
	if str == "" {
		return 0
	}

	var result int64 = 0
	for i := 0; i < len(str); i++ {
		digit := int64(position62[str[i]])

		// 检查非法字符
		if digit < 0 {
			return 0
		}
		prev := result
		result = result*base62 + digit
		// 溢出检查
		if result < prev {
			return 0 // 溢出
		}
	}
	return result
}
