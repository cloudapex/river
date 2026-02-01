package tools

// Args 参数
func Args(params ...interface{}) []interface{} { return params }

// Cast 三元运算符
func Tern[T bool, U any](isTrue T, ifValue U, elseValue U) U {
	if isTrue {
		return ifValue
	} else {
		return elseValue
	}
}

// Cast 三元方法
func Cast(condition bool, trueFun, falseFun func()) {
	if condition {
		if trueFun != nil {
			trueFun()
		}
	} else {
		if falseFun != nil {
			falseFun()
		}
	}
}

// DefaultVal 默认值
func DefaultVal[T any](vars []T) T {
	var zero T
	if len(vars) > 0 {
		return vars[0]
	}
	return zero
}
