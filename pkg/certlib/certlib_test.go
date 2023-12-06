package certlib

import (
	"crypto/x509/pkix"
	"os"
	"testing"
)

func TestGenerateCA(t *testing.T) {
	caCert, err := GenerateCA(CAArgs{
		Subject: pkix.Name{
			CommonName:    "Test",
			Organization:  []string{"Test"},
			Country:       []string{"IT"},
			Province:      []string{"Rome"},
			Locality:      []string{"Rome"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
	})
	if err != nil {
		t.Fatalf(err.Error())
	}

	cert, err := GenerateCertificate(CertificateArgs{
		Subject: pkix.Name{
			CommonName:    "Test",
			Organization:  []string{"Test"},
			Country:       []string{"IT"},
			Province:      []string{"Rome"},
			Locality:      []string{"Rome"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		CACert: caCert.TlsCert,
		CA:     caCert.Cert,
	})

	if err != nil {
		t.Fatalf(err.Error())
	}

	certBytes, keyBytes, err := EncodePEM(cert.TlsCert)

	if err != nil {
		t.Fatalf(err.Error())
	}

	os.WriteFile("./cert.pem", certBytes, 0644)
	os.WriteFile("./key.pem", keyBytes, 0644)

}
