package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	db "wish-list/internal/db/models"
	"wish-list/internal/repositories"

	"github.com/jackc/pgx/v5/pgtype"
)

// Sentinel errors for wishlist-item operations
var (
	ErrItemAlreadyAttached = errors.New("item already attached to this wishlist")
	ErrItemNotInWishlist   = errors.New("item not found in this wishlist")
)

// WishlistItemServiceInterface defines operations for wishlist-item relationships
type WishlistItemServiceInterface interface {
	GetWishlistItems(ctx context.Context, wishlistID string, userID string, page, limit int) (*PaginatedItemsOutput, error)
	AttachItem(ctx context.Context, wishlistID string, itemID string, userID string) error
	CreateItemInWishlist(ctx context.Context, wishlistID string, userID string, input CreateItemInput) (*ItemOutput, error)
	DetachItem(ctx context.Context, wishlistID string, itemID string, userID string) error
}

// WishlistItemService implements WishlistItemServiceInterface
type WishlistItemService struct {
	wishlistRepo     repositories.WishListRepositoryInterface
	itemRepo         repositories.GiftItemRepositoryInterface
	wishlistItemRepo repositories.WishlistItemRepositoryInterface
}

// NewWishlistItemService creates a new WishlistItemService
func NewWishlistItemService(
	wishlistRepo repositories.WishListRepositoryInterface,
	itemRepo repositories.GiftItemRepositoryInterface,
	wishlistItemRepo repositories.WishlistItemRepositoryInterface,
) *WishlistItemService {
	return &WishlistItemService{
		wishlistRepo:     wishlistRepo,
		itemRepo:         itemRepo,
		wishlistItemRepo: wishlistItemRepo,
	}
}

// GetWishlistItems retrieves all items in a wishlist with pagination
func (s *WishlistItemService) GetWishlistItems(ctx context.Context, wishlistID string, userID string, page, limit int) (*PaginatedItemsOutput, error) {
	// Parse IDs
	wlID := pgtype.UUID{}
	if err := wlID.Scan(wishlistID); err != nil {
		return nil, errors.New("invalid wishlist id")
	}

	ownerID := pgtype.UUID{}
	if err := ownerID.Scan(userID); err != nil {
		return nil, errors.New("invalid user id")
	}

	// Get wishlist to check ownership/access
	wishlist, err := s.wishlistRepo.GetByID(ctx, wlID)
	if err != nil {
		return nil, ErrWishListNotFound
	}

	// Check access: must be owner or public
	if wishlist.OwnerID.Bytes != ownerID.Bytes && (!wishlist.IsPublic.Valid || !wishlist.IsPublic.Bool) {
		return nil, ErrWishListForbidden
	}

	// Set defaults
	if limit == 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if page < 1 {
		page = 1
	}

	// Get items
	items, err := s.wishlistItemRepo.GetByWishlist(ctx, wlID, page, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get wishlist items: %w", err)
	}

	// Get total count
	totalCount, err := s.wishlistItemRepo.GetByWishlistCount(ctx, wlID)
	if err != nil {
		return nil, fmt.Errorf("failed to count wishlist items: %w", err)
	}

	// Convert to output
	outputs := make([]*ItemOutput, 0, len(items))
	for _, item := range items {
		outputs = append(outputs, s.convertItemToOutput(item))
	}

	totalPages := int((totalCount + int64(limit) - 1) / int64(limit))

	return &PaginatedItemsOutput{
		Items:      outputs,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// AttachItem attaches an existing item to a wishlist
func (s *WishlistItemService) AttachItem(ctx context.Context, wishlistID string, itemID string, userID string) error {
	// Parse IDs
	wlID := pgtype.UUID{}
	if err := wlID.Scan(wishlistID); err != nil {
		return errors.New("invalid wishlist id")
	}

	itID := pgtype.UUID{}
	if err := itID.Scan(itemID); err != nil {
		return errors.New("invalid item id")
	}

	ownerID := pgtype.UUID{}
	if err := ownerID.Scan(userID); err != nil {
		return errors.New("invalid user id")
	}

	// Get wishlist to check ownership
	wishlist, err := s.wishlistRepo.GetByID(ctx, wlID)
	if err != nil {
		return ErrWishListNotFound
	}

	// Must be wishlist owner
	if wishlist.OwnerID.Bytes != ownerID.Bytes {
		return ErrWishListForbidden
	}

	// Get item to check ownership
	item, err := s.itemRepo.GetByID(ctx, itID)
	if err != nil {
		return ErrItemNotFound
	}

	// Must be item owner
	if item.OwnerID.Bytes != ownerID.Bytes {
		return ErrItemForbidden
	}

	// Check if already attached
	attached, err := s.wishlistItemRepo.IsAttached(ctx, wlID, itID)
	if err != nil {
		return fmt.Errorf("failed to check attachment: %w", err)
	}

	if attached {
		return ErrItemAlreadyAttached
	}

	// Attach
	if err := s.wishlistItemRepo.Attach(ctx, wlID, itID); err != nil {
		return fmt.Errorf("failed to attach item: %w", err)
	}

	return nil
}

// CreateItemInWishlist creates a new item and immediately attaches it to a wishlist
func (s *WishlistItemService) CreateItemInWishlist(ctx context.Context, wishlistID string, userID string, input CreateItemInput) (*ItemOutput, error) {
	// Validate input
	if input.Title == "" {
		return nil, errors.New("title is required")
	}

	// Parse IDs
	wlID := pgtype.UUID{}
	if err := wlID.Scan(wishlistID); err != nil {
		return nil, errors.New("invalid wishlist id")
	}

	ownerID := pgtype.UUID{}
	if err := ownerID.Scan(userID); err != nil {
		return nil, errors.New("invalid user id")
	}

	// Get wishlist to check ownership
	wishlist, err := s.wishlistRepo.GetByID(ctx, wlID)
	if err != nil {
		return nil, ErrWishListNotFound
	}

	// Must be wishlist owner
	if wishlist.OwnerID.Bytes != ownerID.Bytes {
		return nil, ErrWishListForbidden
	}

	// Create item model
	item := db.GiftItem{
		OwnerID:     ownerID,
		Name:        input.Title,
		Description: pgtype.Text{String: input.Description, Valid: input.Description != ""},
		Link:        pgtype.Text{String: input.Link, Valid: input.Link != ""},
		ImageUrl:    pgtype.Text{String: input.ImageURL, Valid: input.ImageURL != ""},
		Priority:    pgtype.Int4{Int32: int32(input.Priority), Valid: true},
		Notes:       pgtype.Text{String: input.Notes, Valid: input.Notes != ""},
	}

	// Set price if provided
	if input.Price > 0 {
		if err := item.Price.Scan(fmt.Sprintf("%f", input.Price)); err != nil {
			return nil, fmt.Errorf("invalid price: %w", err)
		}
	}

	// Create item
	createdItem, err := s.itemRepo.CreateWithOwner(ctx, item)
	if err != nil {
		return nil, fmt.Errorf("failed to create item: %w", err)
	}

	// Attach to wishlist
	if err := s.wishlistItemRepo.Attach(ctx, wlID, createdItem.ID); err != nil {
		// If attachment fails, we could optionally delete the created item
		// For now, we leave it unattached
		return nil, fmt.Errorf("failed to attach item to wishlist: %w", err)
	}

	return s.convertItemToOutput(createdItem), nil
}

// DetachItem removes an item from a wishlist (doesn't delete the item)
func (s *WishlistItemService) DetachItem(ctx context.Context, wishlistID string, itemID string, userID string) error {
	// Parse IDs
	wlID := pgtype.UUID{}
	if err := wlID.Scan(wishlistID); err != nil {
		return errors.New("invalid wishlist id")
	}

	itID := pgtype.UUID{}
	if err := itID.Scan(itemID); err != nil {
		return errors.New("invalid item id")
	}

	ownerID := pgtype.UUID{}
	if err := ownerID.Scan(userID); err != nil {
		return errors.New("invalid user id")
	}

	// Get wishlist to check ownership
	wishlist, err := s.wishlistRepo.GetByID(ctx, wlID)
	if err != nil {
		return ErrWishListNotFound
	}

	// Must be wishlist owner
	if wishlist.OwnerID.Bytes != ownerID.Bytes {
		return ErrWishListForbidden
	}

	// Check if attached
	attached, err := s.wishlistItemRepo.IsAttached(ctx, wlID, itID)
	if err != nil {
		return fmt.Errorf("failed to check attachment: %w", err)
	}

	if !attached {
		return ErrItemNotInWishlist
	}

	// Detach
	if err := s.wishlistItemRepo.Detach(ctx, wlID, itID); err != nil {
		return fmt.Errorf("failed to detach item: %w", err)
	}

	return nil
}

// Helper to convert db.GiftItem to ItemOutput
func (s *WishlistItemService) convertItemToOutput(item *db.GiftItem) *ItemOutput {
	output := &ItemOutput{
		ID:          item.ID.String(),
		OwnerID:     item.OwnerID.String(),
		Title:       item.Name,
		Description: "",
		Link:        "",
		ImageURL:    "",
		Price:       0,
		Priority:    0,
		Notes:       "",
		IsPurchased: item.PurchasedByUserID.Valid,
		IsArchived:  item.ArchivedAt.Valid,
		CreatedAt:   item.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:   item.UpdatedAt.Time.Format(time.RFC3339),
	}

	// Handle nullable fields
	if item.Description.Valid {
		output.Description = item.Description.String
	}
	if item.Link.Valid {
		output.Link = item.Link.String
	}
	if item.ImageUrl.Valid {
		output.ImageURL = item.ImageUrl.String
	}
	if item.Price.Valid {
		if priceValue, err := item.Price.Float64Value(); err == nil && priceValue.Valid {
			output.Price = priceValue.Float64
		}
	}
	if item.Priority.Valid {
		output.Priority = int(item.Priority.Int32)
	}
	if item.Notes.Valid {
		output.Notes = item.Notes.String
	}

	return output
}
