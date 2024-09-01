package jwxlib

import (
	"crypto/rsa"
	"fmt"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwe"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

func JweEncryptMultiRsa(plaintext []byte, rsaPubKeys []*rsa.PublicKey) (ciphertext []byte, err error) {
	pubkeys := make([]jwk.Key, len(rsaPubKeys))

	for i, key := range rsaPubKeys {
		pubkey, e := jwk.FromRaw(key)
		if e != nil {
			err = fmt.Errorf("failed to create jwk: %s", e)
			return
		}
		pubkeys[i] = pubkey
	}

	options := []jwe.EncryptOption{jwe.WithJSON()}
	for _, key := range pubkeys {
		options = append(options, jwe.WithKey(jwa.RSA_OAEP, key))
	}

	ciphertext, e := jwe.Encrypt(plaintext, options...)
	if e != nil {
		err = fmt.Errorf("failed to encrypt: %s", e)
		return
	}

	return
}

func JweEncryptRsa(plaintext []byte, rsaPubKey *rsa.PublicKey, json bool) (ciphertext []byte, err error) {
	pubkey, e := jwk.FromRaw(rsaPubKey)
	if e != nil {
		err = fmt.Errorf("failed to create jwk: %s", e)
		return
	}

	options := []jwe.EncryptOption{jwe.WithKey(jwa.RSA_OAEP, pubkey)}

	if json {
		options = append(options, jwe.WithJSON())
	}

	ciphertext, e = jwe.Encrypt(plaintext, options...)
	if e != nil {
		err = fmt.Errorf("failed to encrypt: %s", e)
		return
	}

	return
}

func JweDecryptRsa(ciphertext []byte, rsaPrivKey *rsa.PrivateKey) (plaintext []byte, err error) {
	privkey, e := jwk.FromRaw(rsaPrivKey)
	if e != nil {
		err = fmt.Errorf("failed to create jwk: %s", e)
		return
	}

	options := []jwe.DecryptOption{jwe.WithKey(jwa.RSA_OAEP, privkey)}

	plaintext, e = jwe.Decrypt(ciphertext, options...)
	if e != nil {
		err = fmt.Errorf("failed to decrypt: %s", e)
		return
	}

	return
}
