package enclib

import "encoding/hex"

func HexEncode(value []byte) string {
	return hex.EncodeToString(value)
}

func HexDecode(value string) ([]byte, error) {
	return hex.DecodeString(value)
}
