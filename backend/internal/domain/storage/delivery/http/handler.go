package http

import (
	"io"
	"mime/multipart"
	nethttp "net/http"
	"path/filepath"
	"strings"
	"wish-list/internal/domain/storage/delivery/http/dto"
	"wish-list/internal/pkg/apperrors"
	"wish-list/internal/pkg/aws"
	"wish-list/internal/pkg/logger"

	"github.com/labstack/echo/v4"
)

// Handler handles S3 storage operations
type Handler struct {
	s3Client *aws.S3Client
}

// NewHandler creates a new storage handler
func NewHandler(s3Client *aws.S3Client) *Handler {
	return &Handler{
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
//	@Success		200		{object}	dto.UploadImageResponse	"Image uploaded successfully, returns URL"
//	@Failure		400		{object}	map[string]string	"Invalid file or file too large"
//	@Failure		401		{object}	map[string]string	"Unauthorized"
//	@Failure		500		{object}	map[string]string	"Internal server error"
//	@Security		BearerAuth
//	@Router			/images/upload [post]
func (h *Handler) UploadImage(c echo.Context) error {
	// Get the file from the form data
	file, err := c.FormFile("image")
	if err != nil {
		return apperrors.BadRequest("Failed to get uploaded file")
	}

	src, err := file.Open()
	if err != nil {
		return apperrors.Internal("Failed to open uploaded file").Wrap(err)
	}
	defer src.Close()

	// Validate file type
	if !aws.IsValidImageExtension(file.Filename) || !aws.IsValidImageContentType(file.Header.Get("Content-Type")) {
		return apperrors.BadRequest("Invalid file type. Only images are allowed.")
	}

	// Validate file size (max 10MB)
	if file.Size > 10*1024*1024 { // 10MB in bytes
		return apperrors.BadRequest("File too large. Maximum size is 10MB.")
	}

	// Handle GIF file processing
	if err := h.processGifFile(src, file.Filename); err != nil {
		return err
	}

	// Upload to S3
	url, err := h.s3Client.UploadFile(c.Request().Context(), src, file.Filename, file.Header.Get("Content-Type"))
	if err != nil {
		return apperrors.Internal("Failed to upload image to S3").Wrap(err)
	}

	return c.JSON(nethttp.StatusOK, dto.UploadImageResponse{
		URL: url,
	})
}

// processGifFile handles GIF-specific processing (animation check)
func (h *Handler) processGifFile(src multipart.File, filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".gif" {
		return nil // Not a GIF, nothing to process
	}

	isAnimated, err := aws.IsAnimatedGif(src)
	if err != nil {
		logger.Warn("could not check if GIF is animated", "error", err, "filename", filename)
		// Reset file pointer to beginning since we read it during animation check
		if seeker, ok := src.(io.Seeker); ok {
			_, seekErr := seeker.Seek(0, 0)
			if seekErr != nil {
				return apperrors.Internal("Failed to process image file").Wrap(seekErr)
			}
		}
		// Non-fatal error, continue with upload
		return nil
	}

	if isAnimated {
		// Log that we have an animated GIF - this is allowed per FR-011
		logger.Info("animated GIF uploaded", "filename", filename)
	}

	return nil
}
