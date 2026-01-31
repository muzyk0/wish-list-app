package handlers

import (
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"wish-list/internal/auth"
	"wish-list/internal/aws"

	"github.com/labstack/echo/v4"
)

type S3Handler struct {
	s3Client *aws.S3Client
}

func NewS3Handler(s3Client *aws.S3Client) *S3Handler {
	return &S3Handler{
		s3Client: s3Client,
	}
}

// UploadImage godoc
//
//	@Summary		Upload an image to S3
//	@Description	Upload an image file to S3 storage. The user must be authenticated.
//	@Tags			S3 Upload
//	@Accept			mpfd
//	@Produce		json
//	@Param			image	formData	file				true	"Image file to upload (max 10MB, only images allowed)"
//	@Success		200		{object}	map[string]string	"Image uploaded successfully, returns URL"
//	@Failure		400		{object}	map[string]string	"Invalid file or file too large"
//	@Failure		401		{object}	map[string]string	"Unauthorized"
//	@Failure		500		{object}	map[string]string	"Internal server error"
//	@Security		BearerAuth
//	@Router			/s3/upload [post]
func (h *S3Handler) UploadImage(c echo.Context) error {
	// Get user from context to ensure they're authenticated
	_, _, _, err := auth.GetUserFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}

	// Get the file from the form data
	file, err := c.FormFile("image")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to get uploaded file")
	}

	src, err := file.Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to open uploaded file")
	}
	defer src.Close()

	// Validate file type
	if !aws.IsValidImageExtension(file.Filename) || !aws.IsValidImageContentType(file.Header.Get("Content-Type")) {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid file type. Only images are allowed.")
	}

	// Validate file size (max 10MB)
	if file.Size > 10*1024*1024 { // 10MB in bytes
		return echo.NewHTTPError(http.StatusBadRequest, "File too large. Maximum size is 10MB.")
	}

	// Handle GIF file processing
	if err := h.processGifFile(src, file.Filename); err != nil {
		return err
	}

	// Upload to S3
	url, err := h.s3Client.UploadFile(c.Request().Context(), src, file.Filename, file.Header.Get("Content-Type"))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to upload image to S3")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"url": url,
	})
}

// processGifFile handles GIF-specific processing (animation check)
func (h *S3Handler) processGifFile(src multipart.File, filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".gif" {
		return nil // Not a GIF, nothing to process
	}

	isAnimated, err := aws.IsAnimatedGif(src)
	if err != nil {
		log.Printf("Warning: Could not check if GIF is animated: %v", err)
		// Reset file pointer to beginning since we read it during animation check
		if seeker, ok := src.(io.Seeker); ok {
			_, seekErr := seeker.Seek(0, 0)
			if seekErr != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process image file")
			}
		}
		// Non-fatal error, continue with upload
		return nil
	}

	if isAnimated {
		// Log that we have an animated GIF - this is allowed per FR-011
		log.Printf("Animated GIF uploaded: %s", filename)
	}

	return nil
}
