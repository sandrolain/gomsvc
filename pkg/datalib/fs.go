package datalib

import (
	"strings"

	"github.com/flytam/filenamify"
)

func SafeDirName(parts ...string) (fn string, err error) {
	filename := strings.Join(parts, "")
	fn, err = filenamify.Filenamify(filename, filenamify.Options{
		Replacement: "_",
	})
	if err != nil {
		return
	}
	return
}

func SafeFilename(filename string, ext ...string) (fn string, err error) {
	fn, err = filenamify.Filenamify(filename, filenamify.Options{
		Replacement: "_",
	})
	if err != nil {
		return
	}
	if len(ext) > 0 {
		fn = fn + "." + ext[0]
	}
	return
}
