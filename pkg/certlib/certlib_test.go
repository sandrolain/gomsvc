package certlib

import (
	"crypto/x509/pkix"
	"os"
	"testing"
	"time"
)

func TestGenerateCA(t *testing.T) {
	rootCA, err := GenerateCertificate(CertificateTypeRootCA, CertificateArgs{
		Subject: pkix.Name{
			CommonName:    "Test",
			Organization:  []string{"Test"},
			Country:       []string{"IT"},
			Province:      []string{"Rome"},
			Locality:      []string{"Rome"},
			StreetAddress: []string{"test"},
			PostalCode:    []string{"12345"},
		},
		Duration: 365 * 24 * time.Hour,
	})
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

		os.WriteFile("./root_ca_cert.pem", certBytes, 0644)
		os.WriteFile("./root_ca_key.pem", keyBytes, 0644)
	})

	ca, err := GenerateCertificate(CertificateTypeIntermediateCA, CertificateArgs{
		Subject: pkix.Name{
			CommonName:    "Test",
			Organization:  []string{"Test"},
			Country:       []string{"IT"},
			Province:      []string{"Rome"},
			Locality:      []string{"Rome"},
			StreetAddress: []string{"test"},
			PostalCode:    []string{"12345"},
		},
		Issuer:   rootCA,
		Duration: 365 * 24 * time.Hour,
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	t.Run("Intermediate CA to File", func(t *testing.T) {
		certBytes, err := EncodeCertificateToPEM(ca.Cert)
		if err != nil {
			t.Fatal(err.Error())
		}

		keyBytes, err := EncodePrivateKeyToPEM(ca.Key)
		if err != nil {
			t.Fatal(err.Error())
		}

		os.WriteFile("./int_ca_cert.pem", certBytes, 0644)
		os.WriteFile("./int_ca_key.pem", keyBytes, 0644)
	})

	client, err := GenerateCertificate(CertificateTypeClient, CertificateArgs{
		Subject: pkix.Name{
			CommonName:    "Test",
			Organization:  []string{"Test"},
			Country:       []string{"IT"},
			Province:      []string{"Rome"},
			Locality:      []string{"Rome"},
			StreetAddress: []string{"test"},
			PostalCode:    []string{"12345"},
		},
		Issuer:   ca,
		Duration: 365 * 24 * time.Hour,
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	t.Run("Client to File", func(t *testing.T) {
		certBytes, err := EncodeCertificateToPEM(client.Cert)
		if err != nil {
			t.Fatal(err.Error())
		}

		keyBytes, err := EncodePrivateKeyToPEM(client.Key)
		if err != nil {
			t.Fatal(err.Error())
		}

		os.WriteFile("./client_cert.pem", certBytes, 0644)
		os.WriteFile("./client_key.pem", keyBytes, 0644)
	})

	server, err := GenerateCertificate(CertificateTypeServer, CertificateArgs{
		Subject: pkix.Name{
			CommonName:    "Test",
			Organization:  []string{"Test"},
			Country:       []string{"IT"},
			Province:      []string{"Rome"},
			Locality:      []string{"Rome"},
			StreetAddress: []string{"test"},
			PostalCode:    []string{"12345"},
		},
		Issuer:   ca,
		Duration: 365 * 24 * time.Hour,
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	t.Run("Server to File", func(t *testing.T) {
		certBytes, err := EncodeCertificateToPEM(server.Cert)
		if err != nil {
			t.Fatal(err.Error())
		}

		keyBytes, err := EncodePrivateKeyToPEM(server.Key)
		if err != nil {
			t.Fatal(err.Error())
		}

		os.WriteFile("./server_cert.pem", certBytes, 0644)
		os.WriteFile("./server_key.pem", keyBytes, 0644)
	})
}
