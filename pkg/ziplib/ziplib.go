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

	if _, err = gz.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write gzip writer: %s", err)
	}

	if err = gz.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %s", err)
	}

	res = buf.Bytes()
	return
}

func GzipDecompress(data []byte) (res []byte, err error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %s", err)
	}
	defer func() {
		if closeErr := gz.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close gzip reader: %s", closeErr)
		}
	}()

	var buf bytes.Buffer
	if _, err = buf.ReadFrom(gz); err != nil {
		return nil, fmt.Errorf("failed to read gzip reader: %s", err)
	}

	res = buf.Bytes()
	return
}

func BrotliCompress(data []byte) (res []byte, err error) {
	var buf bytes.Buffer
	w := brotli.NewWriter(&buf)

	if _, err = w.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write brotli writer: %s", err)
	}

	if err = w.Close(); err != nil {
		return nil, fmt.Errorf("failed to close brotli writer: %s", err)
	}

	res = buf.Bytes()
	return
}

func BrotliDecompress(data []byte) (res []byte, err error) {
	var buf bytes.Buffer
	r := brotli.NewReader(bytes.NewReader(data))
	if _, err = buf.ReadFrom(r); err != nil {
		return nil, fmt.Errorf("failed to read brotli reader: %s", err)
	}

	res = buf.Bytes()
	return
}
