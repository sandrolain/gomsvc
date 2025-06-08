package certlib

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"net"
	"os"
	"testing"
	"time"
)

func TestGenerateCA(t *testing.T) {
	rootCA, err := generateBasicCA("Test Root CA", "Test Organization", "US", 365*24*time.Hour)
	if err != nil {
		t.Fatal(err.Error())
	}

	t.Run("Root CA to File", func(t *testing.T) {
		certBytes, err := EncodeCertificateToPEM(rootCA.Cert)
		if err != nil {
			t.Fatal(err.Error())
		}

		keyBytes, err := EncodePrivateKeyToPEM(rootCA.Key)
		if err != nil {
			t.Fatal(err.Error())
		}

		_ = os.WriteFile("./root_ca_cert.pem", certBytes, 0644)
		_ = os.WriteFile("./root_ca_key.pem", keyBytes, 0644)

		defer func() {
			err := os.Remove("./root_ca_cert.pem")
			if err != nil {
				t.Errorf("Failed to remove root_ca_cert.pem: %v", err)
			}
			err = os.Remove("./root_ca_key.pem")
			if err != nil {
				t.Errorf("Failed to remove root_ca_key.pem: %v", err)
			}
		}()
	})

	ca, err := generateBasicIntermediateCA("Test Intermediate CA", "Test Organization", "US", rootCA, 365*24*time.Hour)
	if err != nil {
		t.Fatal(err.Error())
	}

	defer func() {
		err := os.Remove("./int_ca_cert.pem")
		if err != nil {
			t.Errorf("Failed to remove int_ca_cert.pem: %v", err)
		}
		err = os.Remove("./int_ca_key.pem")
		if err != nil {
			t.Errorf("Failed to remove int_ca_key.pem: %v", err)
		}
		err = os.Remove("./client_cert.pem")
		if err != nil {
			t.Errorf("Failed to remove client_cert.pem: %v", err)
		}
		err = os.Remove("./client_key.pem")
		if err != nil {
			t.Errorf("Failed to remove client_key.pem: %v", err)
		}
		err = os.Remove("./server_cert.pem")
		if err != nil {
			t.Errorf("Failed to remove server_cert.pem: %v", err)
		}
		err = os.Remove("./server_key.pem")
		if err != nil {
			t.Errorf("Failed to remove server_key.pem: %v", err)
		}
	}()

	t.Run("Intermediate CA to File", func(t *testing.T) {
		certBytes, err := EncodeCertificateToPEM(ca.Cert)
		if err != nil {
			t.Fatal(err.Error())
		}

		keyBytes, err := EncodePrivateKeyToPEM(ca.Key)
		if err != nil {
			t.Fatal(err.Error())
		}

		_ = os.WriteFile("./int_ca_cert.pem", certBytes, 0644)
		_ = os.WriteFile("./int_ca_key.pem", keyBytes, 0644)
	})

	client, err := generateBasicClientCert("Test Client", ca, 365*24*time.Hour)
	if err != nil {
		t.Fatal(err.Error())
	}

	t.Run("Client Certificate to File", func(t *testing.T) {
		certBytes, err := EncodeCertificateToPEM(client.Cert)
		if err != nil {
			t.Fatal(err.Error())
		}

		keyBytes, err := EncodePrivateKeyToPEM(client.Key)
		if err != nil {
			t.Fatal(err.Error())
		}

		_ = os.WriteFile("./client_cert.pem", certBytes, 0644)
		_ = os.WriteFile("./client_key.pem", keyBytes, 0644)
	})

	server, err := generateBasicServerCert("Test Server", []string{"localhost"}, ca, 365*24*time.Hour)
	if err != nil {
		t.Fatal(err.Error())
	}

	t.Run("Server Certificate to File", func(t *testing.T) {
		certBytes, err := EncodeCertificateToPEM(server.Cert)
		if err != nil {
			t.Fatal(err.Error())
		}

		keyBytes, err := EncodePrivateKeyToPEM(server.Key)
		if err != nil {
			t.Fatal(err.Error())
		}

		_ = os.WriteFile("./server_cert.pem", certBytes, 0644)
		_ = os.WriteFile("./server_key.pem", keyBytes, 0644)
	})
}

func TestCertificateErrors(t *testing.T) {
	t.Run("Invalid Key Size", func(t *testing.T) {
		_, err := GenerateCertificate(CertificateTypeRootCA, CertificateArgs{
			Subject: pkix.Name{
				CommonName:   "Test Invalid Key Size",
				Organization: []string{"Test Org"},
				Country:      []string{"US"},
			},
			Duration: 24 * time.Hour,
			KeySize:  1024, // Too small
		})
		if err == nil {
			t.Error("Expected error for invalid key size, got nil")
		}
	})

	t.Run("Missing Duration", func(t *testing.T) {
		_, err := GenerateCertificate(CertificateTypeRootCA, CertificateArgs{
			Subject: pkix.Name{
				CommonName:   "Test Missing Duration",
				Organization: []string{"Test Org"},
				Country:      []string{"US"},
			},
		})
		if err == nil {
			t.Error("Expected error for missing duration, got nil")
		}
	})

	t.Run("Missing Issuer for Intermediate CA", func(t *testing.T) {
		_, err := GenerateCertificate(CertificateTypeIntermediateCA, CertificateArgs{
			Subject: pkix.Name{
				CommonName:   "Test Missing Issuer",
				Organization: []string{"Test Org"},
				Country:      []string{"US"},
			},
			Duration: 24 * time.Hour,
		})
		if err == nil {
			t.Error("Expected error for missing issuer, got nil")
		}
	})

	t.Run("Invalid Certificate Chain", func(t *testing.T) {
		// Create a root CA first
		rootCA, err := generateBasicCA("Test Root CA", "Test Org", "US", 24*time.Hour)
		if err != nil {
			t.Fatal(err)
		}

		// Create a client cert directly signed by root (should be signed by intermediate)
		clientCert, err := GenerateCertificate(CertificateTypeClient, CertificateArgs{
			Subject: pkix.Name{
				CommonName: "Test Client",
			},
			Issuer:   rootCA,
			Duration: 24 * time.Hour,
		})
		if err != nil {
			t.Fatal(err)
		}

		// Try to verify the client cert with proper chain validation
		err = VerifyCertificate(VerifyCertificateArgs{
			Type:          CertificateTypeClient,
			Cert:          clientCert.Cert,
			Roots:         []*x509.Certificate{rootCA.Cert},
			Intermediates: []*x509.Certificate{}, // Missing intermediate
		})
		if err == nil {
			t.Error("Expected error for invalid certificate chain, got nil")
		}
	})

	t.Run("Expired Certificate", func(t *testing.T) {
		// Create an already expired certificate
		rootCA, err := GenerateCertificate(CertificateTypeRootCA, CertificateArgs{
			Subject: pkix.Name{
				CommonName:   "Test Expired Root CA",
				Organization: []string{"Test Org"},
				Country:      []string{"US"},
			},
			NotBefore: time.Now().Add(-2 * time.Hour), // 2 hours ago
			Duration:  time.Hour,                      // 1 hour duration
		})
		if err != nil {
			t.Fatal(err)
		}

		err = VerifyCertificate(VerifyCertificateArgs{
			Type:  CertificateTypeRootCA,
			Cert:  rootCA.Cert,
			Roots: []*x509.Certificate{rootCA.Cert},
		})
		if err == nil {
			t.Error("Expected error for expired certificate, got nil")
		} else if err.Error() != "certificate has expired" {
			t.Errorf("Expected 'certificate has expired' error, got: %s", err)
		}
	})

	t.Run("Wrong Certificate Type Usage", func(t *testing.T) {
		rootCA, err := generateBasicCA("Test Root CA", "Test Org", "US", 24*time.Hour)
		if err != nil {
			t.Fatal(err)
		}

		// Create a server certificate
		serverCert, err := generateBasicServerCert("Test Server", []string{"localhost"}, rootCA, 24*time.Hour)
		if err != nil {
			t.Fatal(err)
		}

		// Try to verify it as a client certificate (wrong type)
		err = VerifyCertificate(VerifyCertificateArgs{
			Type:    CertificateTypeClient,
			Cert:    serverCert.Cert,
			DNSName: "localhost",
			Roots:   []*x509.Certificate{rootCA.Cert},
		})
		if err == nil {
			t.Error("Expected error for wrong certificate type usage, got nil")
		}
	})
}

func TestServerIdentityValidation(t *testing.T) {
	rootCA, err := generateBasicCA("Test CA", "Test Org", "US", 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Missing DNS and IP", func(t *testing.T) {
		_, err := GenerateCertificate(CertificateTypeServer, CertificateArgs{
			Subject: pkix.Name{
				CommonName: "Test Server",
			},
			Issuer:   rootCA,
			Duration: 24 * time.Hour,
		})
		if err == nil {
			t.Error("Expected error for missing DNS and IP, got nil")
		}
	})

	t.Run("Empty DNS Name", func(t *testing.T) {
		_, err := GenerateCertificate(CertificateTypeServer, CertificateArgs{
			Subject: pkix.Name{
				CommonName: "Test Server",
			},
			Issuer:   rootCA,
			Duration: 24 * time.Hour,
			DNSNames: []string{""},
		})
		if err == nil {
			t.Error("Expected error for empty DNS name, got nil")
		}
	})

	t.Run("Invalid IP Address", func(t *testing.T) {
		_, err := GenerateCertificate(CertificateTypeServer, CertificateArgs{
			Subject: pkix.Name{
				CommonName: "Test Server",
			},
			Issuer:      rootCA,
			Duration:    24 * time.Hour,
			IPAddresses: []net.IP{net.IPv4zero},
		})
		if err == nil {
			t.Error("Expected error for unspecified IP address, got nil")
		}
	})
}

func TestUtilityFunctions(t *testing.T) {
	// Test basic CA creation
	rootCA, err := generateBasicCA("Test CA", "Test Org", "US", 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// Test basic intermediate CA creation
	intCA, err := generateBasicIntermediateCA("Test Int CA", "Test Org", "US", rootCA, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// Test basic server cert creation
	serverCert, err := generateBasicServerCert("Test Server", []string{"localhost"}, intCA, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// Test basic client cert creation
	clientCert, err := generateBasicClientCert("Test Client", intCA, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// Test cert pool creation
	pool := createCertPool(rootCA.Cert, intCA.Cert)
	if pool == nil {
		t.Error("Expected non-nil cert pool")
	}

	// Test TLS config creation
	tlsConfig := createTLSConfig(serverCert, pool)
	if tlsConfig == nil {
		t.Fatal("Expected non-nil TLS config")
	}
	if len(tlsConfig.Certificates) != 1 {
		t.Error("Expected 1 certificate in TLS config")
	}
	if tlsConfig.RootCAs != pool {
		t.Error("Expected root CAs to match provided pool")
	}

	// Verify the complete chain works
	err = VerifyCertificate(VerifyCertificateArgs{
		Type:          CertificateTypeServer,
		Cert:          serverCert.Cert,
		DNSName:       "localhost",
		Intermediates: []*x509.Certificate{intCA.Cert},
		Roots:         []*x509.Certificate{rootCA.Cert},
	})
	if err != nil {
		t.Errorf("Failed to verify server certificate chain: %v", err)
	}

	err = VerifyCertificate(VerifyCertificateArgs{
		Type:          CertificateTypeClient,
		Cert:          clientCert.Cert,
		Intermediates: []*x509.Certificate{intCA.Cert},
		Roots:         []*x509.Certificate{rootCA.Cert},
	})
	if err != nil {
		t.Errorf("Failed to verify client certificate chain: %v", err)
	}
}
