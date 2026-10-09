package s3store

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/lucast1574/harkai-backend/internal/core"
	"github.com/lucast1574/harkai-backend/internal/security"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Store struct {
	client *minio.Client
	bucket string
}

func Connect(ctx context.Context, endpoint, access, secret, bucket string, secure bool) (*Store, error) {
	client, err := minio.New(endpoint, &minio.Options{Creds: credentials.NewStaticV4(access, secret, ""), Secure: secure})
	if err != nil {
		return nil, errors.New("invalid S3 configuration")
	}
	s := &Store{client: client, bucket: bucket}
	if err = s.Ping(ctx); err != nil {
		return nil, errors.New("S3 bucket unavailable")
	}
	return s, nil
}

func (s *Store) Ping(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil || !exists {
		return core.ErrUnavailable
	}
	return nil
}

func (s *Store) Put(ctx context.Context, id string, data []byte) error {
	_, err := s.client.PutObject(ctx, s.bucket, "photos/"+id+".jpg", bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{ContentType: "image/jpeg"})
	if err != nil {
		return core.ErrUnavailable
	}
	return nil
}

func (s *Store) Get(ctx context.Context, id string) ([]byte, error) {
	object, err := s.client.GetObject(ctx, s.bucket, "photos/"+id+".jpg", minio.GetObjectOptions{})
	if err != nil {
		return nil, core.ErrUnavailable
	}
	defer object.Close()
	data, err := io.ReadAll(io.LimitReader(object, security.MaxImageBytes+1))
	if err != nil || len(data) > security.MaxImageBytes {
		return nil, core.ErrUnavailable
	}
	return data, nil
}
