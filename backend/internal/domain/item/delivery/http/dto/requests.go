package dto

import (
	"wish-list/internal/domain/item/service"
)

// CreateItemRequest represents the request to create a gift item
type CreateItemRequest struct {
	Title       string  `json:"title" validate:"required,min=1,max=255" example:"iPhone 15 Pro"`
	Description string  `json:"description" validate:"max=2000" example:"256GB, Blue Titanium"`
	Link        string  `json:"link" validate:"omitempty,url" example:"https://apple.com/iphone-15-pro"`
	ImageURL    string  `json:"image_url" validate:"omitempty,url" example:"https://example.com/image.jpg"`
	Price       float64 `json:"price" validate:"omitempty,gte=0" example:"999.99"`
	Priority    int32   `json:"priority" validate:"omitempty,gte=0,lte=10" example:"3"`
	Notes       string  `json:"notes" validate:"max=1000" example:"Preferred color: Blue"`
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

// UpdateItemRequest represents the request to update a gift item
type UpdateItemRequest struct {
	Title       *string  `json:"title" validate:"omitempty,min=1,max=255"`
	Description *string  `json:"description" validate:"omitempty,max=2000"`
	Link        *string  `json:"link" validate:"omitempty,url"`
	ImageURL    *string  `json:"image_url" validate:"omitempty,url"`
	Price       *float64 `json:"price" validate:"omitempty,gte=0"`
	Priority    *int32   `json:"priority" validate:"omitempty,gte=0,lte=10"`
	Notes       *string  `json:"notes" validate:"omitempty,max=1000"`
}

// ToDomain converts UpdateItemRequest to service input
func (r *UpdateItemRequest) ToDomain() service.UpdateItemInput {
	return service.UpdateItemInput{
		Title:       r.Title,
		Description: r.Description,
		Link:        r.Link,
		ImageURL:    r.ImageURL,
		Price:       r.Price,
		Priority:    r.Priority,
		Notes:       r.Notes,
	}
}

// MarkPurchasedRequest represents the request to mark item as purchased
type MarkPurchasedRequest struct {
	PurchasedPrice float64 `json:"purchased_price" validate:"required,gte=0" example:"899.99"`
}
