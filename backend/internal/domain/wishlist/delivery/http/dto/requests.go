package dto

import "wish-list/internal/domain/wishlist/service"

type CreateWishListRequest struct {
	Title        string `json:"title" validate:"required,max=200"`
	Description  string `json:"description"`
	Occasion     string `json:"occasion"`
	OccasionDate string `json:"occasion_date"`
	IsPublic     bool   `json:"is_public"`
}

func (r *CreateWishListRequest) ToServiceInput() service.CreateWishListInput {
	return service.CreateWishListInput{
		Title:        r.Title,
		Description:  r.Description,
		Occasion:     r.Occasion,
		OccasionDate: r.OccasionDate,
		IsPublic:     r.IsPublic,
	}
}

type UpdateWishListRequest struct {
	Title        *string `json:"title" validate:"omitempty,max=200"`
	Description  *string `json:"description"`
	Occasion     *string `json:"occasion"`
	OccasionDate *string `json:"occasion_date"`
	IsPublic     *bool   `json:"is_public"`
	PublicSlug   *string `json:"public_slug" validate:"omitempty,max=100"`
}

func (r *UpdateWishListRequest) ToServiceInput() service.UpdateWishListInput {
	return service.UpdateWishListInput{
		Title:        r.Title,
		Description:  r.Description,
		Occasion:     r.Occasion,
		OccasionDate: r.OccasionDate,
		IsPublic:     r.IsPublic,
		PublicSlug:   r.PublicSlug,
	}
}

type CreateGiftItemRequest struct {
	Name        string  `json:"name" validate:"required,max=255"`
	Description string  `json:"description"`
	Link        string  `json:"link" validate:"omitempty,url"`
	ImageURL    string  `json:"image_url" validate:"omitempty,url"`
	Price       float64 `json:"price" validate:"omitempty,min=0"`
	Priority    int     `json:"priority" validate:"omitempty,min=0,max=10"`
	Notes       string  `json:"notes"`
	Position    int     `json:"position" validate:"omitempty,min=0"`
}

func (r *CreateGiftItemRequest) ToServiceInput() service.CreateGiftItemInput {
	return service.CreateGiftItemInput{
		Name:        r.Name,
		Description: r.Description,
		Link:        r.Link,
		ImageURL:    r.ImageURL,
		Price:       r.Price,
		Priority:    r.Priority,
		Notes:       r.Notes,
		Position:    r.Position,
	}
}

type UpdateGiftItemRequest struct {
	Name        *string  `json:"name" validate:"omitempty,max=255"`
	Description *string  `json:"description"`
	Link        *string  `json:"link" validate:"omitempty,url"`
	ImageURL    *string  `json:"image_url" validate:"omitempty,url"`
	Price       *float64 `json:"price" validate:"omitempty,min=0"`
	Priority    *int     `json:"priority" validate:"omitempty,min=0,max=10"`
	Notes       *string  `json:"notes"`
	Position    *int     `json:"position" validate:"omitempty,min=0"`
}

func (r *UpdateGiftItemRequest) ToServiceInput() service.UpdateGiftItemInput {
	return service.UpdateGiftItemInput{
		Name:        r.Name,
		Description: r.Description,
		Link:        r.Link,
		ImageURL:    r.ImageURL,
		Price:       r.Price,
		Priority:    r.Priority,
		Notes:       r.Notes,
		Position:    r.Position,
	}
}

type PurchaseRequest struct {
	PurchasedPrice float64 `json:"purchased_price"`
}
