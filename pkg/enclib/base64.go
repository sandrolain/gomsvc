package enclib

import (
	"encoding/base64"
)

func EncodeBase64(value []byte) string {
	return base64.StdEncoding.EncodeToString(value)
}

func DecodeBase64(value string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(value)
}

func EncodeBase64URL(value []byte) string {
	return base64.URLEncoding.EncodeToString(value)
}

func DecodeBase64URL(value string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(value)
}
