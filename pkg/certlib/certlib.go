package certlib

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"net"
	"os"
	"time"
)

const (
	crtType = "CERTIFICATE"
	rsaType = "RSA PRIVATE KEY"
)

func GetCertificateFromFile(path string) (cert *x509.Certificate, err error) {
	r, err := os.ReadFile(path)
	if err != nil {
		return
	}
	block, _ := pem.Decode(r)
	if block.Type != crtType {
		err = errors.New("invalid certificate ype")
		return
	}
	cert, err = x509.ParseCertificate(block.Bytes)
	return
}

func GetPrivateKeyFromFile(path string) (key *rsa.PrivateKey, err error) {
	r, err := os.ReadFile(path)
	if err != nil {
		return
	}
	block, _ := pem.Decode(r)
	if block.Type != rsaType {
		err = errors.New("invalid key type")
		return
	}
	key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	return
}

func GenerateCertificate(subject pkix.Name, ca *x509.Certificate, caCert *tls.Certificate) (cert *x509.Certificate, tlsCert *tls.Certificate, err error) {
	// set up our server certificate
	cert = &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject:      subject,
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0), // 10 years
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	privKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &privKey.PublicKey, caCert.PrivateKey)
	if err != nil {
		return
	}

	res := tls.Certificate{}
	res.Certificate = append(res.Certificate, certBytes)
	res.PrivateKey = privKey

	tlsCert = &res

	return
}

func GenerateCA(subject pkix.Name) (cert *x509.Certificate, tlsCert *tls.Certificate, err error) {
	// set up our CA certificate
	cert = &x509.Certificate{
		SerialNumber:          big.NewInt(time.Now().Unix()),
		Subject:               subject,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// create our private and public key
	caKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return
	}

	// create the CA
	caBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &caKey.PublicKey, caKey)
	if err != nil {
		return
	}

	res := tls.Certificate{}
	res.Certificate = append(res.Certificate, caBytes)
	res.PrivateKey = caKey

	tlsCert = &res
	return
}

func EncodePEM(tlsCert *tls.Certificate) (certPEMBytes []byte, keyPEMBytes []byte, err error) {
	caPEM := new(bytes.Buffer)
	err = pem.Encode(caPEM, &pem.Block{
		Type:  crtType,
		Bytes: tlsCert.Certificate[0],
	})
	if err != nil {
		return
	}

	caPrivKeyPEM := new(bytes.Buffer)
	err = pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  rsaType,
		Bytes: x509.MarshalPKCS1PrivateKey(tlsCert.PrivateKey.(*rsa.PrivateKey)),
	})
	if err != nil {
		return
	}

	certPEMBytes = caPEM.Bytes()
	keyPEMBytes = caPrivKeyPEM.Bytes()
	return
}

func CreateServerTLSConfig(certBytes []byte, keyBytes []byte) (cfg *tls.Config, err error) {
	serverCert, err := tls.X509KeyPair(certBytes, keyBytes)
	if err != nil {
		return
	}

	cfg = &tls.Config{
		Certificates: []tls.Certificate{serverCert},
	}
	return
}

func CreateClientTLSConfig(caCertBytes []byte) (cfg *tls.Config) {
	certpool := x509.NewCertPool()
	certpool.AppendCertsFromPEM(caCertBytes)
	cfg = &tls.Config{
		RootCAs: certpool,
	}
	return
}
