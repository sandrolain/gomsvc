package certlib

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"time"

	"google.golang.org/grpc/credentials"
)

const (
	crtType = "CERTIFICATE"
	rsaType = "RSA PRIVATE KEY"
)

type Certificate struct {
	Cert    *x509.Certificate
	Key     *rsa.PrivateKey
	TlsCert *tls.Certificate
}

func GetCertificateFromFile(certPath string, keyPath string) (res Certificate, err error) {
	r, err := os.ReadFile(certPath)
	if err != nil {
		return
	}
	kr, err := os.ReadFile(keyPath)
	if err != nil {
		return
	}
	return GetCertificate(r, kr)
}

func GetCertificate(certBytes []byte, keyBytes []byte) (res Certificate, err error) {
	block, _ := pem.Decode(certBytes)
	if block.Type != crtType {
		err = errors.New("invalid certificate ype")
		return
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return
	}

	kblock, _ := pem.Decode(keyBytes)
	if kblock.Type != rsaType {
		err = errors.New("invalid key type")
		return
	}
	key, err := x509.ParsePKCS1PrivateKey(kblock.Bytes)
	if err != nil {
		return
	}

	tlsCert := tls.Certificate{}
	tlsCert.Certificate = append(tlsCert.Certificate, cert.Raw)
	tlsCert.PrivateKey = key

	res.Cert = cert
	res.Key = key
	res.TlsCert = &tlsCert

	return
}

func hashPublicKey(key *rsa.PublicKey) ([]byte, error) {
	b, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return nil, fmt.Errorf("Unable to hash key: %s", err)
	}

	h := sha1.New()
	h.Write(b)
	return h.Sum(nil), nil
}

type CertificateArgs struct {
	Subject        pkix.Name
	CA             *x509.Certificate
	CACert         *tls.Certificate
	Duration       time.Duration
	EmailAddresses []string
	DNSNames       []string
}

func GenerateCertificate(args CertificateArgs) (res Certificate, err error) {
	// set up our server certificate
	cert := &x509.Certificate{
		SerialNumber:   big.NewInt(time.Now().UnixMilli()),
		Subject:        args.Subject,
		NotBefore:      time.Now(),
		NotAfter:       time.Now().Add(args.Duration),
		SubjectKeyId:   []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:    []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:       x509.KeyUsageDigitalSignature,
		EmailAddresses: args.EmailAddresses,
		DNSNames:       args.DNSNames,
	}

	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return
	}

	keyID, err := hashPublicKey(&key.PublicKey)
	if err != nil {
		return
	}

	cert.SubjectKeyId = keyID
	cert.PublicKey = key.Public()

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, args.CA, &key.PublicKey, args.CACert.PrivateKey)
	if err != nil {
		return
	}

	tlsCert := tls.Certificate{}
	tlsCert.Certificate = append(tlsCert.Certificate, certBytes)
	tlsCert.PrivateKey = key

	res.Cert = cert
	res.Key = key
	res.TlsCert = &tlsCert

	return
}

type CAArgs struct {
	Subject        pkix.Name
	Duration       time.Duration
	EmailAddresses []string
	DNSNames       []string
}

func GenerateCA(args CAArgs) (res Certificate, err error) {
	// set up our CA certificate
	cert := &x509.Certificate{
		SerialNumber:          big.NewInt(time.Now().UnixMilli()),
		Subject:               args.Subject,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(args.Duration),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		EmailAddresses:        args.EmailAddresses,
		DNSNames:              args.DNSNames,
	}

	// create our private and public key
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return
	}

	// create the CA
	caBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &key.PublicKey, key)
	if err != nil {
		return
	}

	tlsCert := tls.Certificate{}
	tlsCert.Certificate = append(tlsCert.Certificate, caBytes)
	tlsCert.PrivateKey = key

	res.Cert = cert
	res.Key = key
	res.TlsCert = &tlsCert

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
