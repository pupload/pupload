package s3

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Store struct {
	client   *minio.Client
	location string
	bucket   string
}

func NewS3Store(endpoint, accessKeyID, secretAccessKey, location, bucket string) *S3Store {
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

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure:    true,
		Transport: transport,
	})

	if err != nil {
		log.Fatalln(err)
	}

	return &S3Store{
		client:   minioClient,
		location: location,
		bucket:   bucket,
	}
}

/*
func (s *S3Store) PutFromFile(ctx context.Context, objectName, filePath, content_type string) (*minio.UploadInfo, error) {
	info, err := s.client.FPutObject(ctx, s.bucket, objectName, filePath, minio.PutObjectOptions{
		ContentType: content_type,
	})

	if err != nil {
		return nil, err
	}

	return &info, nil
}

func (s *S3Store) PutFromStream(ctx context.Context, objectName string, data io.Reader) (*minio.UploadInfo, error) {
	info, err := s.client.PutObject(ctx, s.bucket, objectName, data, -1, minio.PutObjectOptions{
		ContentType: "image",
	})

	if err != nil {
		return nil, err
	}

	return &info, nil
}
*/

func (s *S3Store) PutPresignedURL(ctx context.Context, objectName string, expires time.Duration) (u *url.URL, err error) {
	return s.client.PresignedPutObject(ctx, s.bucket, objectName, expires)
}

func (s *S3Store) GetPresignedURL(ctx context.Context, objectName string, expires time.Duration) (u *url.URL, err error) {
	return s.client.PresignedGetObject(ctx, s.bucket, objectName, expires, url.Values{})
}

func (s *S3Store) DeleteObject(ctx context.Context, objectName string) error {
	return s.client.RemoveObject(ctx, s.bucket, objectName, minio.RemoveObjectOptions{})
}

func (s *S3Store) Close() {

}
