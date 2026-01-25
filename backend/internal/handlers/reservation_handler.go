package handlers

import (
	"fmt"
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
	GuestName  *string `json:"guestName" validate:"omitempty,max=200"`
	GuestEmail *string `json:"guestEmail" validate:"omitempty,email"`
}

type CancelReservationRequest struct {
	ReservationToken *string `json:"reservationToken" validate:"omitempty,uuid"`
}

type CreateReservationResponse struct {
	ID               string  `json:"id"`
	GiftItemID       string  `json:"giftItemId"`
	ReservedByUserID *string `json:"reservedByUserId"`
	GuestName        *string `json:"guestName"`
	GuestEmail       *string `json:"guestEmail"`
	ReservationToken string  `json:"reservationToken"`
	Status           string  `json:"status"`
	ReservedAt       string  `json:"reservedAt"`
	ExpiresAt        *string `json:"expiresAt"`
	CanceledAt       *string `json:"canceledAt"`
	CanceledReason   *string `json:"cancelReason"`
	NotificationSent bool    `json:"notificationSent"`
}

type ReservationDetailsResponse struct {
	ID         string          `json:"id"`
	GiftItem   GiftItemSummary `json:"giftItem"`
	Wishlist   WishListSummary `json:"wishlist"`
	Status     string          `json:"status"`
	ReservedAt string          `json:"reservedAt"`
	ExpiresAt  *string         `json:"expiresAt"`
}

type GiftItemSummary struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	ImageURL *string `json:"image_url,omitempty"`
	Price    *string `json:"price,omitempty"`
}

type WishListSummary struct {
	ID             string  `json:"id"`
	Title          string  `json:"title"`
	OwnerFirstName *string `json:"ownerFirstName,omitempty"`
	OwnerLastName  *string `json:"ownerLastName,omitempty"`
}

type ReservationStatusResponse struct {
	IsReserved     bool    `json:"isReserved"`
	ReservedByName *string `json:"reservedByName"`
	ReservedAt     *string `json:"reservedAt"`
	Status         string  `json:"status"`
}

func (h *ReservationHandler) CreateReservation(c echo.Context) error {
	wishListID := c.Param("wishlistId")
	giftItemID := c.Param("itemId")

	var req CreateReservationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
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

func (h *ReservationHandler) CancelReservation(c echo.Context) error {
	wishListID := c.Param("wishlistId")
	giftItemID := c.Param("itemId")

	var req CancelReservationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
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
	reservations, err := h.service.GetUserReservations(ctx, userID, limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Errorf("failed to get user reservations: %w", err).Error(),
		})
	}

	response := struct {
		Data       []ReservationDetailsResponse `json:"data"`
		Pagination interface{}                  `json:"pagination"`
	}{
		Data: []ReservationDetailsResponse{},
		Pagination: map[string]interface{}{
			"page":       page,
			"limit":      limit,
			"total":      len(reservations),
			"totalPages": 1, // Simplified for this example
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
			ID:    res.GiftItemID.String(), // This should be wishlist ID
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
			ID:    res.GiftItemID.String(), // This should be wishlist ID
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
