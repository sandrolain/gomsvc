package api

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/sandrolain/gomsvc/pkg/body"
	"github.com/vmihailenco/msgpack/v5"
	"google.golang.org/protobuf/proto"
)

func loadData[T any](ctx *fiber.Ctx, dest *T) error {
	if err := ctx.ReqHeaderParser(dest); err != nil {
		return err
	}
	if err := ctx.QueryParser(dest); err != nil {
		return err
	}
	if err := ctx.ParamsParser(dest); err != nil {
		return err
	}
	tagName := "req"
	sv := reflect.ValueOf(dest)
	if sv.Kind() == reflect.Ptr {
		sv = sv.Elem()
	}
	st := sv.Type()
	numFields := st.NumField()
	for i := 0; i < numFields; i++ {
		fieldType := st.Field(i)
		fieldValue := sv.Field(i)
		tag := fieldType.Tag.Get(tagName)
		if tag != "" {
			src, key, enc := getTagParts(tag)
			var err error
			switch src {
			case "body":
				err = extractBody(&fieldType, &fieldValue, ctx)
			case "query":
				if str := ctx.Query(key); len(str) > 0 {
					err = convertValue(&fieldType, &fieldValue, str, enc)
				}
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func getTagParts(tag string) (source string, key string, enc string) {
	tagParts := strings.Split(tag, ":")
	source = tagParts[0]
	if len(tagParts) > 1 {
		key = tagParts[1]
	}
	if len(tagParts) > 2 {
		enc = tagParts[2]
	}
	return
}

func extractBody(ft *reflect.StructField, fv *reflect.Value, ctx *fiber.Ctx) (err error) {
	field := *ft
	fieldValue := *fv
	ptr := reflect.New(field.Type).Interface()
	if err = ctx.BodyParser(ptr); err != nil {
		if err != fiber.ErrUnprocessableEntity {
			return
		}
		err = nil
		resType := strings.Split(ctx.Get("Content-Type"), ";")
		switch resType[0] {
		case body.TypeMsgpack, body.TypeXMsgpack:
			err = msgpack.Unmarshal(ctx.Body(), ptr)
		case body.TypeProtobuf:
			d, ok := ptr.(proto.Message)
			if ok {
				err = proto.Unmarshal(ctx.Body(), d)
			} else {
				err = fmt.Errorf("not a protobuf Message")
			}
		}
		if err != nil {
			return
		}
	}
	refVal := reflect.ValueOf(ptr).Elem()
	if fieldValue.CanSet() {
		fieldValue.Set(refVal)
	}
	return nil
}

func convertValue(ft *reflect.StructField, fv *reflect.Value, str string, enc string) error {
	fieldType := *ft
	fieldValue := *fv
	var refVal reflect.Value
	fieldValueType := fieldValue.Type()
	fieldValueTypeName := fieldValueType.Name()
	switch fieldValueTypeName {
	case "int":
		v, err := strconv.Atoi(str)
		if err != nil {
			return err
		}
		refVal = reflect.ValueOf(v)
	case "int32":
		v, err := strconv.ParseInt(str, 10, 32)
		if err != nil {
			return err
		}
		refVal = reflect.ValueOf(v)
	case "int64":
		v, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return err
		}
		refVal = reflect.ValueOf(v)
	case "float32":
		v, err := strconv.ParseFloat(str, 32)
		if err != nil {
			return err
		}
		refVal = reflect.ValueOf(v)
	case "float64":
		v, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return err
		}
		refVal = reflect.ValueOf(v)
	case "string":
		refVal = reflect.ValueOf(str)
	default:
		switch enc {
		case "json":
			ptr := reflect.New(fieldType.Type).Interface()
			err := json.Unmarshal([]byte(str), ptr)
			if err != nil {
				return err
			}
			refVal = reflect.ValueOf(ptr).Elem()
		case "csv":
			parts := strings.Split(str, ",")
			refVal = reflect.ValueOf(parts).Elem()
		}
	}
	fieldValue.Set(refVal)
	return nil
}

type EmptyData struct{}
