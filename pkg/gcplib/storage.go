// Package gcplib provides a client for interacting with GCP storage buckets.
//
// Example usage:
//
//	bucket, err := NewBucket(ctx, "my-bucket")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Upload a file to the bucket
//	w, err := os.Open("path/to/file")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer w.Close()
//	n, err := bucket.Upload(ctx, "path/to/object", w)
//	if err != nil {
//		log.Fatal(err)
//	}
//	log.Printf("uploaded %d bytes to %s", n, bucket.Bucket)
//
//	// Download a file from the bucket
//	r, err := bucket.Download(ctx, "path/to/object")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer r.Close()
//	w, err = os.Create("path/to/downloaded/file")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer w.Close()
//	n, err = io.Copy(w, r)
//	if err != nil {
//		log.Fatal(err)
//	}
//	log.Printf("downloaded %d bytes from %s", n, bucket.Bucket)
//
//	// Check if a file exists in the bucket
//	exists, err := bucket.Exists(ctx, "path/to/object")
//	if err != nil {
//		log.Fatal(err)
//	}
//	log.Printf("file exists in %s: %t", bucket.Bucket, exists)
//
//	// List the objects in the bucket
//	objects, err := bucket.List(ctx, &ListOptions{Prefix: "path/to/prefix"})
//	if err != nil {
//		log.Fatal(err)
//	}
//	for _, o := range objects {
//		log.Printf("%s/%s", bucket.Bucket, o.Name)
//	}
package gcplib

import (
	"context"
	"io"
	"log"

	"cloud.google.com/go/storage"
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"
)

// Bucket represents a GCP storage bucket
type Bucket struct {
	Bucket string
	client *storage.Client
}

// NewBucket returns a new GCP storage bucket
func NewBucket(ctx context.Context, bucket string) (*Bucket, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create GCP storage client")
	}
	return &Bucket{bucket, client}, nil
}

// Upload uploads a file to a GCP bucket
func (b *Bucket) Upload(ctx context.Context, object string, r io.Reader) (int64, error) {
	log.Printf("uploading to %s/%s", b.Bucket, object)
	w := b.client.Bucket(b.Bucket).Object(object).NewWriter(ctx)
	defer func() {
		if err := w.Close(); err != nil {
			log.Printf("error closing GCP storage writer: %s", err)
		}
	}()
	n, err := io.Copy(w, r)
	if err != nil {
		return 0, errors.Wrap(err, "error uploading to GCP storage")
	}
	return n, nil
}

// Download downloads a file from a GCP bucket
func (b *Bucket) Download(ctx context.Context, object string, w io.Writer) (int64, error) {
	log.Printf("downloading from %s/%s", b.Bucket, object)
	r, err := b.client.Bucket(b.Bucket).Object(object).NewReader(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "error downloading from GCP storage")
	}
	defer func() {
		if err := r.Close(); err != nil {
			log.Printf("error closing GCP storage reader: %s", err)
		}
	}()
	n, err := io.Copy(w, r)
	if err != nil {
		return 0, errors.Wrap(err, "error downloading from GCP storage")
	}
	return n, nil
}

// Exists checks if a file exists in a GCP bucket
func (b *Bucket) Exists(ctx context.Context, object string) (bool, error) {
	_, err := b.client.Bucket(b.Bucket).Object(object).Attrs(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return false, nil
		}
		return false, errors.Wrap(err, "error checking if GCP storage file exists")
	}
	return true, nil
}

// Info returns information about an object without downloading it
func (b *Bucket) Info(ctx context.Context, object string) (*storage.ObjectAttrs, error) {
	return b.client.Bucket(b.Bucket).Object(object).Attrs(ctx)
}

// Delete deletes a file from a GCP bucket
func (b *Bucket) Delete(ctx context.Context, object string) error {
	err := b.client.Bucket(b.Bucket).Object(object).Delete(ctx)
	if err != nil {
		return errors.Wrap(err, "error deleting GCP storage file")
	}
	return nil
}

// ListOptions are options for listing objects in a GCP bucket
type ListOptions struct {
	Prefix     string
	Delimiter  string
	MaxResults int
}

// List lists the objects in a GCP bucket with optional filters
func (b *Bucket) List(ctx context.Context, opts *ListOptions) ([]*storage.ObjectAttrs, error) {
	var gcpOpts storage.Query
	if opts != nil {
		if opts.Prefix != "" {
			gcpOpts.Prefix = opts.Prefix
		}
		if opts.Delimiter != "" {
			gcpOpts.Delimiter = opts.Delimiter
		}
	}

	it := b.client.Bucket(b.Bucket).Objects(ctx, &gcpOpts)

	var objects []*storage.ObjectAttrs
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, errors.Wrap(err, "error listing GCP storage objects")
		}

		objects = append(objects, attrs)

		// If MaxResults is specified, stop when we reach the limit
		if opts != nil && opts.MaxResults > 0 && len(objects) >= opts.MaxResults {
			break
		}
	}

	return objects, nil
}
