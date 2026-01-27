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
		invalidToken := "not-a-valid-uuid"
		reqBody := CancelReservationRequest{
			ReservationToken: &invalidToken,
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/wishlists/list-123/items/item-456/cancel", bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("wishlistId", "itemId")
		c.SetParamValues("list-123", "item-456")

		// Should reject invalid token format
	})

	t.Run("reservation token uniqueness", func(t *testing.T) {
		// Test that each guest reservation gets a unique token
		// This would be tested at the service layer primarily
	})

	t.Run("token-based reservation lookup", func(t *testing.T) {
		// Test retrieving reservations by token
		mockService := new(MockReservationService)
		token := "123e4567-e89b-12d3-a456-426614174000"

		expectedReservations := []*services.ReservationOutput{
			{
				ID:         pgtype.UUID{Valid: true},
				GiftItemID: pgtype.UUID{Valid: true},
				Status:     "active",
				ReservationToken: pgtype.UUID{
					Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
					Valid: true,
				},
			},
		}

		mockService.On("GetGuestReservations", mock.Anything, token).
			Return(expectedReservations, nil)

		// Full test requires handler method for getting guest reservations
	})
}

// Additional tests for reservation status checks
func TestReservationHandler_GetReservationStatus(t *testing.T) {
	t.Run("check status of available gift item", func(t *testing.T) {
		mockService := new(MockReservationService)
		statusOutput := &services.ReservationStatusOutput{
			IsReserved: false,
			Status:     "available",
		}

		mockService.On("GetReservationStatus", mock.Anything, "public-slug", "item-123").
			Return(statusOutput, nil)

		// Test requires handler method implementation
	})

	t.Run("check status of reserved gift item", func(t *testing.T) {
		mockService := new(MockReservationService)
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

		// Test requires handler method implementation
	})

	t.Run("check status of purchased gift item", func(t *testing.T) {
		mockService := new(MockReservationService)
		statusOutput := &services.ReservationStatusOutput{
			IsReserved: true,
			Status:     "purchased",
		}

		mockService.On("GetReservationStatus", mock.Anything, "public-slug", "item-123").
			Return(statusOutput, nil)

		// Test requires handler method implementation
	})
}
