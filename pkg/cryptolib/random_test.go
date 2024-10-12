package cryptolib

import (
	"bytes"
	"testing"
)

func TestRandomBytes(t *testing.T) {
	b1, err := RandomBytes(16)
	if err != nil {
		t.Fatal(err)
	}
	b2, err := RandomBytes(16)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(b1, b2) {
		t.Fatal("random bytes should be different")
	}
}

func TestRandomBytesBase64(t *testing.T) {
	s1, err := RandomBytesBase64(16)
	if err != nil {
		t.Fatal(err)
	}
	s2, err := RandomBytesBase64(16)
	if err != nil {
		t.Fatal(err)
	}
	if s1 == s2 {
		t.Fatal("random base64 strings should be different")
	}
}

func TestGenerateAES256Key(t *testing.T) {
	key1, err := GenerateAES256Key()
	if err != nil {
		t.Fatal(err)
	}
	key2, err := GenerateAES256Key()
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(key1, key2) {
		t.Fatal("AES-256 keys should be different")
	}
}
