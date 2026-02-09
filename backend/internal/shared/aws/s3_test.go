package aws

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidImageExtension(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{"Valid JPG", "image.jpg", true},
		{"Valid JPEG", "image.jpeg", true},
		{"Valid PNG", "image.png", true},
		{"Valid GIF", "image.gif", true},
		{"Valid BMP", "image.bmp", true},
		{"Valid WEBP", "image.webp", true},
		{"Invalid TXT", "document.txt", false},
		{"Invalid PDF", "document.pdf", false},
		{"Case insensitive JPG", "image.JPG", true},
		{"Case insensitive jpeg", "image.JPEG", true},
		{"No extension", "image", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidImageExtension(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidImageContentType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		expected    bool
	}{
		{"Valid JPEG", "image/jpeg", true},
		{"Valid JPG", "image/jpg", true},
		{"Valid PNG", "image/png", true},
		{"Valid GIF", "image/gif", true},
		{"Valid BMP", "image/bmp", true},
		{"Valid WEBP", "image/webp", true},
		{"Invalid TXT", "text/plain", false},
		{"Invalid PDF", "application/pdf", false},
		{"Case insensitive", "IMAGE/JPEG", true}, // Our function converts to lowercase
		{"Empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidImageContentType(tt.contentType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewS3Client(t *testing.T) {
	// Note: This test will fail without AWS credentials, but we can test the error case
	region := "us-east-1"
	accessKeyID := ""
	secretAccessKey := ""
	bucketName := "test-bucket"

	client, err := NewS3Client(region, accessKeyID, secretAccessKey, bucketName)

	// Since we don't have valid AWS credentials in test environment,
	// we expect this to fail due to missing credentials
	if err != nil {
		// If there's an error, it should be related to AWS configuration
		assert.Contains(t, err.Error(), "failed to load AWS config")
	} else {
		// If there's no error, the client should be properly initialized
		assert.NotNil(t, client)
		assert.Equal(t, region, client.Region)
		assert.Equal(t, bucketName, client.Bucket)
	}
}

func TestGeneratePresignedURL(t *testing.T) {
	// This test would require a real S3 client and valid credentials
	// For unit testing purposes, we'll just verify the function signature works
	// when we have a mock client

	// Since we can't create a real S3 client without credentials in tests,
	// we'll skip this test or use a mock
	t.Skip("Skipping test that requires real S3 client")
}

func TestUploadFile(t *testing.T) {
	// This test would require a real S3 client and valid credentials
	// For unit testing purposes, we'll just verify the function signature works
	// when we have a mock client

	t.Skip("Skipping test that requires real S3 client")
}

func TestUploadBytes(t *testing.T) {
	// This test would require a real S3 client and valid credentials
	// For unit testing purposes, we'll just verify the function signature works
	// when we have a mock client

	t.Skip("Skipping test that requires real S3 client")
}

func TestDeleteFile(t *testing.T) {
	// This test would require a real S3 client and valid credentials
	// For unit testing purposes, we'll just verify the function signature works
	// when we have a mock client

	t.Skip("Skipping test that requires real S3 client")
}
