package certlib

import (
	"crypto/tls"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateServerTLSConfig(t *testing.T) {
	// Test cases
	tests := []struct {
		name    string
		args    ServerTLSConfigBytes
		wantErr bool
	}{
		{
			name: "Valid certificates",
			args: ServerTLSConfigBytes{
				Cert: []byte(`-----BEGIN CERTIFICATE-----
MIIBhTCCASugAwIBAgIQIRi6zePL6mKjOipn+dNuaTAKBggqhkjOPQQDAjASMRAw
DgYDVQQKEwdBY21lIENvMB4XDTE3MTAyMDE5NDMwNloXDTE4MTAyMDE5NDMwNlow
EjEQMA4GA1UEChMHQWNtZSBDbzBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABD0d
7VNhbWvZLWPuj/RtHFjvtJBEwOkhbN/BnnE8rnZR8+sbwnc/KhCk3FhnpHZnQz7B
5aETbbIgmuvewdjvSBSjYzBhMA4GA1UdDwEB/wQEAwICpDATBgNVHSUEDDAKBggr
BgEFBQcDATAPBgNVHRMBAf8EBTADAQH/MCkGA1UdEQQiMCCCDmxvY2FsaG9zdDo1
NDUzgg4xMjcuMC4wLjE6NTQ1MzAKBggqhkjOPQQDAgNIADBFAiEA2zpJEPQyz6/l
Wf86aX6PepsntZv2GYlA5UpabfT2EZICICpJ5h/iI+i341gBmLiAFQOyTDT+/wQc
6MF9+Yw1Yy0t
-----END CERTIFICATE-----`),
				Key: []byte(`-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIIrYSSNQFaA2Hwf1duRSxKtLYX5CB04fSeQ6tF1aY/PuoAoGCCqGSM49
AwEHoUQDQgAEPR3tU2Fta9ktY+6P9G0cWO+0kETA6SFs38GecTyudlHz6xvCdz8q
EKTcWGekdmdDPsHloRNtsiCa697B2O9IFA==
-----END EC PRIVATE KEY-----`),
				CA: []byte(`-----BEGIN CERTIFICATE-----
MIIBhTCCASugAwIBAgIQIRi6zePL6mKjOipn+dNuaTAKBggqhkjOPQQDAjASMRAw
DgYDVQQKEwdBY21lIENvMB4XDTE3MTAyMDE5NDMwNloXDTE4MTAyMDE5NDMwNlow
EjEQMA4GA1UEChMHQWNtZSBDbzBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABD0d
7VNhbWvZLWPuj/RtHFjvtJBEwOkhbN/BnnE8rnZR8+sbwnc/KhCk3FhnpHZnQz7B
5aETbbIgmuvewdjvSBSjYzBhMA4GA1UdDwEB/wQEAwICpDATBgNVHSUEDDAKBggr
BgEFBQcDATAPBgNVHRMBAf8EBTADAQH/MCkGA1UdEQQiMCCCDmxvY2FsaG9zdDo1
NDUzgg4xMjcuMC4wLjE6NTQ1MzAKBggqhkjOPQQDAgNIADBFAiEA2zpJEPQyz6/l
Wf86aX6PepsntZv2GYlA5UpabfT2EZICICpJ5h/iI+i341gBmLiAFQOyTDT+/wQc
6MF9+Yw1Yy0t
-----END CERTIFICATE-----`),
			},
			wantErr: false,
		},
		{
			name: "Invalid certificate",
			args: ServerTLSConfigBytes{
				Cert: []byte("invalid cert"),
				Key:  []byte("invalid key"),
				CA:   []byte("invalid CA"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateServerTLSConfig(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateServerTLSConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Error("CreateServerTLSConfig() returned nil config when no error was expected")
			}
			if !tt.wantErr {
				if got.ClientAuth != tls.RequireAndVerifyClientCert {
					t.Error("CreateServerTLSConfig() ClientAuth not set to RequireAndVerifyClientCert")
				}
				if got.MinVersion != tls.VersionTLS12 {
					t.Error("CreateServerTLSConfig() MinVersion not set to TLS1.2")
				}
				if len(got.Certificates) != 1 {
					t.Error("CreateServerTLSConfig() expected 1 certificate")
				}
			}
		})
	}
}

func TestLoadServerTLSCredentials(t *testing.T) {
	// Create temporary test files
	tmpDir := t.TempDir()

	validCert := []byte(`-----BEGIN CERTIFICATE-----
MIIBhTCCASugAwIBAgIQIRi6zePL6mKjOipn+dNuaTAKBggqhkjOPQQDAjASMRAw
DgYDVQQKEwdBY21lIENvMB4XDTE3MTAyMDE5NDMwNloXDTE4MTAyMDE5NDMwNlow
EjEQMA4GA1UEChMHQWNtZSBDbzBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABD0d
7VNhbWvZLWPuj/RtHFjvtJBEwOkhbN/BnnE8rnZR8+sbwnc/KhCk3FhnpHZnQz7B
5aETbbIgmuvewdjvSBSjYzBhMA4GA1UdDwEB/wQEAwICpDATBgNVHSUEDDAKBggr
BgEFBQcDATAPBgNVHRMBAf8EBTADAQH/MCkGA1UdEQQiMCCCDmxvY2FsaG9zdDo1
NDUzgg4xMjcuMC4wLjE6NTQ1MzAKBggqhkjOPQQDAgNIADBFAiEA2zpJEPQyz6/l
Wf86aX6PepsntZv2GYlA5UpabfT2EZICICpJ5h/iI+i341gBmLiAFQOyTDT+/wQc
6MF9+Yw1Yy0t
-----END CERTIFICATE-----`)
	validKey := []byte(`-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIIrYSSNQFaA2Hwf1duRSxKtLYX5CB04fSeQ6tF1aY/PuoAoGCCqGSM49
AwEHoUQDQgAEPR3tU2Fta9ktY+6P9G0cWO+0kETA6SFs38GecTyudlHz6xvCdz8q
EKTcWGekdmdDPsHloRNtsiCa697B2O9IFA==
-----END EC PRIVATE KEY-----`)

	certFile := filepath.Join(tmpDir, "cert.pem")
	keyFile := filepath.Join(tmpDir, "key.pem")
	caFile := filepath.Join(tmpDir, "ca.pem")

	if err := os.WriteFile(certFile, validCert, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyFile, validKey, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(caFile, validCert, 0600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		args    ServerTLSConfigFiles
		wantErr bool
	}{
		{
			name: "Valid files",
			args: ServerTLSConfigFiles{
				CertFile: certFile,
				KeyFile:  keyFile,
				CAFile:   caFile,
			},
			wantErr: false,
		},
		{
			name: "Non-existent files",
			args: ServerTLSConfigFiles{
				CertFile: "nonexistent.pem",
				KeyFile:  "nonexistent.pem",
				CAFile:   "nonexistent.pem",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadServerTLSConfig(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadServerTLSCredentials() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Error("LoadServerTLSCredentials() returned nil config when no error was expected")
			}
			if !tt.wantErr {
				if got.ClientAuth != tls.RequireAndVerifyClientCert {
					t.Error("LoadServerTLSCredentials() ClientAuth not set to RequireAndVerifyClientCert")
				}
				if got.MinVersion != tls.VersionTLS12 {
					t.Error("LoadServerTLSCredentials() MinVersion not set to TLS1.2")
				}
				if len(got.Certificates) != 1 {
					t.Error("LoadServerTLSCredentials() expected 1 certificate")
				}
			}
		})
	}
}
