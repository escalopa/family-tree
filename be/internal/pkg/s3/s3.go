package s3

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/escalopa/family-tree/internal/domain"
	"github.com/google/uuid"
)

const (
	MaxImageSize = 3 * 1024 * 1024 // 3MB
)

var allowedImageTypes = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
}

type S3Client struct {
	client *s3.Client
	bucket string
}

func NewS3Client(endpoint, region, accessKey, secretKey, bucket string) (*S3Client, error) {
	resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...any) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               endpoint,
			HostnameImmutable: true,
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
		config.WithEndpointResolverWithOptions(resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		log.Printf("Failed to load S3 config: %v", err)
		return nil, domain.NewInternalError("failed to initialize S3 client", err)
	}

	client := s3.NewFromConfig(cfg)

	return &S3Client{
		client: client,
		bucket: bucket,
	}, nil
}

func (s *S3Client) UploadImage(ctx context.Context, data []byte, filename string) (string, error) {
	// Validate size
	if len(data) > MaxImageSize {
		return "", domain.NewValidationError("image size exceeds maximum allowed size of 3MB")
	}

	// Validate file type
	ext := strings.ToLower(filepath.Ext(filename))
	if !allowedImageTypes[ext] {
		return "", domain.NewValidationError(fmt.Sprintf("unsupported image type: %s", ext))
	}

	// Generate unique filename
	key := fmt.Sprintf("members/%s%s", uuid.New().String(), ext)

	// Upload to S3
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		log.Printf("Failed to upload image to S3: %v", err)
		return "", domain.NewExternalServiceError("S3", err)
	}

	log.Printf("Successfully uploaded image: %s", key)
	// Return S3 key (not public URL, will be served through backend)
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
		log.Printf("Failed to delete image from S3: %v", err)
		return domain.NewExternalServiceError("S3", err)
	}

	log.Printf("Successfully deleted image: %s", key)
	return nil
}

func (s *S3Client) GetImage(ctx context.Context, key string) ([]byte, error) {
	if key == "" {
		return nil, domain.NewNotFoundError("image")
	}

	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		log.Printf("Failed to get image from S3: %v", err)
		return nil, domain.NewExternalServiceError("S3", err)
	}
	defer result.Body.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(result.Body); err != nil {
		log.Printf("Failed to read image data: %v", err)
		return nil, domain.NewInternalError("failed to read image data", err)
	}

	return buf.Bytes(), nil
}


