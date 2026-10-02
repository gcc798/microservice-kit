package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/gcc798/microservice-kit/application/resource-rpc/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Storage struct {
	client *minio.Client
	bucket string
}

func New(cfg config.StorageConf) (*Storage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""), Secure: cfg.UseSSL, Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("initialize object storage: %w", err)
	}
	return &Storage{client: client, bucket: cfg.Bucket}, nil
}

func (s *Storage) Upload(ctx context.Context, key, contentType string, reader io.Reader, size int64) error {
	_, err := s.client.PutObject(ctx, s.bucket, key, reader, size, minio.PutObjectOptions{ContentType: contentType})
	return err
}

func (s *Storage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	return s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
}

func (s *Storage) Delete(ctx context.Context, key string) error {
	return s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
}

func (s *Storage) URL(ctx context.Context, key string, expires time.Duration) (string, error) {
	if expires == 0 {
		return fmt.Sprintf("%s/%s/%s", s.client.EndpointURL(), s.bucket, key), nil
	}
	u, err := s.client.PresignedGetObject(ctx, s.bucket, key, expires, nil)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}
