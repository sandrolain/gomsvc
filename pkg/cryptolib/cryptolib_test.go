package cryptolib

import (
	"bytes"
	"crypto/aes"
	"crypto/rand"
	"testing"
)

func TestEncryptAES(t *testing.T) {
	keyLengths := []int{16, 24, 32} // AES-128, AES-192, AES-256

	for _, keyLength := range keyLengths {
		key := make([]byte, keyLength)
		if _, err := rand.Read(key); err != nil {
			t.Fatalf("Failed to generate key: %v", err)
		}

		plainText := []byte("Test plaintext")
		cipherText, err := EncryptAESGCM(plainText, key)
		if err != nil {
			t.Fatalf("EncryptAES failed for key length %d: %v", keyLength, err)
		}

		if len(cipherText) <= aes.BlockSize {
			t.Fatalf("Ciphertext too short for key length %d: %d", keyLength, len(cipherText))
		}
	}
}

func TestDecryptAES(t *testing.T) {
	keyLengths := []int{16, 24, 32} // AES-128, AES-192, AES-256

	for _, keyLength := range keyLengths {
		key := make([]byte, keyLength)
		if _, err := rand.Read(key); err != nil {
			t.Fatalf("Failed to generate key: %v", err)
		}

		plainText := []byte("Test plaintext")
		cipherText, err := EncryptAESGCM(plainText, key)
		if err != nil {
			t.Fatalf("EncryptAES failed for key length %d: %v", keyLength, err)
		}

		decryptedText, err := DecryptAESGCM(cipherText, key)
		if err != nil {
			t.Fatalf("DecryptAES failed for key length %d: %v", keyLength, err)
		}

		if !bytes.Equal(plainText, decryptedText) {
			t.Fatalf("Decrypted text does not match plaintext for key length %d. Got %s, want %s", keyLength, decryptedText, plainText)
		}
	}
}
