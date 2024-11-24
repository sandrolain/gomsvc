package enclib

import (
	"testing"
)

func TestEncodeURL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "basic url",
			input: "https://example.com",
			want:  "https://example.com",
		},
		{
			name:  "url with spaces",
			input: "https://example.com/path with spaces",
			want:  "https://example.com/path%20with%20spaces",
		},
		{
			name:  "url with special chars",
			input: "https://example.com/path?q=hello&name=world",
			want:  "https://example.com/path?q=hello&name=world",
		},
		{
			name:  "url with unicode",
			input: "https://example.com/世界",
			want:  "https://example.com/%E4%B8%96%E7%95%8C",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EncodeURL(tt.input)
			if err != nil {
				t.Fatalf("EncodeURL() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("EncodeURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecodeURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "basic url",
			input:   "https://example.com",
			want:    "https://example.com",
			wantErr: false,
		},
		{
			name:    "encoded spaces",
			input:   "https://example.com/path%20with%20spaces",
			want:    "https://example.com/path with spaces",
			wantErr: false,
		},
		{
			name:    "encoded unicode",
			input:   "https://example.com/%E4%B8%96%E7%95%8C",
			want:    "https://example.com/世界",
			wantErr: false,
		},
		{
			name:    "invalid encoding",
			input:   "https://example.com/%",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			want:    "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeURL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("DecodeURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEncodeURLComponent(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "basic string",
			input: "hello world",
			want:  "hello%20world",
		},
		{
			name:  "special chars",
			input: "hello!@#$%^&*()",
			want:  "hello%21%40%23%24%25%5E%26%2A%28%29",
		},
		{
			name:  "unicode",
			input: "世界",
			want:  "%E4%B8%96%E7%95%8C",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EncodeURLComponent(tt.input)
			if got != tt.want {
				t.Errorf("EncodeURLComponent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecodeURLComponent(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "basic encoded string",
			input:   "hello%20world",
			want:    "hello world",
			wantErr: false,
		},
		{
			name:    "encoded special chars",
			input:   "hello%21%40%23%24%25%5E%26%2A%28%29",
			want:    "hello!@#$%^&*()",
			wantErr: false,
		},
		{
			name:    "encoded unicode",
			input:   "%E4%B8%96%E7%95%8C",
			want:    "世界",
			wantErr: false,
		},
		{
			name:    "invalid encoding",
			input:   "%",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			want:    "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeURLComponent(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeURLComponent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("DecodeURLComponent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildQueryString(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]string
		want  string
	}{
		{
			name: "basic params",
			input: map[string]string{
				"name": "john",
				"age":  "30",
			},
			want: "age=30&name=john",
		},
		{
			name: "params with spaces",
			input: map[string]string{
				"name": "john doe",
				"job":  "software engineer",
			},
			want: "job=software%20engineer&name=john%20doe",
		},
		{
			name: "params with special chars",
			input: map[string]string{
				"q":    "hello!@#$",
				"lang": "en-US",
			},
			want: "lang=en-US&q=hello%21%40%23%24",
		},
		{
			name:  "empty map",
			input: map[string]string{},
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildQueryString(tt.input)
			if got != tt.want {
				t.Errorf("BuildQueryString() = %v, want %v", got, tt.want)
			}
		})
	}
}
