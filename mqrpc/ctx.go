package mqrpc

import (
	"context"
	"sync"

	"github.com/cloudapex/river/log"
)

// Context value 复合类型的创建函数
type FunMakeCtxValue func() IMarshaler

// 支持rpc trans的Context Keys
var (
	contextKeysMutex    sync.RWMutex
	translatableCtxKeys = map[string]FunMakeCtxValue{}
)

// 默认注册log.RPC_CONTEXT_KEY_TRACE
func init() {
	RegTranslatableCtxKey(log.RPC_CONTEXT_KEY_TRACE, func() IMarshaler {
		return log.CreateRootTrace()
	})
}

// 使用此WithValue方法才能通过Context传递数据
func ContextWithValue(ctx context.Context, key string, val any) context.Context {
	addTranslatableCtxKey(key)
	return context.WithValue(ctx, key, val)
}

// 提前注册复合类型的Context val数据(基本类型不需要注册)
func RegTranslatableCtxKey(key string, makeFun FunMakeCtxValue) {
	translatableCtxKeys[key] = makeFun
}

func addTranslatableCtxKey(key string) {
	if hasTranslatableCtxKey(key) {
		return
	}
	contextKeysMutex.Lock()
	defer contextKeysMutex.Unlock()

	translatableCtxKeys[key] = nil
}
func hasTranslatableCtxKey(key string) bool {
	contextKeysMutex.RLock()
	defer contextKeysMutex.RUnlock()

	_, exists := translatableCtxKeys[key]
	return exists
}
func getTranslatableCtxKeys() []string {
	contextKeysMutex.RLock()
	defer contextKeysMutex.RUnlock()

	ks := make([]string, 0, len(translatableCtxKeys))
	for k := range translatableCtxKeys {
		ks = append(ks, k)
	}
	return ks
}
func getTranslatableCtxValMakeFun(key string) FunMakeCtxValue {
	contextKeysMutex.RLock()
	defer contextKeysMutex.RUnlock()
	return translatableCtxKeys[key]
}
