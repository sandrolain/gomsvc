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

	"google.golang.org/grpc/credentials"
)

const (
	crtType = "CERTIFICATE"
	rsaType = "RSA PRIVATE KEY"
)

func GetCertificateFromFile(certPath string, keyPath string) (cert *x509.Certificate, key *rsa.PrivateKey, tlsCert *tls.Certificate, err error) {
	r, err := os.ReadFile(certPath)
	if err != nil {
		return
	}
	block, _ := pem.Decode(r)
	if block.Type != crtType {
		err = errors.New("invalid certificate ype")
		return
	}
	cert, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		return
	}

	kr, err := os.ReadFile(keyPath)
	if err != nil {
		return
	}
	kblock, _ := pem.Decode(kr)
	if kblock.Type != rsaType {
		err = errors.New("invalid key type")
		return
	}
	key, err = x509.ParsePKCS1PrivateKey(kblock.Bytes)
	if err != nil {
		return
	}

	res := tls.Certificate{}
	res.Certificate = append(res.Certificate, cert.Raw)
	res.PrivateKey = key

	tlsCert = &res

	return
}

func GenerateCertificate(subject pkix.Name, ca *x509.Certificate, caCert *tls.Certificate) (cert *x509.Certificate, key *rsa.PrivateKey, tlsCert *tls.Certificate, err error) {
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

	key, err = rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &key.PublicKey, caCert.PrivateKey)
	if err != nil {
		return
	}

	res := tls.Certificate{}
	res.Certificate = append(res.Certificate, certBytes)
	res.PrivateKey = key

	tlsCert = &res

	return
}

func GenerateCA(subject pkix.Name) (cert *x509.Certificate, key *rsa.PrivateKey, tlsCert *tls.Certificate, err error) {
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
	key, err = rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return
	}

	// create the CA
	caBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &key.PublicKey, key)
	if err != nil {
		return
	}

	res := tls.Certificate{}
	res.Certificate = append(res.Certificate, caBytes)
	res.PrivateKey = key

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

type ServerTLSConfigArgs[T any] struct {
	Cert T
	Key  T
	CA   T
}

func CreateServerTLSConfig(args ServerTLSConfigArgs[[]byte]) (cred credentials.TransportCredentials, err error) {
	serverCert, err := tls.X509KeyPair(args.Cert, args.Key)
	if err != nil {
		return
	}
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(args.CA) {
		err = errors.New("failed to add client CA's certificate")
		return
	}
	cfg := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}
	cred = credentials.NewTLS(cfg)
	return
}

func LoadServerTLSCredentials(args ServerTLSConfigArgs[string]) (cred credentials.TransportCredentials, err error) {
	serverCert, err := tls.LoadX509KeyPair(args.Cert, args.Key)
	if err != nil {
		return
	}
	pemClientCA, err := os.ReadFile(args.CA)
	if err != nil {
		return nil, err
	}
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemClientCA) {
		err = errors.New("failed to add client CA's certificate")
		return
	}
	cred = credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	})
	return
}

type ClientTLSConfigArgs[T any] struct {
	Cert T
	Key  T
	CA   T
}

func CreateClientTLSCredentials(args ClientTLSConfigArgs[[]byte]) (cred credentials.TransportCredentials, err error) {
	clientCert, err := tls.X509KeyPair(args.Cert, args.Key)
	if err != nil {
		return
	}
	certpool := x509.NewCertPool()
	if !certpool.AppendCertsFromPEM(args.CA) {
		err = errors.New("failed to add client CA's certificate")
		return
	}
	cred = credentials.NewTLS(&tls.Config{
		Certificates:       []tls.Certificate{clientCert},
		RootCAs:            certpool,
		InsecureSkipVerify: true,
	})
	return
}

func LoadClientTLSCredentials(args ClientTLSConfigArgs[string]) (cred credentials.TransportCredentials, err error) {
	caCertBytes, err := os.ReadFile(args.CA)
	if err != nil {
		return
	}
	clientCert, err := tls.LoadX509KeyPair(args.Cert, args.Key)
	if err != nil {
		return
	}
	certpool := x509.NewCertPool()
	if !certpool.AppendCertsFromPEM(caCertBytes) {
		err = errors.New("failed to add client CA's certificate")
		return
	}
	cred = credentials.NewTLS(&tls.Config{
		Certificates:       []tls.Certificate{clientCert},
		RootCAs:            certpool,
		InsecureSkipVerify: true,
	})
	return
}
