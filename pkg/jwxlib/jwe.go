// Package jwxlib provides JWT (JSON Web Token) and JWE (JSON Web Encryption) functionality
// using the lestrrat-go/jwx/v2 library. The JWE implementation supports RSA encryption
// for secure data exchange between parties.
//
// Key Features:
//   - RSA-OAEP encryption and decryption
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
//	ciphertext, err := jwxlib.JweEncryptRsa(plaintext, publicKey, false)
//
//	// Decrypt data
//	decrypted, err := jwxlib.JweDecryptRsa(ciphertext, privateKey)
//
// Security Considerations:
//   - Use RSA keys with at least 2048 bits
//   - Store private keys securely
//   - Rotate keys periodically
//   - Validate key permissions and usage
package jwxlib

import (
	"crypto/rsa"
	"fmt"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwe"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

// JweEncryptMultiRsa encrypts plaintext for multiple RSA public keys, allowing
// multiple recipients to decrypt the same content with their respective private keys.
// This is useful in scenarios where the same data needs to be shared with multiple
// parties while maintaining end-to-end encryption.
//
// The function uses RSA-OAEP for key encryption and outputs the result in JSON format,
// which is more suitable for multiple recipients than the compact format.
//
// Parameters:
//   - plaintext: The data to encrypt
//   - rsaPubKeys: Array of RSA public keys for the intended recipients
//
// Returns:
//   - []byte: The encrypted data in JWE JSON format
//   - error: Any error that occurred during encryption
//
// Example:
//
//	pubKeys := []*rsa.PublicKey{recipient1.PublicKey, recipient2.PublicKey}
//	encrypted, err := JweEncryptMultiRsa([]byte("secret"), pubKeys)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Security Considerations:
//   - Ensure all public keys are from trusted sources
//   - Verify key lengths are sufficient (2048 bits minimum recommended)
//   - Consider the performance impact with large numbers of recipients
func JweEncryptMultiRsa(plaintext []byte, rsaPubKeys []*rsa.PublicKey) (ciphertext []byte, err error) {
	pubkeys := make([]jwk.Key, len(rsaPubKeys))

	for i, key := range rsaPubKeys {
		pubkey, e := jwk.FromRaw(key)
		if e != nil {
			err = fmt.Errorf("failed to create jwk: %s", e)
			return
		}
		pubkeys[i] = pubkey
	}

	options := []jwe.EncryptOption{jwe.WithJSON()}
	for _, key := range pubkeys {
		options = append(options, jwe.WithKey(jwa.RSA_OAEP, key))
	}

	ciphertext, e := jwe.Encrypt(plaintext, options...)
	if e != nil {
		err = fmt.Errorf("failed to encrypt: %s", e)
		return
	}

	return
}

// JweEncryptRsa encrypts plaintext using a single RSA public key. The function supports
// both JSON and Compact serialization formats through the json parameter.
//
// The function uses RSA-OAEP for key encryption, which provides strong security
// and is recommended for new applications.
//
// Parameters:
//   - plaintext: The data to encrypt
//   - rsaPubKey: RSA public key of the intended recipient
//   - json: If true, outputs in JSON format; if false, uses compact format
//
// Returns:
//   - []byte: The encrypted data in either JWE JSON or compact format
//   - error: Any error that occurred during encryption
//
// Example:
//
//	// Encrypt using compact format
//	encrypted, err := JweEncryptRsa([]byte("secret"), publicKey, false)
//
//	// Encrypt using JSON format
//	encrypted, err := JweEncryptRsa([]byte("secret"), publicKey, true)
//
// Security Considerations:
//   - Use keys from trusted sources
//   - Ensure key length is sufficient (2048 bits minimum)
//   - JSON format is more verbose but supports additional metadata
//   - Compact format is smaller but supports only one recipient
func JweEncryptRsa(plaintext []byte, rsaPubKey *rsa.PublicKey, json bool) (ciphertext []byte, err error) {
	pubkey, e := jwk.FromRaw(rsaPubKey)
	if e != nil {
		err = fmt.Errorf("failed to create jwk: %s", e)
		return
	}

	options := []jwe.EncryptOption{jwe.WithKey(jwa.RSA_OAEP, pubkey)}

	if json {
		options = append(options, jwe.WithJSON())
	}

	ciphertext, e = jwe.Encrypt(plaintext, options...)
	if e != nil {
		err = fmt.Errorf("failed to encrypt: %s", e)
		return
	}

	return
}

// JweDecryptRsa decrypts JWE content using an RSA private key. This function can
// decrypt content in both JSON and compact formats, automatically detecting the
// correct format.
//
// The function supports content encrypted with RSA-OAEP and will fail if the
// content was encrypted using a different algorithm.
//
// Parameters:
//   - ciphertext: The encrypted data in either JWE JSON or compact format
//   - rsaPrivKey: RSA private key for decryption
//
// Returns:
//   - []byte: The decrypted plaintext
//   - error: Any error that occurred during decryption
//
// Example:
//
//	decrypted, err := JweDecryptRsa(encrypted, privateKey)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Decrypted: %s\n", string(decrypted))
//
// Security Considerations:
//   - Protect private keys from unauthorized access
//   - Validate the source of encrypted content
//   - Consider key rotation policies
//   - Handle decryption errors securely without leaking information
func JweDecryptRsa(ciphertext []byte, rsaPrivKey *rsa.PrivateKey) (plaintext []byte, err error) {
	privkey, e := jwk.FromRaw(rsaPrivKey)
	if e != nil {
		err = fmt.Errorf("failed to create jwk: %s", e)
		return
	}

	options := []jwe.DecryptOption{jwe.WithKey(jwa.RSA_OAEP, privkey)}

	plaintext, e = jwe.Decrypt(ciphertext, options...)
	if e != nil {
		err = fmt.Errorf("failed to decrypt: %s", e)
		return
	}

	return
}
