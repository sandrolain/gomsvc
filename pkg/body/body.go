package body

import (
	"encoding/json"
	"strings"

	"github.com/vmihailenco/msgpack/v5"
)

const (
	TypeJson     = "application/json"
	TypeMsgpack  = "application/msgpack"
	TypeXMsgpack = "application/x-msgpack"
	TypeProtobuf = "appliction/protobuf"
)

func dataAsType[T any, R any](data T) (R, bool) {
	var i interface{} = data
	d, ok := i.(R)
	return d, ok
}

func MarshalBody[T any](typ string, data *T) (reqBytes []byte, err error) {
	switch typ {
	case TypeJson:
		reqBytes, err = json.Marshal(*data)
	case TypeMsgpack, TypeXMsgpack:
		reqBytes, err = msgpack.Marshal(*data)
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
	}
	return
}
