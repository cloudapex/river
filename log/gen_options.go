package log

//go:generate optiongen --option_with_struct_name=false
func OptionsOptionDeclareWithDefault() any {
	return map[string]any{
		"Debug":       false,
		"ProcessID":   "",
		"LogDir":      "",
		"LogFileName": func(logdir, prefix, processID, suffix string) string { return "" },
		"BiDir":       "",
		"BIFileName":  func(logdir, prefix, processID, suffix string) string { return "" },
		"BiSetting":   map[string]any{},
		"LogSetting":  map[string]any{},
	}
}
