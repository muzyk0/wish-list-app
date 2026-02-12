package dto

import (
	"wish-list/internal/domain/reservation/service"

	"github.com/jackc/pgx/v5/pgtype"
)

type CreateReservationRequest struct {
	GuestName  *string `json:"guest_name" validate:"omitempty,max=200"`
	GuestEmail *string `json:"guest_email" validate:"omitempty,email"`
}

func (r *CreateReservationRequest) ToServiceInput(wishListID, giftItemID string, userID pgtype.UUID) service.CreateReservationInput {
	return service.CreateReservationInput{
		WishListID: wishListID,
		GiftItemID: giftItemID,
		UserID:     userID,
		GuestName:  r.GuestName,
		GuestEmail: r.GuestEmail,
	}
}

type CancelReservationRequest struct {
	ReservationToken *string `json:"reservation_token" validate:"omitempty,uuid"`
}
