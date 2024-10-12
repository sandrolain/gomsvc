package datalib

import (
	"fmt"
	"reflect"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/jinzhu/copier"
	"github.com/vincent-petithory/dataurl"
)

func IsEmpty(val interface{}) bool {
	return val == nil || reflect.DeepEqual(val, reflect.Zero(reflect.TypeOf(val)).Interface())
}

func DefaultIfEmpty[T any](val *T, def T) *T {
	if val == nil {
		return &def
	}
	if reflect.ValueOf(*val).IsZero() {
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

func EncodeDataURI(data []byte, contentType string) string {
	d := dataurl.New(data, contentType)
	return d.String()
}

func DecodeDataURI(uri string) (data []byte, ct string, err error) {
	dataURL, err := dataurl.DecodeString(uri)
	if err != nil {
		return
	}
	data = dataURL.Data
	ct = dataURL.ContentType()
	return
}

func NewSet[T comparable]() mapset.Set[T] {
	return mapset.NewSet[T]()
}

func MapSlice[T any, R any](s []T, f func(T) R) []R {
	res := make([]R, len(s))
	for i, item := range s {
		res[i] = f(item)
	}
	return res
}

func ReduceSlice[T any, R any](s []T, res R, f func(R, T) R) R {
	for _, item := range s {
		res = f(res, item)
	}
	return res
}

func MapToSlice[K comparable, V any, R any](m map[K]V, f func(K, V) R) []R {
	res := make([]R, len(m))
	for k, v := range m {
		res = append(res, f(k, v))
	}
	return res
}

func MapKeys[K comparable, V any](m map[K]V) []K {
	res := make([]K, len(m))
	i := 0
	for k := range m {
		res[i] = k
		i++
	}
	return res
}

func MapValues[K comparable, V any](m map[K]V) []V {
	res := make([]V, len(m))
	i := 0
	for _, v := range m {
		res[i] = v
		i++
	}
	return res
}

func ConvertType[T interface{}](all []interface{}) ([]T, error) {
	res := make([]T, len(all))
	for i, v := range all {
		val, ok := v.(T)
		if !ok {
			return res, fmt.Errorf("item %v not convertible", i)
		}
		res[i] = val
	}
	return res, nil
}
