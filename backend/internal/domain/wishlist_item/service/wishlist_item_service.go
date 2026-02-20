//go:generate go run github.com/matryer/moq@latest -out mock_crossdomain_test.go -pkg service . WishListRepositoryInterface GiftItemRepositoryInterface

package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	itemmodels "wish-list/internal/domain/item/models"
	wishlistmodels "wish-list/internal/domain/wishlist/models"
	"wish-list/internal/domain/wishlist_item/repository"

	"github.com/jackc/pgx/v5/pgtype"
)

// Sentinel errors for wishlist-item operations
var (
	ErrItemAlreadyAttached       = errors.New("item already attached to this wishlist")
	ErrItemNotInWishlist         = errors.New("item not found in this wishlist")
	ErrInvalidWishlistItemWLID   = errors.New("invalid wishlist id")
	ErrInvalidWishlistItemID     = errors.New("invalid item id")
	ErrInvalidWishlistItemUser   = errors.New("invalid user id")
	ErrWishlistItemTitleRequired = errors.New("title is required")
	ErrWishListNotFound          = errors.New("wishlist not found")
	ErrWishListForbidden         = errors.New("not authorized to access this wishlist")
	ErrItemNotFound              = errors.New("item not found")
	ErrItemForbidden             = errors.New("not authorized to access this item")
)

// WishListRepositoryInterface defines what the wishlist_item service needs from wishlist repository (cross-domain)
type WishListRepositoryInterface interface {
	GetByID(ctx context.Context, id pgtype.UUID) (*wishlistmodels.WishList, error)
}

// GiftItemRepositoryInterface defines what the wishlist_item service needs from item repository (cross-domain)
type GiftItemRepositoryInterface interface {
	GetByID(ctx context.Context, id pgtype.UUID) (*itemmodels.GiftItem, error)
	CreateWithOwner(ctx context.Context, giftItem itemmodels.GiftItem) (*itemmodels.GiftItem, error)
}

// Input/Output types

// CreateItemInput represents input for creating an item in a wishlist
type CreateItemInput struct {
	Title       string
	Description *string
	Link        *string
	ImageURL    *string
	Price       *float64
	Priority    *int32
	Notes       *string
}

// ItemOutput represents an item in service responses
type ItemOutput struct {
	ID          string
	OwnerID     string
	Name        string
	Description string
	Link        string
	ImageURL    string
	Price       float64
	Priority    int
	Notes       string
	IsPurchased bool
	IsReserved  bool
	IsArchived  bool
	CreatedAt   string
	UpdatedAt   string
}

// PaginatedItemsOutput represents paginated list of items
type PaginatedItemsOutput struct {
	Items      []*ItemOutput
	TotalCount int64
	Page       int
	Limit      int
	TotalPages int
}

// WishlistItemServiceInterface defines operations for wishlist-item relationships
type WishlistItemServiceInterface interface {
	GetWishlistItems(ctx context.Context, wishlistID string, userID string, page, limit int) (*PaginatedItemsOutput, error)
	AttachItem(ctx context.Context, wishlistID string, itemID string, userID string) error
	CreateItemInWishlist(ctx context.Context, wishlistID string, userID string, input CreateItemInput) (*ItemOutput, error)
	DetachItem(ctx context.Context, wishlistID string, itemID string, userID string) error
}

// WishlistItemService implements WishlistItemServiceInterface
type WishlistItemService struct {
	wishlistRepo     WishListRepositoryInterface
	itemRepo         GiftItemRepositoryInterface
	wishlistItemRepo repository.WishlistItemRepositoryInterface
}

// NewWishlistItemService creates a new WishlistItemService
func NewWishlistItemService(
	wishlistRepo WishListRepositoryInterface,
	itemRepo GiftItemRepositoryInterface,
	wishlistItemRepo repository.WishlistItemRepositoryInterface,
) *WishlistItemService {
	return &WishlistItemService{
		wishlistRepo:     wishlistRepo,
		itemRepo:         itemRepo,
		wishlistItemRepo: wishlistItemRepo,
	}
}

// GetWishlistItems retrieves all items in a wishlist with pagination
func (s *WishlistItemService) GetWishlistItems(ctx context.Context, wishlistID, userID string, page, limit int) (*PaginatedItemsOutput, error) {
	// Parse IDs
	wlID := pgtype.UUID{}
	if err := wlID.Scan(wishlistID); err != nil {
		return nil, ErrInvalidWishlistItemWLID
	}

	ownerID := pgtype.UUID{}
	if err := ownerID.Scan(userID); err != nil {
		return nil, ErrInvalidWishlistItemUser
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
func (s *WishlistItemService) AttachItem(ctx context.Context, wishlistID, itemID, userID string) error {
	// Parse IDs
	wlID := pgtype.UUID{}
	if err := wlID.Scan(wishlistID); err != nil {
		return ErrInvalidWishlistItemWLID
	}

	itID := pgtype.UUID{}
	if err := itID.Scan(itemID); err != nil {
		return ErrInvalidWishlistItemID
	}

	ownerID := pgtype.UUID{}
	if err := ownerID.Scan(userID); err != nil {
		return ErrInvalidWishlistItemUser
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
func (s *WishlistItemService) CreateItemInWishlist(ctx context.Context, wishlistID, userID string, input CreateItemInput) (*ItemOutput, error) {
	// Validate input
	if input.Title == "" {
		return nil, ErrWishlistItemTitleRequired
	}

	// Parse IDs
	wlID := pgtype.UUID{}
	if err := wlID.Scan(wishlistID); err != nil {
		return nil, ErrInvalidWishlistItemWLID
	}

	ownerID := pgtype.UUID{}
	if err := ownerID.Scan(userID); err != nil {
		return nil, ErrInvalidWishlistItemUser
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
	item := itemmodels.GiftItem{
		OwnerID: ownerID,
		Name:    input.Title,
	}

	if input.Description != nil && *input.Description != "" {
		item.Description = pgtype.Text{String: *input.Description, Valid: true}
	}
	if input.Link != nil && *input.Link != "" {
		item.Link = pgtype.Text{String: *input.Link, Valid: true}
	}
	if input.ImageURL != nil && *input.ImageURL != "" {
		item.ImageUrl = pgtype.Text{String: *input.ImageURL, Valid: true}
	}
	if input.Priority != nil {
		item.Priority = pgtype.Int4{Int32: *input.Priority, Valid: true}
	}
	if input.Notes != nil && *input.Notes != "" {
		item.Notes = pgtype.Text{String: *input.Notes, Valid: true}
	}

	// Set price if provided
	if input.Price != nil && *input.Price > 0 {
		if err := item.Price.Scan(fmt.Sprintf("%f", *input.Price)); err != nil {
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
func (s *WishlistItemService) DetachItem(ctx context.Context, wishlistID, itemID, userID string) error {
	// Parse IDs
	wlID := pgtype.UUID{}
	if err := wlID.Scan(wishlistID); err != nil {
		return ErrInvalidWishlistItemWLID
	}

	itID := pgtype.UUID{}
	if err := itID.Scan(itemID); err != nil {
		return ErrInvalidWishlistItemID
	}

	ownerID := pgtype.UUID{}
	if err := ownerID.Scan(userID); err != nil {
		return ErrInvalidWishlistItemUser
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

// Helper to convert itemmodels.GiftItem to ItemOutput
func (s *WishlistItemService) convertItemToOutput(item *itemmodels.GiftItem) *ItemOutput {
	output := &ItemOutput{
		ID:          item.ID.String(),
		OwnerID:     item.OwnerID.String(),
		Name:        item.Name,
		Description: "",
		Link:        "",
		ImageURL:    "",
		Price:       0,
		Priority:    0,
		Notes:       "",
		IsPurchased: item.PurchasedByUserID.Valid,
		IsReserved:  item.ReservedByUserID.Valid,
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
