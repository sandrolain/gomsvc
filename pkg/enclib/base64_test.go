package enclib

import (
	"reflect"
	"testing"
)

func TestEncodeBase64(t *testing.T) {
	tests := []struct {
		name string
		arg  []byte
		want string
	}{
		{"empty", []byte(""), ""},
		{"not empty", []byte("Hello world"), "SGVsbG8gd29ybGQ="},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EncodeBase64(tt.arg); got != tt.want {
				t.Errorf("EncodeBase64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecodeBase64(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    []byte
		wantErr bool
	}{
		{"empty", "", []byte{}, false},
		{"not base64", "not base64", nil, true},
		{"base64", "SGVsbG8gd29ybGQ=", []byte("Hello world"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeBase64(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeBase64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DecodeBase64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecodeBase64URL(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    []byte
		wantErr bool
	}{
		{"empty", "", []byte{}, false},
		{"not base64", "not base64", nil, true},
		{"base64", "SGVsbG8gd29ybGQ=", []byte("Hello world"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeBase64URL(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeBase64URL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DecodeBase64URL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEncodeBase64URL(t *testing.T) {
	tests := []struct {
		name string
		arg  []byte
		want string
	}{
		{"empty", []byte{}, ""},
		{"not empty", []byte("Hello world"), "SGVsbG8gd29ybGQ="},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EncodeBase64URL(tt.arg); got != tt.want {
				t.Errorf("EncodeBase64URL() = %v, want %v", got, tt.want)
			}
		})
	}
}
