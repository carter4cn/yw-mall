package minioutil

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	Bucket     string
	PublicHost string
	UseSSL     bool
}

type Client struct {
	raw    *minio.Client
	bucket string
	host   string
}

func New(cfg Config) (*Client, error) {
	raw, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}
	host := cfg.PublicHost
	if host == "" {
		scheme := "http"
		if cfg.UseSSL {
			scheme = "https"
		}
		host = fmt.Sprintf("%s://%s", scheme, cfg.Endpoint)
	}
	host = strings.TrimRight(host, "/")
	return &Client{raw: raw, bucket: cfg.Bucket, host: host}, nil
}

func (c *Client) PutFile(ctx context.Context, key string, r io.Reader, size int64, contentType string) (string, error) {
	_, err := c.raw.PutObject(ctx, c.bucket, key, r, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", err
	}
	return c.PublicURL(key), nil
}

func (c *Client) PutBytes(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	return c.PutFile(ctx, key, bytes.NewReader(data), int64(len(data)), contentType)
}

func (c *Client) Exists(ctx context.Context, key string) bool {
	_, err := c.raw.StatObject(ctx, c.bucket, key, minio.StatObjectOptions{})
	return err == nil
}

func (c *Client) PublicURL(key string) string {
	return fmt.Sprintf("%s/%s/%s", c.host, c.bucket, key)
}

func (c *Client) Bucket() string { return c.bucket }
