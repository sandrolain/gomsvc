package repo

import (
	"fmt"
	"reflect"
	"strings"
)

func getIdValue[T any, K any](s *T) (*K, bool) {
	sv := reflect.ValueOf(s)
	if sv.Kind() == reflect.Ptr {
		sv = sv.Elem()
	}
	st := sv.Type()
	numFields := st.NumField()
	for i := 0; i < numFields; i++ {
		fieldType := st.Field(i)
		fieldValue := sv.Field(i)
		tag := fieldType.Tag.Get("bson")
		if tag == "" {
			continue
		}
		tagParts := strings.Split(tag, ",")
		if tagParts[0] != "_id" {
			continue
		}
		if fieldValue.IsZero() {
			return nil, false
		}
		el := fieldValue.Interface()
		res, ok := el.(K)
		return &res, ok
	}
	return nil, false
}

func setIdValue[T any, K any](s *T, v *K) (bool, error) {
	sv := reflect.ValueOf(s)
	if sv.Kind() == reflect.Ptr {
		sv = sv.Elem()
	}
	st := sv.Type()
	numFields := st.NumField()
	for i := 0; i < numFields; i++ {
		fieldType := st.Field(i)
		fieldValue := sv.Field(i)
		tag := fieldType.Tag.Get("bson")
		if tag == "" {
			continue
		}
		tagParts := strings.Split(tag, ",")
		if tagParts[0] != "_id" {
			continue
		}
		if !fieldValue.IsZero() {
			return false, nil
		}
		if !fieldValue.CanSet() {
			return false, fmt.Errorf("value could not be set")
		}
		val := reflect.ValueOf(v).Elem()
		fieldValue.Set(val)
		return true, nil
	}
	return false, nil
}
