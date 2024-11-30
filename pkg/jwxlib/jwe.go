// Package jwxlib provides JWT (JSON Web Token) and JWE (JSON Web Encryption) functionality
// using the lestrrat-go/jwx/v2 library. The JWE implementation supports RSA and ECDSA encryption
// for secure data exchange between parties.
//
// Key Features:
//   - RSA-OAEP-256 encryption and decryption
//   - ECDH-ES with A256KW key wrapping
//   - AES-256-GCM content encryption
//   - Support for multiple recipients (multi-key encryption)
//   - JSON and Compact serialization formats
//   - Strong security with modern cryptographic algorithms
//
// Example Usage:
//
//	// Generate RSA key pair
//	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
//	publicKey := &privateKey.PublicKey
//
//	// Encrypt data
//	plaintext := []byte("sensitive data")
//	ciphertext, err := jwxlib.JweEncrypt(plaintext, []interface{}{publicKey})
//
//	// Decrypt data
//	decrypted, err := jwxlib.JweDecrypt(ciphertext, []interface{}{privateKey})
//
// Security Considerations:
//   - Use RSA keys with at least 2048 bits
//   - Store private keys securely
//   - Rotate keys periodically
//   - Validate key permissions and usage
package jwxlib

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"fmt"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwe"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

// JweEncrypt encrypts plaintext for multiple keys, allowing
// multiple recipients to decrypt the same content with their respective private keys.
// This is useful in scenarios where the same data needs to be shared with multiple
// parties while maintaining end-to-end encryption.
//
// The function supports:
//   - RSA keys: Uses RSA-OAEP-256 for key encryption
//   - ECDSA keys: Uses ECDH-ES with A256KW key wrapping
//   - Content encryption: AES-256-GCM
//
// If multiple keys are provided, the output will be in JSON format, which is more
// suitable for multiple recipients than the compact format.
//
// Parameters:
//   - plaintext: The data to encrypt
//   - keys: Array of public keys for the intended recipients (RSA or ECDSA)
//
// Returns:
//   - []byte: The encrypted data in JWE JSON or compact format
//   - error: Any error that occurred during encryption
//
// Example:
//
//	pubKeys := []interface{}{recipient1.PublicKey, recipient2.PublicKey}
//	encrypted, err := JweEncrypt([]byte("secret"), pubKeys)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Security Considerations:
//   - Ensure all public keys are from trusted sources
//   - Verify key lengths are sufficient (2048 bits minimum for RSA)
//   - Consider the performance impact with large numbers of recipients
func JweEncrypt(plaintext []byte, keys []interface{}) (ciphertext []byte, err error) {
	options := []jwe.EncryptOption{
		jwe.WithContentEncryption(jwa.A256GCM),
	}

	if len(keys) > 1 {
		options = append(options, jwe.WithJSON())
	}

	for _, key := range keys {
		var k jwk.Key
		var e error
		var algo jwa.KeyEncryptionAlgorithm
		switch keyT := key.(type) {
		case *rsa.PublicKey:
			k, e = jwk.FromRaw(keyT)
			algo = jwa.RSA_OAEP_256
		case *ecdsa.PublicKey:
			k, e = jwk.FromRaw(keyT)
			algo = jwa.ECDH_ES_A256KW
		default:
			e = fmt.Errorf("unsupported public key type: %T", keyT)
		}
		if e != nil {
			return nil, fmt.Errorf("failed to create jwk: %s", e)
		}
		options = append(options, jwe.WithKey(algo, k))
	}

	ciphertext, err = jwe.Encrypt(plaintext, options...)
	if err != nil {
		err = fmt.Errorf("failed to encrypt: %s", err)
		return
	}

	return
}

// JweDecrypt decrypts JWE content using the given keys. This function can
// decrypt content in both JSON and compact formats, automatically detecting the
// correct format.
//
// The function supports:
//   - RSA keys: Uses RSA-OAEP-256 for key decryption
//   - ECDSA keys: Uses ECDH-ES with A256KW key unwrapping
//   - Content encryption: AES-256-GCM
//
// Parameters:
//   - ciphertext: The encrypted JWE data (in JSON or compact format)
//   - keys: Array of private keys to attempt decryption with (RSA or ECDSA)
//
// Returns:
//   - []byte: The decrypted plaintext
//   - error: Any error that occurred during decryption
//
// Example:
//
//	privKeys := []interface{}{recipient1.PrivateKey, recipient2.PrivateKey}
//	decrypted, err := JweDecrypt(ciphertext, privKeys)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Security Considerations:
//   - Validate the source of encrypted content
//   - Consider key rotation policies
//   - Handle decryption errors securely without leaking information
func JweDecrypt(ciphertext []byte, keys []interface{}) (plaintext []byte, err error) {
	options := []jwe.DecryptOption{}

	for _, key := range keys {
		var k jwk.Key
		var e error
		var algo jwa.KeyEncryptionAlgorithm
		switch keyT := key.(type) {
		case *rsa.PrivateKey:
			k, e = jwk.FromRaw(keyT)
			algo = jwa.RSA_OAEP_256
		case *ecdsa.PrivateKey:
			k, e = jwk.FromRaw(keyT)
			algo = jwa.ECDH_ES_A256KW
		default:
			e = fmt.Errorf("unsupported private key type: %T", keyT)
		}
		if e != nil {
			return nil, fmt.Errorf("failed to create jwk: %s", e)
		}
		options = append(options, jwe.WithKey(algo, k))
	}

	plaintext, err = jwe.Decrypt(ciphertext, options...)
	if err != nil {
		err = fmt.Errorf("failed to decrypt: %s", err)
		return
	}

	return
}
