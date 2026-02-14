package dto

// UploadImageResponse represents the response after successful image upload
type UploadImageResponse struct {
	URL string `json:"url" example:"https://s3.amazonaws.com/bucket/images/uuid.jpg" validate:"required"`
}
