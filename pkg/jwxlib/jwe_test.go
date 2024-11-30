package jwxlib

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
)

func TestJweEncrypt(t *testing.T) {
	privkeys := make([]*rsa.PrivateKey, 3)
	pubkeys := make([]interface{}, 3)

	for i := range privkeys {
		var err error
		privkeys[i], err = rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			t.Fatal(err.Error())
		}
		pubkeys[i] = &privkeys[i].PublicKey
	}

	plain := "Lorem ipsum dolor sit amet, consectetur adipiscing elit."

	ct, err := JweEncrypt([]byte(plain), pubkeys)
	if err != nil {
		t.Fatalf("error encrypting JWE: %s", err.Error())
	}

	t.Logf("ct: %s", string(ct))

	for i, privkey := range privkeys {
		pt, err := JweDecrypt(ct, []interface{}{privkey})
		if err != nil {
			t.Fatalf("error decrypting with key %d: %s", i, err)
		}
		if string(pt) != plain {
			t.Errorf("decrypted text mismatch for key %d: expected %q, got %q", i, plain, pt)
		}
	}
}

func TestJweDecrypt(t *testing.T) {
	privkey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err.Error())
	}

	plain := "Lorem ipsum dolor sit amet, consectetur adipiscing elit."

	ct, err := JweEncrypt([]byte(plain), []interface{}{&privkey.PublicKey})
	if err != nil {
		t.Fatalf("error encrypting JWE: %s", err.Error())
	}

	t.Logf("ct: %s", string(ct))

	pt, err := JweDecrypt(ct, []interface{}{privkey})
	if err != nil {
		t.Fatalf("error decrypting with key: %s", err)
	}
	if string(pt) != plain {
		t.Errorf("decrypted text mismatch: expected %q, got %q", plain, pt)
	}

	wrongKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	_, err = JweDecrypt(ct, []interface{}{wrongKey})
	if err == nil {
		t.Fatal("expected error decrypting with wrong key")
	}

	// Test with a different key
	anotherKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	ct, err = JweEncrypt([]byte(plain), []interface{}{&anotherKey.PublicKey})
	if err != nil {
		t.Fatalf("error encrypting JWE: %s", err.Error())
	}

	t.Logf("ct: %s", string(ct))

	pt, err = JweDecrypt(ct, []interface{}{anotherKey})
	if err != nil {
		t.Fatalf("error decrypting with key: %s", err)
	}
	if string(pt) != plain {
		t.Errorf("decrypted text mismatch: expected %q, got %q", plain, pt)
	}
}
