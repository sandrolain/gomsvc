package enclib

import (
	"testing"
)

func TestEncodeQuotedString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "basic string",
			input: "hello world",
			want:  `"hello world"`,
		},
		{
			name:  "string with quotes",
			input: `hello "world"`,
			want:  `"hello \"world\""`,
		},
		{
			name:  "string with special chars",
			input: "hello\nworld\t!",
			want:  `"hello\nworld\t!"`,
		},
		{
			name:  "empty string",
			input: "",
			want:  `""`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EncodeQuotedString(tt.input)
			if got != tt.want {
				t.Errorf("EncodeQuotedString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecodeQuotedString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "basic quoted string",
			input:   `"hello world"`,
			want:    "hello world",
			wantErr: false,
		},
		{
			name:    "string with escaped quotes",
			input:   `"hello \"world\""`,
			want:    `hello "world"`,
			wantErr: false,
		},
		{
			name:    "string with escaped chars",
			input:   `"hello\nworld\t!"`,
			want:    "hello\nworld\t!",
			wantErr: false,
		},
		{
			name:    "empty quoted string",
			input:   `""`,
			want:    "",
			wantErr: false,
		},
		{
			name:    "unmatched quotes",
			input:   `"hello`,
			wantErr: true,
		},
		{
			name:    "not quoted",
			input:   `hello`,
			wantErr: true,
		},
		{
			name:    "invalid escape",
			input:   `"hello\x"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeQuotedString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeQuotedString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("DecodeQuotedString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEncodeNonPrintable(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "printable only",
			input: "Hello, World!",
			want:  "Hello, World!",
		},
		{
			name:  "with newline",
			input: "Hello\nWorld",
			want:  `Hello\nWorld`,
		},
		{
			name:  "with tab",
			input: "Hello\tWorld",
			want:  `Hello\tWorld`,
		},
		{
			name:  "with carriage return",
			input: "Hello\rWorld",
			want:  `Hello\rWorld`,
		},
		{
			name:  "with null byte",
			input: "Hello\x00World",
			want:  `Hello\x00World`,
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EncodeNonPrintable(tt.input)
			if got != tt.want {
				t.Errorf("EncodeNonPrintable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSanitizeControlChars(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no control chars",
			input: "Hello, World!",
			want:  "Hello, World!",
		},
		{
			name:  "with newline",
			input: "Hello\nWorld",
			want:  "Hello World",
		},
		{
			name:  "with tab",
			input: "Hello\tWorld",
			want:  "Hello World",
		},
		{
			name:  "with multiple controls",
			input: "Hello\n\t\rWorld",
			want:  "Hello World",
		},
		{
			name:  "with null byte",
			input: "Hello\x00World",
			want:  "HelloWorld",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeControlChars(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeControlChars() = %v, want %v", got, tt.want)
			}
		})
	}
}
