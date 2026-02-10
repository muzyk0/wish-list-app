package handlers

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"wish-list/internal/auth"
	"wish-list/internal/services"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

type ReservationHandler struct {
	service services.ReservationServiceInterface
}

func NewReservationHandler(service services.ReservationServiceInterface) *ReservationHandler {
	return &ReservationHandler{
		service: service,
	}
}

type CreateReservationRequest struct {
	GuestName  *string `json:"guest_name" validate:"omitempty,max=200"`
	GuestEmail *string `json:"guest_email" validate:"omitempty,email"`
}

type CancelReservationRequest struct {
	ReservationToken *string `json:"reservation_token" validate:"omitempty,uuid"`
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

type ReservationDetailsResponse struct {
	ID         string          `json:"id" validate:"required"`
	GiftItem   GiftItemSummary `json:"gift_item" validate:"required"`
	Wishlist   WishListSummary `json:"wishlist" validate:"required"`
	Status     string          `json:"status" validate:"required"`
	ReservedAt string          `json:"reserved_at" validate:"required"`
	ExpiresAt  *string         `json:"expires_at"`
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

type ReservationStatusResponse struct {
	IsReserved     bool    `json:"is_reserved" validate:"required"`
	ReservedByName *string `json:"reserved_by_name"`
	ReservedAt     *string `json:"reserved_at"`
	Status         string  `json:"status" validate:"required"`
}

type UserReservationsResponse struct {
	Data       []ReservationDetailsResponse `json:"data" validate:"required"`
	Pagination any                          `json:"pagination" validate:"required"`
}

// CreateReservation godoc
//
//	@Summary		Create a reservation for a gift item
//	@Description	Create a reservation for a gift item. Can be done by authenticated users or guests (with name and email).
//	@Tags			Reservations
//	@Accept			json
//	@Produce		json
//	@Param			wishlistId			path		string						true	"Wish List ID"
//	@Param			itemId				path		string						true	"Gift Item ID"
//	@Param			reservation_request	body		CreateReservationRequest	false	"Reservation information (required for guests)"
//	@Success		200					{object}	CreateReservationResponse	"Reservation created successfully"
//	@Failure		400					{object}	map[string]string			"Invalid request body or validation error"
//	@Failure		401					{object}	map[string]string			"Unauthorized (guests need name and email)"
//	@Failure		500					{object}	map[string]string			"Internal server error"
//	@Security		BearerAuth
//	@Router			/reservations/wishlist/{wishlistId}/item/{itemId} [post]
func (h *ReservationHandler) CreateReservation(c echo.Context) error {
	wishListID := c.Param("wishlistId")
	giftItemID := c.Param("itemId")

	var req CreateReservationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}
	ctx := c.Request().Context()

	// Check if user is authenticated
	userIDStr, _, _, authErr := auth.GetUserFromContext(c)

	var reservation *services.ReservationOutput
	var err error

	if authErr == nil {
		// Parse the user ID string into a UUID
		userID := pgtype.UUID{}
		if err := userID.Scan(userIDStr); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Invalid user ID format",
			})
		}

		// Authenticated user reservation
		reservation, err = h.service.CreateReservation(ctx, services.CreateReservationInput{
			WishListID: wishListID,
			GiftItemID: giftItemID,
			UserID:     userID,
			GuestName:  nil,
			GuestEmail: nil,
		})
	} else {
		// Guest reservation
		if req.GuestName == nil || req.GuestEmail == nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Guest name and email are required for unauthenticated reservations",
			})
		}

		reservation, err = h.service.CreateReservation(ctx, services.CreateReservationInput{
			WishListID: wishListID,
			GiftItemID: giftItemID,
			UserID:     pgtype.UUID{Valid: false},
			GuestName:  req.GuestName,
			GuestEmail: req.GuestEmail,
		})
	}

	if err != nil {
		if errors.Is(err, services.ErrInvalidGiftItemID) || errors.Is(err, services.ErrInvalidReservationWishlist) || errors.Is(err, services.ErrGuestInfoRequired) {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		}
		if errors.Is(err, services.ErrGiftItemNotInWishlist) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": err.Error(),
			})
		}
		if errors.Is(err, services.ErrGiftItemAlreadyReserved) {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": err.Error(),
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Errorf("failed to create reservation: %w", err).Error(),
		})
	}

	response := CreateReservationResponse{
		ID:               reservation.ID.String(),
		GiftItemID:       reservation.GiftItemID.String(),
		ReservedByUserID: nil,
		GuestName:        reservation.GuestName,
		GuestEmail:       reservation.GuestEmail,
		ReservationToken: reservation.ReservationToken.String(),
		Status:           reservation.Status,
		ReservedAt:       reservation.ReservedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		ExpiresAt:        nil,
		CanceledAt:       nil,
		CanceledReason:   nil,
		NotificationSent: reservation.NotificationSent.Bool,
	}

	if reservation.ReservedByUserID.Valid {
		userIDStr := reservation.ReservedByUserID.String()
		response.ReservedByUserID = &userIDStr
	}

	if reservation.ExpiresAt.Valid {
		expiresAtStr := reservation.ExpiresAt.Time.Format("2006-01-02T15:04:05Z07:00")
		response.ExpiresAt = &expiresAtStr
	}

	if reservation.CanceledAt.Valid {
		canceledAtStr := reservation.CanceledAt.Time.Format("2006-01-02T15:04:05Z07:00")
		response.CanceledAt = &canceledAtStr
	}

	if reservation.CancelReason.Valid {
		reason := reservation.CancelReason.String
		response.CanceledReason = &reason
	}

	return c.JSON(http.StatusOK, response)
}

// CancelReservation godoc
//
//	@Summary		Cancel a reservation for a gift item
//	@Description	Cancel a reservation for a gift item. Can be done by authenticated users or guests (with reservation token).
//	@Tags			Reservations
//	@Accept			json
//	@Produce		json
//	@Param			wishlistId		path		string						true	"Wish List ID"
//	@Param			itemId			path		string						true	"Gift Item ID"
//	@Param			cancel_request	body		CancelReservationRequest	false	"Cancellation information (required for guests)"
//	@Success		200				{object}	CreateReservationResponse	"Reservation canceled successfully"
//	@Failure		400				{object}	map[string]string			"Invalid request body or validation error"
//	@Failure		401				{object}	map[string]string			"Unauthorized (guests need reservation token)"
//	@Failure		500				{object}	map[string]string			"Internal server error"
//	@Security		BearerAuth
//	@Router			/reservations/wishlist/{wishlistId}/item/{itemId} [delete]
func (h *ReservationHandler) CancelReservation(c echo.Context) error {
	wishListID := c.Param("wishlistId")
	giftItemID := c.Param("itemId")

	var req CancelReservationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	ctx := c.Request().Context()

	// Check if user is authenticated
	userIDStr, _, _, authErr := auth.GetUserFromContext(c)

	var reservation *services.ReservationOutput
	var err error

	if authErr == nil {
		// Parse the user ID string into a UUID
		userID := pgtype.UUID{}
		if err := userID.Scan(userIDStr); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Invalid user ID format",
			})
		}

		// Authenticated user cancellation
		reservation, err = h.service.CancelReservation(ctx, services.CancelReservationInput{
			WishListID:       wishListID,
			GiftItemID:       giftItemID,
			UserID:           userID,
			ReservationToken: nil,
		})
	} else {
		// Guest cancellation with token
		if req.ReservationToken == nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Reservation token is required for unauthenticated cancellations",
			})
		}

		token := pgtype.UUID{}
		if err := token.Scan(*req.ReservationToken); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid reservation token format",
			})
		}

		reservation, err = h.service.CancelReservation(ctx, services.CancelReservationInput{
			WishListID:       wishListID,
			GiftItemID:       giftItemID,
			UserID:           pgtype.UUID{Valid: false},
			ReservationToken: &token,
		})
	}

	if err != nil {
		if errors.Is(err, services.ErrInvalidGiftItemID) || errors.Is(err, services.ErrInvalidReservationWishlist) || errors.Is(err, services.ErrMissingUserOrToken) {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		}
		if errors.Is(err, services.ErrGiftItemNotInWishlist) || errors.Is(err, services.ErrReservationNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": err.Error(),
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Errorf("failed to cancel reservation: %w", err).Error(),
		})
	}

	response := CreateReservationResponse{
		ID:               reservation.ID.String(),
		GiftItemID:       reservation.GiftItemID.String(),
		ReservedByUserID: nil,
		GuestName:        reservation.GuestName,
		GuestEmail:       reservation.GuestEmail,
		ReservationToken: reservation.ReservationToken.String(),
		Status:           reservation.Status,
		ReservedAt:       reservation.ReservedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		ExpiresAt:        nil,
		CanceledAt:       nil,
		CanceledReason:   nil,
		NotificationSent: reservation.NotificationSent.Bool,
	}

	if reservation.ReservedByUserID.Valid {
		userIDStr := reservation.ReservedByUserID.String()
		response.ReservedByUserID = &userIDStr
	}

	if reservation.ExpiresAt.Valid {
		expiresAtStr := reservation.ExpiresAt.Time.Format("2006-01-02T15:04:05Z07:00")
		response.ExpiresAt = &expiresAtStr
	}

	if reservation.CanceledAt.Valid {
		canceledAtStr := reservation.CanceledAt.Time.Format("2006-01-02T15:04:05Z07:00")
		response.CanceledAt = &canceledAtStr
	}

	if reservation.CancelReason.Valid {
		reason := reservation.CancelReason.String
		response.CanceledReason = &reason
	}

	return c.JSON(http.StatusOK, response)
}

// GetUserReservations godoc
//
//	@Summary		Get all reservations made by the authenticated user
//	@Description	Get all reservations made by the authenticated user with pagination.
//	@Tags			Reservations
//	@Produce		json
//	@Param			page	query		int							false	"Page number (default 1)"
//	@Param			limit	query		int							false	"Items per page (default 10, max 100)"
//	@Success		200		{object}	UserReservationsResponse	"List of user reservations retrieved successfully"
//	@Failure		401		{object}	map[string]string			"Unauthorized"
//	@Failure		500		{object}	map[string]string			"Internal server error"
//	@Security		BearerAuth
//	@Router			/reservations/user [get]
func (h *ReservationHandler) GetUserReservations(c echo.Context) error {
	pageStr := c.QueryParam("page")
	limitStr := c.QueryParam("limit")

	page := 1
	if pageStr != "" {
		if parsedPage, err := strconv.Atoi(pageStr); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	limit := 10
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	offset := (page - 1) * limit

	// Get user from context
	userIDStr, _, _, err := auth.GetUserFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Unauthorized",
		})
	}

	// Parse the user ID string into a UUID
	userID := pgtype.UUID{}
	if err := userID.Scan(userIDStr); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Invalid user ID format",
		})
	}

	ctx := c.Request().Context()

	// Get total count for accurate pagination
	totalCount, err := h.service.CountUserReservations(ctx, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Errorf("failed to count user reservations: %w", err).Error(),
		})
	}

	reservations, err := h.service.GetUserReservations(ctx, userID, limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Errorf("failed to get user reservations: %w", err).Error(),
		})
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))
	if totalPages == 0 {
		totalPages = 1
	}

	response := UserReservationsResponse{
		Data: []ReservationDetailsResponse{},
		Pagination: map[string]any{
			"page":       page,
			"limit":      limit,
			"total":      totalCount,
			"totalPages": totalPages,
		},
	}

	for _, res := range reservations {
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

		detailResponse := ReservationDetailsResponse{
			ID:         res.ID.String(),
			GiftItem:   itemSummary,
			Wishlist:   listSummary,
			Status:     res.Status,
			ReservedAt: res.ReservedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		}

		if res.ExpiresAt.Valid {
			expiresAtStr := res.ExpiresAt.Time.Format("2006-01-02T15:04:05Z07:00")
			detailResponse.ExpiresAt = &expiresAtStr
		}

		response.Data = append(response.Data, detailResponse)
	}

	return c.JSON(http.StatusOK, response)
}

// GetGuestReservations godoc
//
//	@Summary		Get reservations made by a guest using a token
//	@Description	Get all reservations made by a guest using their reservation token.
//	@Tags			Reservations
//	@Produce		json
//	@Param			token	query		string						true	"Reservation token"
//	@Success		200		{array}		ReservationDetailsResponse	"List of guest reservations retrieved successfully"
//	@Failure		400		{object}	map[string]string			"Invalid request parameters"
//	@Failure		500		{object}	map[string]string			"Internal server error"
//	@Router			/guest/reservations [get]
func (h *ReservationHandler) GetGuestReservations(c echo.Context) error {
	tokenStr := c.QueryParam("token")
	if tokenStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Token parameter is required",
		})
	}

	token := pgtype.UUID{}
	if err := token.Scan(tokenStr); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid reservation token format",
		})
	}

	ctx := c.Request().Context()
	reservations, err := h.service.GetGuestReservations(ctx, token)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Errorf("failed to get guest reservations: %w", err).Error(),
		})
	}

	response := []ReservationDetailsResponse{}

	for _, res := range reservations {
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

		detailResponse := ReservationDetailsResponse{
			ID:         res.ID.String(),
			GiftItem:   itemSummary,
			Wishlist:   listSummary,
			Status:     res.Status,
			ReservedAt: res.ReservedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		}

		if res.ExpiresAt.Valid {
			expiresAtStr := res.ExpiresAt.Time.Format("2006-01-02T15:04:05Z07:00")
			detailResponse.ExpiresAt = &expiresAtStr
		}

		response = append(response, detailResponse)
	}

	return c.JSON(http.StatusOK, response)
}

// GetReservationStatus godoc
//
//	@Summary		Get the reservation status for a gift item in a public wish list
//	@Description	Get the reservation status for a specific gift item in a public wish list.
//	@Tags			Reservations
//	@Produce		json
//	@Param			slug	path		string						true	"Public wish list slug"
//	@Param			itemId	path		string						true	"Gift Item ID"
//	@Success		200		{object}	ReservationStatusResponse	"Reservation status retrieved successfully"
//	@Failure		500		{object}	map[string]string			"Internal server error"
//	@Router			/public/reservations/list/{slug}/item/{itemId} [get]
func (h *ReservationHandler) GetReservationStatus(c echo.Context) error {
	publicSlug := c.Param("slug")
	giftItemID := c.Param("itemId")

	ctx := c.Request().Context()
	status, err := h.service.GetReservationStatus(ctx, publicSlug, giftItemID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Errorf("failed to get reservation status: %w", err).Error(),
		})
	}

	response := ReservationStatusResponse{
		IsReserved: status.IsReserved,
		Status:     status.Status,
	}

	if status.ReservedByName != nil {
		response.ReservedByName = status.ReservedByName
	}

	if status.ReservedAt != nil {
		reservedAtStr := status.ReservedAt.Format("2006-01-02T15:04:05Z07:00")
		response.ReservedAt = &reservedAtStr
	}

	return c.JSON(http.StatusOK, response)
}
