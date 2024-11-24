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
		{"basic string", []byte("Hello world"), "SGVsbG8gd29ybGQ="},
		{"binary data", []byte{0x00, 0xFF, 0x42}, "AP9C"},
		{"unicode", []byte("Hello 世界"), "SGVsbG8g5LiW55WM"},
		{"with padding 1", []byte("a"), "YQ=="},
		{"with padding 2", []byte("ab"), "YWI="},
		{"no padding", []byte("abc"), "YWJj"},
		{"special chars", []byte("!@#$%^&*()"), "IUAjJCVeJiooKQ=="},
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
		{"basic string", "SGVsbG8gd29ybGQ=", []byte("Hello world"), false},
		{"binary data", "AP9C", []byte{0x00, 0xFF, 0x42}, false},
		{"unicode", "SGVsbG8g5LiW55WM", []byte("Hello 世界"), false},
		{"with padding 1", "YQ==", []byte("a"), false},
		{"with padding 2", "YWI=", []byte("ab"), false},
		{"no padding", "YWJj", []byte("abc"), false},
		{"special chars", "IUAjJCVeJiooKQ==", []byte("!@#$%^&*()"), false},
		{"invalid base64", "not base64", nil, true},
		{"invalid padding", "YQ=", nil, true},
		{"incomplete input", "YWJjZ", nil, true},
		{"whitespace", "SGVs bG8=", nil, true},
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

func TestEncodeBase64URL(t *testing.T) {
	tests := []struct {
		name string
		arg  []byte
		want string
	}{
		{"empty", []byte{}, ""},
		{"basic string", []byte("Hello world"), "SGVsbG8gd29ybGQ"},
		{"url unsafe chars", []byte("?&="), "PyY9"},
		{"binary data", []byte{0x00, 0xFF, 0x42}, "AP9C"},
		{"unicode", []byte("Hello 世界"), "SGVsbG8g5LiW55WM"},
		{"with padding 1", []byte("a"), "YQ"},
		{"with padding 2", []byte("ab"), "YWI"},
		{"no padding", []byte("abc"), "YWJj"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EncodeBase64URL(tt.arg); got != tt.want {
				t.Errorf("EncodeBase64URL() = %v, want %v", got, tt.want)
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
		{"basic string", "SGVsbG8gd29ybGQ", []byte("Hello world"), false},
		{"url unsafe chars", "PyY9", []byte("?&="), false},
		{"binary data", "AP9C", []byte{0x00, 0xFF, 0x42}, false},
		{"unicode", "SGVsbG8g5LiW55WM", []byte("Hello 世界"), false},
		{"with padding 1", "YQ", []byte("a"), false},
		{"with padding 2", "YWI", []byte("ab"), false},
		{"no padding", "YWJj", []byte("abc"), false},
		{"invalid base64", "not base64", nil, true},
		{"invalid padding", "YQ=", nil, true},
		{"incomplete input", "YWJjZ", nil, true},
		{"whitespace", "SGVs bG8=", nil, true},
		{"standard base64 chars", "+/+/", nil, true}, // Should fail with URL-safe base64
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
