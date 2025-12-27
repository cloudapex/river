// Copyright 2014 loolgame Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package mqrpc

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/cloudapex/river/tools"
	"github.com/vmihailenco/msgpack/v5"
)

var (
	NULL    = "null"    // nil null
	BOOL    = "bool"    // bool
	INT     = "int"     // int
	LONG    = "long"    // int64
	FLOAT   = "float"   // float32
	DOUBLE  = "double"  // float64
	BYTES   = "bytes"   // []byte
	STRING  = "string"  // string
	JSMAP   = "map"     // map[string]any
	CONTEXT = "context" // context
	MARSHAL = "marshal" // mqrpc.Marshaler
	MSGPACK = "msgpack" // msgpack
)

func ArgToData(arg any) (string, []byte, error) {
	if arg == nil {
		return NULL, nil, nil
	}

	switch v2 := arg.(type) {
	case string:
		return STRING, []byte(v2), nil
	case bool:
		return BOOL, tools.BoolToBytes(v2), nil
	case int32:
		return INT, tools.Int32ToBytes(v2), nil
	case int64:
		return LONG, tools.Int64ToBytes(v2), nil
	case float32:
		return FLOAT, tools.Float32ToBytes(v2), nil
	case float64:
		return DOUBLE, tools.Float64ToBytes(v2), nil
	case []byte:
		return BYTES, v2, nil
	case map[string]any:
		bytes, err := tools.MapToBytes(v2)
		return JSMAP, bytes, err
	case context.Context:
		maps := map[string]any{} // 把支持trans的kv序列化到map中再编码进行传输
		for _, k := range getTranslatableCtxKeys() {
			v := v2.Value(k)
			if v == nil {
				continue
			}

			// can Marshaler value
			val, ok := v.(IMarshaler)
			if ok {
				b, err := val.Marshal()
				if err != nil {
					return "", nil, fmt.Errorf("ArgToData args [%s] contextValue.marshal error %v", reflect.TypeOf(arg), err)
				}
				maps[string(k)] = b
			} else { // basic value
				maps[string(k)] = v
			}
		}
		bytes, err := tools.MapToBytes(maps)
		return CONTEXT, bytes, err
	default:

		// 下面必须是struct
		rv := reflect.ValueOf(arg)
		if rv.Kind() != reflect.Ptr {
			return "", nil, fmt.Errorf("ArgToData [%v] not pointer type", reflect.TypeOf(arg))
		}
		if rv.IsNil() { //如果是nil则直接返回
			return NULL, nil, nil
		}
		if rv.Elem().Kind() != reflect.Struct {
			return "", nil, fmt.Errorf("ArgToData [%v] not struct type", reflect.TypeOf(arg))
		}

		// 1 struct for mqrpc.Marshaler
		if v2, ok := arg.(IMarshaler); ok {
			b, err := v2.Marshal()
			if err != nil {
				return "", nil, fmt.Errorf("args [%s] marshal error %v", reflect.TypeOf(arg), err)
			}
			return fmt.Sprintf("%v@%v", MARSHAL, reflect.TypeOf(arg)), b, nil
		}
		// 2 struct for msgpack (default)
		b, err := msgpack.Marshal(arg)
		if err != nil {
			return "", nil, fmt.Errorf("args [%s] msgpack encode(default) error %v", reflect.TypeOf(arg), err)
		}
		return fmt.Sprintf("%v@%v", MSGPACK, reflect.TypeOf(arg)), b, nil
	}
}

func DataToArg(argType string, argData []byte) (any, error) {
	switch {
	case argType == NULL:
		return nil, nil
	case argType == STRING:
		return string(argData), nil
	case argType == BOOL:
		return tools.BytesToBool(argData), nil
	case argType == INT:
		return tools.BytesToInt32(argData), nil
	case argType == LONG:
		return tools.BytesToInt64(argData), nil
	case argType == FLOAT:
		return tools.BytesToFloat32(argData), nil
	case argType == DOUBLE:
		return tools.BytesToFloat64(argData), nil
	case argType == BYTES:
		return argData, nil
	case argType == JSMAP:
		mps, err := tools.BytesToMap(argData)
		if err != nil {
			return nil, err
		}
		return mps, nil
	case argType == CONTEXT:
		mps, err := tools.BytesToMap(argData)
		if err != nil {
			return nil, err
		}

		ctx := context.Background()
		for k, v := range mps {
			makefun := getTranslatableCtxValMakeFun(k)
			if makefun == nil {
				ctx = ContextWithValue(ctx, k, v)
				continue
			}
			obj := makefun()
			if err := Marshal(obj, RpcResult(v, nil)); err != nil {
				return nil, err
			}
			ctx = ContextWithValue(ctx, k, obj)
		}
		return ctx, nil
	case strings.HasPrefix(argType, MARSHAL): // 不能直接解出对象, 让外面解析
		return argData, nil

	case strings.HasPrefix(argType, MSGPACK): // 不能直接解出对象, 让外面解析
		return argData, nil
	}
	return nil, fmt.Errorf("DataToArg [%s] unsupported argType", argType)
}
