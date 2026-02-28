package dto

import (
	"wish-list/internal/domain/wishlist_item/service"
)

// AttachItemRequest represents the request to attach an existing item to a wishlist
type AttachItemRequest struct {
	ItemID string `json:"item_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// CreateItemRequest represents the request to create a gift item in a wishlist
type CreateItemRequest struct {
	Title       string   `json:"title" validate:"required,min=1,max=255" example:"iPhone 15 Pro"`
	Description *string  `json:"description" validate:"omitempty,max=2000" example:"256GB, Blue Titanium"`
	Link        *string  `json:"link" validate:"omitempty,url" example:"https://apple.com/iphone-15-pro"`
	ImageURL    *string  `json:"image_url" validate:"omitempty,url" example:"https://example.com/image.jpg"`
	Price       *float64 `json:"price" validate:"omitempty,gte=0" example:"999.99"`
	Priority    *int32   `json:"priority" validate:"omitempty,gte=0,lte=10" example:"3"`
	Notes       *string  `json:"notes" validate:"omitempty,max=1000" example:"Preferred color: Blue"`
}

// MarkManualReservationRequest represents the request to manually mark a wishlist item as reserved
type MarkManualReservationRequest struct {
	ReservedByName string  `json:"reserved_by_name" validate:"required,min=1,max=255" example:"Бабушка и дедушка"`
	Note           *string `json:"note" validate:"omitempty,max=1000" example:"Сказали что купят велосипед"`
}

// ToDomain converts CreateItemRequest to service input
func (r *CreateItemRequest) ToDomain() service.CreateItemInput {
	return service.CreateItemInput{
		Title:       r.Title,
		Description: r.Description,
		Link:        r.Link,
		ImageURL:    r.ImageURL,
		Price:       r.Price,
		Priority:    r.Priority,
		Notes:       r.Notes,
	}
}
