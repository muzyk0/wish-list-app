package dto

import (
	"wish-list/internal/domain/item/service"
)

// ItemResponse represents a gift item in API responses
type ItemResponse struct {
	ID          string  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	OwnerID     string  `json:"ownerId" example:"550e8400-e29b-41d4-a716-446655440001"`
	Title       string  `json:"title" example:"iPhone 15 Pro"`
	Description string  `json:"description" example:"256GB, Blue Titanium"`
	Link        string  `json:"link" example:"https://apple.com/iphone-15-pro"`
	ImageURL    string  `json:"imageUrl" example:"https://example.com/image.jpg"`
	Price       float64 `json:"price" example:"999.99"`
	Priority    int     `json:"priority" example:"3"`
	Notes       string  `json:"notes" example:"Preferred color: Blue"`
	IsPurchased bool    `json:"isPurchased" example:"false"`
	IsArchived  bool    `json:"isArchived" example:"false"`
	CreatedAt   string  `json:"createdAt" example:"2024-01-01T12:00:00Z"`
	UpdatedAt   string  `json:"updatedAt" example:"2024-01-01T12:00:00Z"`
}

// ItemResponseFromService converts service output to API response
func ItemResponseFromService(item *service.ItemOutput) ItemResponse {
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
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

// PaginatedItemsResponse represents paginated list of items
type PaginatedItemsResponse struct {
	Items      []ItemResponse `json:"items"`
	TotalCount int64          `json:"totalCount" example:"42"`
	Page       int            `json:"page" example:"1"`
	Limit      int            `json:"limit" example:"10"`
	TotalPages int            `json:"totalPages" example:"5"`
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
