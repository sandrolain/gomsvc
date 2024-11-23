package httplib

import (
	"regexp"
	"strings"
)

var reSpaces = regexp.MustCompile(`\s+`)

func parsePath(parts ...string) (method string, path string) {
	partsNum := len(parts)
	if partsNum == 1 {
		parts = reSpaces.Split(parts[0], -1)
		partsNum = len(parts)
	}
	switch partsNum {
	case 1:
		path = parts[0]
	default:
		method = strings.ToUpper(parts[0])
		path = parts[1]
	}
	return
}
