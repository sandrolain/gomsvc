package pwdlib

import (
	"fmt"
	"net/url"

	pwdgen "github.com/sethvargo/go-password/password"
	pwdval "github.com/wagslane/go-password-validator"
)

const (
	MinEntropy = 60
)

func ValidatePasswordEntropy(password string) error {
	return pwdval.Validate(password, MinEntropy)
}

func GeneratePassword(len int) (string, error) {
	dig := len / 4
	sim := len / 4
	return pwdgen.Generate(len, dig, sim, false, false)
}

func GetPasswordInURI(uri string) (pwd string, err error) {
	u, err := url.Parse(uri)
	if err != nil {
		return
	}
	pwd, ok := u.User.Password()
	if !ok {
		err = fmt.Errorf("the URI does not contain a password")
	}
	return
}
