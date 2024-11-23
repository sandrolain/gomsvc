package datalib

import (
	"testing"
)

func TestSafeDirName(t *testing.T) {
	tests := []struct {
		name    string
		parts   []string
		want    string
		wantErr bool
	}{
		{
			name:    "simple path",
			parts:   []string{"test", "dir"},
			want:    "testdir",
			wantErr: false,
		},
		{
			name:    "path with spaces",
			parts:   []string{"test ", " dir"},
			want:    "test  dir",
			wantErr: false,
		},
		{
			name:    "path with special characters",
			parts:   []string{"test/", "/dir"},
			want:    "test_dir",
			wantErr: false,
		},
		{
			name:    "path with unicode",
			parts:   []string{"téşt", "dïr"},
			want:    "téştdïr",
			wantErr: false,
		},
		{
			name:    "empty parts",
			parts:   []string{"", ""},
			want:    "",
			wantErr: false,
		},
		{
			name:    "path with multiple special chars",
			parts:   []string{"test/\\*:?\"<>|", "dir"},
			want:    "test_dir",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SafeDirName(tt.parts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("SafeDirName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SafeDirName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSafeFileName(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		ext      []string
		want     string
		wantErr  bool
	}{
		{
			name:     "simple filename",
			filename: "test",
			ext:      []string{"txt"},
			want:     "test.txt",
			wantErr:  false,
		},
		{
			name:     "filename with spaces",
			filename: "test file",
			ext:      []string{"txt"},
			want:     "test file.txt",
			wantErr:  false,
		},
		{
			name:     "filename with special characters",
			filename: "test/file",
			ext:      []string{"txt"},
			want:     "test_file.txt",
			wantErr:  false,
		},
		{
			name:     "filename with unicode",
			filename: "téşt fïle",
			ext:      []string{"txt"},
			want:     "téşt fïle.txt",
			wantErr:  false,
		},
		{
			name:     "empty filename",
			filename: "",
			ext:      []string{"txt"},
			want:     ".txt",
			wantErr:  false,
		},
		{
			name:     "filename without extension",
			filename: "test",
			ext:      []string{},
			want:     "test",
			wantErr:  false,
		},
		{
			name:     "filename with multiple special chars",
			filename: "test/\\*:?\"<>|file",
			ext:      []string{"txt"},
			want:     "test_file.txt",
			wantErr:  false,
		},
		{
			name:     "filename with multiple dots",
			filename: "test.backup.old",
			ext:      []string{"txt"},
			want:     "test.backup.old.txt",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SafeFileName(tt.filename, tt.ext...)
			if (err != nil) != tt.wantErr {
				t.Errorf("SafeFileName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SafeFileName() = %v, want %v", got, tt.want)
			}
		})
	}
}
