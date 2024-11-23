// Package cryptolib provides cryptographic utilities including AES encryption and decryption.
package cryptolib

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

// EncryptAESGCM encrypts data using AES in Galois/Counter Mode (GCM).
// GCM provides both confidentiality and authenticity (AEAD - Authenticated Encryption with Associated Data).
//
// The function automatically generates a random nonce (number used once) for each encryption,
// which is prepended to the ciphertext. This ensures that the same plaintext will encrypt
// to different ciphertexts, even when using the same key.
//
// Parameters:
//   - plainText: The data to encrypt
//   - key: Must be 16, 24, or 32 bytes for AES-128, AES-192, or AES-256 respectively
//
// Returns:
//   - []byte: The encrypted data with the nonce prepended
//   - error: Any error that occurred during encryption
//
// Security Considerations:
//   - The key must be kept secret and should be generated using GenerateAES256Key()
//   - Each encryption generates a unique nonce, making the encryption non-deterministic
//   - The output includes authentication data to detect tampering
//
// Example:
//
//	key, err := GenerateAES256Key() // Generate a secure 32-byte key
//	if err != nil {
//	    log.Fatal(err)
//	}
//	encrypted, err := EncryptAESGCM([]byte("secret message"), key)
//	if err != nil {
//	    log.Fatal(err)
//	}
func EncryptAESGCM(plainText, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Generate a random nonce for each encryption
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Seal will append the ciphertext to the nonce
	// The nonce is prepended to the ciphertext and will be used for decryption
	return gcm.Seal(nonce, nonce, plainText, nil), nil
}

// DecryptAESGCM decrypts data that was encrypted using EncryptAESGCM.
// It uses AES in Galois/Counter Mode (GCM) and automatically extracts the nonce
// from the beginning of the ciphertext.
//
// Parameters:
//   - cipherText: The encrypted data (including the prepended nonce)
//   - key: Must be the same key used for encryption (16, 24, or 32 bytes)
//
// Returns:
//   - []byte: The decrypted data
//   - error: Any error that occurred during decryption, including authentication failures
//
// Security Considerations:
//   - If the ciphertext has been tampered with, decryption will fail with an error
//   - The function verifies the authenticity of the data before decryption
//   - The key must match the one used for encryption exactly
//
// Example:
//
//	decrypted, err := DecryptAESGCM(encrypted, key)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(string(decrypted)) // "secret message"
//
// Common Errors:
//   - "ciphertext is too short": The input doesn't contain enough bytes for the nonce
//   - "invalid nonce size": The nonce extracted from the ciphertext is invalid
//   - "encrypted data is empty": No encrypted data after the nonce
func DecryptAESGCM(cipherText, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(cipherText) < nonceSize {
		return nil, errors.New("ciphertext is too short")
	}

	// Extract the nonce from the ciphertext
	nonce := cipherText[:nonceSize]
	if len(nonce) != nonceSize {
		return nil, errors.New("invalid nonce size")
	}

	// Extract the actual ciphertext after the nonce
	encryptedData := cipherText[nonceSize:]
	if len(encryptedData) < 1 {
		return nil, errors.New("encrypted data is empty")
	}

	// #nosec G407 -- nonce is not hardcoded, it is extracted from the ciphertext
	return gcm.Open(nil, nonce, encryptedData, nil)
}
