package services

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// GetTemplates returns all available templates
func (s *WishListService) GetTemplates(ctx context.Context) ([]*TemplateOutput, error) {
	templates, err := s.templateRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get templates from repository: %w", err)
	}

	var outputs []*TemplateOutput
	for _, template := range templates {
		output := &TemplateOutput{
			ID:              template.ID,
			Name:            template.Name,
			Description:     template.Description.String,
			PreviewImageUrl: template.PreviewImageUrl.String,
			Config:          template.Config,
			IsDefault:       template.IsDefault.Bool,
			CreatedAt:       template.CreatedAt.Time.Format(time.RFC3339),
			UpdatedAt:       template.UpdatedAt.Time.Format(time.RFC3339),
		}
		outputs = append(outputs, output)
	}

	return outputs, nil
}

// GetDefaultTemplate returns the default template
func (s *WishListService) GetDefaultTemplate(ctx context.Context) (*TemplateOutput, error) {
	template, err := s.templateRepo.GetDefault(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get default template from repository: %w", err)
	}

	output := &TemplateOutput{
		ID:              template.ID,
		Name:            template.Name,
		Description:     template.Description.String,
		PreviewImageUrl: template.PreviewImageUrl.String,
		Config:          template.Config,
		IsDefault:       template.IsDefault.Bool,
		CreatedAt:       template.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:       template.UpdatedAt.Time.Format(time.RFC3339),
	}

	return output, nil
}

// UpdateWishListTemplate updates the template for a wish list
func (s *WishListService) UpdateWishListTemplate(ctx context.Context, wishListID, userID, templateID string) (*WishListOutput, error) {
	// Parse UUIDs
	listID := pgtype.UUID{}
	if err := listID.Scan(wishListID); err != nil {
		return nil, fmt.Errorf("invalid wishlist id: %w", err)
	}

	userIDParsed := pgtype.UUID{}
	if err := userIDParsed.Scan(userID); err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	// First, get the existing wishlist to verify ownership
	existingWishList, err := s.wishListRepo.GetByID(ctx, listID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wishlist: %w", err)
	}

	// Check if the user owns this wishlist
	if existingWishList.OwnerID != userIDParsed {
		return nil, ErrWishListForbidden
	}

	// Verify the template exists
	template, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("template not found: %w", err)
	}

	// Update the wishlist with the new template
	updatedWishList := *existingWishList
	updatedWishList.TemplateID = template.ID

	result, err := s.wishListRepo.Update(ctx, updatedWishList)
	if err != nil {
		return nil, fmt.Errorf("failed to update wishlist in repository: %w", err)
	}

	// Convert to output format
	output := &WishListOutput{
		ID:          result.ID.String(),
		OwnerID:     result.OwnerID.String(),
		Title:       result.Title,
		Description: result.Description.String,
		Occasion:    result.Occasion.String,
		TemplateID:  result.TemplateID,
		IsPublic:    result.IsPublic.Bool,
		PublicSlug:  result.PublicSlug.String,
		ViewCount:   int64(result.ViewCount.Int32),
		CreatedAt:   result.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:   result.UpdatedAt.Time.Format(time.RFC3339),
	}

	return output, nil
}
