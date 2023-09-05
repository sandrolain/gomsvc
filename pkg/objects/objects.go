package objects

import (
	"bytes"
	"context"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type EnvObjectsConfig struct {
	ObjectsEndpoint     string `env:"OBJECTS_ENDPOINT" validate:"required"`
	ObjectsAccessId     string `env:"OBJECTS_ACCESS_ID" validate:"required"`
	ObjectsAccessSecret string `env:"OBJECTS_ACCESS_SECRET" validate:"required"`
	ObjectsSSL          bool   `env:"OBJECTS_SSL"`
}

func (e *EnvObjectsConfig) GetObjectsClientConfig() Config {
	return Config{
		Endpoint:     e.ObjectsEndpoint,
		AccessId:     e.ObjectsAccessId,
		AccessSecret: e.ObjectsAccessSecret,
		SSL:          e.ObjectsSSL,
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

func (c *Client) PutFile(ctx context.Context, bucketName string, objectName string, filePath string, contentType string) (info minio.UploadInfo, err error) {
	info, err = c.MinIO.FPutObject(ctx, bucketName, objectName, filePath, minio.PutObjectOptions{ContentType: contentType})
	return
}

func (c *Client) PutData(ctx context.Context, bucketName string, objectName string, data []byte, contentType string) (info minio.UploadInfo, err error) {
	reader := bytes.NewReader(data)
	objectSize := int64(len(data))
	info, err = c.MinIO.PutObject(ctx, bucketName, objectName, reader, objectSize, minio.PutObjectOptions{ContentType: contentType})
	return
}
