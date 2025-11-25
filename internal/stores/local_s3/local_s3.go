package locals3

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type LocalS3Store struct {
	server *httptest.Server
	client *minio.Client
	bucket string
}

type LocalS3StoreInput struct {
	BucketName string
}

func NewLocalS3Store(input LocalS3StoreInput) (*LocalS3Store, error) {
	backend := s3mem.New()
	faker := gofakes3.New(backend)

	ts := httptest.NewServer(faker.Server())

	u, err := url.Parse(ts.URL)
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		MaxIdleConns:        200,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 5 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
	}

	minioClient, err := minio.New(u.Host, &minio.Options{
		Creds:     credentials.NewStaticV4("ACCESS_KEY", "SECRET_KEY", ""),
		Secure:    false,
		Transport: transport,
	})

	if err != nil {
		return nil, err
	}

	minioClient.MakeBucket(context.TODO(), input.BucketName, minio.MakeBucketOptions{})

	return &LocalS3Store{
		server: ts,
		client: minioClient,
		bucket: input.BucketName,
	}, nil

}

func (s *LocalS3Store) PutURL(ctx context.Context, objectName string, expires time.Duration) (u *url.URL, err error) {
	return s.client.PresignedPutObject(ctx, s.bucket, objectName, expires)
}

func (s *LocalS3Store) GetURL(ctx context.Context, objectName string, expires time.Duration) (u *url.URL, err error) {
	return s.client.PresignedGetObject(ctx, s.bucket, objectName, expires, url.Values{})
}

func (s *LocalS3Store) DeleteObject(ctx context.Context, objectName string) error {
	return s.client.RemoveObject(ctx, s.bucket, objectName, minio.RemoveObjectOptions{})
}

func (s *LocalS3Store) Close() {
	s.server.Close()
}

func (s *LocalS3Store) Exists(objectName string) bool {
	_, err := s.client.StatObject(context.TODO(), s.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return false
	}

	return true
}
