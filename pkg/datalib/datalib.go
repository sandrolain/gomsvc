package datalib

import (
	"encoding/base64"
	"fmt"
	"reflect"

	"github.com/jinzhu/copier"
)

func IsEmpty(val interface{}) bool {
	return val == nil || reflect.DeepEqual(val, reflect.Zero(reflect.TypeOf(val)).Interface())
}

func DefaultIfEmpty[T any](val *T, def T) *T {
	if val == nil {
		return &def
	}
	if reflect.ValueOf(val).IsZero() {
		return &def
	}
	return val
}

func Convert[T any, F any](from *F) (T, error) {
	var res T
	err := copier.Copy(&res, from)
	return res, err
}

func Copy[T any, F any](to *T, from *F) error {
	return copier.CopyWithOption(to, from, copier.Option{IgnoreEmpty: true, DeepCopy: true})
}

func DataURI(data []byte, contentType string) string {
	b64 := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:%s;base64,%s", contentType, b64)
}
