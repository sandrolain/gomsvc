package body

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/vmihailenco/msgpack/v5"
	"google.golang.org/protobuf/proto"
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
	case TypeProtobuf:
		d, ok := dataAsType[T, proto.Message](*data)
		if ok {
			reqBytes, err = proto.Marshal(d)
		} else {
			err = fmt.Errorf("not a protobuf Message")
		}
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
	case TypeProtobuf:
		d, ok := dataAsType[R, proto.Message](data)
		if ok {
			err = proto.Unmarshal(resBody, d)
		} else {
			err = fmt.Errorf("not a protobuf Message")
		}
	}
	return
}
