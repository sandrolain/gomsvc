package datalib

import "github.com/flytam/filenamify"

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
