package storage

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Client owns the MinIO connection and the bucket used by PresensiGo.
type Client struct {
	client *minio.Client
	bucket string
}

// NewClient connects to MinIO and creates the bucket on first startup.
func NewClient(ctx context.Context, endpoint, accessKey, secretKey, bucket string, useSSL bool) (*Client, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("create MinIO client: %w", err)
	}

	exists, err := minioClient.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("check MinIO bucket: %w", err)
	}
	if !exists {
		if err := minioClient.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("create MinIO bucket: %w", err)
		}
	}

	return &Client{client: minioClient, bucket: bucket}, nil
}

// PutImage uploads an image and returns a signed URL valid for seven days.
func (c *Client) PutImage(ctx context.Context, objectName, contentType string, data []byte) (string, error) {
	_, err := c.client.PutObject(ctx, c.bucket, objectName, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("upload image to MinIO: %w", err)
	}

	signedURL, err := c.client.PresignedGetObject(ctx, c.bucket, objectName, 7*24*time.Hour, url.Values{})
	if err != nil {
		return "", fmt.Errorf("create signed selfie URL: %w", err)
	}
	return signedURL.String(), nil
}

func (c *Client) RemoveObject(ctx context.Context, objectName string) error {
	return c.client.RemoveObject(ctx, c.bucket, objectName, minio.RemoveObjectOptions{})
}
