//go:generate go run github.com/matryer/moq@latest -out mock_cross_domain_test.go -pkg service . GiftItemRepositoryInterface ReservationRepositoryInterface EmailServiceInterface CacheInterface

package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"math/big"
	"regexp"
	"strings"
	"time"

	"wish-list/internal/app/database"
	itemmodels "wish-list/internal/domain/item/models"
	itemrepository "wish-list/internal/domain/item/repository"
	reservationmodels "wish-list/internal/domain/reservation/models"
	"wish-list/internal/domain/wishlist/models"
	"wish-list/internal/domain/wishlist/repository"
	"wish-list/internal/pkg/logger"

	"github.com/jackc/pgx/v5/pgtype"
)

// slugPattern accepts only lowercase letters, digits, and hyphens.
var slugPattern = regexp.MustCompile(`^[a-z0-9-]+$`)

// Cross-domain interfaces - only methods actually used by WishListService

// GiftItemRepositoryInterface defines gift item repository methods used by wishlist service
type GiftItemRepositoryInterface interface {
	CreateWithOwner(ctx context.Context, giftItem itemmodels.GiftItem) (*itemmodels.GiftItem, error)
	GetByID(ctx context.Context, id pgtype.UUID) (*itemmodels.GiftItem, error)
	GetByWishList(ctx context.Context, wishlistID pgtype.UUID) ([]*itemmodels.GiftItem, error)
	GetPublicWishListGiftItemsPaginated(ctx context.Context, publicSlug string, limit, offset int) ([]*itemmodels.GiftItem, int, error)
	GetPublicWishListGiftItemsFiltered(ctx context.Context, publicSlug string, filters itemrepository.PublicItemFilters) ([]*itemmodels.GiftItem, int, error)
	Update(ctx context.Context, giftItem itemmodels.GiftItem) (*itemmodels.GiftItem, error)
}

// GiftItemReservationRepositoryInterface defines gift item reservation repository methods used by wishlist service
type GiftItemReservationRepositoryInterface interface {
	DeleteWithReservationNotification(ctx context.Context, giftItemID pgtype.UUID) ([]*reservationmodels.Reservation, error)
}

// GiftItemPurchaseRepositoryInterface defines gift item purchase repository methods used by wishlist service
type GiftItemPurchaseRepositoryInterface interface {
	MarkAsPurchased(ctx context.Context, giftItemID, userID pgtype.UUID, purchasedPrice pgtype.Numeric) (*itemmodels.GiftItem, error)
}

// ReservationRepositoryInterface defines reservation repository methods used by wishlist service
type ReservationRepositoryInterface interface {
	GetActiveReservationForGiftItem(ctx context.Context, giftItemID pgtype.UUID) (*reservationmodels.Reservation, error)
}

// EmailServiceInterface defines email service methods used by wishlist service
type EmailServiceInterface interface {
	SendReservationRemovedEmail(ctx context.Context, recipientEmail, giftItemName, wishlistTitle string) error
	SendGiftPurchasedConfirmationEmail(ctx context.Context, recipientEmail, giftItemName, wishlistTitle, guestName string) error
}

// CacheInterface defines cache methods used by wishlist service
type CacheInterface interface {
	Get(ctx context.Context, key string, dest any) error
	Set(ctx context.Context, key string, value any) error
	Delete(ctx context.Context, key string) error
}

// Sentinel errors
var (
	ErrWishListNotFound        = errors.New("wishlist not found")
	ErrWishListForbidden       = errors.New("not authorized to access this wishlist")
	ErrWishListTitleRequired   = errors.New("title is required")
	ErrInvalidWishListUserID   = errors.New("invalid user id")
	ErrInvalidWishListID       = errors.New("invalid wishlist id")
	ErrInvalidWishListGiftItem = errors.New("invalid gift item id")
	ErrActiveReservationsExist = errors.New("cannot delete wishlist with active reservations - please remove or cancel all reservations first")
	ErrNameRequired            = errors.New("name is required")
	ErrPriorityOutOfRange      = errors.New("priority value out of int32 range")
	ErrPositionOutOfRange      = errors.New("position value out of int32 range")
	ErrGiftItemIDRequired      = errors.New("gift item ID is required")
	ErrUserIDRequired          = errors.New("user ID is required")
	ErrSlugTaken               = errors.New("public slug is already taken by another wishlist")
	ErrSlugInvalid             = errors.New("public slug must contain only lowercase letters, digits, and hyphens")
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
	GetGiftItemsByPublicSlugPaginated(ctx context.Context, publicSlug string, limit, offset int) ([]*GiftItemOutput, int, error)
	GetGiftItemsByPublicSlugFiltered(ctx context.Context, publicSlug string, filters PublicItemFiltersInput) ([]*GiftItemOutput, int, error)
	UpdateGiftItem(ctx context.Context, giftItemID string, input UpdateGiftItemInput) (*GiftItemOutput, error)
	DeleteGiftItem(ctx context.Context, giftItemID string) error
	MarkGiftItemAsPurchased(ctx context.Context, giftItemID, userID string, purchasedPrice float64) (*GiftItemOutput, error)
}

type WishListService struct {
	wishListRepo            repository.WishListRepositoryInterface
	giftItemRepo            GiftItemRepositoryInterface
	giftItemReservationRepo GiftItemReservationRepositoryInterface
	giftItemPurchaseRepo    GiftItemPurchaseRepositoryInterface
	emailService            EmailServiceInterface
	reservationRepo         ReservationRepositoryInterface
	cache                   CacheInterface
}

func NewWishListService(
	wishListRepo repository.WishListRepositoryInterface,
	giftItemRepo GiftItemRepositoryInterface,
	giftItemReservationRepo GiftItemReservationRepositoryInterface,
	giftItemPurchaseRepo GiftItemPurchaseRepositoryInterface,
	emailService EmailServiceInterface,
	reservationRepo ReservationRepositoryInterface,
	cacheService CacheInterface,
) *WishListService {
	return &WishListService{
		wishListRepo:            wishListRepo,
		giftItemRepo:            giftItemRepo,
		giftItemReservationRepo: giftItemReservationRepo,
		giftItemPurchaseRepo:    giftItemPurchaseRepo,
		emailService:            emailService,
		reservationRepo:         reservationRepo,
		cache:                   cacheService,
	}
}

type CreateWishListInput struct {
	Title        string
	Description  string
	Occasion     string
	OccasionDate string
	IsPublic     bool
}

type UpdateWishListInput struct {
	Title        *string
	Description  *string
	Occasion     *string
	OccasionDate *string
	IsPublic     *bool
	PublicSlug   *string // nil = no change; empty string = clear slug; non-empty = set custom slug
}

type WishListOutput struct {
	ID           string
	OwnerID      string
	Title        string
	Description  string
	Occasion     string
	OccasionDate string
	IsPublic     bool
	PublicSlug   string
	ViewCount    int64
	ItemCount    int64 // Number of gift items in this wishlist
	CreatedAt    string
	UpdatedAt    string
}

// PublicItemFiltersInput contains filter and pagination parameters for server-side filtered public item queries
type PublicItemFiltersInput struct {
	Limit  int
	Offset int
	Search string
	Status string
	SortBy string
}

type CreateGiftItemInput struct {
	Name        string
	Description string
	Link        string
	ImageURL    string
	Price       float64
	Priority    int
	Notes       string
	Position    int
}

type UpdateGiftItemInput struct {
	Name        *string
	Description *string
	Link        *string
	ImageURL    *string
	Price       *float64
	Priority    *int
	Notes       *string
	Position    *int
}

type GiftItemOutput struct {
	ID                string
	WishlistID        string
	OwnerID           string // Items now belong to users, not wishlists
	Name              string
	Description       string
	Link              string
	ImageURL          string
	Price             float64
	Priority          int
	ReservedByUserID  string
	ReservedAt        string
	IsReserved        bool
	PurchasedByUserID string
	PurchasedAt       string
	PurchasedPrice    float64
	Notes             string
	Position          int
	CreatedAt         string
	UpdatedAt         string
}

func isGiftItemReserved(item *itemmodels.GiftItem) bool {
	if item == nil {
		return false
	}

	if item.PurchasedByUserID.Valid || item.PurchasedAt.Valid {
		return false
	}

	return item.ReservedByUserID.Valid || item.ReservedAt.Valid || item.ManualReservedByName.Valid
}

func (s *WishListService) CreateWishList(ctx context.Context, userID string, input CreateWishListInput) (*WishListOutput, error) {
	// Validate input
	if input.Title == "" {
		return nil, ErrWishListTitleRequired
	}

	// Parse user ID
	ownerID := pgtype.UUID{}
	if err := ownerID.Scan(userID); err != nil {
		return nil, ErrInvalidWishListUserID
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
	wishList := models.WishList{
		OwnerID:      ownerID,
		Title:        input.Title,
		Description:  pgtype.Text{String: input.Description, Valid: input.Description != ""},
		Occasion:     pgtype.Text{String: input.Occasion, Valid: input.Occasion != ""},
		OccasionDate: occasionDate,
		IsPublic:     pgtype.Bool{Bool: input.IsPublic, Valid: true},
		PublicSlug:   publicSlug,
	}

	createdWishList, err := s.wishListRepo.Create(ctx, wishList)
	if err != nil {
		return nil, fmt.Errorf("failed to create wishlist in repository: %w", err)
	}

	output := &WishListOutput{
		ID:        createdWishList.ID.String(),
		OwnerID:   createdWishList.OwnerID.String(),
		Title:     createdWishList.Title,
		CreatedAt: createdWishList.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt: createdWishList.UpdatedAt.Time.Format(time.RFC3339),
	}

	// Handle nullable fields
	if createdWishList.Description.Valid {
		output.Description = createdWishList.Description.String
	}
	if createdWishList.Occasion.Valid {
		output.Occasion = createdWishList.Occasion.String
	}
	if createdWishList.OccasionDate.Valid {
		output.OccasionDate = createdWishList.OccasionDate.Time.Format(time.RFC3339)
	}
	if createdWishList.IsPublic.Valid {
		output.IsPublic = createdWishList.IsPublic.Bool
	}
	if createdWishList.PublicSlug.Valid {
		output.PublicSlug = createdWishList.PublicSlug.String
	}
	if createdWishList.ViewCount.Valid {
		output.ViewCount = int64(createdWishList.ViewCount.Int32)
	}

	return output, nil
}

func (s *WishListService) GetWishList(ctx context.Context, wishListID string) (*WishListOutput, error) {
	id := pgtype.UUID{}
	if err := id.Scan(wishListID); err != nil {
		return nil, ErrInvalidWishListID
	}

	wishList, err := s.wishListRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get wishlist from repository: %w", err)
	}

	output := &WishListOutput{
		ID:        wishList.ID.String(),
		OwnerID:   wishList.OwnerID.String(),
		Title:     wishList.Title,
		CreatedAt: wishList.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt: wishList.UpdatedAt.Time.Format(time.RFC3339),
	}

	// Handle nullable fields
	if wishList.Description.Valid {
		output.Description = wishList.Description.String
	}
	if wishList.Occasion.Valid {
		output.Occasion = wishList.Occasion.String
	}
	if wishList.OccasionDate.Valid {
		output.OccasionDate = wishList.OccasionDate.Time.Format(time.RFC3339)
	}
	if wishList.IsPublic.Valid {
		output.IsPublic = wishList.IsPublic.Bool
	}
	if wishList.PublicSlug.Valid {
		output.PublicSlug = wishList.PublicSlug.String
	}
	if wishList.ViewCount.Valid {
		output.ViewCount = int64(wishList.ViewCount.Int32)
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
		ID:        wishList.ID.String(),
		OwnerID:   wishList.OwnerID.String(),
		Title:     wishList.Title,
		CreatedAt: wishList.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt: wishList.UpdatedAt.Time.Format(time.RFC3339),
	}

	// Handle nullable fields
	if wishList.Description.Valid {
		output.Description = wishList.Description.String
	}
	if wishList.Occasion.Valid {
		output.Occasion = wishList.Occasion.String
	}
	if wishList.OccasionDate.Valid {
		output.OccasionDate = wishList.OccasionDate.Time.Format(time.RFC3339)
	}
	if wishList.IsPublic.Valid {
		output.IsPublic = wishList.IsPublic.Bool
	}
	if wishList.PublicSlug.Valid {
		output.PublicSlug = wishList.PublicSlug.String
	}
	if wishList.ViewCount.Valid {
		output.ViewCount = int64(wishList.ViewCount.Int32)
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
		return nil, ErrInvalidWishListUserID
	}

	// Use the efficient method that gets wishlists with item counts in a single query
	wishLists, err := s.wishListRepo.GetByOwnerWithItemCount(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get wish lists by owner with item count from repository: %w", err)
	}

	var outputs []*WishListOutput
	for _, wishListWithCount := range wishLists {
		output := &WishListOutput{
			ID:        wishListWithCount.ID.String(),
			OwnerID:   wishListWithCount.OwnerID.String(),
			Title:     wishListWithCount.Title,
			ItemCount: wishListWithCount.ItemCount,
			CreatedAt: wishListWithCount.CreatedAt.Time.Format(time.RFC3339),
			UpdatedAt: wishListWithCount.UpdatedAt.Time.Format(time.RFC3339),
		}

		// Handle nullable fields
		if wishListWithCount.Description.Valid {
			output.Description = wishListWithCount.Description.String
		}
		if wishListWithCount.Occasion.Valid {
			output.Occasion = wishListWithCount.Occasion.String
		}
		if wishListWithCount.OccasionDate.Valid {
			output.OccasionDate = wishListWithCount.OccasionDate.Time.Format(time.RFC3339)
		}
		if wishListWithCount.IsPublic.Valid {
			output.IsPublic = wishListWithCount.IsPublic.Bool
		}
		if wishListWithCount.PublicSlug.Valid {
			output.PublicSlug = wishListWithCount.PublicSlug.String
		}
		if wishListWithCount.ViewCount.Valid {
			output.ViewCount = int64(wishListWithCount.ViewCount.Int32)
		}

		outputs = append(outputs, output)
	}

	return outputs, nil
}

func (s *WishListService) UpdateWishList(ctx context.Context, wishListID, userID string, input UpdateWishListInput) (*WishListOutput, error) {
	id := pgtype.UUID{}
	if err := id.Scan(wishListID); err != nil {
		return nil, ErrInvalidWishListID
	}

	// Verify ownership
	wishList, err := s.wishListRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrWishListNotFound, err)
	}

	ownerID := pgtype.UUID{}
	if err := ownerID.Scan(userID); err != nil {
		return nil, ErrInvalidWishListUserID
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

	// Handle custom public slug provided by the user
	if input.PublicSlug != nil {
		customSlug := strings.TrimSpace(*input.PublicSlug)
		if customSlug != "" {
			// Validate format: lowercase letters, digits, hyphens only
			if !slugPattern.MatchString(customSlug) {
				return nil, ErrSlugInvalid
			}
			// Check uniqueness (exclude current wishlist)
			taken, err := s.wishListRepo.IsSlugTaken(ctx, customSlug, id)
			if err != nil {
				return nil, fmt.Errorf("failed to check slug uniqueness: %w", err)
			}
			if taken {
				return nil, ErrSlugTaken
			}
			updatedWishList.PublicSlug = pgtype.Text{String: customSlug, Valid: true}
		}
		// empty string → keep existing slug (do not clear it)
	}

	// Auto-generate slug if making the list public and it still has no slug
	currentIsPublic := input.IsPublic != nil && *input.IsPublic
	if currentIsPublic && !updatedWishList.PublicSlug.Valid {
		titleToUse := updatedWishList.Title
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
		ID:        updated.ID.String(),
		OwnerID:   updated.OwnerID.String(),
		Title:     updated.Title,
		CreatedAt: updated.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt: updated.UpdatedAt.Time.Format(time.RFC3339),
	}

	// Handle nullable fields
	if updated.Description.Valid {
		output.Description = updated.Description.String
	}
	if updated.Occasion.Valid {
		output.Occasion = updated.Occasion.String
	}
	if updated.OccasionDate.Valid {
		output.OccasionDate = updated.OccasionDate.Time.Format(time.RFC3339)
	}
	if updated.IsPublic.Valid {
		output.IsPublic = updated.IsPublic.Bool
	}
	if updated.PublicSlug.Valid {
		output.PublicSlug = updated.PublicSlug.String
	}
	if updated.ViewCount.Valid {
		output.ViewCount = int64(updated.ViewCount.Int32)
	}

	return output, nil
}

func (s *WishListService) DeleteWishList(ctx context.Context, wishListID, userID string) error {
	id := pgtype.UUID{}
	if err := id.Scan(wishListID); err != nil {
		return ErrInvalidWishListID
	}

	// Verify ownership
	wishList, err := s.wishListRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get wishlist from repository: %w", err)
	}

	ownerID := pgtype.UUID{}
	if err := ownerID.Scan(userID); err != nil {
		return ErrInvalidWishListUserID
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
		return ErrActiveReservationsExist
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
		return nil, ErrNameRequired
	}

	// Validate int32 bounds for Priority and Position
	if input.Priority < math.MinInt32 || input.Priority > math.MaxInt32 {
		return nil, ErrPriorityOutOfRange
	}
	if input.Position < math.MinInt32 || input.Position > math.MaxInt32 {
		return nil, ErrPositionOutOfRange
	}

	// Parse wishlist ID
	listID := pgtype.UUID{}
	if err := listID.Scan(wishListID); err != nil {
		return nil, ErrInvalidWishListID
	}

	// Resolve wishlist owner; item owner must be a user ID.
	wishList, err := s.wishListRepo.GetByID(ctx, listID)
	if err != nil {
		if errors.Is(err, repository.ErrWishListNotFound) {
			return nil, ErrWishListNotFound
		}
		return nil, fmt.Errorf("failed to get wishlist: %w", err)
	}

	// Create price numeric
	priceBig := new(big.Int)
	priceBig.SetInt64(int64(input.Price * 100)) // Convert to cents

	// Create gift item
	giftItem := itemmodels.GiftItem{
		OwnerID:     wishList.OwnerID,
		Name:        input.Name,
		Description: pgtype.Text{String: input.Description, Valid: input.Description != ""},
		Link:        pgtype.Text{String: input.Link, Valid: input.Link != ""},
		ImageUrl:    pgtype.Text{String: input.ImageURL, Valid: input.ImageURL != ""},
		Price:       pgtype.Numeric{Int: priceBig, Exp: -2, Valid: input.Price > 0},
		Priority:    pgtype.Int4{Int32: int32(input.Priority), Valid: true},
		Notes:       pgtype.Text{String: input.Notes, Valid: input.Notes != ""},
		Position:    pgtype.Int4{Int32: int32(input.Position), Valid: true},
	}

	createdGiftItem, err := s.giftItemRepo.CreateWithOwner(ctx, giftItem)
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
		ID:         createdGiftItem.ID.String(),
		WishlistID: wishListID,
		OwnerID:    createdGiftItem.OwnerID.String(),
		Name:       createdGiftItem.Name,
		Price:      price,
		IsReserved: isGiftItemReserved(createdGiftItem),
		CreatedAt:  createdGiftItem.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:  createdGiftItem.UpdatedAt.Time.Format(time.RFC3339),
	}

	// Handle nullable fields
	if createdGiftItem.Description.Valid {
		output.Description = createdGiftItem.Description.String
	}
	if createdGiftItem.Link.Valid {
		output.Link = createdGiftItem.Link.String
	}
	if createdGiftItem.ImageUrl.Valid {
		output.ImageURL = createdGiftItem.ImageUrl.String
	}
	if createdGiftItem.Priority.Valid {
		output.Priority = int(createdGiftItem.Priority.Int32)
	}
	if createdGiftItem.Notes.Valid {
		output.Notes = createdGiftItem.Notes.String
	}
	if createdGiftItem.Position.Valid {
		output.Position = int(createdGiftItem.Position.Int32)
	}

	return output, nil
}

func (s *WishListService) GetGiftItem(ctx context.Context, giftItemID string) (*GiftItemOutput, error) {
	id := pgtype.UUID{}
	if err := id.Scan(giftItemID); err != nil {
		return nil, ErrInvalidWishListGiftItem
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
		ID:         giftItem.ID.String(),
		WishlistID: "",
		OwnerID:    giftItem.OwnerID.String(),
		Name:       giftItem.Name,
		Price:      price,
		IsReserved: isGiftItemReserved(giftItem),
		CreatedAt:  giftItem.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:  giftItem.UpdatedAt.Time.Format(time.RFC3339),
	}

	// Handle nullable fields
	if giftItem.Description.Valid {
		output.Description = giftItem.Description.String
	}
	if giftItem.Link.Valid {
		output.Link = giftItem.Link.String
	}
	if giftItem.ImageUrl.Valid {
		output.ImageURL = giftItem.ImageUrl.String
	}
	if giftItem.Priority.Valid {
		output.Priority = int(giftItem.Priority.Int32)
	}
	if giftItem.Notes.Valid {
		output.Notes = giftItem.Notes.String
	}
	if giftItem.Position.Valid {
		output.Position = int(giftItem.Position.Int32)
	}

	return output, nil
}

func (s *WishListService) GetGiftItemsByWishList(ctx context.Context, wishListID string) ([]*GiftItemOutput, error) {
	listID := pgtype.UUID{}
	if err := listID.Scan(wishListID); err != nil {
		return nil, ErrInvalidWishListID
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
			ID:         giftItem.ID.String(),
			WishlistID: wishListID,
			OwnerID:    giftItem.OwnerID.String(),
			Name:       giftItem.Name,
			Price:      price,
			IsReserved: isGiftItemReserved(giftItem),
			CreatedAt:  giftItem.CreatedAt.Time.Format(time.RFC3339),
			UpdatedAt:  giftItem.UpdatedAt.Time.Format(time.RFC3339),
		}

		// Handle nullable fields
		if giftItem.Description.Valid {
			output.Description = giftItem.Description.String
		}
		if giftItem.Link.Valid {
			output.Link = giftItem.Link.String
		}
		if giftItem.ImageUrl.Valid {
			output.ImageURL = giftItem.ImageUrl.String
		}
		if giftItem.Priority.Valid {
			output.Priority = int(giftItem.Priority.Int32)
		}
		if giftItem.Notes.Valid {
			output.Notes = giftItem.Notes.String
		}
		if giftItem.Position.Valid {
			output.Position = int(giftItem.Position.Int32)
		}
		if giftItem.ReservedByUserID.Valid {
			output.ReservedByUserID = giftItem.ReservedByUserID.String()
		}
		if giftItem.ReservedAt.Valid {
			output.ReservedAt = giftItem.ReservedAt.Time.Format(time.RFC3339)
		}
		if giftItem.PurchasedByUserID.Valid {
			output.PurchasedByUserID = giftItem.PurchasedByUserID.String()
		}
		if giftItem.PurchasedAt.Valid {
			output.PurchasedAt = giftItem.PurchasedAt.Time.Format(time.RFC3339)
		}
		if giftItem.PurchasedPrice.Valid {
			purchasedPriceValue, err := giftItem.PurchasedPrice.Float64Value()
			if err == nil && purchasedPriceValue.Valid {
				output.PurchasedPrice = purchasedPriceValue.Float64
			}
		}

		outputs = append(outputs, output)
	}

	return outputs, nil
}

func (s *WishListService) GetGiftItemsByPublicSlugPaginated(ctx context.Context, publicSlug string, limit, offset int) ([]*GiftItemOutput, int, error) {
	wishList, err := s.wishListRepo.GetByPublicSlug(ctx, publicSlug)
	if err != nil {
		if errors.Is(err, repository.ErrWishListNotFound) {
			return nil, 0, ErrWishListNotFound
		}
		return nil, 0, fmt.Errorf("failed to get wishlist by public slug: %w", err)
	}

	giftItems, totalCount, err := s.giftItemRepo.GetPublicWishListGiftItemsPaginated(ctx, publicSlug, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get gift items from repository: %w", err)
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
			ID:         giftItem.ID.String(),
			WishlistID: wishList.ID.String(),
			OwnerID:    giftItem.OwnerID.String(),
			Name:       giftItem.Name,
			Price:      price,
			IsReserved: isGiftItemReserved(giftItem),
			CreatedAt:  giftItem.CreatedAt.Time.Format(time.RFC3339),
			UpdatedAt:  giftItem.UpdatedAt.Time.Format(time.RFC3339),
		}

		// Handle nullable fields
		if giftItem.Description.Valid {
			output.Description = giftItem.Description.String
		}
		if giftItem.Link.Valid {
			output.Link = giftItem.Link.String
		}
		if giftItem.ImageUrl.Valid {
			output.ImageURL = giftItem.ImageUrl.String
		}
		if giftItem.Priority.Valid {
			output.Priority = int(giftItem.Priority.Int32)
		}
		if giftItem.Position.Valid {
			output.Position = int(giftItem.Position.Int32)
		}
		if giftItem.ReservedAt.Valid {
			output.ReservedAt = giftItem.ReservedAt.Time.Format(time.RFC3339)
		}
		if giftItem.PurchasedAt.Valid {
			output.PurchasedAt = giftItem.PurchasedAt.Time.Format(time.RFC3339)
		}
		if giftItem.PurchasedPrice.Valid {
			purchasedPriceValue, err := giftItem.PurchasedPrice.Float64Value()
			if err == nil && purchasedPriceValue.Valid {
				output.PurchasedPrice = purchasedPriceValue.Float64
			}
		}

		outputs = append(outputs, output)
	}

	return outputs, totalCount, nil
}

func (s *WishListService) GetGiftItemsByPublicSlugFiltered(ctx context.Context, publicSlug string, filters PublicItemFiltersInput) ([]*GiftItemOutput, int, error) {
	wishList, err := s.wishListRepo.GetByPublicSlug(ctx, publicSlug)
	if err != nil {
		if errors.Is(err, repository.ErrWishListNotFound) {
			return nil, 0, ErrWishListNotFound
		}
		return nil, 0, fmt.Errorf("failed to get wishlist by public slug: %w", err)
	}

	repoFilters := itemrepository.PublicItemFilters{
		Limit:  filters.Limit,
		Offset: filters.Offset,
		Search: filters.Search,
		Status: filters.Status,
		SortBy: filters.SortBy,
	}

	giftItems, totalCount, err := s.giftItemRepo.GetPublicWishListGiftItemsFiltered(ctx, publicSlug, repoFilters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get gift items from repository: %w", err)
	}

	var outputs []*GiftItemOutput

	for _, giftItem := range giftItems {
		if giftItem == nil {
			continue
		}

		var price float64
		if giftItem.Price.Valid {
			priceValue, err := giftItem.Price.Float64Value()
			if err == nil && priceValue.Valid {
				price = priceValue.Float64
			}
		}

		output := &GiftItemOutput{
			ID:         giftItem.ID.String(),
			WishlistID: wishList.ID.String(),
			OwnerID:    giftItem.OwnerID.String(),
			Name:       giftItem.Name,
			Price:      price,
			IsReserved: isGiftItemReserved(giftItem),
			CreatedAt:  giftItem.CreatedAt.Time.Format(time.RFC3339),
			UpdatedAt:  giftItem.UpdatedAt.Time.Format(time.RFC3339),
		}

		if giftItem.Description.Valid {
			output.Description = giftItem.Description.String
		}
		if giftItem.Link.Valid {
			output.Link = giftItem.Link.String
		}
		if giftItem.ImageUrl.Valid {
			output.ImageURL = giftItem.ImageUrl.String
		}
		if giftItem.Priority.Valid {
			output.Priority = int(giftItem.Priority.Int32)
		}
		if giftItem.Notes.Valid {
			output.Notes = giftItem.Notes.String
		}
		if giftItem.Position.Valid {
			output.Position = int(giftItem.Position.Int32)
		}
		if giftItem.ReservedByUserID.Valid {
			output.ReservedByUserID = giftItem.ReservedByUserID.String()
		}
		if giftItem.ReservedAt.Valid {
			output.ReservedAt = giftItem.ReservedAt.Time.Format(time.RFC3339)
		}
		if giftItem.PurchasedByUserID.Valid {
			output.PurchasedByUserID = giftItem.PurchasedByUserID.String()
		}
		if giftItem.PurchasedAt.Valid {
			output.PurchasedAt = giftItem.PurchasedAt.Time.Format(time.RFC3339)
		}
		if giftItem.PurchasedPrice.Valid {
			purchasedPriceValue, err := giftItem.PurchasedPrice.Float64Value()
			if err == nil && purchasedPriceValue.Valid {
				output.PurchasedPrice = purchasedPriceValue.Float64
			}
		}

		outputs = append(outputs, output)
	}

	return outputs, totalCount, nil
}

func (s *WishListService) UpdateGiftItem(ctx context.Context, giftItemID string, input UpdateGiftItemInput) (*GiftItemOutput, error) {
	// Validate int32 bounds for Priority and Position if provided
	if input.Priority != nil {
		if *input.Priority < math.MinInt32 || *input.Priority > math.MaxInt32 {
			return nil, ErrPriorityOutOfRange
		}
	}
	if input.Position != nil {
		if *input.Position < math.MinInt32 || *input.Position > math.MaxInt32 {
			return nil, ErrPositionOutOfRange
		}
	}

	id := pgtype.UUID{}
	if err := id.Scan(giftItemID); err != nil {
		return nil, ErrInvalidWishListGiftItem
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
	s.invalidatePublicWishlistsCacheByOwner(ctx, updated.OwnerID)

	// Convert price to float64
	var price float64
	if updated.Price.Valid {
		priceValue, err := updated.Price.Float64Value()
		if err == nil && priceValue.Valid {
			price = priceValue.Float64
		}
	}

	output := &GiftItemOutput{
		ID:         updated.ID.String(),
		WishlistID: "",
		OwnerID:    updated.OwnerID.String(),
		Name:       updated.Name,
		Price:      price,
		IsReserved: isGiftItemReserved(updated),
		CreatedAt:  updated.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:  updated.UpdatedAt.Time.Format(time.RFC3339),
	}

	// Handle nullable fields
	if updated.Description.Valid {
		output.Description = updated.Description.String
	}
	if updated.Link.Valid {
		output.Link = updated.Link.String
	}
	if updated.ImageUrl.Valid {
		output.ImageURL = updated.ImageUrl.String
	}
	if updated.Priority.Valid {
		output.Priority = int(updated.Priority.Int32)
	}
	if updated.Notes.Valid {
		output.Notes = updated.Notes.String
	}
	if updated.Position.Valid {
		output.Position = int(updated.Position.Int32)
	}

	return output, nil
}

func (s *WishListService) DeleteGiftItem(ctx context.Context, giftItemID string) error {
	id := pgtype.UUID{}
	if err := id.Scan(giftItemID); err != nil {
		return ErrInvalidWishListGiftItem
	}

	// Get gift item before deletion to get wishlist ID for cache invalidation
	giftItemForCache, err := s.giftItemRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get gift item from repository: %w", err)
	}

	// Delete the gift item and get any active reservations for notification purposes
	activeReservations, err := s.giftItemReservationRepo.DeleteWithReservationNotification(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete gift item in repository: %w", err)
	}

	s.invalidatePublicWishlistsCacheByOwner(ctx, giftItemForCache.OwnerID)

	// If there were active reservations, send notifications to the reservation holders
	if len(activeReservations) > 0 {
		wishlistTitles := make(map[string]string, len(activeReservations))

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

			if recipientEmail == "" {
				continue
			}

			wishlistTitle := ""
			if reservation.WishlistID.Valid {
				wishlistID := reservation.WishlistID.String()
				if cachedTitle, ok := wishlistTitles[wishlistID]; ok {
					wishlistTitle = cachedTitle
				} else {
					wishList, err := s.wishListRepo.GetByID(ctx, reservation.WishlistID)
					if err != nil {
						logger.Warn(
							"failed to get wishlist details for reservation removal notification",
							"error",
							err,
							"wishlist_id",
							wishlistID,
						)
					} else {
						wishlistTitle = wishList.Title
						wishlistTitles[wishlistID] = wishlistTitle
					}
				}
			}

			err := s.emailService.SendReservationRemovedEmail(ctx, recipientEmail, giftItemForCache.Name, wishlistTitle)
			if err != nil {
				// Log the error but don't fail the deletion
				logger.Warn(
					"failed to send reservation removal notification",
					"error",
					err,
					"reservation_id",
					reservation.ID.String(),
					"item_id",
					id.String(),
				)
			}
		}
	}

	return nil
}

// MarkGiftItemAsPurchased marks a gift item as purchased
func (s *WishListService) MarkGiftItemAsPurchased(ctx context.Context, giftItemID, userID string, purchasedPrice float64) (*GiftItemOutput, error) {
	// Validate input
	if giftItemID == "" {
		return nil, ErrGiftItemIDRequired
	}
	if userID == "" {
		return nil, ErrUserIDRequired
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
	updatedGiftItem, err := s.giftItemPurchaseRepo.MarkAsPurchased(ctx, itemID, userUUID, priceValue)
	if err != nil {
		return nil, fmt.Errorf("failed to mark gift item as purchased in repository: %w", err)
	}

	// Send email notification to the person who reserved the gift
	if s.emailService != nil {
		// Check if there's an active reservation for this gift item
		reservation, err := s.reservationRepo.GetActiveReservationForGiftItem(ctx, updatedGiftItem.ID)
		if err == nil && reservation != nil {
			var recipientEmail, guestName string
			if reservation.GuestEmail.Valid {
				recipientEmail = reservation.GuestEmail.String
			}
			if reservation.GuestName.Valid {
				guestName = reservation.GuestName.String
			}

			if recipientEmail != "" {
				wishlistTitle := ""
				if reservation.WishlistID.Valid {
					wishList, err := s.wishListRepo.GetByID(ctx, reservation.WishlistID)
					if err != nil {
						logger.Warn(
							"failed to get wishlist details for purchase confirmation notification",
							"error",
							err,
							"wishlist_id",
							reservation.WishlistID.String(),
						)
					} else {
						wishlistTitle = wishList.Title
					}
				}

				err := s.emailService.SendGiftPurchasedConfirmationEmail(
					ctx,
					recipientEmail,
					updatedGiftItem.Name,
					wishlistTitle,
					guestName,
				)
				if err != nil {
					// Log the error but don't fail the purchase marking
					logger.Warn(
						"failed to send gift purchased notification",
						"error",
						err,
						"reservation_id",
						reservation.ID.String(),
						"item_id",
						updatedGiftItem.ID.String(),
					)
				}
			}
		}
	}

	s.invalidatePublicWishlistsCacheByOwner(ctx, updatedGiftItem.OwnerID)

	// Convert to output format
	output := &GiftItemOutput{
		ID:             updatedGiftItem.ID.String(),
		WishlistID:     "",
		OwnerID:        updatedGiftItem.OwnerID.String(),
		Name:           updatedGiftItem.Name,
		Price:          database.NumericToFloat64(updatedGiftItem.Price),
		IsReserved:     isGiftItemReserved(updatedGiftItem),
		PurchasedPrice: database.NumericToFloat64(updatedGiftItem.PurchasedPrice),
		CreatedAt:      updatedGiftItem.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:      updatedGiftItem.UpdatedAt.Time.Format(time.RFC3339),
	}

	// Handle nullable fields
	if updatedGiftItem.Description.Valid {
		output.Description = updatedGiftItem.Description.String
	}
	if updatedGiftItem.Link.Valid {
		output.Link = updatedGiftItem.Link.String
	}
	if updatedGiftItem.ImageUrl.Valid {
		output.ImageURL = updatedGiftItem.ImageUrl.String
	}
	if updatedGiftItem.Priority.Valid {
		output.Priority = int(updatedGiftItem.Priority.Int32)
	}
	if updatedGiftItem.Notes.Valid {
		output.Notes = updatedGiftItem.Notes.String
	}
	if updatedGiftItem.Position.Valid {
		output.Position = int(updatedGiftItem.Position.Int32)
	}
	if updatedGiftItem.ReservedByUserID.Valid {
		output.ReservedByUserID = updatedGiftItem.ReservedByUserID.String()
	}
	if updatedGiftItem.ReservedAt.Valid {
		output.ReservedAt = updatedGiftItem.ReservedAt.Time.Format(time.RFC3339)
	}
	if updatedGiftItem.PurchasedByUserID.Valid {
		output.PurchasedByUserID = updatedGiftItem.PurchasedByUserID.String()
	}
	if updatedGiftItem.PurchasedAt.Valid {
		output.PurchasedAt = updatedGiftItem.PurchasedAt.Time.Format(time.RFC3339)
	}

	return output, nil
}

func (s *WishListService) invalidatePublicWishlistsCacheByOwner(ctx context.Context, ownerID pgtype.UUID) {
	if s.cache == nil || !ownerID.Valid {
		return
	}

	wishLists, err := s.wishListRepo.GetByOwner(ctx, ownerID)
	if err != nil {
		logger.Warn("failed to get wishlists for cache invalidation", "error", err, "owner_id", ownerID.String())
		return
	}

	for _, wishList := range wishLists {
		if wishList == nil || !wishList.PublicSlug.Valid || wishList.PublicSlug.String == "" {
			continue
		}

		cacheKey := fmt.Sprintf("wishlist:public:%s", wishList.PublicSlug.String)
		if err := s.cache.Delete(ctx, cacheKey); err != nil {
			logger.Warn("failed to invalidate wishlist cache", "error", err, "cache_key", cacheKey)
		}
	}
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
