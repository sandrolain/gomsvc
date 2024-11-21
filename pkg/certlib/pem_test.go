package certlib

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestCertificate(t *testing.T) (*x509.Certificate, *rsa.PrivateKey) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "Test Cert",
			Organization: []string{"Test Org"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(certBytes)
	require.NoError(t, err)

	return cert, key
}

func TestEncodeCertificateToPEM(t *testing.T) {
	cert, _ := createTestCertificate(t)

	pemBytes, err := EncodeCertificateToPEM(cert)
	require.NoError(t, err)
	assert.NotEmpty(t, pemBytes)

	// Test decoding the PEM back to a certificate
	decodedCert, err := ParseCertificateFromPEM(pemBytes)
	require.NoError(t, err)
	assert.Equal(t, cert.SerialNumber, decodedCert.SerialNumber)
}

func TestEncodePrivateKeyToPEM(t *testing.T) {
	_, key := createTestCertificate(t)

	pemBytes, err := EncodePrivateKeyToPEM(key)
	require.NoError(t, err)
	assert.NotEmpty(t, pemBytes)

	// Test decoding the PEM back to a private key
	decodedKey, err := ParsePrivateKeyFromPEM(pemBytes)
	require.NoError(t, err)
	assert.Equal(t, key.D, decodedKey.D)
}

func TestEncodePublicKeyToPEM(t *testing.T) {
	_, key := createTestCertificate(t)

	pemBytes, err := EncodePublicKeyToPEM(&key.PublicKey)
	require.NoError(t, err)
	assert.NotEmpty(t, pemBytes)

	// Verify the PEM format
	block, _ := pem.Decode(pemBytes)
	require.NotNil(t, block)
	assert.Equal(t, pubPemType, block.Type)
}

func TestEncodeRSAKeysToPEM(t *testing.T) {
	_, key := createTestCertificate(t)

	t.Run("RSA Private Key", func(t *testing.T) {
		pemBytes, err := EncodeRSAPrivateKeyToPEM(key)
		require.NoError(t, err)
		assert.NotEmpty(t, pemBytes)

		// Test decoding
		decodedKey, err := ParsePrivateKeyFromPEM(pemBytes)
		require.NoError(t, err)
		assert.Equal(t, key.D, decodedKey.D)
	})

	t.Run("RSA Public Key", func(t *testing.T) {
		pemBytes, err := EncodeRSAPublicKeyToPEM(&key.PublicKey)
		require.NoError(t, err)
		assert.NotEmpty(t, pemBytes)

		// Verify the PEM format
		block, _ := pem.Decode(pemBytes)
		require.NotNil(t, block)
		assert.Equal(t, pubRsaPemType, block.Type)
	})
}

func TestParsePrivateKeyFromPEM(t *testing.T) {
	_, key := createTestCertificate(t)

	t.Run("PKCS8 Private Key", func(t *testing.T) {
		pemBytes, err := EncodePrivateKeyToPEM(key)
		require.NoError(t, err)

		decodedKey, err := ParsePrivateKeyFromPEM(pemBytes)
		require.NoError(t, err)
		assert.Equal(t, key.D, decodedKey.D)
	})

	t.Run("RSA Private Key", func(t *testing.T) {
		pemBytes, err := EncodeRSAPrivateKeyToPEM(key)
		require.NoError(t, err)

		decodedKey, err := ParsePrivateKeyFromPEM(pemBytes)
		require.NoError(t, err)
		assert.Equal(t, key.D, decodedKey.D)
	})

	t.Run("Invalid PEM", func(t *testing.T) {
		_, err := ParsePrivateKeyFromPEM([]byte("invalid pem"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse PEM block")
	})

	t.Run("Invalid Key Type", func(t *testing.T) {
		pemBytes := []byte(`-----BEGIN UNKNOWN KEY-----
MIIEowIBAAKCAQEA0u1dgrWYj8p2UUvkDDLtNwNDB9eRAAAAAAAAAAAAAAA=
-----END UNKNOWN KEY-----`)
		_, err := ParsePrivateKeyFromPEM(pemBytes)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid key type")
	})
}

func TestParseFromFile(t *testing.T) {
	cert, key := createTestCertificate(t)
	tempDir := t.TempDir()

	t.Run("Certificate File", func(t *testing.T) {
		certPath := filepath.Join(tempDir, "cert.pem")
		certPEM, err := EncodeCertificateToPEM(cert)
		require.NoError(t, err)
		err = os.WriteFile(certPath, certPEM, 0600)
		require.NoError(t, err)

		parsedCert, err := ParseCertificateFromFile(certPath)
		require.NoError(t, err)
		assert.Equal(t, cert.SerialNumber, parsedCert.SerialNumber)

		// Test non-existent file
		_, err = ParseCertificateFromFile(filepath.Join(tempDir, "nonexistent.pem"))
		assert.Error(t, err)
	})

	t.Run("Private Key File", func(t *testing.T) {
		keyPath := filepath.Join(tempDir, "key.pem")
		keyPEM, err := EncodePrivateKeyToPEM(key)
		require.NoError(t, err)
		err = os.WriteFile(keyPath, keyPEM, 0600)
		require.NoError(t, err)

		parsedKey, err := ParsePrivateKeyFromFile(keyPath)
		require.NoError(t, err)
		assert.Equal(t, key.D, parsedKey.D)

		// Test non-existent file
		_, err = ParsePrivateKeyFromFile(filepath.Join(tempDir, "nonexistent.pem"))
		assert.Error(t, err)
	})
}
