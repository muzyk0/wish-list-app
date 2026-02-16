package dto

import (
	"fmt"

	"wish-list/internal/domain/reservation/repository"
	"wish-list/internal/domain/reservation/service"
)

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
		ID: res.GiftItemID.String(),
	}
	if res.GiftItemName.Valid {
		itemSummary.Name = res.GiftItemName.String
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
		ID: res.WishlistID.String(),
	}
	if res.WishlistTitle.Valid {
		listSummary.Title = res.WishlistTitle.String
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

type UserReservationsResponse struct {
	Data       []ReservationDetailsResponse `json:"data" validate:"required"`
	Pagination any                          `json:"pagination" validate:"required"`
}
