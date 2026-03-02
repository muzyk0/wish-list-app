package dto

import (
	"wish-list/internal/domain/item/service"
)

// ErrorResponse represents a standard error API response.
type ErrorResponse struct {
	Error string `json:"error" validate:"required" example:"error message"`
}

// ItemResponse represents a gift item in API responses
type ItemResponse struct {
	ID          string   `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	OwnerID     string   `json:"owner_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	Title       string   `json:"title" example:"iPhone 15 Pro"`
	Description string   `json:"description" example:"256GB, Blue Titanium"`
	Link        string   `json:"link" example:"https://apple.com/iphone-15-pro"`
	ImageURL    string   `json:"image_url" example:"https://example.com/image.jpg"`
	Price       float64  `json:"price" example:"999.99"`
	Priority    int      `json:"priority" example:"3"`
	Notes       string   `json:"notes" example:"Preferred color: Blue"`
	IsPurchased bool     `json:"is_purchased" example:"false"`
	IsArchived  bool     `json:"is_archived" example:"false"`
	WishlistIDs []string `json:"wishlist_ids" example:"550e8400-e29b-41d4-a716-446655440002"`
	CreatedAt   string   `json:"created_at" example:"2024-01-01T12:00:00Z"`
	UpdatedAt   string   `json:"updated_at" example:"2024-01-01T12:00:00Z"`
}

// ItemResponseFromService converts service output to API response
func ItemResponseFromService(item *service.ItemOutput) ItemResponse {
	wishlistIDs := item.WishlistIDs
	if wishlistIDs == nil {
		wishlistIDs = []string{}
	}
	return ItemResponse{
		ID:          item.ID,
		OwnerID:     item.OwnerID,
		Title:       item.Name,
		Description: item.Description,
		Link:        item.Link,
		ImageURL:    item.ImageURL,
		Price:       item.Price,
		Priority:    item.Priority,
		Notes:       item.Notes,
		IsPurchased: item.IsPurchased,
		IsArchived:  item.IsArchived,
		WishlistIDs: wishlistIDs,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

// HomeStatsResponse represents aggregate item counts for the home screen
type HomeStatsResponse struct {
	TotalItems int64 `json:"total_items" example:"12" validate:"required"`
	Reserved   int64 `json:"reserved" example:"3" validate:"required"`
	Purchased  int64 `json:"purchased" example:"1" validate:"required"`
}

// PaginatedItemsResponse represents paginated list of items
type PaginatedItemsResponse struct {
	Items      []ItemResponse `json:"items"`
	TotalCount int64          `json:"total_count" example:"42"`
	Page       int            `json:"page" example:"1"`
	Limit      int            `json:"limit" example:"10"`
	TotalPages int            `json:"total_pages" example:"5"`
}

// PaginatedItemsResponseFromService converts service output to API response
func PaginatedItemsResponseFromService(result *service.PaginatedItemsOutput) PaginatedItemsResponse {
	items := make([]ItemResponse, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, ItemResponseFromService(item))
	}
	return PaginatedItemsResponse{
		Items:      items,
		TotalCount: result.TotalCount,
		Page:       result.Page,
		Limit:      result.Limit,
		TotalPages: result.TotalPages,
	}
}
