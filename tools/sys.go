package tools

import (
	"fmt"
	"runtime"
)

// catch panic
func Catch(panicRet any) error {
	if panicRet == nil {
		return nil
	}
	buf := make([]byte, 512*3)
	size := runtime.Stack(buf, false)
	stack := string(buf[0:size])
	return fmt.Errorf("%v, stack:\n%v", panicRet, stack)
}
