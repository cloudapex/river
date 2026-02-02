package tools

import (
	"errors"
	"fmt"
	"math"
)

const (
	// base34 进制位
	base34 = 34

	// 34进制字母表, 由0-9a-z(去掉o和l)组成
	digits34 = "ub4zpfwj8ra72ynvgh3ism09ex5d6cq1kt"
)

// in init func
var position34 [256]int8

func init() {
	// 初始化为-1表示未使用
	for i := range position34 {
		position34[i] = -1
	}

	// 设置有效字符的位置
	for i, c := range []byte(digits34) {
		position34[c] = int8(i)
	}
}

// ToBase34 ...
func ToBase34(d uint64) string {
	if d == 0 {
		return string(digits34[0])
	}
	// 预分配固定大小数组
	var buf [13]byte // log34(2^64) ≈ 12.47
	i := len(buf)

	for v := d; v > 0; v /= base34 {
		i--
		buf[n] = digits34[v%base34]
	}

	return string(buf[i:])
}

// FromBase34 ...
func FromBase34(s string) (uint64, error) {
	var out uint64
	for _, c := range s {
		// 检查字符是否有效
		if c >= 256 || position34[byte(c)] < 0 {
			return 0, fmt.Errorf("invalid base34 char: %c", c)
		}

		d := position34[byte(c)]
		// 溢出检查
		if out > (math.MaxUint64-uint64(d))/base34 {
			return 0, errors.New("value exceeds uint64 range")
		}
		out = out*base34 + uint64(d)
	}
	return out, nil
}
