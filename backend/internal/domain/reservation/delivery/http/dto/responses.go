package dto

import (
	"fmt"

	"wish-list/internal/domain/reservation/repository"
	"wish-list/internal/domain/reservation/service"
)

// ErrorResponse represents a standard error API response.
type ErrorResponse struct {
	Error string `json:"error" validate:"required" example:"error message"`
}

type CreateReservationResponse struct {
	ID               string  `json:"id" validate:"required"`
	GiftItemID       string  `json:"gift_item_id" validate:"required"`
	ReservedByUserID *string `json:"reserved_by_user_id"`
	GuestName        *string `json:"guest_name"`
	GuestEmail       *string `json:"guest_email" validate:"email"`
	ReservationToken string  `json:"reservation_token" validate:"required"`
	Status           string  `json:"status" validate:"required"`
	ReservedAt       string  `json:"reserved_at" validate:"required"`
	ExpiresAt        *string `json:"expires_at"`
	CanceledAt       *string `json:"canceled_at"`
	CanceledReason   *string `json:"cancel_reason"`
	NotificationSent bool    `json:"notification_sent" validate:"required"`
}

func FromReservationOutput(r *service.ReservationOutput) *CreateReservationResponse {
	if r == nil {
		return nil
	}

	resp := &CreateReservationResponse{
		ID:               r.ID.String(),
		GiftItemID:       r.GiftItemID.String(),
		GuestName:        r.GuestName,
		GuestEmail:       r.GuestEmail,
		ReservationToken: r.ReservationToken.String(),
		Status:           r.Status,
		ReservedAt:       r.ReservedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		NotificationSent: r.NotificationSent.Bool,
	}

	if r.ReservedByUserID.Valid {
		userIDStr := r.ReservedByUserID.String()
		resp.ReservedByUserID = &userIDStr
	}

	if r.ExpiresAt.Valid {
		expiresAtStr := r.ExpiresAt.Time.Format("2006-01-02T15:04:05Z07:00")
		resp.ExpiresAt = &expiresAtStr
	}

	if r.CanceledAt.Valid {
		canceledAtStr := r.CanceledAt.Time.Format("2006-01-02T15:04:05Z07:00")
		resp.CanceledAt = &canceledAtStr
	}

	if r.CancelReason.Valid {
		reason := r.CancelReason.String
		resp.CanceledReason = &reason
	}

	return resp
}

type GiftItemSummary struct {
	ID       string  `json:"id" validate:"required"`
	Name     string  `json:"name" validate:"required"`
	ImageURL *string `json:"image_url,omitempty"`
	Price    *string `json:"price,omitempty"`
}

type WishListSummary struct {
	ID             string  `json:"id" validate:"required"`
	Title          string  `json:"title" validate:"required"`
	OwnerFirstName *string `json:"owner_first_name,omitempty"`
	OwnerLastName  *string `json:"owner_last_name,omitempty"`
}

type ReservationDetailsResponse struct {
	ID         string          `json:"id" validate:"required"`
	GiftItem   GiftItemSummary `json:"gift_item" validate:"required"`
	Wishlist   WishListSummary `json:"wishlist" validate:"required"`
	Status     string          `json:"status" validate:"required"`
	ReservedAt string          `json:"reserved_at" validate:"required"`
	ExpiresAt  *string         `json:"expires_at"`
}

func FromReservationDetail(res repository.ReservationDetail) ReservationDetailsResponse {
	itemSummary := GiftItemSummary{
		ID:   res.GiftItemID.String(),
		Name: res.GiftItemName.String,
	}
	if res.GiftItemImageURL.Valid {
		itemSummary.ImageURL = &res.GiftItemImageURL.String
	}
	if res.GiftItemPrice.Valid {
		priceFloat, err := res.GiftItemPrice.Float64Value()
		if err == nil {
			priceStr := fmt.Sprintf("%.2f", priceFloat.Float64)
			itemSummary.Price = &priceStr
		}
	}

	listSummary := WishListSummary{
		ID:    res.WishlistID.String(),
		Title: res.WishlistTitle.String,
	}
	if res.OwnerFirstName.Valid {
		listSummary.OwnerFirstName = &res.OwnerFirstName.String
	}
	if res.OwnerLastName.Valid {
		listSummary.OwnerLastName = &res.OwnerLastName.String
	}

	detail := ReservationDetailsResponse{
		ID:         res.ID.String(),
		GiftItem:   itemSummary,
		Wishlist:   listSummary,
		Status:     res.Status,
		ReservedAt: res.ReservedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}

	if res.ExpiresAt.Valid {
		expiresAtStr := res.ExpiresAt.Time.Format("2006-01-02T15:04:05Z07:00")
		detail.ExpiresAt = &expiresAtStr
	}

	return detail
}

func FromReservationDetails(details []repository.ReservationDetail) []ReservationDetailsResponse {
	responses := make([]ReservationDetailsResponse, 0, len(details))
	for _, d := range details {
		responses = append(responses, FromReservationDetail(d))
	}
	return responses
}

type ReservationStatusResponse struct {
	IsReserved     bool    `json:"is_reserved" validate:"required"`
	ReservedByName *string `json:"reserved_by_name"`
	ReservedAt     *string `json:"reserved_at"`
	Status         string  `json:"status" validate:"required"`
}

func FromReservationStatusOutput(s *service.ReservationStatusOutput) *ReservationStatusResponse {
	if s == nil {
		return nil
	}

	resp := &ReservationStatusResponse{
		IsReserved:     s.IsReserved,
		ReservedByName: s.ReservedByName,
		Status:         s.Status,
	}

	if s.ReservedAt != nil {
		reservedAtStr := s.ReservedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.ReservedAt = &reservedAtStr
	}

	return resp
}

// PaginationResponse is the standard pagination envelope returned by paginated endpoints.
type PaginationResponse struct {
	Page       int `json:"page" validate:"required" example:"1"`
	Limit      int `json:"limit" validate:"required" example:"10"`
	Total      int `json:"total" validate:"required" example:"42"`
	TotalPages int `json:"total_pages" validate:"required" example:"5"`
}

type UserReservationsResponse struct {
	Data       []ReservationDetailsResponse `json:"data" validate:"required"`
	Pagination PaginationResponse           `json:"pagination" validate:"required"`
}

// WishlistOwnerReservationResponse is the "My Wishes" view: items from the owner's wishlists
// that have been reserved. The identity of the reserver is intentionally hidden — only the
// fact that the item is reserved (and its status) is shown.
type WishlistOwnerReservationResponse struct {
	ID         string          `json:"id" validate:"required" format:"uuid"`
	GiftItem   GiftItemSummary `json:"gift_item" validate:"required"`
	Wishlist   WishListSummary `json:"wishlist" validate:"required"`
	Status     string          `json:"status" validate:"required"`
	ReservedAt string          `json:"reserved_at" validate:"required" format:"date-time"`
	ExpiresAt  *string         `json:"expires_at" format:"date-time"`
}

type WishlistOwnerReservationsResponse struct {
	Data       []WishlistOwnerReservationResponse `json:"data" validate:"required"`
	Pagination PaginationResponse                 `json:"pagination" validate:"required"`
}

// WishlistOwnerReservationsUnauthorizedResponse documents 401 response shape for owner reservations endpoint.
type WishlistOwnerReservationsUnauthorizedResponse struct {
	Error string `json:"error" validate:"required" example:"Unauthorized"`
}

// WishlistOwnerReservationsInternalResponse documents 500 response shape for owner reservations endpoint.
type WishlistOwnerReservationsInternalResponse struct {
	Error string `json:"error" validate:"required" example:"Failed to get wishlist owner reservations"`
}

func FromWishlistOwnerReservationDetail(res repository.ReservationDetail) WishlistOwnerReservationResponse {
	itemSummary := GiftItemSummary{
		ID:   res.GiftItemID.String(),
		Name: res.GiftItemName.String,
	}
	if res.GiftItemImageURL.Valid {
		itemSummary.ImageURL = &res.GiftItemImageURL.String
	}
	if res.GiftItemPrice.Valid {
		priceFloat, err := res.GiftItemPrice.Float64Value()
		if err == nil {
			priceStr := fmt.Sprintf("%.2f", priceFloat.Float64)
			itemSummary.Price = &priceStr
		}
	}

	listSummary := WishListSummary{
		ID:    res.WishlistID.String(),
		Title: res.WishlistTitle.String,
	}

	detail := WishlistOwnerReservationResponse{
		ID:         res.ID.String(),
		GiftItem:   itemSummary,
		Wishlist:   listSummary,
		Status:     res.Status,
		ReservedAt: res.ReservedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}

	if res.ExpiresAt.Valid {
		expiresAtStr := res.ExpiresAt.Time.Format("2006-01-02T15:04:05Z07:00")
		detail.ExpiresAt = &expiresAtStr
	}

	return detail
}

func FromWishlistOwnerReservationDetails(details []repository.ReservationDetail) []WishlistOwnerReservationResponse {
	responses := make([]WishlistOwnerReservationResponse, 0, len(details))
	for _, d := range details {
		responses = append(responses, FromWishlistOwnerReservationDetail(d))
	}
	return responses
}
