package postgis

import (
	"fmt"
	"regexp"

	"github.com/twpayne/go-geom/encoding/ewkbhex"
)

func HexCoordsToPoint(hex string) (res string, err error) {
	g, err := ewkbhex.Decode(hex)
	if err != nil {
		return
	}
	coords := g.FlatCoords()
	res = fmt.Sprintf("POINT(%v %v)", coords[0], coords[1])
	return
}

var latLngPoint = regexp.MustCompile(`^\s*POINT\s*\(\s*(\d+\.\d+)\s+(\d+\.\d+)\s*\)\s*$`)
var latLngOnly = regexp.MustCompile(`^\s*(\d+\.\d+)[\s,]+(\d+\.\d+)\s*$`)

func ValidatePoint(pnt string) (string, error) {
	if latLngPoint.MatchString(pnt) {
		return pnt, nil
	}
	if latLngOnly.MatchString(pnt) {
		return fmt.Sprintf("POINT(%s)", pnt), nil
	}
	return "", fmt.Errorf("invalid point %s", pnt)
}
