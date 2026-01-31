package tools

import (
	"math"
	"math/rand"
	"testing"
	"time"
)

func TestToBase34(t *testing.T) {
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
			if got := ToBase34(tt.args.userId); got != tt.want {
				t.Errorf("TestToBase34() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromBase34(t *testing.T) {
	type args struct {
		acctId string
	}
	tests := []struct {
		name string
		args args
		want uint64
	}{
		{"1", args{"bb5j2mg"}, 0},
		{"1", args{"u"}, 0},
		{"2", args{"3k0ec9"}, 33},
		{"3", args{"b85rtak"}, 34},
		{"3", args{"43z5deh"}, 2000000},
		{"3", args{"ikzuaa"}, 2502145},
		{"3", args{"b6vg7sv"}, 250214500},
		{"3", args{"4m5mihr"}, 250214500},
		{"3", args{"b6q2ti7"}, 250214500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := FromBase34(tt.args.acctId); got != tt.want {
				t.Errorf("TestFromBase34() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapping(t *testing.T) {
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
		acct := ToBase34(id)
		got, _ := FromBase34(acct)
		if got != id {
			t.Fatalf("Boundary test failed: id=%d, acctId=%s, got=%d", id, acct, got)
		}
	}

	// 随机采样测试
	const sampleCount = 10000000 // 1000 万次随机测试，秒级完成
	for i := 0; i < sampleCount; i++ {
		id := rand.Uint64()
		acct := ToBase34(id)
		got, _ := FromBase34(acct)
		if got != id {
			t.Fatalf("Random test failed: id=%d, acctId=%s, got=%d", id, acct, got)
		}
	}
}
