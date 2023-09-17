package envtype

import "encoding/base64"

type Password []byte

func (p *Password) UnmarshalText(text []byte) (err error) {
	out, err := base64.StdEncoding.DecodeString(string(text))
	if err == nil {
		*p = out
	}
	return
}
