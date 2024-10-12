package enclib

import (
	"reflect"
	"testing"
)

func TestHexDecode(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    []byte
		wantErr bool
	}{
		{"empty", "", []byte(""), false},
		{"not empty", "48656c6c6f", []byte("Hello"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HexDecode(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("HexDecode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HexDecode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHexEncode(t *testing.T) {
	tests := []struct {
		name string
		arg  []byte
		want string
	}{
		{"empty", []byte(""), ""},
		{"not empty", []byte("Hello"), "48656c6c6f"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HexEncode(tt.arg); got != tt.want {
				t.Errorf("HexEncode() = %v, want %v", got, tt.want)
			}
		})
	}
}
