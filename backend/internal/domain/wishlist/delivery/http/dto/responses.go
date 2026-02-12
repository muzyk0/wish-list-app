package dto

import (
	"fmt"

	"wish-list/internal/domain/wishlist/service"
)

// WishListResponse is the handler-level DTO for wishlist data
type WishListResponse struct {
	ID           string `json:"id" validate:"required"`
	OwnerID      string `json:"owner_id" validate:"required"`
	Title        string `json:"title" validate:"required"`
	Description  string `json:"description"`
	Occasion     string `json:"occasion"`
	OccasionDate string `json:"occasion_date"`
	TemplateID   string `json:"template_id"`
	IsPublic     bool   `json:"is_public"`
	PublicSlug   string `json:"public_slug"`
	ViewCount    string `json:"view_count" validate:"required"`
	ItemCount    int    `json:"item_count" example:"5"`
	CreatedAt    string `json:"created_at" validate:"required"`
	UpdatedAt    string `json:"updated_at" validate:"required"`
}

func FromWishListOutput(wl *service.WishListOutput) *WishListResponse {
	if wl == nil {
		return nil
	}
	return &WishListResponse{
		ID:           wl.ID,
		OwnerID:      wl.OwnerID,
		Title:        wl.Title,
		Description:  wl.Description,
		Occasion:     wl.Occasion,
		OccasionDate: wl.OccasionDate,
		TemplateID:   wl.TemplateID,
		IsPublic:     wl.IsPublic,
		PublicSlug:   wl.PublicSlug,
		ViewCount:    fmt.Sprintf("%d", wl.ViewCount),
		ItemCount:    int(wl.ItemCount),
		CreatedAt:    wl.CreatedAt,
		UpdatedAt:    wl.UpdatedAt,
	}
}

func FromWishListOutputs(wishlists []*service.WishListOutput) []*WishListResponse {
	if wishlists == nil {
		return nil
	}
	responses := make([]*WishListResponse, len(wishlists))
	for i, wl := range wishlists {
		responses[i] = FromWishListOutput(wl)
	}
	return responses
}

// GiftItemResponse is the handler-level DTO for gift item data
type GiftItemResponse struct {
	ID                string  `json:"id" validate:"required"`
	WishlistID        string  `json:"wishlist_id" validate:"required"`
	Name              string  `json:"name" validate:"required"`
	Description       string  `json:"description"`
	Link              string  `json:"link"`
	ImageURL          string  `json:"image_url"`
	Price             float64 `json:"price"`
	Priority          int     `json:"priority"`
	ReservedByUserID  string  `json:"reserved_by_user_id"`
	ReservedAt        string  `json:"reserved_at"`
	PurchasedByUserID string  `json:"purchased_by_user_id"`
	PurchasedAt       string  `json:"purchased_at"`
	PurchasedPrice    float64 `json:"purchased_price"`
	Notes             string  `json:"notes"`
	Position          int     `json:"position"`
	CreatedAt         string  `json:"created_at" validate:"required"`
	UpdatedAt         string  `json:"updated_at" validate:"required"`
}

func FromGiftItemOutput(item *service.GiftItemOutput) *GiftItemResponse {
	if item == nil {
		return nil
	}
	return &GiftItemResponse{
		ID:                item.ID,
		WishlistID:        item.OwnerID,
		Name:              item.Name,
		Description:       item.Description,
		Link:              item.Link,
		ImageURL:          item.ImageURL,
		Price:             item.Price,
		Priority:          item.Priority,
		ReservedByUserID:  item.ReservedByUserID,
		ReservedAt:        item.ReservedAt,
		PurchasedByUserID: item.PurchasedByUserID,
		PurchasedAt:       item.PurchasedAt,
		PurchasedPrice:    item.PurchasedPrice,
		Notes:             item.Notes,
		Position:          item.Position,
		CreatedAt:         item.CreatedAt,
		UpdatedAt:         item.UpdatedAt,
	}
}

func FromGiftItemOutputs(items []*service.GiftItemOutput) []*GiftItemResponse {
	if items == nil {
		return nil
	}
	responses := make([]*GiftItemResponse, len(items))
	for i, item := range items {
		responses[i] = FromGiftItemOutput(item)
	}
	return responses
}

type GetGiftItemsResponse struct {
	Items []*GiftItemResponse `json:"items" validate:"required"`
	Total int                 `json:"total" validate:"required"`
	Page  int                 `json:"page" validate:"required"`
	Limit int                 `json:"limit" validate:"required"`
	Pages int                 `json:"pages" validate:"required"`
}
