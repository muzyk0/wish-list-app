package storage

import (
	"wish-list/internal/domains/storage/handlers"
	"wish-list/internal/pkg/aws"
)

// NewS3Handler creates a new S3 handler for file upload endpoints
func NewS3Handler(s3Client *aws.S3Client) *handlers.S3Handler {
	return handlers.NewS3Handler(s3Client)
}
