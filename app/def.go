package app

// default app instance
var defaultApp IApp = nil

func Default(set ...IApp) IApp {
	if defaultApp != nil {
		return defaultApp
	}

	if len(set) != 0 {
		defaultApp = set[0]
	}

	return defaultApp
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
