package cryptolib

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

// EncryptAESGCM encrypts the plaintext using the provided key with AES-GCM mode.
// It generates a secure random nonce for each encryption operation.
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

// DecryptAESGCM decrypts the ciphertext using the provided key with AES-GCM mode.
// The nonce is expected to be prepended to the ciphertext as performed by EncryptAESGCM.
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
