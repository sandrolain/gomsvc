package ziplib

import (
	"bytes"
	"compress/gzip"
	"fmt"

	"github.com/andybalholm/brotli"
)

func GzipCompress(data []byte) (res []byte, err error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, e := gz.Write(data)
	if e != nil {
		gz.Close()
		err = fmt.Errorf("failed to write gzip writer: %s", e)
		return
	}
	if e := gz.Close(); e != nil {
		err = fmt.Errorf("failed to close gzip writer: %s", e)
		return
	}
	res = buf.Bytes()
	return
}

func GzipDecompress(data []byte) (res []byte, err error) {
	var buf bytes.Buffer
	gz, e := gzip.NewReader(bytes.NewReader(data))
	if e != nil {
		err = fmt.Errorf("failed to create gzip reader: %s", e)
		return
	}

	_, e = buf.ReadFrom(gz)
	if e != nil {
		gz.Close()
		err = fmt.Errorf("failed to read gzip reader: %s", e)
		return
	}

	if e := gz.Close(); e != nil {
		err = fmt.Errorf("failed to close gzip reader: %s", e)
		return
	}

	res = buf.Bytes()
	return
}

func BrotliCompress(data []byte) (res []byte, err error) {
	var buf bytes.Buffer
	w := brotli.NewWriter(&buf)
	_, e := w.Write(data)
	if e != nil {
		err = fmt.Errorf("failed to write brotli writer: %s", e)
		return
	}
	if e := w.Close(); e != nil {
		err = fmt.Errorf("failed to close brotli writer: %s", e)
		return
	}
	res = buf.Bytes()
	return
}

func BrotliDecompress(data []byte) (res []byte, err error) {
	var buf bytes.Buffer
	r := brotli.NewReader(bytes.NewReader(data))
	_, e := buf.ReadFrom(r)
	if e != nil {
		err = fmt.Errorf("failed to read brotli reader: %s", e)
		return
	}
	res = buf.Bytes()
	return
}
