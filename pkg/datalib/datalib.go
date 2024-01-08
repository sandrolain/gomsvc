package datalib

import (
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
