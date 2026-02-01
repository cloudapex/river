package timer

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/cloudapex/river/log"
	"github.com/cloudapex/river/tools"
)

func NewTimer(tickerInterval ...time.Duration) ITimer {
	t := &timer{}
	t.Init(tickerInterval...)
	return t
}

// timer
type timer struct {
	mutex    tools.RWLocker
	cronjobs map[string]*cronjob

	interval time.Duration // 定时器tick时间间隔

	sgExit   chan int
	wgExit   sync.WaitGroup
	restorer ITimerRestorer // 外部状态存储器
}

func (this *timer) Init(tickerInterval ...time.Duration) {
	this.cronjobs = make(map[string]*cronjob)

	this.interval = tools.DefaultVal(tickerInterval)
	tools.Cast(this.interval == 0, func() { this.interval = C_TIMER_TICK_INTERVAL }, nil)
}
func (this *timer) StartTimer(restorer ...ITimerRestorer) {
	this.sgExit = make(chan int)

	tools.Cast(tools.DefaultVal(restorer) != nil, func() { this.restore(restorer[0]) }, nil)

	go this.loop()
	return
}
func (this *timer) CloseTimer() {
	tools.Cast(this.sgExit != nil, func() { this.sgExit <- 0 }, nil)

	this.wgExit.Wait()
}

func (this *timer) AddTimer(interval time.Duration, handle TTimerHandFunc, opt ...*TimerOpt) {
	this.addCronjob(interval, handle, false, opt...)
}
func (this *timer) AddDailyTimer(absolute time.Duration, handle TTimerHandFunc, opt ...*TimerOpt) {
	this.addCronjob(absolute, handle, true, opt...)
}
func (this *timer) DelTimer(handle TTimerHandFunc) {
	defer this.mutex.UnLock(this.mutex.Lock())
	delete(this.cronjobs, tools.FuncFullNameRef(reflect.ValueOf(handle)))
}
func (this *timer) DelTimerByName(name string) {
	defer this.mutex.UnLock(this.mutex.Lock())
	delete(this.cronjobs, name)
}

// --------------------

func (this *timer) tick() {
	defer this.mutex.RUnLock(this.mutex.RLock())
	if len(this.cronjobs) == 0 {
		return
	}

	curTime := time.Now()
	for name, t := range this.cronjobs {
		lstTime := t.last

		if t.daily {
			if curTime.YearDay() == t.last.YearDay() {
				continue
			}
			lstTime = time.Date(curTime.Year(), curTime.Month(), curTime.Day(), 0, 0, 0, 0, curTime.Location())
		}

		if curTime.Sub(lstTime) <= t.interval {
			continue
		}

		t.last = curTime
		if t.store && this.restorer != nil {
			this.restorer.Save(t.name, t.last.Unix())
		}

		_name, _fun := name, t.fun
		go func(wg ...*sync.WaitGroup) {
			if w := tools.DefaultVal(wg); w != nil {
				w.Add(1)
				defer func() { w.Done() }()
			}
			defer func() {
				if err := tools.Catch(fmt.Sprintf("cronjob[%q]", name), recover()); err != nil {
					log.Error(err.Error())
				}
			}()

			if ret := _fun.Call([]reflect.Value{reflect.ValueOf(curTime)}); !ret[0].Bool() {
				this.DelTimerByName(_name)
			}

		}(&this.wgExit)

	}
}
func (this *timer) loop() {
	t := time.NewTicker(this.interval)

	defer func() {
		if err := tools.Catch("Timer.loop() panic and it will resume", recover()); err != nil {
			go this.loop()
		}
	}()

	for {
		select {
		case <-t.C:
			this.tick()
		case <-this.sgExit:
			return
		}
	}
}
func (this *timer) addCronjob(interval time.Duration, handle TTimerHandFunc, daily bool, opts ...*TimerOpt) {
	var tOpt TimerOpt
	if opt := tools.DefaultVal(opts); opt != nil {
		tOpt = *opt
	}

	last := time.Now()
	if tOpt.Right {
		if daily { // 当天时间只要满足则执行,否则次日才会开始执行
			last = time.Time{}
		} else if !handle(time.Now()) { // 立即执行
			return
		}
	}

	job := &cronjob{tOpt.Name, last, daily, tOpt.Store, interval, reflect.ValueOf(handle)}
	if job.name == "" {
		job.name = tools.FuncFullNameRef(job.fun)
	}

	defer this.mutex.UnLock(this.mutex.Lock())
	this.cronjobs[job.name] = job
}
func (this *timer) restore(restorer ITimerRestorer) {
	this.restorer = restorer

	for name, last := range restorer.Load() {
		for name_, it := range this.cronjobs {
			if name_ == name {
				it.last = time.Unix(last, 0)
				break
			}
		}
	}
}
