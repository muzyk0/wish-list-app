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

// Sentinel errors for items
var (
	ErrItemNotFound      = errors.New("item not found")
	ErrItemForbidden     = errors.New("not authorized to access this item")
	ErrInvalidItemUser   = errors.New("invalid user id")
	ErrItemTitleRequired = errors.New("title is required")
)

// ItemServiceInterface defines the interface for item-related operations
type ItemServiceInterface interface {
	GetMyItems(ctx context.Context, userID string, filters repositories.ItemFilters) (*PaginatedItemsOutput, error)
	CreateItem(ctx context.Context, userID string, input CreateItemInput) (*ItemOutput, error)
	GetItem(ctx context.Context, itemID string, userID string) (*ItemOutput, error)
	UpdateItem(ctx context.Context, itemID string, userID string, input UpdateItemInput) (*ItemOutput, error)
	SoftDeleteItem(ctx context.Context, itemID string, userID string) error
	MarkPurchased(ctx context.Context, itemID string, userID string, purchasedPrice float64) (*ItemOutput, error)
}

// ItemService implements ItemServiceInterface
type ItemService struct {
	itemRepo         repositories.GiftItemRepositoryInterface
	wishlistItemRepo repositories.WishlistItemRepositoryInterface
}

// NewItemService creates a new ItemService
func NewItemService(
	itemRepo repositories.GiftItemRepositoryInterface,
	wishlistItemRepo repositories.WishlistItemRepositoryInterface,
) *ItemService {
	return &ItemService{
		itemRepo:         itemRepo,
		wishlistItemRepo: wishlistItemRepo,
	}
}

// Input/Output types

// CreateItemInput represents input for creating an item
type CreateItemInput struct {
	Title       string
	Description string
	Link        string
	ImageURL    string
	Price       float64
	Priority    int
	Notes       string
}

// UpdateItemInput represents input for updating an item
type UpdateItemInput struct {
	Title       *string
	Description *string
	Link        *string
	ImageURL    *string
	Price       *float64
	Priority    *int
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

// GetMyItems retrieves all items owned by the user with filters
func (s *ItemService) GetMyItems(ctx context.Context, userID string, filters repositories.ItemFilters) (*PaginatedItemsOutput, error) {
	// Parse user ID
	ownerID := pgtype.UUID{}
	if err := ownerID.Scan(userID); err != nil {
		return nil, ErrInvalidItemUser
	}

	// Set defaults
	if filters.Limit == 0 {
		filters.Limit = 10
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.Sort == "" {
		filters.Sort = "created_at"
	}
	if filters.Order == "" {
		filters.Order = "desc"
	}

	// Get items from repository
	result, err := s.itemRepo.GetByOwnerPaginated(ctx, ownerID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get items: %w", err)
	}

	// Convert to output
	items := make([]*ItemOutput, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, s.convertToOutput(item))
	}

	totalPages := int((result.TotalCount + int64(filters.Limit) - 1) / int64(filters.Limit))

	return &PaginatedItemsOutput{
		Items:      items,
		TotalCount: result.TotalCount,
		Page:       filters.Page,
		Limit:      filters.Limit,
		TotalPages: totalPages,
	}, nil
}

// CreateItem creates a new item without attaching it to a wishlist
func (s *ItemService) CreateItem(ctx context.Context, userID string, input CreateItemInput) (*ItemOutput, error) {
	// Validate input
	if input.Title == "" {
		return nil, ErrItemTitleRequired
	}

	// Parse user ID
	ownerID := pgtype.UUID{}
	if err := ownerID.Scan(userID); err != nil {
		return nil, ErrInvalidItemUser
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

	// Create in repository
	createdItem, err := s.itemRepo.Create(ctx, item)
	if err != nil {
		return nil, fmt.Errorf("failed to create item: %w", err)
	}

	return s.convertToOutput(createdItem), nil
}

// GetItem retrieves a specific item by ID
func (s *ItemService) GetItem(ctx context.Context, itemID string, userID string) (*ItemOutput, error) {
	// Parse IDs
	id := pgtype.UUID{}
	if err := id.Scan(itemID); err != nil {
		return nil, ErrItemNotFound
	}

	ownerID := pgtype.UUID{}
	if err := ownerID.Scan(userID); err != nil {
		return nil, ErrInvalidItemUser
	}

	// Get item from repository
	item, err := s.itemRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrItemNotFound
	}

	// Check ownership
	if item.OwnerID.Bytes != ownerID.Bytes {
		return nil, ErrItemForbidden
	}

	return s.convertToOutput(item), nil
}

// UpdateItem updates an existing item
func (s *ItemService) UpdateItem(ctx context.Context, itemID string, userID string, input UpdateItemInput) (*ItemOutput, error) {
	// Parse IDs
	id := pgtype.UUID{}
	if err := id.Scan(itemID); err != nil {
		return nil, ErrItemNotFound
	}

	ownerID := pgtype.UUID{}
	if err := ownerID.Scan(userID); err != nil {
		return nil, ErrInvalidItemUser
	}

	// Get existing item
	item, err := s.itemRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrItemNotFound
	}

	// Check ownership
	if item.OwnerID.Bytes != ownerID.Bytes {
		return nil, ErrItemForbidden
	}

	// Update fields
	if input.Title != nil {
		item.Name = *input.Title
	}
	if input.Description != nil {
		item.Description = pgtype.Text{String: *input.Description, Valid: *input.Description != ""}
	}
	if input.Link != nil {
		item.Link = pgtype.Text{String: *input.Link, Valid: *input.Link != ""}
	}
	if input.ImageURL != nil {
		item.ImageUrl = pgtype.Text{String: *input.ImageURL, Valid: *input.ImageURL != ""}
	}
	if input.Price != nil {
		if err := item.Price.Scan(fmt.Sprintf("%f", *input.Price)); err != nil {
			return nil, fmt.Errorf("invalid price: %w", err)
		}
	}
	if input.Priority != nil {
		item.Priority = pgtype.Int4{Int32: int32(*input.Priority), Valid: true}
	}
	if input.Notes != nil {
		item.Notes = pgtype.Text{String: *input.Notes, Valid: *input.Notes != ""}
	}

	// Update in repository
	updatedItem, err := s.itemRepo.UpdateWithNewSchema(ctx, item)
	if err != nil {
		return nil, fmt.Errorf("failed to update item: %w", err)
	}

	return s.convertToOutput(updatedItem), nil
}

// SoftDeleteItem marks an item as archived
func (s *ItemService) SoftDeleteItem(ctx context.Context, itemID string, userID string) error {
	// Parse IDs
	id := pgtype.UUID{}
	if err := id.Scan(itemID); err != nil {
		return ErrItemNotFound
	}

	ownerID := pgtype.UUID{}
	if err := ownerID.Scan(userID); err != nil {
		return ErrInvalidItemUser
	}

	// Get existing item
	item, err := s.itemRepo.GetByID(ctx, id)
	if err != nil {
		return ErrItemNotFound
	}

	// Check ownership
	if item.OwnerID.Bytes != ownerID.Bytes {
		return ErrItemForbidden
	}

	// Soft delete in repository
	if err := s.itemRepo.SoftDelete(ctx, id); err != nil {
		return fmt.Errorf("failed to archive item: %w", err)
	}

	return nil
}

// MarkPurchased marks an item as purchased with the actual price
func (s *ItemService) MarkPurchased(ctx context.Context, itemID string, userID string, purchasedPrice float64) (*ItemOutput, error) {
	// Parse IDs
	id := pgtype.UUID{}
	if err := id.Scan(itemID); err != nil {
		return nil, ErrItemNotFound
	}

	purchasedByUserID := pgtype.UUID{}
	if err := purchasedByUserID.Scan(userID); err != nil {
		return nil, ErrInvalidItemUser
	}

	// Get existing item
	item, err := s.itemRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrItemNotFound
	}

	// Update purchase fields
	item.PurchasedByUserID = purchasedByUserID
	item.PurchasedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}
	if err := item.PurchasedPrice.Scan(fmt.Sprintf("%f", purchasedPrice)); err != nil {
		return nil, fmt.Errorf("invalid purchased price: %w", err)
	}

	// Update in repository
	updatedItem, err := s.itemRepo.UpdateWithNewSchema(ctx, item)
	if err != nil {
		return nil, fmt.Errorf("failed to mark item as purchased: %w", err)
	}

	return s.convertToOutput(updatedItem), nil
}

// Helper function to convert db.GiftItem to ItemOutput
func (s *ItemService) convertToOutput(item *db.GiftItem) *ItemOutput {
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
