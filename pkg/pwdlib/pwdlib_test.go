package pwdlib

import (
	"testing"
)

func TestValidatePasswordEntropy(t *testing.T) {
	tests := []struct {
		name    string
		pwd     string
		wantErr bool
	}{
		{
			name:    "empty",
			pwd:     "",
			wantErr: true,
		},
		{
			name:    "short",
			pwd:     "short",
			wantErr: true,
		},
		{
			name:    "long",
			pwd:     "a1b2c3d4e5f6",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidatePasswordEntropy(tt.pwd); (err != nil) != tt.wantErr {
				t.Errorf("ValidatePasswordEntropy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
func TestGeneratePassword(t *testing.T) {
	tests := []struct {
		name    string
		length  int
		wantLen int
	}{
		{
			name:    "short",
			length:  8,
			wantLen: 8,
		},
		{
			name:    "long",
			length:  16,
			wantLen: 16,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pwd, err := GeneratePassword(tt.length)
			if err != nil {
				t.Errorf("GeneratePassword() error = %v", err)
			}
			if len(pwd) != tt.wantLen {
				t.Errorf("GeneratePassword() = %v, want length %v", len(pwd), tt.wantLen)
			}
		})
	}
}

func TestGetPasswordInURI(t *testing.T) {
	tests := []struct {
		name    string
		uri     string
		wantPwd string
		wantErr bool
	}{
		{
			name:    "no password",
			uri:     "postgresql://user@localhost:5432/database",
			wantPwd: "",
			wantErr: true,
		},
		{
			name:    "password",
			uri:     "postgresql://user:password@localhost:5432/database",
			wantPwd: "password",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPwd, err := GetPasswordInURI(tt.uri)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPasswordInURI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotPwd != tt.wantPwd {
				t.Errorf("GetPasswordInURI() = %v, want %v", gotPwd, tt.wantPwd)
			}
		})
	}
}
