package devlib

import (
	"testing"
)

func TestFileLine(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"simple"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FileLine(); got == "" {
				t.Errorf("FileLine() = %v, want %v", got, "non-empty string")
			}
		})
	}
}

func TestFileLineError(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"simple", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := FileLineError("test"); (err != nil) != tt.wantErr {
				t.Errorf("FileLineError() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
