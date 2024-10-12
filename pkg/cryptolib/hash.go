package cryptolib

import (
	"bytes"
	"crypto/sha256"

	"golang.org/x/crypto/bcrypt"
)

func HashBCrypt(value []byte) ([]byte, error) {
	return bcrypt.GenerateFromPassword(value, 14)
}

func CompareBCrypt(value []byte, hash []byte) bool {
	err := bcrypt.CompareHashAndPassword(hash, value)
	return err == nil
}

func HashSHA256(value []byte) []byte {
	h := sha256.New()
	h.Write(value)
	return h.Sum(nil)
}

func CompareSHA256(value []byte, hash []byte) bool {
	h := HashSHA256(value)
	return bytes.Equal(hash, h)
}
