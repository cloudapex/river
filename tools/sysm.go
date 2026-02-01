package tools

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

// catch panic
func Catch(desc string, x interface{}) error {
	if x == nil {
		return nil
	}
	head := fmt.Sprintf("%s panic: %v\n", desc, x)

	buf := make([]byte, 256*10)
	size := runtime.Stack(buf, true)
	stack := string(buf[0:size])
	return fmt.Errorf("%v, stack:\n%v", head, stack)
}

// FuncName
func FuncName(fun interface{}) string {
	return FuncFullName(fun, '.')
}
func FuncFullName(fun interface{}, seps ...rune) string {
	return FuncFullNameRef(reflect.ValueOf(fun), seps...)
}
func FuncFullNameRef(valFun reflect.Value, seps ...rune) string {
	fn := runtime.FuncForPC(valFun.Pointer()).Name()
	if len(seps) == 0 {
		return fn
	}

	fields := strings.FieldsFunc(fn, func(sep rune) bool {
		for _, s := range seps {
			if sep == s {
				return true
			}
		}
		return false
	})
	if size := len(fields); size > 0 {
		return strings.Split(fields[size-1], "-")[0]
	}
	return fn
}
