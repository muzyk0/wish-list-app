package dto

import (
	"wish-list/internal/domain/wishlist_item/service"
)

// ItemResponse represents a gift item in API responses
type ItemResponse struct {
	ID                    string  `json:"id" validate:"required" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	OwnerID               string  `json:"owner_id" validate:"required" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440001"`
	Title                 string  `json:"title" validate:"required" example:"iPhone 15 Pro"`
	Description           string  `json:"description" example:"256GB, Blue Titanium"`
	Link                  string  `json:"link" example:"https://apple.com/iphone-15-pro"`
	ImageURL              string  `json:"image_url" example:"https://example.com/image.jpg"`
	Price                 float64 `json:"price" validate:"required" example:"999.99"`
	Priority              int     `json:"priority" validate:"required" example:"3"`
	Notes                 string  `json:"notes" example:"Preferred color: Blue"`
	IsPurchased           bool    `json:"is_purchased" validate:"required" example:"false"`
	IsReserved            bool    `json:"is_reserved" validate:"required" example:"false"`
	IsManuallyReserved    bool    `json:"is_manually_reserved" validate:"required" example:"false"`
	ManualReservedByName  string  `json:"manual_reserved_by_name" validate:"required" example:"Бабушка и дедушка"`
	ManualReservationNote string  `json:"manual_reservation_note" validate:"required" example:"Сказали что купят велосипед"`
	IsArchived            bool    `json:"is_archived" validate:"required" example:"false"`
	CreatedAt             string  `json:"created_at" validate:"required" format:"date-time" example:"2024-01-01T12:00:00Z"`
	UpdatedAt             string  `json:"updated_at" validate:"required" format:"date-time" example:"2024-01-01T12:00:00Z"`
}

// ItemResponseFromService converts service output to API response
func ItemResponseFromService(item *service.ItemOutput) ItemResponse {
	return ItemResponse{
		ID:                    item.ID,
		OwnerID:               item.OwnerID,
		Title:                 item.Name,
		Description:           item.Description,
		Link:                  item.Link,
		ImageURL:              item.ImageURL,
		Price:                 item.Price,
		Priority:              item.Priority,
		Notes:                 item.Notes,
		IsPurchased:           item.IsPurchased,
		IsReserved:            item.IsReserved,
		IsManuallyReserved:    item.IsManuallyReserved,
		ManualReservedByName:  item.ManualReservedByName,
		ManualReservationNote: item.ManualReservationNote,
		IsArchived:            item.IsArchived,
		CreatedAt:             item.CreatedAt,
		UpdatedAt:             item.UpdatedAt,
	}
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
