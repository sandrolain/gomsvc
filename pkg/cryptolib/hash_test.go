package cryptolib

import (
	"testing"
)

func TestHashBCrypt(t *testing.T) {
	hash, err := HashBCrypt([]byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
	if len(hash) == 0 {
		t.Fatal("BCrypt hash should not be empty")
	}
}

func TestCompareBCrypt(t *testing.T) {
	hash, err := HashBCrypt([]byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
	if !CompareBCrypt([]byte("hello"), hash) {
		t.Fatal("CompareBCrypt should return true")
	}
}

func TestHashSHA256(t *testing.T) {
	hash := HashSHA256([]byte("hello"))
	if len(hash) == 0 {
		t.Fatal("sha256 hash should not be empty")
	}
}

func TestCompareSHA256(t *testing.T) {
	hash := HashSHA256([]byte("hello"))
	if !CompareSHA256([]byte("hello"), hash) {
		t.Fatal("CompareSHA256 should return true")
	}
}
