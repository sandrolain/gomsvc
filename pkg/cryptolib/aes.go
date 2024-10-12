package cryptolib

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

// EncryptAES encrypts the plaintext using the provided key
func EncryptAESGCM(plainText, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	cipherText := make([]byte, gcm.NonceSize()+len(plainText))
	nonce := cipherText[:gcm.NonceSize()]
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(cipherText[:gcm.NonceSize()], nonce, plainText, nil), nil
}

// DecryptAES decrypts the ciphertext using the provided key
func DecryptAESGCM(cipherText, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(cipherText) < aes.BlockSize {
		return nil, errors.New("ciphertext must be at least AES.BlockSize bytes long")
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(cipherText) < nonceSize {
		return nil, errors.New("ciphertext is too short")
	}

	nonce, cipherText := cipherText[:nonceSize], cipherText[nonceSize:]
	plainText, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return nil, err
	}

	return plainText, nil
}
