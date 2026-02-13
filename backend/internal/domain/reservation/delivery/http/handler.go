package http

import (
	"errors"
	"fmt"
	"math"
	nethttp "net/http"

	"wish-list/internal/domain/reservation/delivery/http/dto"
	"wish-list/internal/domain/reservation/service"
	"wish-list/internal/pkg/auth"
	"wish-list/internal/pkg/helpers"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

// Handler handles HTTP requests for reservations
type Handler struct {
	service service.ReservationServiceInterface
}

// NewHandler creates a new Handler
func NewHandler(svc service.ReservationServiceInterface) *Handler {
	return &Handler{
		service: svc,
	}
}

// CreateReservation godoc
//
//	@Summary		Create a reservation for a gift item
//	@Description	Create a reservation for a gift item. Can be done by authenticated users or guests (with name and email).
//	@Tags			Reservations
//	@Accept			json
//	@Produce		json
//	@Param			wishlistId			path		string							true	"Wish List ID"
//	@Param			itemId				path		string							true	"Gift Item ID"
//	@Param			reservation_request	body		dto.CreateReservationRequest		false	"Reservation information (required for guests)"
//	@Success		200					{object}	dto.CreateReservationResponse	"Reservation created successfully"
//	@Failure		400					{object}	map[string]string				"Invalid request body or validation error"
//	@Failure		401					{object}	map[string]string				"Unauthorized (guests need name and email)"
//	@Failure		500					{object}	map[string]string				"Internal server error"
//	@Security		BearerAuth
//	@Router			/reservations/wishlist/{wishlistId}/item/{itemId} [post]
func (h *Handler) CreateReservation(c echo.Context) error {
	wishListID := c.Param("wishlistId")
	giftItemID := c.Param("itemId")

	var req dto.CreateReservationRequest
	if err := helpers.BindAndValidate(c, &req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	// Check if user is authenticated (NOT an error - used to detect guest vs authenticated)
	userIDStr, _, _, authErr := auth.GetUserFromContext(c)

	var reservation *service.ReservationOutput
	var err error

	if authErr == nil {
		// Authenticated user reservation
		userID, err := helpers.ParseUUID(c, userIDStr)
		if err != nil {
			return err
		}

		reservation, err = h.service.CreateReservation(ctx, req.ToServiceInput(wishListID, giftItemID, userID))
	} else {
		// Guest reservation
		if req.GuestName == nil || req.GuestEmail == nil {
			return c.JSON(nethttp.StatusBadRequest, map[string]string{
				"error": "Guest name and email are required for unauthenticated reservations",
			})
		}

		reservation, err = h.service.CreateReservation(ctx, req.ToServiceInput(wishListID, giftItemID, pgtype.UUID{Valid: false}))
	}

	if err != nil {
		if errors.Is(err, service.ErrInvalidGiftItemID) || errors.Is(err, service.ErrInvalidReservationWishlist) || errors.Is(err, service.ErrGuestInfoRequired) {
			return c.JSON(nethttp.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		}
		if errors.Is(err, service.ErrGiftItemNotInWishlist) {
			return c.JSON(nethttp.StatusNotFound, map[string]string{
				"error": err.Error(),
			})
		}
		if errors.Is(err, service.ErrGiftItemAlreadyReserved) {
			return c.JSON(nethttp.StatusConflict, map[string]string{
				"error": err.Error(),
			})
		}
		return c.JSON(nethttp.StatusInternalServerError, map[string]string{
			"error": fmt.Errorf("failed to create reservation: %w", err).Error(),
		})
	}

	return c.JSON(nethttp.StatusOK, dto.FromReservationOutput(reservation))
}

// CancelReservation godoc
//
//	@Summary		Cancel a reservation for a gift item
//	@Description	Cancel a reservation for a gift item. Can be done by authenticated users or guests (with reservation token).
//	@Tags			Reservations
//	@Accept			json
//	@Produce		json
//	@Param			wishlistId		path		string							true	"Wish List ID"
//	@Param			itemId			path		string							true	"Gift Item ID"
//	@Param			cancel_request	body		dto.CancelReservationRequest		false	"Cancellation information (required for guests)"
//	@Success		200				{object}	dto.CreateReservationResponse	"Reservation canceled successfully"
//	@Failure		400				{object}	map[string]string				"Invalid request body or validation error"
//	@Failure		401				{object}	map[string]string				"Unauthorized (guests need reservation token)"
//	@Failure		500				{object}	map[string]string				"Internal server error"
//	@Security		BearerAuth
//	@Router			/reservations/wishlist/{wishlistId}/item/{itemId} [delete]
func (h *Handler) CancelReservation(c echo.Context) error {
	wishListID := c.Param("wishlistId")
	giftItemID := c.Param("itemId")

	var req dto.CancelReservationRequest
	if err := helpers.BindAndValidate(c, &req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	// Check if user is authenticated (NOT an error - used to detect guest vs authenticated)
	userIDStr, _, _, authErr := auth.GetUserFromContext(c)

	var reservation *service.ReservationOutput
	var err error

	if authErr == nil {
		// Authenticated user cancellation
		userID, err := helpers.ParseUUID(c, userIDStr)
		if err != nil {
			return err
		}

		reservation, err = h.service.CancelReservation(ctx, service.CancelReservationInput{
			WishListID:       wishListID,
			GiftItemID:       giftItemID,
			UserID:           userID,
			ReservationToken: nil,
		})
	} else {
		// Guest cancellation with token
		if req.ReservationToken == nil {
			return c.JSON(nethttp.StatusBadRequest, map[string]string{
				"error": "Reservation token is required for unauthenticated cancellations",
			})
		}

		token, err := helpers.ParseUUID(c, *req.ReservationToken)
		if err != nil {
			return err
		}

		reservation, err = h.service.CancelReservation(ctx, service.CancelReservationInput{
			WishListID:       wishListID,
			GiftItemID:       giftItemID,
			UserID:           pgtype.UUID{Valid: false},
			ReservationToken: &token,
		})
	}

	if err != nil {
		if errors.Is(err, service.ErrInvalidGiftItemID) || errors.Is(err, service.ErrInvalidReservationWishlist) || errors.Is(err, service.ErrMissingUserOrToken) {
			return c.JSON(nethttp.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		}
		if errors.Is(err, service.ErrGiftItemNotInWishlist) || errors.Is(err, service.ErrReservationNotFound) {
			return c.JSON(nethttp.StatusNotFound, map[string]string{
				"error": err.Error(),
			})
		}
		return c.JSON(nethttp.StatusInternalServerError, map[string]string{
			"error": fmt.Errorf("failed to cancel reservation: %w", err).Error(),
		})
	}

	return c.JSON(nethttp.StatusOK, dto.FromReservationOutput(reservation))
}

// GetUserReservations godoc
//
//	@Summary		Get all reservations made by the authenticated user
//	@Description	Get all reservations made by the authenticated user with pagination.
//	@Tags			Reservations
//	@Produce		json
//	@Param			page	query		int								false	"Page number (default 1)"
//	@Param			limit	query		int								false	"Items per page (default 10, max 100)"
//	@Success		200		{object}	dto.UserReservationsResponse		"List of user reservations retrieved successfully"
//	@Failure		401		{object}	map[string]string				"Unauthorized"
//	@Failure		500		{object}	map[string]string				"Internal server error"
//	@Security		BearerAuth
//	@Router			/reservations/user [get]
func (h *Handler) GetUserReservations(c echo.Context) error {
	userIDStr := auth.MustGetUserID(c)
	pagination := helpers.ParsePagination(c)

	userID, err := helpers.ParseUUID(c, userIDStr)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	// Get total count for accurate pagination
	totalCount, err := h.service.CountUserReservations(ctx, userID)
	if err != nil {
		return c.JSON(nethttp.StatusInternalServerError, map[string]string{
			"error": fmt.Errorf("failed to count user reservations: %w", err).Error(),
		})
	}

	reservations, err := h.service.GetUserReservations(ctx, userID, pagination.Limit, pagination.Offset)
	if err != nil {
		return c.JSON(nethttp.StatusInternalServerError, map[string]string{
			"error": fmt.Errorf("failed to get user reservations: %w", err).Error(),
		})
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(pagination.Limit)))
	if totalPages == 0 {
		totalPages = 1
	}

	response := dto.UserReservationsResponse{
		Data: dto.FromReservationDetails(reservations),
		Pagination: map[string]any{
			"page":       pagination.Page,
			"limit":      pagination.Limit,
			"total":      totalCount,
			"totalPages": totalPages,
		},
	}

	return c.JSON(nethttp.StatusOK, response)
}

// GetGuestReservations godoc
//
//	@Summary		Get reservations made by a guest using a token
//	@Description	Get all reservations made by a guest using their reservation token.
//	@Tags			Reservations
//	@Produce		json
//	@Param			token	query		string								true	"Reservation token"
//	@Success		200		{array}		dto.ReservationDetailsResponse		"List of guest reservations retrieved successfully"
//	@Failure		400		{object}	map[string]string					"Invalid request parameters"
//	@Failure		500		{object}	map[string]string					"Internal server error"
//	@Router			/guest/reservations [get]
func (h *Handler) GetGuestReservations(c echo.Context) error {
	tokenStr := c.QueryParam("token")
	if tokenStr == "" {
		return c.JSON(nethttp.StatusBadRequest, map[string]string{
			"error": "Token parameter is required",
		})
	}

	token, err := helpers.ParseUUID(c, tokenStr)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()
	reservations, err := h.service.GetGuestReservations(ctx, token)
	if err != nil {
		return c.JSON(nethttp.StatusInternalServerError, map[string]string{
			"error": fmt.Errorf("failed to get guest reservations: %w", err).Error(),
		})
	}

	return c.JSON(nethttp.StatusOK, dto.FromReservationDetails(reservations))
}

// GetReservationStatus godoc
//
//	@Summary		Get the reservation status for a gift item in a public wish list
//	@Description	Get the reservation status for a specific gift item in a public wish list.
//	@Tags			Reservations
//	@Produce		json
//	@Param			slug	path		string							true	"Public wish list slug"
//	@Param			itemId	path		string							true	"Gift Item ID"
//	@Success		200		{object}	dto.ReservationStatusResponse	"Reservation status retrieved successfully"
//	@Failure		500		{object}	map[string]string				"Internal server error"
//	@Router			/public/reservations/list/{slug}/item/{itemId} [get]
func (h *Handler) GetReservationStatus(c echo.Context) error {
	publicSlug := c.Param("slug")
	giftItemID := c.Param("itemId")

	ctx := c.Request().Context()
	status, err := h.service.GetReservationStatus(ctx, publicSlug, giftItemID)
	if err != nil {
		return c.JSON(nethttp.StatusInternalServerError, map[string]string{
			"error": fmt.Errorf("failed to get reservation status: %w", err).Error(),
		})
	}

	return c.JSON(nethttp.StatusOK, dto.FromReservationStatusOutput(status))
}
