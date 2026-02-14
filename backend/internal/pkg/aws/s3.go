package aws

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Client wraps the AWS S3 client with helper methods
type S3Client struct {
	Client *s3.Client
	Bucket string
	Region string
}

// NewS3Client creates a new S3 client
func NewS3Client(region, accessKeyID, secretAccessKey, bucketName string) (*S3Client, error) {
	var cfg aws.Config
	var err error

	if accessKeyID != "" && secretAccessKey != "" {
		// Use provided credentials
		cfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
		)
	} else {
		// Use default credential chain (for production deployments)
		cfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(region),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	return &S3Client{
		Client: client,
		Bucket: bucketName,
		Region: region,
	}, nil
}

// UploadFile uploads a file to S3
func (s *S3Client) UploadFile(ctx context.Context, file multipart.File, fileName, contentType string) (string, error) {
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Sanitize filename: use basename and replace spaces to prevent path traversal and collisions
	safeName := filepath.Base(fileName)
	safeName = strings.ReplaceAll(safeName, " ", "_")
	key := fmt.Sprintf("uploads/%d/%s", time.Now().UnixNano(), safeName)

	uploadParams := &s3.PutObjectInput{
		Bucket:      aws.String(s.Bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(fileBytes),
		ContentType: aws.String(contentType),
	}

	_, err = s.Client.PutObject(ctx, uploadParams)
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	// Construct the public URL for the uploaded file
	publicURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.Bucket, s.Region, key)

	return publicURL, nil
}

// UploadBytes uploads byte data to S3
func (s *S3Client) UploadBytes(ctx context.Context, data []byte, fileName, contentType string) (string, error) {
	// Sanitize filename: use basename and replace spaces to prevent path traversal and collisions
	safeName := filepath.Base(fileName)
	safeName = strings.ReplaceAll(safeName, " ", "_")
	key := fmt.Sprintf("uploads/%d/%s", time.Now().UnixNano(), safeName)

	uploadParams := &s3.PutObjectInput{
		Bucket:      aws.String(s.Bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	}

	_, err := s.Client.PutObject(ctx, uploadParams)
	if err != nil {
		return "", fmt.Errorf("failed to upload data to S3: %w", err)
	}

	// Construct the public URL for the uploaded file
	publicURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.Bucket, s.Region, key)

	return publicURL, nil
}

// DeleteFile deletes a file from S3
func (s *S3Client) DeleteFile(ctx context.Context, fileKey string) error {
	deleteParams := &s3.DeleteObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(fileKey),
	}

	_, err := s.Client.DeleteObject(ctx, deleteParams)
	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}

	return nil
}

// GeneratePresignedURL generates a presigned URL for temporary access to a file
func (s *S3Client) GeneratePresignedURL(ctx context.Context, fileKey string, duration time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.Client)

	req, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(fileKey),
	}, s3.WithPresignExpires(duration))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return req.URL, nil
}

// IsValidImageExtension checks if a file has a valid image extension
func IsValidImageExtension(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
		".webp": true,
	}

	return validExtensions[ext]
}

// IsValidImageContentType checks if a content type is a valid image type
func IsValidImageContentType(contentType string) bool {
	contentType = strings.ToLower(contentType)
	validTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true, // Support both static and animated GIFs
		"image/bmp":  true,
		"image/webp": true,
	}

	return validTypes[contentType]
}

// IsAnimatedGifExtension checks if a file extension indicates a GIF file
func IsAnimatedGifExtension(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".gif"
}

// IsAnimatedGif checks if a GIF file is animated by examining its content
func IsAnimatedGif(file multipart.File) (bool, error) {
	// Read the first 1024 bytes to check for animation markers
	buffer := make([]byte, 1024)
	_, err := file.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		return false, fmt.Errorf("failed to read file for animation check: %w", err)
	}

	// Reset file pointer to beginning
	_, err = file.Seek(0, 0)
	if err != nil {
		return false, fmt.Errorf("failed to reset file pointer: %w", err)
	}

	// Look for multiple frame markers in GIF files
	// GIF87a or GIF89a signature at start
	// Then look for multiple image descriptors (0x2C) which indicate frames
	hasMultipleFrames := false
	frameCount := 0

	// Search for image descriptor markers in the buffer using integer range (Go 1.22+)
	for i := range len(buffer) - 10 {
		if buffer[i] == 0x2C { // GIF image descriptor marker
			frameCount++
			if frameCount > 1 {
				hasMultipleFrames = true
				break
			}
		}
	}

	return hasMultipleFrames, nil
}
