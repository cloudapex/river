package mqrpc

import (
	"context"
	"sync"
)

// 支持rpc trans的Context Keys
var (
	contextKeysMutex sync.RWMutex
	transContextKeys = map[string]func() IMarshaler{}
)

// 使用此WithValue方法才能通过Context传递数据
func ContextWithValue(ctx context.Context, key string, val any) context.Context {
	addTransContextKey(key)
	return context.WithValue(ctx, key, val)
}

// 提前注册复合类型的Context val数据(基本类型不需要注册)
func RegTransContextKey(key string, makeFun func() IMarshaler) {
	transContextKeys[key] = makeFun
}

func addTransContextKey(key string) {
	if hasTransContextKey(key) {
		return
	}
	contextKeysMutex.Lock()
	defer contextKeysMutex.Unlock()

	transContextKeys[key] = nil
}
func hasTransContextKey(key string) bool {
	contextKeysMutex.RLock()
	defer contextKeysMutex.RUnlock()

	_, exists := transContextKeys[key]
	return exists
}
func getTransContextKeys() []string {
	contextKeysMutex.RLock()
	defer contextKeysMutex.RUnlock()

	ks := make([]string, len(transContextKeys))
	for k := range transContextKeys {
		ks = append(ks, k)
	}
	return ks
}
func getTransContextKeyItem(key string) func() IMarshaler {
	contextKeysMutex.RLock()
	defer contextKeysMutex.RUnlock()
	return transContextKeys[key]
}
