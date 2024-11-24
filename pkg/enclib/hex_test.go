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
		{"basic string", "48656c6c6f", []byte("Hello"), false},
		{"uppercase hex", "48656C6C6F", []byte("Hello"), false},
		{"mixed case", "48656c6C6f", []byte("Hello"), false},
		{"binary data", "00ff42", []byte{0x00, 0xFF, 0x42}, false},
		{"unicode", "e4b896e7958c", []byte("世界"), false},
		{"special chars", "21402324255E262A2829", []byte("!@#$%^&*()"), false},
		{"odd length", "4", nil, true},
		{"invalid hex", "not hex", nil, true},
		{"invalid chars", "4x4x", nil, true},
		{"with spaces", "48 65 6c", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeHex(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("HexDecode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
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
		{"basic string", []byte("Hello"), "48656c6c6f"},
		{"binary data", []byte{0x00, 0xFF, 0x42}, "00ff42"},
		{"unicode", []byte("世界"), "e4b896e7958c"},
		{"special chars", []byte("!@#$%^&*()"), "21402324255e262a2829"},
		{"null byte", []byte{0x00}, "00"},
		{"all bytes", func() []byte {
			b := make([]byte, 256)
			for i := range b {
				b[i] = byte(i)
			}
			return b
		}(), func() string {
			var s string
			for i := 0; i < 256; i++ {
				s += string([]byte{hexDigits[i>>4], hexDigits[i&0x0f]})
			}
			return s
		}()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EncodeHex(tt.arg); got != tt.want {
				t.Errorf("HexEncode() = %v, want %v", got, tt.want)
			}
		})
	}
}

// hexDigits is used by the all bytes test case
var hexDigits = []byte("0123456789abcdef")
