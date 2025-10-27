package mqrpc

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/vmihailenco/msgpack/v5"
)

// ErrNil ErrNil
var ErrNil = errors.New("mqrpc: nil returned")

type callResult struct {
	Reply any
	Error error
}

// 拼装CallResult
func RpcResult(reply any, err error) callResult {
	return callResult{
		Reply: reply,
		Error: err,
	}
}

// Int Int
func Int(reply any, err error) (int, error) {
	if err != nil {
		return 0, err
	}

	switch reply := reply.(type) {
	case int64:
		x := int(reply)
		if int64(x) != reply {
			return 0, strconv.ErrRange
		}
		return x, nil
	case nil:
		return 0, ErrNil
	}
	return 0, fmt.Errorf("mqrpc: unexpected type for Int, got type %T", reply)
}

// Int64 is a helper that converts a command reply to 64 bit integer. If err is
// not equal to nil, then Int returns 0, err. Otherwise, Int64 converts the
// reply to an int64 as follows:
//
//	Reply type    Result
//	integer       reply, nil
//	bulk string   parsed reply, nil
//	nil           0, ErrNil
//	other         0, error
func Int64(reply any, err error) (int64, error) {
	if err != nil {
		return 0, err
	}

	switch reply := reply.(type) {
	case int64:
		return reply, nil
	case nil:
		return 0, ErrNil
	}
	return 0, fmt.Errorf("mqrpc: unexpected type for Int64, got type %T", reply)
}

// Float64 is a helper that converts a command reply to 64 bit float. If err is
// not equal to nil, then Float64 returns 0, err. Otherwise, Float64 converts
// the reply to an int as follows:
//
//	Reply type    Result
//	bulk string   parsed reply, nil
//	nil           0, ErrNil
//	other         0, error
func Float64(reply any, err error) (float64, error) {
	if err != nil {
		return 0, err
	}

	switch reply := reply.(type) {
	case float64:
		return reply, nil
	case nil:
		return 0, ErrNil
	}
	return 0, fmt.Errorf("mqrpc: unexpected type for Float64, got type %T", reply)
}

// String is a helper that converts a command reply to a string. If err is not
// equal to nil, then String returns "", err. Otherwise String converts the
// reply to a string as follows:
//
//	Reply type      Result
//	bulk string     string(reply), nil
//	simple string   reply, nil
//	nil             "",  ErrNil
//	other           "",  error
func String(reply any, err error) (string, error) {
	if err != nil {
		return "", err
	}

	switch reply := reply.(type) {
	case string:
		return reply, nil
	case nil:
		return "", ErrNil
	}
	return "", fmt.Errorf("mqrpc: unexpected type for String, got type %T", reply)
}

// Bytes is a helper that converts a command reply to a slice of bytes. If err
// is not equal to nil, then Bytes returns nil, err. Otherwise Bytes converts
// the reply to a slice of bytes as follows:
//
//	Reply type      Result
//	bulk string     reply, nil
//	simple string   []byte(reply), nil
//	nil             nil, ErrNil
//	other           nil, error
func Bytes(reply any, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}
	switch reply := reply.(type) {
	case []byte:
		return reply, nil
	case nil:
		return nil, ErrNil
	}
	return nil, fmt.Errorf("mqrpc: unexpected type for Bytes, got type %T", reply)
}

func Bool(reply any, err error) (bool, error) {
	if err != nil {
		return false, err
	}

	switch reply := reply.(type) {
	case bool:
		return reply, nil
	case nil:
		return false, ErrNil
	}
	return false, fmt.Errorf("mqrpc: unexpected type for Bool, got type %T", reply)
}

// JsMap JsMap
func JsMap(reply any, err error) (map[string]any, error) {
	if err != nil {
		return nil, err
	}

	switch reply := reply.(type) {
	case map[string]any:
		return reply, nil
	case nil:
		return nil, ErrNil
	}
	return nil, fmt.Errorf("mqrpc: unexpected type for Bool, got type %T", reply)
}

// Marshal Marshal
func Marshal(pObj any, ret callResult) error {
	if ret.Error != nil {
		return ret.Error
	}

	rv := reflect.ValueOf(pObj)
	if rv.Kind() != reflect.Ptr {
		//不是指针
		return fmt.Errorf("pObj [%v] not *mqrpc.marshaler pointer type", rv.Type())
	}
	if v2, ok := pObj.(Marshaler); ok {
		switch r := ret.Reply.(type) {
		case []byte:
			err := v2.Unmarshal(r)
			if err != nil {
				return err
			}
			return nil
		case nil:
			return ErrNil
		}
	} else {
		return fmt.Errorf("pObj [%v] not *mqrpc.marshaler type", rv.Type())
	}
	return fmt.Errorf("mqrpc: unexpected type for %v, got type %T", reflect.ValueOf(ret.Reply), ret.Reply)
}

// MsgPack MsgPack
func MsgPack(pObj any, ret callResult) error {
	if ret.Error != nil {
		return ret.Error
	}

	rv := reflect.ValueOf(pObj)
	if rv.Kind() != reflect.Ptr { //不是指针
		return fmt.Errorf("pObj [%v] not struct pointer type", rv.Type())
	}

	switch r := ret.Reply.(type) {
	case []byte:
		if err := msgpack.Unmarshal(r, pObj); err != nil {
			return fmt.Errorf("msgpack unmarshal error: %v", err)
		}
		return nil
	case nil:
		return ErrNil
	}
	return fmt.Errorf("mqrpc: unexpected type for %v, got type %T", reflect.ValueOf(ret.Reply), ret.Reply)
}
