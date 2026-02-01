package modules

import (
	"time"

	"github.com/cloudapex/river/app"
	"github.com/cloudapex/river/conf"
	timewheel "github.com/cloudapex/river/module/modules/timer"
)

var TimerModule = func() app.IModule {
	Timer := new(Timer)
	return Timer
}

// 定时器模块(会被自动内置到框架中运行)
type Timer struct {
	app.IModule
}

// 换掉
func (m *Timer) GetType() string {
	// 很关键,需要与配置文件中的Module配置对应
	return "Timer"
}
func (m *Timer) Version() string { return "1.0.0" }

func (m *Timer) OnInit(settings *conf.ModuleSettings) {
	timewheel.SetTimeWheel(timewheel.New(10*time.Millisecond, 36))
	// 时间轮使用方式:
	// 执行过的定时器会自动被删除
	//timewheel.GetTimeWheel().AddTimer(66 * time.Millisecond , nil,self.Update)
	//
	//timewheel.GetTimeWheel().AddTimerCustom(66 * time.Millisecond ,"baba", nil,self.Update)
	// 删除一个为执行的定时器, 参数为添加定时器传递的唯一标识
	//timewheel.GetTimeWheel().RemoveTimer("baba")
}

func (m *Timer) Run(closeSig chan bool) {
	timewheel.GetTimeWheel().Start(closeSig)
}

func (m *Timer) OnDestroy() {
}
