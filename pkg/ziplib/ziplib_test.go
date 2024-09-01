package ziplib

import "testing"

func TestGzipCompress(t *testing.T) {
	data := []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Lorem ipsum dolor sit amet, consectetur adipiscing elit.")

	res, err := GzipCompress(data)
	if err != nil {
		t.Fatal(err)
	}

	if len(res) >= len(data) {
		t.Fatalf("expected compressed data to be smaller than original, got %d vs %d", len(res), len(data))
	}

	t.Logf("compressed %d bytes to %d bytes", len(data), len(res))

	res, err = GzipDecompress(res)
	if err != nil {
		t.Fatal(err)
	}

	if string(res) != string(data) {
		t.Fatalf("expected decompressed data to match original, got %q vs %q", res, data)
	}
}

func TestBrotliCompress(t *testing.T) {
	data := []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Lorem ipsum dolor sit amet, consectetur adipiscing elit.")

	res, err := BrotliCompress(data)
	if err != nil {
		t.Fatal(err)
	}

	if len(res) >= len(data) {
		t.Fatalf("expected compressed data to be smaller than original, got %d vs %d", len(res), len(data))
	}

	t.Logf("compressed %d bytes to %d bytes", len(data), len(res))

	res, err = BrotliDecompress(res)
	if err != nil {
		t.Fatal(err)
	}

	if string(res) != string(data) {
		t.Fatalf("expected decompressed data to match original, got %q vs %q", res, data)
	}
}
