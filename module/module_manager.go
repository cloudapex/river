// Package basemodule  模块管理器
package module

import (
	"fmt"
	"sync"

	"github.com/cloudapex/river/app"
	"github.com/cloudapex/river/conf"
	"github.com/cloudapex/river/log"
	"github.com/cloudapex/river/tools"
)

// NewModuleManager 新建模块管理器
func NewModuleManager() *ModuleManager {
	return &ModuleManager{}
}

// moduleUnit 模块结构
type moduleUnit struct {
	mi       app.IModule
	settings *conf.ModuleSettings // from Config.Settings
	closeSig chan bool
	wg       sync.WaitGroup
}

// ModuleManager 模块管理器
type ModuleManager struct {
	mods    []*moduleUnit // 注册的modules
	runMods []*moduleUnit // 真正运行的modules
}

// Register 注册模块
func (this *ModuleManager) Register(mi app.IModule) {
	this.mods = append(this.mods, &moduleUnit{
		mi:       mi,
		closeSig: make(chan bool, 1),
	})
}

// RegisterRun 注册需要运行的模块
func (this *ModuleManager) RegisterRun(mi app.IModule) {
	this.runMods = append(this.runMods, &moduleUnit{
		mi:       mi,
		closeSig: make(chan bool, 1),
	})
}

// Init 初始化
func (this *ModuleManager) Init(processEnv string) {
	log.Info("This server app process run ProcessEnvId is [%s]", processEnv)

	// 配置文件规则检查(没通过的话直接panic)
	this.checkModuleSettings()

	// 程序注册的module与配置中的module进行匹配,得到最终runMods
	for i := 0; i < len(this.mods); i++ {
		for typ, modSettings := range app.App().Config().Module {
			if this.mods[i].mi.GetType() == typ { // this.mods[i]匹配Conf.ModuleType
				for _, setting := range modSettings {
					// 这里可能有BUG 公网IP和局域网IP处理方式可能不一样,先不管
					if processEnv == setting.ProcessEnv { // 有匹配到
						this.mods[i].settings = setting
						this.runMods = append(this.runMods, this.mods[i]) // 加入到运行列表中
						break
					}
				}
				break // 跳出内部循环
			}
		}
	}

	// 初始化并运行模块
	for i := 0; i < len(this.runMods); i++ {
		m := this.runMods[i]
		m.mi.OnInit(m.settings)

		if app.App().GetModuleInited() != nil {
			app.App().GetModuleInited()(m.mi)
		}

		m.wg.Add(1)
		go func(unit *moduleUnit) {
			defer func() {
				if err := tools.Catch("module run", recover()); err != nil {
					log.Error("module[%q] run panic: %v", unit.mi.GetType(), err)
				}
			}()
			unit.mi.Run(unit.closeSig)
			unit.wg.Done()
		}(m)
	}
	//timer.SetTimer(3, this.ReportStatistics, nil) //统计数据定时任务
}

// Destroy 停止模块(倒序)
func (this *ModuleManager) Destroy() {
	for i := len(this.runMods) - 1; i >= 0; i-- {
		m := this.runMods[i]
		m.closeSig <- true
		m.wg.Wait()
		func(unit *moduleUnit) {
			defer func() {
				if err := tools.Catch("module destroy", recover()); err != nil {
					log.Error("module[%q] destroy panic: %v", unit.mi.GetType(), err)
				}
			}()
			unit.mi.OnDestroy()
		}(m)
	}
}

// checkModuleSettings module配置文件规则检查(ID全局必须唯一) 且 每个类型的Module在同一个ProcessEnv中只能配置一个
func (this *ModuleManager) checkModuleSettings() {
	gid := map[string]string{} // 用来保存全局ID:ModuleType
	for typ, modSettings := range app.App().Config().Module {
		pid := map[string]string{} // 用来保存模块中的 ProcessEnv:ID
		for _, setting := range modSettings {
			if Stype, ok := gid[setting.ID]; ok {
				//如果Id已经存在,说明有两个相同Id的模块,这种情况不能被允许,这里就直接抛异常 强制崩溃以免以后调试找不到问题
				panic(fmt.Sprintf("Module.ID (%s) been used in modules of type [%s] and cannot be reused", setting.ID, Stype))
			} else {
				gid[setting.ID] = typ
			}

			if id, ok := pid[setting.ProcessEnv]; ok {
				//如果ProcessEnv已经存在,说明有两个相同ProcessEnv的模块,这种情况不能被允许,这里就直接抛异常 强制崩溃以免以后调试找不到问题
				panic(fmt.Sprintf("In the list of modules of type [%s], ProcessEnv (%s) has been used for ID module for (%s)", typ, setting.ProcessEnv, id))
			} else {
				pid[setting.ProcessEnv] = setting.ID
			}
		}
	}
}
