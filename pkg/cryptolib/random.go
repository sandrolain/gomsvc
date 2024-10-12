package cryptolib

import (
	"crypto/rand"
	"encoding/base64"
	"io"
)

func RandomBytes(length int) (res []byte, err error) {
	res = make([]byte, length)
	_, err = rand.Read(res)
	return
}

func RandomBytesBase64(length int) (res string, err error) {
	b, err := RandomBytes(length)
	if err != nil {
		return
	}
	res = base64.StdEncoding.EncodeToString(b)
	return
}

func GenerateAES256Key() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	return key, nil
}
