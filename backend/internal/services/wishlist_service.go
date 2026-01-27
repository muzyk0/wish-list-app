package services

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"
	db "wish-list/internal/db/models"

	"wish-list/internal/cache"
	"wish-list/internal/repositories"

	"github.com/jackc/pgx/v5/pgtype"
)

// Sentinel errors
var (
	ErrWishListNotFound  = errors.New("wishlist not found")
	ErrWishListForbidden = errors.New("not authorized to access this wishlist")
)

// WishListServiceInterface defines the interface for wishlist-related operations
type WishListServiceInterface interface {
	CreateWishList(ctx context.Context, userID string, input CreateWishListInput) (*WishListOutput, error)
	GetWishList(ctx context.Context, wishListID string) (*WishListOutput, error)
	GetWishListByPublicSlug(ctx context.Context, publicSlug string) (*WishListOutput, error)
	GetWishListsByOwner(ctx context.Context, userID string) ([]*WishListOutput, error)
	UpdateWishList(ctx context.Context, wishListID, userID string, input UpdateWishListInput) (*WishListOutput, error)
	DeleteWishList(ctx context.Context, wishListID, userID string) error
	CreateGiftItem(ctx context.Context, wishListID string, input CreateGiftItemInput) (*GiftItemOutput, error)
	GetGiftItem(ctx context.Context, giftItemID string) (*GiftItemOutput, error)
	GetGiftItemsByWishList(ctx context.Context, wishListID string) ([]*GiftItemOutput, error)
	UpdateGiftItem(ctx context.Context, giftItemID string, input UpdateGiftItemInput) (*GiftItemOutput, error)
	DeleteGiftItem(ctx context.Context, giftItemID string) error
	MarkGiftItemAsPurchased(ctx context.Context, giftItemID, userID string, purchasedPrice float64) (*GiftItemOutput, error)
	GetTemplates(ctx context.Context) ([]*TemplateOutput, error)
	GetDefaultTemplate(ctx context.Context) (*TemplateOutput, error)
	UpdateWishListTemplate(ctx context.Context, wishListID, userID, templateID string) (*WishListOutput, error)
}

type WishListService struct {
	wishListRepo    repositories.WishListRepositoryInterface
	giftItemRepo    repositories.GiftItemRepositoryInterface
	templateRepo    repositories.TemplateRepositoryInterface
	emailService    EmailServiceInterface
	reservationRepo repositories.ReservationRepositoryInterface
	cache           cache.CacheInterface
}

func NewWishListService(
	wishListRepo repositories.WishListRepositoryInterface,
	giftItemRepo repositories.GiftItemRepositoryInterface,
	templateRepo repositories.TemplateRepositoryInterface,
	emailService EmailServiceInterface,
	reservationRepo repositories.ReservationRepositoryInterface,
	cacheService cache.CacheInterface,
) *WishListService {
	return &WishListService{
		wishListRepo:    wishListRepo,
		giftItemRepo:    giftItemRepo,
		templateRepo:    templateRepo,
		emailService:    emailService,
		reservationRepo: reservationRepo,
		cache:           cacheService,
	}
}

type CreateWishListInput struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	Occasion     string `json:"occasion"`
	OccasionDate string `json:"occasion_date"`
	TemplateID   string `json:"template_id"`
	IsPublic     bool   `json:"is_public"`
}

type UpdateWishListInput struct {
	Title        *string `json:"title"`
	Description  *string `json:"description"`
	Occasion     *string `json:"occasion"`
	OccasionDate *string `json:"occasion_date"`
	TemplateID   *string `json:"template_id"`
	IsPublic     *bool   `json:"is_public"`
}

type WishListOutput struct {
	ID           string `json:"id"`
	OwnerID      string `json:"owner_id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Occasion     string `json:"occasion"`
	OccasionDate string `json:"occasion_date"`
	TemplateID   string `json:"template_id"`
	IsPublic     bool   `json:"is_public"`
	PublicSlug   string `json:"public_slug"`
	ViewCount    int64  `json:"view_count"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type CreateGiftItemInput struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Link        string  `json:"link"`
	ImageURL    string  `json:"image_url"`
	Price       float64 `json:"price"`
	Priority    int     `json:"priority"`
	Notes       string  `json:"notes"`
	Position    int     `json:"position"`
}

type UpdateGiftItemInput struct {
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Link        *string  `json:"link,omitempty"`
	ImageURL    *string  `json:"image_url,omitempty"`
	Price       *float64 `json:"price,omitempty"`
	Priority    *int     `json:"priority,omitempty"`
	Notes       *string  `json:"notes,omitempty"`
	Position    *int     `json:"position,omitempty"`
}

type GiftItemOutput struct {
	ID                string  `json:"id"`
	WishlistID        string  `json:"wishlist_id"`
	Name              string  `json:"name"`
	Description       string  `json:"description"`
	Link              string  `json:"link"`
	ImageURL          string  `json:"image_url"`
	Price             float64 `json:"price"`
	Priority          int     `json:"priority"`
	ReservedByUserID  string  `json:"reserved_by_user_id"`
	ReservedAt        string  `json:"reserved_at"`
	PurchasedByUserID string  `json:"purchased_by_user_id"`
	PurchasedAt       string  `json:"purchased_at"`
	PurchasedPrice    float64 `json:"purchased_price"`
	Notes             string  `json:"notes"`
	Position          int     `json:"position"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
}

type TemplateOutput struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	PreviewImageUrl string `json:"preview_image_url"`
	Config          []byte `json:"config"` // JSONB stored as bytes
	IsDefault       bool   `json:"is_default"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

func (s *WishListService) CreateWishList(ctx context.Context, userID string, input CreateWishListInput) (*WishListOutput, error) {
	// Validate input
	if input.Title == "" {
		return nil, errors.New("title is required")
	}

	// Parse user ID
	ownerID := pgtype.UUID{}
	if err := ownerID.Scan(userID); err != nil {
		return nil, errors.New("invalid user id")
	}

	// Generate public slug if public
	var publicSlug pgtype.Text
	if input.IsPublic {
		publicSlug = pgtype.Text{
			String: generatePublicSlug(input.Title),
			Valid:  true,
		}
	} else {
		publicSlug = pgtype.Text{Valid: false}
	}

	// Parse OccasionDate if provided
	var occasionDate pgtype.Date
	if input.OccasionDate != "" {
		if parsedDate, err := time.Parse(time.RFC3339, input.OccasionDate); err == nil {
			occasionDate = pgtype.Date{
				Time:  parsedDate,
				Valid: true,
			}
		} else {
			occasionDate = pgtype.Date{Valid: false}
		}
	} else {
		occasionDate = pgtype.Date{Valid: false}
	}

	// Create wishlist
	wishList := db.WishList{
		OwnerID:      ownerID,
		Title:        input.Title,
		Description:  pgtype.Text{String: input.Description, Valid: input.Description != ""},
		Occasion:     pgtype.Text{String: input.Occasion, Valid: input.Occasion != ""},
		OccasionDate: occasionDate,
		TemplateID:   input.TemplateID,
		IsPublic:     pgtype.Bool{Bool: input.IsPublic, Valid: true},
		PublicSlug:   publicSlug,
	}

	createdWishList, err := s.wishListRepo.Create(ctx, wishList)
	if err != nil {
		return nil, fmt.Errorf("failed to create wishlist in repository: %w", err)
	}

	output := &WishListOutput{
		ID:           createdWishList.ID.String(),
		OwnerID:      createdWishList.OwnerID.String(),
		Title:        createdWishList.Title,
		Description:  createdWishList.Description.String,
		Occasion:     createdWishList.Occasion.String,
		OccasionDate: createdWishList.OccasionDate.Time.Format(time.RFC3339),
		TemplateID:   createdWishList.TemplateID,
		IsPublic:     createdWishList.IsPublic.Bool,
		PublicSlug:   createdWishList.PublicSlug.String,
		ViewCount:    int64(createdWishList.ViewCount.Int32),
		CreatedAt:    createdWishList.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:    createdWishList.UpdatedAt.Time.Format(time.RFC3339),
	}

	return output, nil
}

func (s *WishListService) GetWishList(ctx context.Context, wishListID string) (*WishListOutput, error) {
	id := pgtype.UUID{}
	if err := id.Scan(wishListID); err != nil {
		return nil, errors.New("invalid wishlist id")
	}

	wishList, err := s.wishListRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get wishlist from repository: %w", err)
	}

	output := &WishListOutput{
		ID:           wishList.ID.String(),
		OwnerID:      wishList.OwnerID.String(),
		Title:        wishList.Title,
		Description:  wishList.Description.String,
		Occasion:     wishList.Occasion.String,
		OccasionDate: wishList.OccasionDate.Time.Format(time.RFC3339),
		TemplateID:   wishList.TemplateID,
		IsPublic:     wishList.IsPublic.Bool,
		PublicSlug:   wishList.PublicSlug.String,
		ViewCount:    int64(wishList.ViewCount.Int32),
		CreatedAt:    wishList.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:    wishList.UpdatedAt.Time.Format(time.RFC3339),
	}

	return output, nil
}

func (s *WishListService) GetWishListByPublicSlug(ctx context.Context, publicSlug string) (*WishListOutput, error) {
	// Try to get from cache if cache is available
	if s.cache != nil {
		cacheKey := fmt.Sprintf("wishlist:public:%s", publicSlug)
		var cached WishListOutput
		if err := s.cache.Get(ctx, cacheKey, &cached); err == nil {
			return &cached, nil
		}
	}

	wishList, err := s.wishListRepo.GetByPublicSlug(ctx, publicSlug)
	if err != nil {
		return nil, fmt.Errorf("failed to get wishlist by public slug from repository: %w", err)
	}

	output := &WishListOutput{
		ID:           wishList.ID.String(),
		OwnerID:      wishList.OwnerID.String(),
		Title:        wishList.Title,
		Description:  wishList.Description.String,
		Occasion:     wishList.Occasion.String,
		OccasionDate: wishList.OccasionDate.Time.Format(time.RFC3339),
		TemplateID:   wishList.TemplateID,
		IsPublic:     wishList.IsPublic.Bool,
		PublicSlug:   wishList.PublicSlug.String,
		ViewCount:    int64(wishList.ViewCount.Int32),
		CreatedAt:    wishList.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:    wishList.UpdatedAt.Time.Format(time.RFC3339),
	}

	// Store in cache if cache is available
	if s.cache != nil {
		cacheKey := fmt.Sprintf("wishlist:public:%s", publicSlug)
		_ = s.cache.Set(ctx, cacheKey, output)
	}

	return output, nil
}

func (s *WishListService) GetWishListsByOwner(ctx context.Context, userID string) ([]*WishListOutput, error) {
	id := pgtype.UUID{}
	if err := id.Scan(userID); err != nil {
		return nil, errors.New("invalid user id")
	}

	wishLists, err := s.wishListRepo.GetByOwner(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get wish lists by owner from repository: %w", err)
	}

	var outputs []*WishListOutput
	for _, wishList := range wishLists {
		output := &WishListOutput{
			ID:           wishList.ID.String(),
			OwnerID:      wishList.OwnerID.String(),
			Title:        wishList.Title,
			Description:  wishList.Description.String,
			Occasion:     wishList.Occasion.String,
			OccasionDate: wishList.OccasionDate.Time.Format(time.RFC3339),
			TemplateID:   wishList.TemplateID,
			IsPublic:     wishList.IsPublic.Bool,
			PublicSlug:   wishList.PublicSlug.String,
			ViewCount:    int64(wishList.ViewCount.Int32),
			CreatedAt:    wishList.CreatedAt.Time.Format(time.RFC3339),
			UpdatedAt:    wishList.UpdatedAt.Time.Format(time.RFC3339),
		}
		outputs = append(outputs, output)
	}

	return outputs, nil
}

func (s *WishListService) UpdateWishList(ctx context.Context, wishListID, userID string, input UpdateWishListInput) (*WishListOutput, error) {
	id := pgtype.UUID{}
	if err := id.Scan(wishListID); err != nil {
		return nil, errors.New("invalid wishlist id")
	}

	// Verify ownership
	wishList, err := s.wishListRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrWishListNotFound, err)
	}

	ownerID := pgtype.UUID{}
	if err := ownerID.Scan(userID); err != nil {
		return nil, errors.New("invalid user id")
	}

	if wishList.OwnerID != ownerID {
		return nil, ErrWishListForbidden
	}

	// Update wishlist - only update fields that are provided in the input
	updatedWishList := *wishList

	if input.Title != nil {
		updatedWishList.Title = *input.Title
	}

	if input.Description != nil {
		updatedWishList.Description = pgtype.Text{String: *input.Description, Valid: *input.Description != ""}
	} else if input.Description == nil {
		// Keep the original description if not provided
		updatedWishList.Description = wishList.Description
	}

	if input.Occasion != nil {
		updatedWishList.Occasion = pgtype.Text{String: *input.Occasion, Valid: *input.Occasion != ""}
	} else if input.Occasion == nil {
		// Keep the original occasion if not provided
		updatedWishList.Occasion = wishList.Occasion
	}

	if input.TemplateID != nil {
		updatedWishList.TemplateID = *input.TemplateID
	} else if input.TemplateID == nil {
		// Keep the original template ID if not provided
		updatedWishList.TemplateID = wishList.TemplateID
	}

	if input.IsPublic != nil {
		updatedWishList.IsPublic = pgtype.Bool{Bool: *input.IsPublic, Valid: true}
	} else if input.IsPublic == nil {
		// Keep the original is_public value if not provided
		updatedWishList.IsPublic = wishList.IsPublic
	}

	if input.OccasionDate != nil {
		// Parse the date string to pgtype.Date
		if parsedDate, err := time.Parse(time.RFC3339, *input.OccasionDate); err == nil {
			updatedWishList.OccasionDate = pgtype.Date{
				Time:  parsedDate,
				Valid: true,
			}
		} else {
			// If parsing fails, keep the original date
			updatedWishList.OccasionDate = wishList.OccasionDate
		}
	} else if input.OccasionDate == nil {
		// Keep the original occasion date if not provided
		updatedWishList.OccasionDate = wishList.OccasionDate
	}

	// Generate public slug if making the list public and no slug exists
	currentIsPublic := input.IsPublic != nil && *input.IsPublic
	if currentIsPublic && !updatedWishList.PublicSlug.Valid {
		titleToUse := updatedWishList.Title // Use the updated title
		if input.Title != nil {
			titleToUse = *input.Title
		}
		updatedWishList.PublicSlug = pgtype.Text{
			String: generatePublicSlug(titleToUse),
			Valid:  true,
		}
	}

	updated, err := s.wishListRepo.Update(ctx, updatedWishList)
	if err != nil {
		return nil, fmt.Errorf("failed to update wishlist in repository: %w", err)
	}

	// Invalidate cache if cache is available
	if s.cache != nil && updated.PublicSlug.Valid {
		cacheKey := fmt.Sprintf("wishlist:public:%s", updated.PublicSlug.String)
		_ = s.cache.Delete(ctx, cacheKey)
	}

	output := &WishListOutput{
		ID:          updated.ID.String(),
		OwnerID:     updated.OwnerID.String(),
		Title:       updated.Title,
		Description: updated.Description.String,
		Occasion:    updated.Occasion.String,
		TemplateID:  updated.TemplateID,
		IsPublic:    updated.IsPublic.Bool,
		PublicSlug:  updated.PublicSlug.String,
		ViewCount:   int64(updated.ViewCount.Int32),
		CreatedAt:   updated.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:   updated.UpdatedAt.Time.Format(time.RFC3339),
	}

	return output, nil
}

func (s *WishListService) DeleteWishList(ctx context.Context, wishListID, userID string) error {
	id := pgtype.UUID{}
	if err := id.Scan(wishListID); err != nil {
		return errors.New("invalid wishlist id")
	}

	// Verify ownership
	wishList, err := s.wishListRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get wishlist from repository: %w", err)
	}

	ownerID := pgtype.UUID{}
	if err := ownerID.Scan(userID); err != nil {
		return errors.New("invalid user id")
	}

	if wishList.OwnerID != ownerID {
		return ErrWishListForbidden
	}

	// Check if there are any active reservations for gift items in this wishlist
	giftItems, err := s.giftItemRepo.GetByWishList(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check for active reservations: %w", err)
	}

	hasActiveReservations := false
	for _, item := range giftItems {
		// Check if this gift item has any active reservations
		reservation, err := s.reservationRepo.GetActiveReservationForGiftItem(ctx, item.ID)
		if err == nil && reservation != nil {
			hasActiveReservations = true
			break
		}
	}

	if hasActiveReservations {
		return errors.New("cannot delete wishlist with active reservations - please remove or cancel all reservations first")
	}

	// Invalidate cache if cache is available
	if s.cache != nil && wishList.PublicSlug.Valid {
		cacheKey := fmt.Sprintf("wishlist:public:%s", wishList.PublicSlug.String)
		_ = s.cache.Delete(ctx, cacheKey)
	}

	return s.wishListRepo.Delete(ctx, id)
}

func (s *WishListService) CreateGiftItem(ctx context.Context, wishListID string, input CreateGiftItemInput) (*GiftItemOutput, error) {
	// Validate input
	if input.Name == "" {
		return nil, errors.New("name is required")
	}

	// Validate int32 bounds for Priority and Position
	if input.Priority < math.MinInt32 || input.Priority > math.MaxInt32 {
		return nil, errors.New("priority value out of int32 range")
	}
	if input.Position < math.MinInt32 || input.Position > math.MaxInt32 {
		return nil, errors.New("position value out of int32 range")
	}

	// Parse wishlist ID
	listID := pgtype.UUID{}
	if err := listID.Scan(wishListID); err != nil {
		return nil, errors.New("invalid wishlist id")
	}

	// Create price numeric
	priceBig := new(big.Int)
	priceBig.SetInt64(int64(input.Price * 100)) // Convert to cents

	// Create gift item
	giftItem := db.GiftItem{
		WishlistID:  listID,
		Name:        input.Name,
		Description: pgtype.Text{String: input.Description, Valid: input.Description != ""},
		Link:        pgtype.Text{String: input.Link, Valid: input.Link != ""},
		ImageUrl:    pgtype.Text{String: input.ImageURL, Valid: input.ImageURL != ""},
		Price:       pgtype.Numeric{Int: priceBig, Exp: -2, Valid: input.Price > 0},
		Priority:    pgtype.Int4{Int32: int32(input.Priority), Valid: true}, //nolint:gosec // Bounds checking performed above, conversion is saf
		Notes:       pgtype.Text{String: input.Notes, Valid: input.Notes != ""},
		Position:    pgtype.Int4{Int32: int32(input.Position), Valid: true}, //nolint:gosec // Bounds checking performed above, conversion is saf
	}

	createdGiftItem, err := s.giftItemRepo.Create(ctx, giftItem)
	if err != nil {
		return nil, fmt.Errorf("failed to create gift item in repository: %w", err)
	}

	// Convert price to float64
	var price float64
	if createdGiftItem.Price.Valid {
		priceValue, err := createdGiftItem.Price.Float64Value()
		if err == nil && priceValue.Valid {
			price = priceValue.Float64
		}
	}

	output := &GiftItemOutput{
		ID:          createdGiftItem.ID.String(),
		WishlistID:  createdGiftItem.WishlistID.String(),
		Name:        createdGiftItem.Name,
		Description: createdGiftItem.Description.String,
		Link:        createdGiftItem.Link.String,
		ImageURL:    createdGiftItem.ImageUrl.String,
		Price:       price,
		Priority:    int(createdGiftItem.Priority.Int32),
		Notes:       createdGiftItem.Notes.String,
		Position:    int(createdGiftItem.Position.Int32),
		CreatedAt:   createdGiftItem.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:   createdGiftItem.UpdatedAt.Time.Format(time.RFC3339),
	}

	return output, nil
}

func (s *WishListService) GetGiftItem(ctx context.Context, giftItemID string) (*GiftItemOutput, error) {
	id := pgtype.UUID{}
	if err := id.Scan(giftItemID); err != nil {
		return nil, errors.New("invalid gift item id")
	}

	giftItem, err := s.giftItemRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get gift item from repository: %w", err)
	}

	// Convert price to float64
	var price float64
	if giftItem.Price.Valid {
		priceValue, err := giftItem.Price.Float64Value()
		if err == nil && priceValue.Valid {
			price = priceValue.Float64
		}
	}

	output := &GiftItemOutput{
		ID:          giftItem.ID.String(),
		WishlistID:  giftItem.WishlistID.String(),
		Name:        giftItem.Name,
		Description: giftItem.Description.String,
		Link:        giftItem.Link.String,
		ImageURL:    giftItem.ImageUrl.String,
		Price:       price,
		Priority:    int(giftItem.Priority.Int32),
		Notes:       giftItem.Notes.String,
		Position:    int(giftItem.Position.Int32),
		CreatedAt:   giftItem.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:   giftItem.UpdatedAt.Time.Format(time.RFC3339),
	}

	return output, nil
}

func (s *WishListService) GetGiftItemsByWishList(ctx context.Context, wishListID string) ([]*GiftItemOutput, error) {
	listID := pgtype.UUID{}
	if err := listID.Scan(wishListID); err != nil {
		return nil, errors.New("invalid wishlist id")
	}

	giftItems, err := s.giftItemRepo.GetByWishList(ctx, listID)
	if err != nil {
		return nil, fmt.Errorf("failed to get gift items from repository: %w", err)
	}

	var outputs []*GiftItemOutput

	for _, giftItem := range giftItems {
		if giftItem == nil {
			continue // Skip nil items to avoid panic
		}

		// Convert price to float64
		var price float64
		if giftItem.Price.Valid {
			priceValue, err := giftItem.Price.Float64Value()
			if err == nil && priceValue.Valid {
				price = priceValue.Float64
			}
		}

		output := &GiftItemOutput{
			ID:          giftItem.ID.String(),
			WishlistID:  giftItem.WishlistID.String(),
			Name:        giftItem.Name,
			Description: giftItem.Description.String,
			Link:        giftItem.Link.String,
			ImageURL:    giftItem.ImageUrl.String,
			Price:       price,
			Priority:    int(giftItem.Priority.Int32),
			Notes:       giftItem.Notes.String,
			Position:    int(giftItem.Position.Int32),
			CreatedAt:   giftItem.CreatedAt.Time.Format(time.RFC3339),
			UpdatedAt:   giftItem.UpdatedAt.Time.Format(time.RFC3339),
		}
		outputs = append(outputs, output)
	}

	return outputs, nil
}

func (s *WishListService) UpdateGiftItem(ctx context.Context, giftItemID string, input UpdateGiftItemInput) (*GiftItemOutput, error) {
	// Validate int32 bounds for Priority and Position if provided
	if input.Priority != nil {
		if *input.Priority < math.MinInt32 || *input.Priority > math.MaxInt32 {
			return nil, errors.New("priority value out of int32 range")
		}
	}
	if input.Position != nil {
		if *input.Position < math.MinInt32 || *input.Position > math.MaxInt32 {
			return nil, errors.New("position value out of int32 range")
		}
	}

	id := pgtype.UUID{}
	if err := id.Scan(giftItemID); err != nil {
		return nil, errors.New("invalid gift item id")
	}

	giftItem, err := s.giftItemRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get gift item from repository: %w", err)
	}

	// Update gift item - only update fields that are provided (non-nil)
	updatedGiftItem := *giftItem
	if input.Name != nil {
		updatedGiftItem.Name = *input.Name
	}
	if input.Description != nil {
		updatedGiftItem.Description = pgtype.Text{String: *input.Description, Valid: *input.Description != ""}
	}
	if input.Link != nil {
		updatedGiftItem.Link = pgtype.Text{String: *input.Link, Valid: *input.Link != ""}
	}
	if input.ImageURL != nil {
		updatedGiftItem.ImageUrl = pgtype.Text{String: *input.ImageURL, Valid: *input.ImageURL != ""}
	}
	if input.Price != nil {
		priceBig := new(big.Int)
		priceBig.SetInt64(int64(*input.Price * 100)) // Convert to cents
		updatedGiftItem.Price = pgtype.Numeric{Int: priceBig, Exp: -2, Valid: *input.Price > 0}
	}
	if input.Priority != nil {
		updatedGiftItem.Priority = pgtype.Int4{Int32: int32(*input.Priority), Valid: true} //nolint:gosec // Bounds checking performed above, conversion is safe
	}
	if input.Notes != nil {
		updatedGiftItem.Notes = pgtype.Text{String: *input.Notes, Valid: *input.Notes != ""}
	}
	if input.Position != nil {
		updatedGiftItem.Position = pgtype.Int4{Int32: int32(*input.Position), Valid: true} //nolint:gosec // Bounds checking performed above, conversion is safe
	}

	updated, err := s.giftItemRepo.Update(ctx, updatedGiftItem)
	if err != nil {
		return nil, fmt.Errorf("failed to update gift item in repository: %w", err)
	}

	// Invalidate wishlist cache if cache is available
	if s.cache != nil {
		wishList, err := s.wishListRepo.GetByID(ctx, updated.WishlistID)
		if err == nil && wishList.PublicSlug.Valid {
			cacheKey := fmt.Sprintf("wishlist:public:%s", wishList.PublicSlug.String)
			_ = s.cache.Delete(ctx, cacheKey)
		}
	}

	// Convert price to float64
	var price float64
	if updated.Price.Valid {
		priceValue, err := updated.Price.Float64Value()
		if err == nil && priceValue.Valid {
			price = priceValue.Float64
		}
	}

	output := &GiftItemOutput{
		ID:          updated.ID.String(),
		WishlistID:  updated.WishlistID.String(),
		Name:        updated.Name,
		Description: updated.Description.String,
		Link:        updated.Link.String,
		ImageURL:    updated.ImageUrl.String,
		Price:       price,
		Priority:    int(updated.Priority.Int32),
		Notes:       updated.Notes.String,
		Position:    int(updated.Position.Int32),
		CreatedAt:   updated.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:   updated.UpdatedAt.Time.Format(time.RFC3339),
	}

	return output, nil
}

func (s *WishListService) DeleteGiftItem(ctx context.Context, giftItemID string) error {
	id := pgtype.UUID{}
	if err := id.Scan(giftItemID); err != nil {
		return errors.New("invalid gift item id")
	}

	// Get gift item before deletion to get wishlist ID for cache invalidation
	giftItemForCache, err := s.giftItemRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get gift item from repository: %w", err)
	}

	// Delete the gift item and get any active reservations for notification purposes
	activeReservations, err := s.giftItemRepo.DeleteWithReservationNotification(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete gift item in repository: %w", err)
	}

	// Invalidate wishlist cache if cache is available
	if s.cache != nil {
		wishList, err := s.wishListRepo.GetByID(ctx, giftItemForCache.WishlistID)
		if err == nil && wishList.PublicSlug.Valid {
			cacheKey := fmt.Sprintf("wishlist:public:%s", wishList.PublicSlug.String)
			_ = s.cache.Delete(ctx, cacheKey)
		}
	}

	// If there were active reservations, send notifications to the reservation holders
	if len(activeReservations) > 0 {
		// Get the wish list details using the gift item we fetched before deletion
		wishList, err := s.wishListRepo.GetByID(ctx, giftItemForCache.WishlistID)
		if err != nil {
			// Log the error but continue with the notifications
			fmt.Printf("Warning: failed to get wish list details for notification: %v\n", err)
		} else {
			// Send notification emails to all reservation holders
			for _, reservation := range activeReservations {
				var recipientEmail string
				if reservation.GuestEmail.Valid {
					recipientEmail = reservation.GuestEmail.String
				} else if reservation.ReservedByUserID.Valid {
					// For authenticated users, we would need to fetch their email
					// For now, we'll skip sending to authenticated users in this implementation
					continue
				}

				if recipientEmail != "" {
					err := s.emailService.SendReservationRemovedEmail(ctx, recipientEmail, giftItemForCache.Name, wishList.Title)
					if err != nil {
						// Log the error but don't fail the deletion
						fmt.Printf("Warning: failed to send reservation removal notification: %v\n", err)
					}
				}
			}
		}
	}

	return nil
}

// MarkGiftItemAsPurchased marks a gift item as purchased
// MarkGiftItemAsPurchased marks a gift item as purchased
func (s *WishListService) MarkGiftItemAsPurchased(ctx context.Context, giftItemID, userID string, purchasedPrice float64) (*GiftItemOutput, error) {
	// Validate input
	if giftItemID == "" {
		return nil, errors.New("gift item ID is required")
	}
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	// Parse UUIDs
	itemID := pgtype.UUID{}
	if err := itemID.Scan(giftItemID); err != nil {
		return nil, fmt.Errorf("invalid gift item id: %w", err)
	}

	userUUID := pgtype.UUID{}
	if err := userUUID.Scan(userID); err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	// Create price value
	priceValue := pgtype.Numeric{}
	if err := priceValue.Scan(purchasedPrice); err != nil {
		return nil, fmt.Errorf("invalid price value: %w", err)
	}

	// Mark as purchased in repository
	updatedGiftItem, err := s.giftItemRepo.MarkAsPurchased(ctx, itemID, userUUID, priceValue)
	if err != nil {
		return nil, fmt.Errorf("failed to mark gift item as purchased in repository: %w", err)
	}

	// Send email notification to the person who reserved the gift
	if s.emailService != nil {
		// Check if there's an active reservation for this gift item
		reservation, err := s.reservationRepo.GetActiveReservationForGiftItem(ctx, updatedGiftItem.ID)
		if err == nil && reservation != nil {
			// Get the wishlist details for the email
			wishList, err := s.wishListRepo.GetByID(ctx, updatedGiftItem.WishlistID)
			if err == nil {
				var recipientEmail, guestName string

				if reservation.GuestEmail.Valid {
					recipientEmail = reservation.GuestEmail.String
					guestName = reservation.GuestName.String
				} else if reservation.ReservedByUserID.Valid {
					// For authenticated users, we would need to fetch their email from user repository
					// Skipping for now as per implementation
				}

				if recipientEmail != "" {
					err := s.emailService.SendGiftPurchasedConfirmationEmail(ctx, recipientEmail, updatedGiftItem.Name, wishList.Title, guestName)
					if err != nil {
						// Log the error but don't fail the purchase marking
						fmt.Printf("Warning: failed to send gift purchased notification: %v\n", err)
					}
				}
			}
		}
	}

	// Invalidate wishlist cache if cache is available
	if s.cache != nil {
		wishList, err := s.wishListRepo.GetByID(ctx, updatedGiftItem.WishlistID)
		if err == nil && wishList.PublicSlug.Valid {
			cacheKey := fmt.Sprintf("wishlist:public:%s", wishList.PublicSlug.String)
			_ = s.cache.Delete(ctx, cacheKey)
		}
	}

	// Convert to output format
	output := &GiftItemOutput{
		ID:                updatedGiftItem.ID.String(),
		WishlistID:        updatedGiftItem.WishlistID.String(),
		Name:              updatedGiftItem.Name,
		Description:       updatedGiftItem.Description.String,
		Link:              updatedGiftItem.Link.String,
		ImageURL:          updatedGiftItem.ImageUrl.String,
		Price:             db.NumericToFloat64(updatedGiftItem.Price),
		Priority:          int(updatedGiftItem.Priority.Int32),
		ReservedByUserID:  updatedGiftItem.ReservedByUserID.String(),
		ReservedAt:        updatedGiftItem.ReservedAt.Time.Format(time.RFC3339),
		PurchasedByUserID: updatedGiftItem.PurchasedByUserID.String(),
		PurchasedAt:       updatedGiftItem.PurchasedAt.Time.Format(time.RFC3339),
		PurchasedPrice:    db.NumericToFloat64(updatedGiftItem.PurchasedPrice),
		Notes:             updatedGiftItem.Notes.String,
		Position:          int(updatedGiftItem.Position.Int32),
		CreatedAt:         updatedGiftItem.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:         updatedGiftItem.UpdatedAt.Time.Format(time.RFC3339),
	}

	return output, nil
}

// Helper function to generate a public slug from title
func generatePublicSlug(title string) string {
	// 1. Initial cleanup: lowercasing and replacing spaces
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")

	// 2. Refactored Loop: Using strings.Builder to fix efficiency (Modernize)
	var sb strings.Builder
	sb.Grow(len(slug)) // Optimization: Allocate memory once

	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			sb.WriteRune(r) // Efficiently builds the string
		}
	}
	cleanSlug := sb.String()

	// 3. Unique Suffix Generation using crypto/rand
	randomMax := big.NewInt(10000)
	n, err := rand.Int(rand.Reader, randomMax)

	var suffix string
	if err != nil {
		suffix = "-0000" // Safe fallback
	} else {
		suffix = fmt.Sprintf("-%04d", n.Int64())
	}

	return cleanSlug + suffix
}

//// GetTemplates returns all available templates
//func (s *WishListService) GetTemplates(ctx context.Context) ([]*TemplateOutput, error) {
//	templates, err := s.templateRepo.GetAll(ctx)
//	if err != nil {
//		return nil, fmt.Errorf("failed to get templates from repository: %w", err)
//	}
//
//	var outputs []*TemplateOutput
//	for _, template := range templates {
//		output := &TemplateOutput{
//			ID:              template.ID,
//			Name:            template.Name,
//			Description:     template.Description.String,
//			PreviewImageUrl: template.PreviewImageUrl.String,
//			Config:          template.Config,
//			IsDefault:       template.IsDefault.Bool,
//			CreatedAt:       template.CreatedAt.Time.Format(time.RFC3339),
//			UpdatedAt:       template.UpdatedAt.Time.Format(time.RFC3339),
//		}
//		outputs = append(outputs, output)
//	}
//
//	return outputs, nil
//}
//
//// GetDefaultTemplate returns the default template
//func (s *WishListService) GetDefaultTemplate(ctx context.Context) (*TemplateOutput, error) {
//	template, err := s.templateRepo.GetDefault(ctx)
//	if err != nil {
//		return nil, fmt.Errorf("failed to get default template from repository: %w", err)
//	}
//
//	output := &TemplateOutput{
//		ID:              template.ID,
//		Name:            template.Name,
//		Description:     template.Description.String,
//		PreviewImageUrl: template.PreviewImageUrl.String,
//		Config:          template.Config,
//		IsDefault:       template.IsDefault.Bool,
//		CreatedAt:       template.CreatedAt.Time.Format(time.RFC3339),
//		UpdatedAt:       template.UpdatedAt.Time.Format(time.RFC3339),
//	}
//
//	return output, nil
//}
//
//// UpdateWishListTemplate updates the template for a wish list
//func (s *WishListService) UpdateWishListTemplate(ctx context.Context, wishListID, userID, templateID string) (*WishListOutput, error) {
//	// Parse UUIDs
//	listID := pgtype.UUID{}
//	if err := listID.Scan(wishListID); err != nil {
//		return nil, fmt.Errorf("invalid wishlist id: %w", err)
//	}
//
//	userIDParsed := pgtype.UUID{}
//	if err := userIDParsed.Scan(userID); err != nil {
//		return nil, fmt.Errorf("invalid user id: %w", err)
//	}
//
//	// First, get the existing wishlist to verify ownership
//	existingWishList, err := s.wishListRepo.GetByID(ctx, listID)
//	if err != nil {
//		return nil, fmt.Errorf("failed to get wishlist: %w", err)
//	}
//
//	// Check if the user owns this wishlist
//	if existingWishList.OwnerID != userIDParsed {
//		return nil, fmt.Errorf("not authorized to update this wishlist")
//	}
//
//	// Verify the template exists
//	template, err := s.templateRepo.GetByID(ctx, templateID)
//	if err != nil {
//		return nil, fmt.Errorf("template not found: %w", err)
//	}
//
//	// Update the wishlist with the new template
//	updatedWishList := *existingWishList
//	updatedWishList.TemplateID = template.ID
//
//	result, err := s.wishListRepo.Update(ctx, updatedWishList)
//	if err != nil {
//		return nil, fmt.Errorf("failed to update wishlist in repository: %w", err)
//	}
//
//	// Convert to output format
//	output := &WishListOutput{
//		ID:           result.ID.String(),
//		OwnerID:      result.OwnerID.String(),
//		Title:        result.Title,
//		Description:  result.Description.String,
//		Occasion:     result.Occasion.String,
//		OccasionDate: result.OccasionDate.Time.Format(time.RFC3339),
//		TemplateID:   result.TemplateID,
//		IsPublic:     result.IsPublic.Bool,
//		PublicSlug:   result.PublicSlug.String,
//		ViewCount:    int64(result.ViewCount.Int32),
//		CreatedAt:    result.CreatedAt.Time.Format(time.RFC3339),
//		UpdatedAt:    result.UpdatedAt.Time.Format(time.RFC3339),
//	}
//
//	return output, nil
//}
