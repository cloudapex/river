package tools

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"
)

func TestMath11(t *testing.T) {
	for n := range 100 {
		i := uint64(2000000 + n)
		s := ToAcctId(i)
		fmt.Println(2000000+n, "ToAcctId:", s)
		if _s := ToUserId(s); _s != i {
			t.Errorf("missmatch s=%s _i=%v i=%d", s, _s, i)
		}
	}
	// if !BasicConversion() {
	// 	t.Errorf("TestBasicConversion faild ")
	// }

	// s := ToAcctId(2000000)
	// if id := ToUserId(s); id != 2000000 {
	// 	t.Errorf("TestMath faild:s = %v id =%v", s, id)

	// }

	mixed := feistelEncrypt(2000000)
	if x := feistelDecrypt(mixed); x != 2000000 {
		t.Errorf("TestMath faild: mixed = %d , x != 2000000, x=%v", mixed, x)
	}
}
func BasicConversion() bool {
	testCases := []uint64{0, 1, 2, 33, 34, 1000, 1000000, 0x7FFFFFFFFFFFFF}

	for _, userId := range testCases {
		acctId := ToAcctId(userId)
		decoded := ToUserId(acctId)
		if decoded != userId {
			return false
		}
	}
	return true
}
func TestToAcctId(t *testing.T) {
	type args struct {
		userId uint64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"1", args{0}, "u"},               // 零值
		{"2", args{33}, "3k0ec9"},         // 最大值单字符
		{"3", args{34}, "b85rtak"},        // 进位
		{"3", args{2000000}, "43z5deh"},   // 进位
		{"3", args{2502145}, "ikzuaa"},    // 进位
		{"3", args{250214500}, "b6vg7sv"}, // 进位

		{"4", args{math.MaxUint64 / 34}, "jedxwz4f74ak"}, // 最大值
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToAcctId(tt.args.userId); got != tt.want {
				t.Errorf("ToAcctId() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToUserId(t *testing.T) {
	type args struct {
		acctId string
	}
	tests := []struct {
		name string
		args args
		want uint64
	}{
		{"1", args{"u"}, 0},
		{"2", args{"3k0ec9"}, 33},
		{"3", args{"b85rtak"}, 34},
		{"3", args{"43z5deh"}, 2000000},
		{"3", args{"ikzuaa"}, 2502145},
		{"3", args{"b6vg7sv"}, 250214500},
		{"3", args{"4m5mihr"}, 250214500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToUserId(tt.args.acctId); got != tt.want {
				t.Errorf("ToUserId() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAcctIdMapping(t *testing.T) {
	// 固定种子确保可重复
	rand.Seed(time.Now().UnixNano())

	// 边界测试
	boundaries := []uint64{
		0,
		1,
		34,
		123456,
		math.MaxUint32 - 1,
		math.MaxUint32,
		uint64(math.MaxUint32) + 1,
		math.MaxUint64 / 2,
		math.MaxInt64, // 注意 MaxInt64 转换成 uint64
	}

	for _, id := range boundaries {
		acct := ToAcctId(id)
		got := ToUserId(acct)
		if got != id {
			t.Fatalf("Boundary test failed: id=%d, acctId=%s, got=%d", id, acct, got)
		}
	}

	// 随机采样测试
	const sampleCount = 10000000 // 1000 万次随机测试，秒级完成
	for i := 0; i < sampleCount; i++ {
		id := rand.Uint64()
		acct := ToAcctId(id)
		got := ToUserId(acct)
		if got != id {
			t.Fatalf("Random test failed: id=%d, acctId=%s, got=%d", id, acct, got)
		}
	}
}
