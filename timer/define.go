package timer

import (
	"reflect"
	"time"
)

const C_TIMER_TICK_INTERVAL = 100 * time.Millisecond // 定时器默认tick精度

// ==================== Timer

// 状态存储接口
type ITimerRestorer interface {

	// 恢复状态
	Load() map[string]int64

	// 保存状态
	Save(key string, valAt int64)
}

// 定时器接口
type ITimer interface {

	// 启动
	StartTimer(restorer ...ITimerRestorer)

	// 关闭
	CloseTimer()

	// 添加定时器[当handle返回false时自动删除此TimerHandler]
	AddTimer(d time.Duration, handle TTimerHandFunc, opt ...*TimerOpt)

	// 添加每天一次的定时器[同上](after0:每天超过零点多少时间)
	AddDailyTimer(after0 time.Duration, handle TTimerHandFunc, opt ...*TimerOpt)

	// 移除定时器
	DelTimer(handle TTimerHandFunc)
	DelTimerByName(name string)
}
type TTimerHandFunc func(now time.Time) (keep bool)

// ==================== cronjob
type cronjob struct {
	name         string        // 定时任务的名称
	last         time.Time     // 上次执行的时间
	daily, store bool          // 是否是日常类型,是否需要保存状态
	interval     time.Duration // 调用间隔时间
	fun          reflect.Value // 调用函数
}

// ==================== TimerOpt(可选参数)
type TimerOpt struct {
	Name  string // 自定义名称(默认为函数名util.FuncFullNameRef)
	Right bool   // 是否立刻执行(daily类:当天时间满足的话则执行)
	Store bool   // 是否需要保存状态
}
