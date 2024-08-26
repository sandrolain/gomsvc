package certlib

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"testing"
	"time"
)

func TestCertificateVerify(t *testing.T) {
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

	err = VerifyCertificate(VerifyCertificateArgs{
		Type: CertificateTypeRootCA,
		Cert: rootCA.Cert,
		Roots: []*x509.Certificate{
			rootCA.Cert,
		},
	})
	if err != nil {
		t.Fatal(err.Error())
	}

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

	err = VerifyCertificate(VerifyCertificateArgs{
		Type: CertificateTypeIntermediateCA,
		Cert: rootCA.Cert,
		Roots: []*x509.Certificate{
			rootCA.Cert,
		},
	})
	if err != nil {
		t.Fatal(err.Error())
	}

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

	err = VerifyCertificate(VerifyCertificateArgs{
		Type: CertificateTypeClient,
		Cert: client.Cert,
		Roots: []*x509.Certificate{
			rootCA.Cert,
		},
		Intermediates: []*x509.Certificate{
			ca.Cert,
		},
	})

	if err != nil {
		t.Fatal(err.Error())
	}
}
