package tools

import (
	"math"
)

const (
	// 34进制字母表, 由0-9a-z(去掉o和l)组成
	digits34      = "ub4zpfwj8ra72ynvgh3ism09ex5d6cq1kt"
	bitMask       = uint64(0xFFFFFFFF) // 32位掩码
	feistelRounds = 4
	//magicNumber   = 0x123456789ABCDEF1 // Feistel轮函数使用的魔术数
)

// in init func
var position34 [256]uint8

func init() {
	for i, c := range []byte(digits34) {
		position34[c] = uint8(i)
	}
}

// ToBase34 ...
func ToBase34(d uint64) string {
	if d == 0 {
		return string(digits34[0])
	}
	// 计算位数
	n := 0
	for v := d; v > 0; v /= 34 {
		n++
	}
	out := make([]byte, n)
	for v := d; v > 0; v /= 34 {
		n--
		out[n] = digits34[v%34]
	}
	return string(out)
}

// FromBase34 ...
func FromBase34(s string) uint64 {
	var out uint64
	for _, c := range s {
		if c >= 256 || position34[byte(c)] == 0 && c != rune(digits34[0]) {
			return 0 //, fmt.Errorf("invalid base34 char: %c", c)
		}
		d := position34[c]
		if out > (math.MaxUint64-uint64(d))/34 {
			return 0 //, errors.New("value exceeds uint64 range")
		}
		out = out*34 + uint64(d)
	}
	return out //, nil
}

// userId to accountId
func ToAcctId(userId uint64) string {
	if userId == 0 {
		return ""
		// return string(digits34[0])
	}
	//return ToBase34(userId)

	mixed := feistelEncrypt(uint32(userId & bitMask))
	return ToBase34(uint64(userId>>32<<32 | uint64(mixed)))
}

// accountId to userId
func ToUserId(acctId string) uint64 {
	if acctId == "" {
		return 0
	}
	// return FromBase34(acctId)
	mixed := FromBase34(acctId)

	right := feistelDecrypt(uint32(mixed & bitMask))
	return mixed>>32<<32 | uint64(right)
}

// ---------------
// 轮函数 - 对16位输入进行非线性变换
func roundFunc(input uint16, key uint16) uint16 {
	// 使用位旋转、异或和加法的组合
	temp := uint32(input) * uint32(key)
	temp ^= temp >> 8
	temp += uint32(input) << 3
	temp ^= temp >> 4
	return uint16(temp & 0xFFFF)
}

// 密钥生成 - 为每轮生成不同的16位密钥
func generateKey(round int) uint16 {
	// 使用常数和轮数生成密钥
	key := uint32(round*0x9E37 + 0xB7E1)
	return uint16((key ^ (key >> 16)) & 0xFFFF)
}

// Feistel网络加密 - 将int32混淆
func feistelEncrypt(num uint32) uint32 {
	// 转换为无符号数进行处理
	value := uint32(num)

	// 分割成两个16位半部分
	left := uint16(value >> 16)     // 高16位
	right := uint16(value & 0xFFFF) // 低16位

	// 执行Feistel轮次
	for i := 0; i < feistelRounds; i++ {
		key := generateKey(i)
		// Feistel结构: newLeft = oldRight, newRight = oldLeft XOR F(oldRight, key)
		newRight := left ^ roundFunc(right, key)
		left = right
		right = newRight
	}

	// 合并结果
	result := (uint32(left) << 16) | uint32(right)
	return uint32(result)
}

// Feistel网络解密 - 还原原始int32
func feistelDecrypt(num uint32) uint32 {
	// 转换为无符号数进行处理
	value := uint32(num)

	// 分割加密后的值
	left := uint16(value >> 16)     // 高16位
	right := uint16(value & 0xFFFF) // 低16位

	// 逆向执行Feistel轮次 - 关键是要逆向操作
	for i := feistelRounds - 1; i >= 0; i-- {
		key := generateKey(i)
		// 逆向Feistel: newRight = oldLeft, newLeft = oldRight XOR F(oldLeft, key)
		newLeft := right ^ roundFunc(left, key)
		right = left
		left = newLeft
	}

	// 合并结果
	result := (uint32(left) << 16) | uint32(right)
	return uint32(result)
}
