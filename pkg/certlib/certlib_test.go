package certlib

import (
	"crypto/x509/pkix"
	"os"
	"testing"
)

func TestGenerateCA(t *testing.T) {
	caCert, caTlsCert, err := GenerateCA(pkix.Name{
		CommonName:    "Test",
		Organization:  []string{"Test"},
		Country:       []string{"IT"},
		Province:      []string{"Rome"},
		Locality:      []string{"Rome"},
		StreetAddress: []string{""},
		PostalCode:    []string{""},
	})
	if err != nil {
		t.Fatalf(err.Error())
	}

	_, tlsCert, err := GenerateCertificate(pkix.Name{
		CommonName:    "Test",
		Organization:  []string{"Test"},
		Country:       []string{"IT"},
		Province:      []string{"Rome"},
		Locality:      []string{"Rome"},
		StreetAddress: []string{""},
		PostalCode:    []string{""},
	}, caCert, caTlsCert)

	if err != nil {
		t.Fatalf(err.Error())
	}

	certBytes, keyBytes, err := EncodePEM(tlsCert)

	if err != nil {
		t.Fatalf(err.Error())
	}

	os.WriteFile("./cert.pem", certBytes, 0644)
	os.WriteFile("./key.pem", keyBytes, 0644)

}
