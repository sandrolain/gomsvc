package jwxlib

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
)

func TestJweEncryptMultiRsa(t *testing.T) {
	privkeys := make([]*rsa.PrivateKey, 3)
	pubkeys := make([]*rsa.PublicKey, 3)

	for i := range privkeys {
		var err error
		privkeys[i], err = rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			t.Fatal(err.Error())
		}
		pubkeys[i] = &privkeys[i].PublicKey
	}

	plain := "Lorem ipsum dolor sit amet, consectetur adipiscing elit."

	ct, err := JweEncryptMultiRsa([]byte(plain), pubkeys)
	if err != nil {
		t.Fatalf("error encrypting JWE: %s", err.Error())
	}

	t.Logf("ct: %s", ct)

	for i, privkey := range privkeys {
		pt, err := JweDecryptRsa([]byte(ct), privkey)
		if err != nil {
			t.Fatalf("error decrypting with key %d: %s", i, err)
		}
		if string(pt) != plain {
			t.Errorf("decrypted text mismatch for key %d: expected %q, got %q", i, plain, pt)
		}
	}
}

func TestJweDecryptRsa(t *testing.T) {
	privkey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err.Error())
	}

	plain := "Lorem ipsum dolor sit amet, consectetur adipiscing elit."

	ct, err := JweEncryptRsa([]byte(plain), &privkey.PublicKey, false)
	if err != nil {
		t.Fatalf("error encrypting JWE: %s", err.Error())
	}

	t.Logf("ct: %s", ct)

	pt, err := JweDecryptRsa([]byte(ct), privkey)
	if err != nil {
		t.Fatalf("error decrypting with key: %s", err)
	}
	if string(pt) != plain {
		t.Errorf("decrypted text mismatch: expected %q, got %q", plain, pt)
	}

	_, err = JweDecryptRsa([]byte(ct), &rsa.PrivateKey{})
	if err == nil {
		t.Fatal("expected error decrypting with wrong key")
	}

	ct, err = JweEncryptRsa([]byte(plain), &privkey.PublicKey, true)
	if err != nil {
		t.Fatalf("error encrypting JWE: %s", err.Error())
	}

	t.Logf("ct: %s", ct)

	pt, err = JweDecryptRsa([]byte(ct), privkey)
	if err != nil {
		t.Fatalf("error decrypting with key: %s", err)
	}
	if string(pt) != plain {
		t.Errorf("decrypted text mismatch: expected %q, got %q", plain, pt)
	}
}
