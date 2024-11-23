package datalib

import (
	"reflect"
	"testing"
)

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name string
		val  interface{}
		want bool
	}{
		{
			name: "nil value",
			val:  nil,
			want: true,
		},
		{
			name: "empty string",
			val:  "",
			want: true,
		},
		{
			name: "zero int",
			val:  0,
			want: true,
		},
		{
			name: "non-empty string",
			val:  "hello",
			want: false,
		},
		{
			name: "non-zero int",
			val:  42,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsEmpty(tt.val); got != tt.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultIfEmpty(t *testing.T) {
	tests := []struct {
		name string
		val  *string
		def  string
		want string
	}{
		{
			name: "nil value",
			val:  nil,
			def:  "default",
			want: "default",
		},
		{
			name: "empty string",
			val:  ptr(""),
			def:  "default",
			want: "default",
		},
		{
			name: "non-empty string",
			val:  ptr("value"),
			def:  "default",
			want: "value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DefaultIfEmpty(tt.val, tt.def)
			if *got != tt.want {
				t.Errorf("DefaultIfEmpty() = %v, want %v", *got, tt.want)
			}
		})
	}
}

func TestConvert(t *testing.T) {
	type Source struct {
		Name string
		Age  int
	}
	type Dest struct {
		Name string
		Age  int
	}

	src := &Source{Name: "John", Age: 30}
	dest, err := Convert[Dest](src)
	if err != nil {
		t.Errorf("Convert() error = %v", err)
		return
	}
	if dest.Name != src.Name || dest.Age != src.Age {
		t.Errorf("Convert() = %v, want %v", dest, src)
	}
}

func TestCopy(t *testing.T) {
	type Source struct {
		Name string
		Age  int
	}
	type Dest struct {
		Name string
		Age  int
	}

	src := &Source{Name: "John", Age: 30}
	var dest Dest
	err := Copy(&dest, src)
	if err != nil {
		t.Errorf("Copy() error = %v", err)
		return
	}
	if dest.Name != src.Name || dest.Age != src.Age {
		t.Errorf("Copy() = %v, want %v", dest, src)
	}
}

func TestEncodeDecodeDataURI(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		contentType string
	}{
		{
			name:        "text data",
			data:        []byte("Hello, World!"),
			contentType: "text/plain",
		},
		{
			name:        "binary data",
			data:        []byte{0x00, 0x01, 0x02, 0x03},
			contentType: "application/octet-stream",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uri := EncodeDataURI(tt.data, tt.contentType)
			gotData, gotCT, err := DecodeDataURI(uri)
			if err != nil {
				t.Errorf("DecodeDataURI() error = %v", err)
				return
			}
			if !reflect.DeepEqual(gotData, tt.data) {
				t.Errorf("DecodeDataURI() gotData = %v, want %v", gotData, tt.data)
			}
			if gotCT != tt.contentType {
				t.Errorf("DecodeDataURI() gotCT = %v, want %v", gotCT, tt.contentType)
			}
		})
	}
}

func TestNewSet(t *testing.T) {
	set := NewSet[string]()
	if set == nil {
		t.Error("NewSet() returned nil")
	}
	set.Add("test")
	if !set.Contains("test") {
		t.Error("Set should contain added element")
	}
}

func TestMapSlice(t *testing.T) {
	input := []int{1, 2, 3, 4, 5}
	want := []int{2, 4, 6, 8, 10}
	got := MapSlice(input, func(x int) int { return x * 2 })
	if !reflect.DeepEqual(got, want) {
		t.Errorf("MapSlice() = %v, want %v", got, want)
	}
}

func TestReduceSlice(t *testing.T) {
	input := []int{1, 2, 3, 4, 5}
	want := 15
	got := ReduceSlice(input, 0, func(acc, x int) int { return acc + x })
	if got != want {
		t.Errorf("ReduceSlice() = %v, want %v", got, want)
	}
}

func TestMapToSlice(t *testing.T) {
	input := map[string]int{"a": 1, "b": 2, "c": 3}
	got := MapToSlice(input, func(k string, v int) string { return k + ":" + string(rune('0'+v)) })
	if len(got) != 3 {
		t.Errorf("MapToSlice() length = %v, want 3", len(got))
		return
	}
	// Since map iteration order is not guaranteed, we just check that all expected values are present
	expectedValues := map[string]bool{
		"a:1": true,
		"b:2": true,
		"c:3": true,
	}
	for _, v := range got {
		if !expectedValues[v] {
			t.Errorf("MapToSlice() unexpected value = %v", v)
		}
	}
}

func TestMapKeys(t *testing.T) {
	input := map[string]int{"a": 1, "b": 2, "c": 3}
	got := MapKeys(input)
	if len(got) != len(input) {
		t.Errorf("MapKeys() length = %v, want %v", len(got), len(input))
	}
	for _, k := range got {
		if _, ok := input[k]; !ok {
			t.Errorf("MapKeys() unexpected key = %v", k)
		}
	}
}

func TestMapValues(t *testing.T) {
	input := map[string]int{"a": 1, "b": 2, "c": 3}
	got := MapValues(input)
	if len(got) != len(input) {
		t.Errorf("MapValues() length = %v, want %v", len(got), len(input))
	}
	sum := 0
	for _, v := range got {
		sum += v
	}
	if sum != 6 {
		t.Errorf("MapValues() sum = %v, want 6", sum)
	}
}

func TestConvertType(t *testing.T) {
	input := []interface{}{1, 2, 3, 4, 5}
	got, err := ConvertType[int](input)
	if err != nil {
		t.Errorf("ConvertType() error = %v", err)
		return
	}
	if len(got) != len(input) {
		t.Errorf("ConvertType() length = %v, want %v", len(got), len(input))
	}
	for i, v := range got {
		if v != input[i] {
			t.Errorf("ConvertType() index %d = %v, want %v", i, v, input[i])
		}
	}

	// Test error case
	invalidInput := []interface{}{1, "2", 3}
	_, err = ConvertType[int](invalidInput)
	if err == nil {
		t.Error("ConvertType() expected error for invalid input")
	}
}

// Helper function to create pointer to string
func ptr(s string) *string {
	return &s
}
