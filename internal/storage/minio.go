package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
	client     *minio.Client
	bucketName string
	endpoint   string
	useSSL     bool
}

func NewMinioClient() (*MinioClient, error) {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" {
		endpoint = "minio:9000"
	}

	accessKey := os.Getenv("MINIO_ROOT_USER")
	if accessKey == "" {
		accessKey = "minioadmin"
	}

	secretKey := os.Getenv("MINIO_ROOT_PASSWORD")
	if secretKey == "" {
		secretKey = "minioadmin"
	}

	bucketName := os.Getenv("MINIO_BUCKET")
	if bucketName == "" {
		bucketName = "chat-uploads"
	}

	useSSL := os.Getenv("MINIO_USE_SSL") == "true"

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	mc := &MinioClient{
		client:     client,
		bucketName: bucketName,
		endpoint:   endpoint,
		useSSL:     useSSL,
	}

	// Create bucket if not exists
	if err := mc.ensureBucket(); err != nil {
		return nil, err
	}

	return mc, nil
}

func (m *MinioClient) ensureBucket() error {
	ctx := context.Background()
	exists, err := m.client.BucketExists(ctx, m.bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket: %w", err)
	}

	if !exists {
		err = m.client.MakeBucket(ctx, m.bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}

		// Set bucket policy to allow public read
		policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Effect": "Allow",
				"Principal": {"AWS": ["*"]},
				"Action": ["s3:GetObject"],
				"Resource": ["arn:aws:s3:::%s/*"]
			}]
		}`, m.bucketName)

		err = m.client.SetBucketPolicy(ctx, m.bucketName, policy)
		if err != nil {
			return fmt.Errorf("failed to set bucket policy: %w", err)
		}
	}

	return nil
}

func (m *MinioClient) Upload(file io.Reader, originalName string, size int64, contentType string) (string, error) {
	ctx := context.Background()

	// Generate unique filename
	ext := filepath.Ext(originalName)
	filename := uuid.New().String() + ext

	_, err := m.client.PutObject(ctx, m.bucketName, filename, file, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// Return backend-accessible URL
	url := "/files/" + filename
	return url, nil
}

func (m *MinioClient) GetFile(filename string) (io.ReadCloser, string, error) {
	ctx := context.Background()

	obj, err := m.client.GetObject(ctx, m.bucketName, filename, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", fmt.Errorf("failed to get file: %w", err)
	}

	stat, err := obj.Stat()
	if err != nil {
		obj.Close()
		return nil, "", fmt.Errorf("failed to get file info: %w", err)
	}

	return obj, stat.ContentType, nil
}

func (m *MinioClient) Delete(filename string) error {
	ctx := context.Background()
	return m.client.RemoveObject(ctx, m.bucketName, filename, minio.RemoveObjectOptions{})
}

var allowedTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

var allowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
}

func IsAllowedImageType(contentType string) bool {
	return allowedTypes[strings.ToLower(contentType)]
}

func IsAllowedExtension(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return allowedExtensions[ext]
}

const MaxFileSize = 10 * 1024 * 1024 // 10MB
