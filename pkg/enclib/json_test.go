package enclib

import (
	"testing"
)

type testStruct struct {
	String  string   `json:"string"`
	Int     int      `json:"int"`
	Float   float64  `json:"float"`
	Bool    bool     `json:"bool"`
	Slice   []string `json:"slice"`
	Pointer *string  `json:"pointer,omitempty"`
}

func TestEncodeJSON(t *testing.T) {
	str := "pointer value"
	tests := []struct {
		name    string
		input   testStruct
		want    string
		wantErr bool
	}{
		{
			name: "basic struct",
			input: testStruct{
				String: "test",
				Int:    42,
				Float:  3.14,
				Bool:   true,
				Slice:  []string{"a", "b", "c"},
			},
			want:    `{"string":"test","int":42,"float":3.14,"bool":true,"slice":["a","b","c"]}`,
			wantErr: false,
		},
		{
			name: "with pointer",
			input: testStruct{
				String:  "test",
				Pointer: &str,
			},
			want:    `{"string":"test","int":0,"float":0,"bool":false,"slice":null,"pointer":"pointer value"}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EncodeJSON(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncodeJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("EncodeJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecodeJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    testStruct
		wantErr bool
	}{
		{
			name:  "basic struct",
			input: `{"string":"test","int":42,"float":3.14,"bool":true,"slice":["a","b","c"]}`,
			want: testStruct{
				String: "test",
				Int:    42,
				Float:  3.14,
				Bool:   true,
				Slice:  []string{"a", "b", "c"},
			},
			wantErr: false,
		},
		{
			name:    "invalid json",
			input:   `{"string":}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeJSON[testStruct](tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !compareTestStruct(got, tt.want) {
				t.Errorf("DecodeJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEncodeJSONPretty(t *testing.T) {
	tests := []struct {
		name    string
		input   testStruct
		want    string
		wantErr bool
	}{
		{
			name: "basic struct",
			input: testStruct{
				String: "test",
				Int:    42,
			},
			want: `{
  "string": "test",
  "int": 42,
  "float": 0,
  "bool": false,
  "slice": null
}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EncodeJSONPretty(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncodeJSONPretty() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("EncodeJSONPretty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompactJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name: "basic json",
			input: `{
				"string": "test",
				"int": 42
			}`,
			want:    `{"string":"test","int":42}`,
			wantErr: false,
		},
		{
			name:    "invalid json",
			input:   `{"string":}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CompactJSON(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("CompactJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("CompactJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid json",
			input:   `{"string":"test","int":42}`,
			wantErr: false,
		},
		{
			name:    "invalid json",
			input:   `{"string":}`,
			wantErr: true,
		},
		{
			name:    "empty json object",
			input:   `{}`,
			wantErr: false,
		},
		{
			name:    "empty json array",
			input:   `[]`,
			wantErr: false,
		},
		{
			name:    "simple value",
			input:   `"test"`,
			wantErr: false,
		},
		{
			name:    "number value",
			input:   `42`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateJSON(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Helper function to compare test structs
func compareTestStruct(a, b testStruct) bool {
	if a.String != b.String ||
		a.Int != b.Int ||
		a.Float != b.Float ||
		a.Bool != b.Bool {
		return false
	}

	if len(a.Slice) != len(b.Slice) {
		return false
	}
	for i := range a.Slice {
		if a.Slice[i] != b.Slice[i] {
			return false
		}
	}

	if (a.Pointer == nil) != (b.Pointer == nil) {
		return false
	}
	if a.Pointer != nil && *a.Pointer != *b.Pointer {
		return false
	}

	return true
}
