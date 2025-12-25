package s3

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/escalopa/family-tree/internal/domain"
	"github.com/google/uuid"
)

type S3Client struct {
	client            *s3.Client
	bucket            string
	maxImageSize      int64
	allowedImageTypes map[string]bool
}

func NewS3Client(ctx context.Context, endpoint, region, accessKey, secretKey, bucket string, maxImageSize int64, allowedExts []string) (*S3Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		slog.Error("S3Client.NewS3Client: load AWS config", "error", err)
		return nil, domain.NewInternalError(err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true // Required for MinIO and some S3-compatible services
	})

	allowedTypes := make(map[string]bool)
	for _, ext := range allowedExts {
		allowedTypes[strings.ToLower(ext)] = true
	}

	return &S3Client{
		client:            client,
		bucket:            bucket,
		maxImageSize:      maxImageSize,
		allowedImageTypes: allowedTypes,
	}, nil
}

func (s *S3Client) UploadImage(ctx context.Context, data []byte, filename string) (string, error) {
	if int64(len(data)) > s.maxImageSize {
		slog.Warn("S3Client.UploadImage: image size exceeds maximum", "size", len(data), "max_size", s.maxImageSize, "filename", filename)
		return "", domain.NewValidationError("error.validation.file_too_large")
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if !s.allowedImageTypes[ext] {
		slog.Warn("S3Client.UploadImage: unsupported image type", "extension", ext, "filename", filename)
		return "", domain.NewValidationError("error.validation.invalid_file")
	}

	if err := s.validateImageContent(data); err != nil {
		return "", err
	}

	key := fmt.Sprintf("members/%s%s", uuid.New().String(), ext)

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		slog.Error("S3Client.UploadImage: upload to S3", "error", err, "key", key)
		return "", domain.NewExternalServiceError(err)
	}

	slog.Info("S3Client.UploadImage: uploaded", "key", key)
	return key, nil
}

func (s *S3Client) DeleteImage(ctx context.Context, key string) error {
	if key == "" {
		return nil
	}

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		slog.Error("S3Client.DeleteImage: delete from S3", "error", err, "key", key)
		return domain.NewExternalServiceError(err)
	}

	slog.Info("S3Client.DeleteImage: deleted", "key", key)
	return nil
}

func (s *S3Client) GetImage(ctx context.Context, key string) ([]byte, error) {
	if key == "" {
		slog.Warn("S3Client.GetImage: empty key provided")
		return nil, domain.NewNotFoundError("image")
	}

	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		slog.Error("S3Client.GetImage: get from S3", "error", err, "key", key)
		return nil, domain.NewExternalServiceError(err)
	}
	defer result.Body.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(result.Body); err != nil {
		slog.Error("S3Client.GetImage: read response body", "error", err, "key", key)
		return nil, domain.NewInternalError(err)
	}

	return buf.Bytes(), nil
}

func (s *S3Client) validateImageContent(data []byte) error {
	if len(data) < 12 {
		return domain.NewValidationError("error.validation.invalid_file")
	}

	// Check magic numbers for common image formats
	// JPEG: FF D8 FF
	if len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return nil
	}

	// PNG: 89 50 4E 47 0D 0A 1A 0A
	if len(data) >= 8 && data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 &&
		data[4] == 0x0D && data[5] == 0x0A && data[6] == 0x1A && data[7] == 0x0A {
		return nil
	}

	// GIF: 47 49 46 38 (GIF8)
	if len(data) >= 4 && data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x38 {
		return nil
	}

	// WebP: 52 49 46 46 ... 57 45 42 50 (RIFF...WEBP)
	if len(data) >= 12 && data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 &&
		data[8] == 0x57 && data[9] == 0x45 && data[10] == 0x42 && data[11] == 0x50 {
		return nil
	}

	slog.Warn("S3Client.validateImageContent: invalid image content", "first_bytes", fmt.Sprintf("%X", data[:min(12, len(data))]))
	return domain.NewValidationError("error.validation.invalid_file")
}
