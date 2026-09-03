package storage

import (
	"context"
)

type Client struct{}

func NewClient(endpoint, accessKey, secretKey string, useSSL bool) (*Client, error) {
	return &Client{}, nil
}

func (c *Client) PutObject(ctx context.Context, bucketName, objectName string, data []byte) error {
	return nil
}