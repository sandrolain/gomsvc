package objects

import (
	"bytes"
	"context"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type EnvConfig struct {
	Endpoint     string `env:"OBJECTS_ENDPOINT" validate:"required"`
	AccessId     string `env:"OBJECTS_ACCESS_ID" validate:"required"`
	AccessSecret string `env:"OBJECTS_ACCESS_SECRET" validate:"required"`
	SSL          bool   `env:"OBJECTS_SSL"`
}

func (e *EnvConfig) GetClientConfig() Config {
	return Config{
		Endpoint:     e.Endpoint,
		AccessId:     e.AccessId,
		AccessSecret: e.AccessSecret,
		SSL:          e.SSL,
	}
}

type Config struct {
	Endpoint     string
	AccessId     string
	AccessSecret string
	SSL          bool
}

func NewClient(cfg Config) (res *Client, err error) {
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessId, cfg.AccessSecret, ""),
		Secure: cfg.SSL,
	})
	if err != nil {
		return
	}
	res = &Client{
		MinIO: minioClient,
	}
	return
}

type Client struct {
	MinIO *minio.Client
}

func (c *Client) AssureBucket(ctx context.Context, bucketName string) (err error) {
	exists, err := c.MinIO.BucketExists(ctx, bucketName)
	if exists || err != nil {
		return
	}
	err = c.MinIO.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	return
}

type Object struct {
	BucketName  string
	ObjectName  string
	FilePath    string
	Data        []byte
	ContentType string
}

func (c *Client) PutObjects(ctx context.Context, objects []Object) (infos []minio.UploadInfo, err error) {
	infos = make([]minio.UploadInfo, len(objects))
	for i, o := range objects {
		var info minio.UploadInfo
		switch {
		case o.FilePath != "":
			info, err = c.MinIO.FPutObject(ctx, o.BucketName, o.ObjectName, o.FilePath, minio.PutObjectOptions{ContentType: o.ContentType})
			if err != nil {
				return
			}
			infos[i] = info
		case len(o.Data) > 0:
			reader := bytes.NewReader(o.Data)
			objectSize := int64(len(o.Data))
			info, err = c.MinIO.PutObject(ctx, o.BucketName, o.ObjectName, reader, objectSize, minio.PutObjectOptions{ContentType: o.ContentType})
			if err != nil {
				return
			}
			infos[i] = info
		}
	}
	return
}
