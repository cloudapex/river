package app

// default app instance
var app IApp = nil

func App(set ...IApp) IApp {
	if app != nil {
		return app
	}

	if len(set) != 0 {
		app = set[0]
	}

	return app
}

// 启动参数(命令行优先级比环境变量高)
type startUpArgs struct {
	WordDir    string `env:"wd" env-default:""`
	ProcessEnv string `env:"env" env-default:"dev"`
	ConsulAddr string `env:"consul" env-default:"127.0.0.1:8500"` // use configKey: config/{env}/server
	LogPath    string `env:"log" env-default:""`
	BiPath     string `env:"bi" env-default:""`
	PProfAddr  string `env:"pprof" env-default:""`
}
