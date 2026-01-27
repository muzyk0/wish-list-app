package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"wish-list/internal/repositories"
	"wish-list/internal/services"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockReservationService implements the ReservationServiceInterface for testing
type MockReservationService struct {
	mock.Mock
}

func (m *MockReservationService) CreateReservation(ctx context.Context, input services.CreateReservationInput) (*services.ReservationOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.ReservationOutput), args.Error(1)
}

func (m *MockReservationService) CancelReservation(ctx context.Context, input services.CancelReservationInput) (*services.ReservationOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.ReservationOutput), args.Error(1)
}

func (m *MockReservationService) GetReservationStatus(ctx context.Context, publicSlug, giftItemID string) (*services.ReservationStatusOutput, error) {
	args := m.Called(ctx, publicSlug, giftItemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.ReservationStatusOutput), args.Error(1)
}

func (m *MockReservationService) GetUserReservations(ctx context.Context, userID pgtype.UUID, limit, offset int) ([]repositories.ReservationDetail, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repositories.ReservationDetail), args.Error(1)
}

func (m *MockReservationService) GetGuestReservations(ctx context.Context, token pgtype.UUID) ([]repositories.ReservationDetail, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repositories.ReservationDetail), args.Error(1)
}

func (m *MockReservationService) CountUserReservations(ctx context.Context, userID pgtype.UUID) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

// T062a: Unit tests for reservation cancellation endpoint (valid cancellation, unauthorized cancellation)
func TestReservationHandler_CancelReservation(t *testing.T) {
	t.Run("valid cancellation by authenticated user", func(t *testing.T) {
		e := setupTestEcho()
		mockService := new(MockReservationService)
		handler := NewReservationHandler(mockService)

		userID := pgtype.UUID{
			Bytes: [16]byte{0x12, 0x3e, 0x45, 0x67, 0xe8, 0x9b, 0x12, 0xd3, 0xa4, 0x56, 0x42, 0x66, 0x14, 0x17, 0x40, 0x00},
			Valid: true,
		}

		expectedReservation := &services.ReservationOutput{
			ID:               pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true},
			GiftItemID:       pgtype.UUID{Valid: true},
			ReservedByUserID: userID,
			Status:           "cancelled",
			ReservedAt:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
			CanceledAt:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
			CancelReason:     pgtype.Text{String: "User cancelled reservation", Valid: true},
			NotificationSent: pgtype.Bool{Bool: false, Valid: true},
		}

		mockService.On("CancelReservation", mock.Anything, mock.AnythingOfType("services.CancelReservationInput")).
			Return(expectedReservation, nil)

		// Create request without token (authenticated user)
		req := CancelReservationRequest{}
		jsonBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/wishlists/list-123/items/item-456/cancel", bytes.NewReader(jsonBody))
		httpReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(httpReq, rec)
		c.SetParamNames("wishlistId", "itemId")
		c.SetParamValues("list-123", "item-456")

		// Set auth context
		c.Set("user_id", "123e4567-e89b-12d3-a456-426614174000")
		c.Set("email", "user@example.com")
		c.Set("user_type", "user")

		err := handler.CancelReservation(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response CreateReservationResponse
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "cancelled", response.Status)

		mockService.AssertExpectations(t)
	})

	t.Run("valid cancellation by guest with token", func(t *testing.T) {
		e := setupTestEcho()
		mockService := new(MockReservationService)
		handler := NewReservationHandler(mockService)

		tokenStr := "123e4567-e89b-12d3-a456-426614174000"
		req := CancelReservationRequest{
			ReservationToken: &tokenStr,
		}

		expectedReservation := &services.ReservationOutput{
			ID:         pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true},
			GiftItemID: pgtype.UUID{Valid: true},
			Status:     "cancelled",
			ReservationToken: pgtype.UUID{
				Bytes: [16]byte{0x12, 0x3e, 0x45, 0x67, 0xe8, 0x9b, 0x12, 0xd3, 0xa4, 0x56, 0x42, 0x66, 0x14, 0x17, 0x40, 0x00},
				Valid: true,
			},
			ReservedAt:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
			CanceledAt:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
			CancelReason:     pgtype.Text{String: "Guest cancelled reservation", Valid: true},
			NotificationSent: pgtype.Bool{Bool: false, Valid: true},
		}

		mockService.On("CancelReservation", mock.Anything, mock.AnythingOfType("services.CancelReservationInput")).
			Return(expectedReservation, nil)

		jsonBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/wishlists/list-123/items/item-456/cancel", bytes.NewReader(jsonBody))
		httpReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(httpReq, rec)
		c.SetParamNames("wishlistId", "itemId")
		c.SetParamValues("list-123", "item-456")

		err := handler.CancelReservation(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response CreateReservationResponse
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "cancelled", response.Status)

		mockService.AssertExpectations(t)
	})

	t.Run("unauthorized cancellation attempt", func(t *testing.T) {
		e := setupTestEcho()
		mockService := new(MockReservationService)
		handler := NewReservationHandler(mockService)

		// No token provided and no auth context
		req := CancelReservationRequest{}

		jsonBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/wishlists/list-123/items/item-456/cancel", bytes.NewReader(jsonBody))
		httpReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(httpReq, rec)
		c.SetParamNames("wishlistId", "itemId")
		c.SetParamValues("list-123", "item-456")

		err := handler.CancelReservation(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var response map[string]string
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["error"], "token is required")
	})

	t.Run("cancel non-existent reservation", func(t *testing.T) {
		e := setupTestEcho()
		mockService := new(MockReservationService)
		handler := NewReservationHandler(mockService)

		tokenStr := "123e4567-e89b-12d3-a456-426614174001"
		req := CancelReservationRequest{
			ReservationToken: &tokenStr,
		}

		mockService.On("CancelReservation", mock.Anything, mock.AnythingOfType("services.CancelReservationInput")).
			Return((*services.ReservationOutput)(nil), assert.AnError)

		jsonBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/wishlists/list-123/items/item-456/cancel", bytes.NewReader(jsonBody))
		httpReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(httpReq, rec)
		c.SetParamNames("wishlistId", "itemId")
		c.SetParamValues("list-123", "item-456")

		err := handler.CancelReservation(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("cancel already canceled reservation", func(t *testing.T) {
		// Idempotency test - should not error, just return current state
		t.Skip("Idempotency depends on service implementation - to be verified")
	})

	t.Run("cancel expired reservation", func(t *testing.T) {
		// Expired reservations should already be in a terminal state
		t.Skip("Expiration handling depends on service implementation - to be verified")
	})
}

// T068a: Unit tests for guest reservation token generation and validation
func TestReservationHandler_GuestReservationToken(t *testing.T) {
	t.Run("guest reservation generates unique token", func(t *testing.T) {
		e := setupTestEcho()
		mockService := new(MockReservationService)
		handler := NewReservationHandler(mockService)

		guestName := "John Doe"
		guestEmail := "john@example.com"
		reqBody := CreateReservationRequest{
			GuestName:  &guestName,
			GuestEmail: &guestEmail,
		}

		generatedToken := pgtype.UUID{
			Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			Valid: true,
		}

		expectedReservation := &services.ReservationOutput{
			ID: pgtype.UUID{
				Bytes: [16]byte{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17},
				Valid: true,
			},
			GiftItemID:       pgtype.UUID{Valid: true},
			GuestName:        &guestName,
			GuestEmail:       &guestEmail,
			ReservationToken: generatedToken,
			Status:           "active",
			ReservedAt: pgtype.Timestamptz{
				Time:  time.Now(),
				Valid: true,
			},
			NotificationSent: pgtype.Bool{Bool: false, Valid: true},
		}

		mockService.On("CreateReservation", mock.Anything, mock.AnythingOfType("services.CreateReservationInput")).
			Return(expectedReservation, nil)

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/wishlists/list-123/items/item-456/reserve", bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("wishlistId", "itemId")
		c.SetParamValues("list-123", "item-456")

		err := handler.CreateReservation(c)

		// Verify token is returned
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response CreateReservationResponse
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.NotEmpty(t, response.ReservationToken)
		assert.Equal(t, generatedToken.String(), response.ReservationToken)

		mockService.AssertExpectations(t)
	})

	t.Run("guest reservation requires name and email", func(t *testing.T) {
		e := setupTestEcho()
		mockService := new(MockReservationService)
		handler := NewReservationHandler(mockService)

		// Missing guest email
		guestName := "John Doe"
		reqBody := CreateReservationRequest{
			GuestName: &guestName,
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/wishlists/list-123/items/item-456/reserve", bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("wishlistId", "itemId")
		c.SetParamValues("list-123", "item-456")

		err := handler.CreateReservation(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var response map[string]string
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["error"], "Guest name and email are required")
	})

	t.Run("invalid reservation token format", func(t *testing.T) {
		// Test that invalid UUID format is rejected
		e := setupTestEcho()
		mockService := new(MockReservationService)
		handler := NewReservationHandler(mockService)

		invalidToken := "not-a-valid-uuid"
		reqBody := CancelReservationRequest{
			ReservationToken: &invalidToken,
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodDelete, "/api/reservations/wishlist/list-123/item/item-456", bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("wishlistId", "itemId")
		c.SetParamValues("list-123", "item-456")

		// Invoke handler - should reject invalid token format
		err := handler.CancelReservation(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var response map[string]string
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["error"], "UUID")
	})

	t.Run("reservation token uniqueness", func(t *testing.T) {
		t.Skip("Token uniqueness is tested in the service layer")
	})

	t.Run("token-based reservation lookup", func(t *testing.T) {
		// Test retrieving reservations by token
		e := setupTestEcho()
		mockService := new(MockReservationService)
		handler := NewReservationHandler(mockService)

		tokenStr := "123e4567-e89b-12d3-a456-426614174000"
		tokenUUID := pgtype.UUID{}
		err := tokenUUID.Scan(tokenStr)
		require.NoError(t, err)

		expectedReservations := []repositories.ReservationDetail{
			{
				ID:               pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true},
				GiftItemID:       pgtype.UUID{Bytes: [16]byte{2, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true},
				Status:           "active",
				ReservationToken: tokenUUID,
				ReservedAt:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
				GiftItemName:     pgtype.Text{String: "Test Item", Valid: true},
				WishlistID:       pgtype.UUID{Bytes: [16]byte{3, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true},
				WishlistTitle:    pgtype.Text{String: "Test Wishlist", Valid: true},
			},
		}

		mockService.On("GetGuestReservations", mock.Anything, tokenUUID).
			Return(expectedReservations, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/guest/reservations?token="+tokenStr, http.NoBody)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err = handler.GetGuestReservations(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response []ReservationDetailsResponse
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Len(t, response, 1)
		assert.Equal(t, "active", response[0].Status)
		assert.Equal(t, "Test Item", response[0].GiftItem.Name)
		assert.Equal(t, "Test Wishlist", response[0].Wishlist.Title)

		mockService.AssertExpectations(t)
	})
}

// Additional tests for reservation status checks
func TestReservationHandler_GetReservationStatus(t *testing.T) {
	t.Run("check status of available gift item", func(t *testing.T) {
		e := setupTestEcho()
		mockService := new(MockReservationService)
		handler := NewReservationHandler(mockService)

		statusOutput := &services.ReservationStatusOutput{
			IsReserved: false,
			Status:     "available",
		}

		mockService.On("GetReservationStatus", mock.Anything, "public-slug", "item-123").
			Return(statusOutput, nil)

		c, rec := CreateTestContextWithParams(e, http.MethodGet, "/wishlists/:slug/items/:itemId/status", nil,
			[]string{"slug", "itemId"}, []string{"public-slug", "item-123"}, nil)

		err := handler.GetReservationStatus(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response ReservationStatusResponse
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response.IsReserved)
		assert.Equal(t, "available", response.Status)
		assert.Nil(t, response.ReservedByName)
		assert.Nil(t, response.ReservedAt)

		mockService.AssertExpectations(t)
	})

	t.Run("check status of reserved gift item", func(t *testing.T) {
		e := setupTestEcho()
		mockService := new(MockReservationService)
		handler := NewReservationHandler(mockService)

		reservedBy := "Jane Doe"
		reservedAt := time.Now()
		statusOutput := &services.ReservationStatusOutput{
			IsReserved:     true,
			ReservedByName: &reservedBy,
			ReservedAt:     &reservedAt,
			Status:         "active",
		}

		mockService.On("GetReservationStatus", mock.Anything, "public-slug", "item-123").
			Return(statusOutput, nil)

		c, rec := CreateTestContextWithParams(e, http.MethodGet, "/wishlists/:slug/items/:itemId/status", nil,
			[]string{"slug", "itemId"}, []string{"public-slug", "item-123"}, nil)

		err := handler.GetReservationStatus(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response ReservationStatusResponse
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.IsReserved)
		assert.Equal(t, "active", response.Status)
		assert.NotNil(t, response.ReservedByName)
		assert.Equal(t, "Jane Doe", *response.ReservedByName)
		assert.NotNil(t, response.ReservedAt)

		mockService.AssertExpectations(t)
	})

	t.Run("check status of purchased gift item", func(t *testing.T) {
		e := setupTestEcho()
		mockService := new(MockReservationService)
		handler := NewReservationHandler(mockService)

		statusOutput := &services.ReservationStatusOutput{
			IsReserved: true,
			Status:     "purchased",
		}

		mockService.On("GetReservationStatus", mock.Anything, "public-slug", "item-123").
			Return(statusOutput, nil)

		c, rec := CreateTestContextWithParams(e, http.MethodGet, "/wishlists/:slug/items/:itemId/status", nil,
			[]string{"slug", "itemId"}, []string{"public-slug", "item-123"}, nil)

		err := handler.GetReservationStatus(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response ReservationStatusResponse
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.IsReserved)
		assert.Equal(t, "purchased", response.Status)

		mockService.AssertExpectations(t)
	})
}
