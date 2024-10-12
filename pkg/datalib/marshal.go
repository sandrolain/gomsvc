package datalib

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/vmihailenco/msgpack/v5"
)

const (
	TypeJson     = "application/json"
	TypeMsgpack  = "application/msgpack"
	TypeXMsgpack = "application/x-msgpack"
	TypeProtobuf = "application/protobuf"
)

func MarshalBody[T any](typ string, data *T) (reqBytes []byte, err error) {
	switch typ {
	case TypeJson:
		reqBytes, err = json.Marshal(*data)
	case TypeMsgpack, TypeXMsgpack:
		reqBytes, err = msgpack.Marshal(*data)
	default:
		err = fmt.Errorf("unknown type: %s", typ)
	}
	return
}

func UnmarshalBody[R any](typ string, resBody []byte) (data R, err error) {
	resType := strings.Split(typ, ";")
	switch resType[0] {
	case TypeJson:
		err = json.Unmarshal(resBody, &data)
	case TypeMsgpack, TypeXMsgpack:
		err = msgpack.Unmarshal(resBody, &data)
	default:
		err = fmt.Errorf("unknown type: %s", typ)
	}
	return
}
