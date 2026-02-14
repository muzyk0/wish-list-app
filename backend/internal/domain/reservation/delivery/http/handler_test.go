package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	nethttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"wish-list/internal/domain/reservation/delivery/http/dto"
	"wish-list/internal/domain/reservation/repository"
	"wish-list/internal/domain/reservation/service"
	"wish-list/internal/pkg/validation"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// setupTestEcho creates a new Echo instance with validator for testing
func setupTestEcho() *echo.Echo {
	e := echo.New()
	e.Validator = validation.NewValidator()
	return e
}

// AuthContext contains authentication context for testing
type AuthContext struct {
	UserID   string
	Email    string
	UserType string
}

// SetAuthContext sets the authentication context on an Echo context
func SetAuthContext(c echo.Context, auth AuthContext) {
	c.Set("user_id", auth.UserID)
	c.Set("email", auth.Email)
	c.Set("user_type", auth.UserType)
}

// CreateTestContextWithParams creates an Echo context with params and optional auth context
func CreateTestContextWithParams(e *echo.Echo, method, path string, body any, paramNames, paramValues []string, auth *AuthContext) (echo.Context, *httptest.ResponseRecorder) {
	var req *nethttp.Request
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		req = httptest.NewRequest(method, path, bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	} else {
		req = httptest.NewRequest(method, path, nethttp.NoBody)
	}

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if auth != nil {
		SetAuthContext(c, *auth)
	}

	c.SetParamNames(paramNames...)
	c.SetParamValues(paramValues...)
	return c, rec
}

// MockReservationService implements the ReservationServiceInterface for testing
type MockReservationService struct {
	mock.Mock
}

func (m *MockReservationService) CreateReservation(ctx context.Context, input service.CreateReservationInput) (*service.ReservationOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.ReservationOutput), args.Error(1)
}

func (m *MockReservationService) CancelReservation(ctx context.Context, input service.CancelReservationInput) (*service.ReservationOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.ReservationOutput), args.Error(1)
}

func (m *MockReservationService) GetReservationStatus(ctx context.Context, publicSlug, giftItemID string) (*service.ReservationStatusOutput, error) {
	args := m.Called(ctx, publicSlug, giftItemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.ReservationStatusOutput), args.Error(1)
}

func (m *MockReservationService) GetUserReservations(ctx context.Context, userID pgtype.UUID, limit, offset int) ([]repository.ReservationDetail, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repository.ReservationDetail), args.Error(1)
}

func (m *MockReservationService) GetGuestReservations(ctx context.Context, token pgtype.UUID) ([]repository.ReservationDetail, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repository.ReservationDetail), args.Error(1)
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
		handler := NewHandler(mockService)

		userID := pgtype.UUID{
			Bytes: [16]byte{0x12, 0x3e, 0x45, 0x67, 0xe8, 0x9b, 0x12, 0xd3, 0xa4, 0x56, 0x42, 0x66, 0x14, 0x17, 0x40, 0x00},
			Valid: true,
		}

		expectedReservation := &service.ReservationOutput{
			ID:               pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true},
			GiftItemID:       pgtype.UUID{Valid: true},
			ReservedByUserID: userID,
			Status:           "canceled",
			ReservedAt:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
			CanceledAt:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
			CancelReason:     pgtype.Text{String: "User canceled reservation", Valid: true},
			NotificationSent: pgtype.Bool{Bool: false, Valid: true},
		}

		mockService.On("CancelReservation", mock.Anything, mock.AnythingOfType("service.CancelReservationInput")).
			Return(expectedReservation, nil)

		// Create request without token (authenticated user)
		req := dto.CancelReservationRequest{}
		jsonBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(nethttp.MethodPost, "/wishlists/list-123/items/item-456/cancel", bytes.NewReader(jsonBody))
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
		assert.Equal(t, nethttp.StatusOK, rec.Code)

		var response dto.CreateReservationResponse
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "canceled", response.Status)

		mockService.AssertExpectations(t)
	})

	t.Run("valid cancellation by guest with token", func(t *testing.T) {
		e := setupTestEcho()
		mockService := new(MockReservationService)
		handler := NewHandler(mockService)

		tokenStr := "123e4567-e89b-12d3-a456-426614174000" // #nosec G101 -- test value, not a credential
		req := dto.CancelReservationRequest{
			ReservationToken: &tokenStr,
		}

		expectedReservation := &service.ReservationOutput{
			ID:         pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true},
			GiftItemID: pgtype.UUID{Valid: true},
			Status:     "canceled",
			ReservationToken: pgtype.UUID{
				Bytes: [16]byte{0x12, 0x3e, 0x45, 0x67, 0xe8, 0x9b, 0x12, 0xd3, 0xa4, 0x56, 0x42, 0x66, 0x14, 0x17, 0x40, 0x00},
				Valid: true,
			},
			ReservedAt:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
			CanceledAt:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
			CancelReason:     pgtype.Text{String: "Guest canceled reservation", Valid: true},
			NotificationSent: pgtype.Bool{Bool: false, Valid: true},
		}

		mockService.On("CancelReservation", mock.Anything, mock.AnythingOfType("service.CancelReservationInput")).
			Return(expectedReservation, nil)

		jsonBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(nethttp.MethodPost, "/wishlists/list-123/items/item-456/cancel", bytes.NewReader(jsonBody))
		httpReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(httpReq, rec)
		c.SetParamNames("wishlistId", "itemId")
		c.SetParamValues("list-123", "item-456")

		err := handler.CancelReservation(c)

		require.NoError(t, err)
		assert.Equal(t, nethttp.StatusOK, rec.Code)

		var response dto.CreateReservationResponse
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "canceled", response.Status)

		mockService.AssertExpectations(t)
	})

	t.Run("unauthorized cancellation attempt", func(t *testing.T) {
		e := setupTestEcho()
		mockService := new(MockReservationService)
		handler := NewHandler(mockService)

		// No token provided and no auth context
		req := dto.CancelReservationRequest{}

		jsonBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(nethttp.MethodPost, "/wishlists/list-123/items/item-456/cancel", bytes.NewReader(jsonBody))
		httpReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(httpReq, rec)
		c.SetParamNames("wishlistId", "itemId")
		c.SetParamValues("list-123", "item-456")

		err := handler.CancelReservation(c)

		require.NoError(t, err)
		assert.Equal(t, nethttp.StatusBadRequest, rec.Code)

		var response map[string]string
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["error"], "token is required")
	})

	t.Run("cancel non-existent reservation", func(t *testing.T) {
		e := setupTestEcho()
		mockService := new(MockReservationService)
		handler := NewHandler(mockService)

		tokenStr := "123e4567-e89b-12d3-a456-426614174001" // #nosec G101 -- test value, not a credential
		req := dto.CancelReservationRequest{
			ReservationToken: &tokenStr,
		}

		mockService.On("CancelReservation", mock.Anything, mock.AnythingOfType("service.CancelReservationInput")).
			Return((*service.ReservationOutput)(nil), assert.AnError)

		jsonBody, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(nethttp.MethodPost, "/wishlists/list-123/items/item-456/cancel", bytes.NewReader(jsonBody))
		httpReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(httpReq, rec)
		c.SetParamNames("wishlistId", "itemId")
		c.SetParamValues("list-123", "item-456")

		err := handler.CancelReservation(c)

		require.NoError(t, err)
		assert.Equal(t, nethttp.StatusInternalServerError, rec.Code)

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
		handler := NewHandler(mockService)

		guestName := "John Doe"
		guestEmail := "john@example.com"
		reqBody := dto.CreateReservationRequest{
			GuestName:  &guestName,
			GuestEmail: &guestEmail,
		}

		generatedToken := pgtype.UUID{
			Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			Valid: true,
		}

		expectedReservation := &service.ReservationOutput{
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

		mockService.On("CreateReservation", mock.Anything, mock.AnythingOfType("service.CreateReservationInput")).
			Return(expectedReservation, nil)

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(nethttp.MethodPost, "/wishlists/list-123/items/item-456/reserve", bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("wishlistId", "itemId")
		c.SetParamValues("list-123", "item-456")

		err := handler.CreateReservation(c)

		// Verify token is returned
		require.NoError(t, err)
		assert.Equal(t, nethttp.StatusOK, rec.Code)

		var response dto.CreateReservationResponse
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.NotEmpty(t, response.ReservationToken)
		assert.Equal(t, generatedToken.String(), response.ReservationToken)

		mockService.AssertExpectations(t)
	})

	t.Run("guest reservation requires name and email", func(t *testing.T) {
		e := setupTestEcho()
		mockService := new(MockReservationService)
		handler := NewHandler(mockService)

		// Missing guest email
		guestName := "John Doe"
		reqBody := dto.CreateReservationRequest{
			GuestName: &guestName,
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(nethttp.MethodPost, "/wishlists/list-123/items/item-456/reserve", bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("wishlistId", "itemId")
		c.SetParamValues("list-123", "item-456")

		err := handler.CreateReservation(c)

		require.NoError(t, err)
		assert.Equal(t, nethttp.StatusBadRequest, rec.Code)

		var response map[string]string
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["error"], "Guest name and email are required")
	})

	t.Run("invalid reservation token format", func(t *testing.T) {
		// Test that invalid UUID format is rejected
		e := setupTestEcho()
		mockService := new(MockReservationService)
		handler := NewHandler(mockService)

		invalidToken := "not-a-valid-uuid" // #nosec G101 -- test value, not a credential
		reqBody := dto.CancelReservationRequest{
			ReservationToken: &invalidToken,
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(nethttp.MethodDelete, "/api/reservations/wishlist/list-123/item/item-456", bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("wishlistId", "itemId")
		c.SetParamValues("list-123", "item-456")

		// Invoke handler - should reject invalid token format
		err := handler.CancelReservation(c)
		require.Error(t, err, "Expected error for invalid UUID")

		var httpErr *echo.HTTPError
		require.True(t, errors.As(err, &httpErr), "Error should be echo.HTTPError")
		assert.Equal(t, nethttp.StatusBadRequest, httpErr.Code)
	})

	t.Run("reservation token uniqueness", func(t *testing.T) {
		t.Skip("Token uniqueness is tested in the service layer")
	})

	t.Run("token-based reservation lookup", func(t *testing.T) {
		// Test retrieving reservations by token
		e := setupTestEcho()
		mockService := new(MockReservationService)
		handler := NewHandler(mockService)

		tokenStr := "123e4567-e89b-12d3-a456-426614174000" // #nosec G101 -- test value, not a credential
		tokenUUID := pgtype.UUID{}
		err := tokenUUID.Scan(tokenStr)
		require.NoError(t, err)

		expectedReservations := []repository.ReservationDetail{
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

		req := httptest.NewRequest(nethttp.MethodGet, "/api/guest/reservations?token="+tokenStr, nethttp.NoBody)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err = handler.GetGuestReservations(c)
		require.NoError(t, err)
		assert.Equal(t, nethttp.StatusOK, rec.Code)

		var response []dto.ReservationDetailsResponse
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
		handler := NewHandler(mockService)

		statusOutput := &service.ReservationStatusOutput{
			IsReserved: false,
			Status:     "available",
		}

		mockService.On("GetReservationStatus", mock.Anything, "public-slug", "item-123").
			Return(statusOutput, nil)

		c, rec := CreateTestContextWithParams(e, nethttp.MethodGet, "/wishlists/:slug/items/:itemId/status", nil,
			[]string{"slug", "itemId"}, []string{"public-slug", "item-123"}, nil)

		err := handler.GetReservationStatus(c)
		require.NoError(t, err)
		assert.Equal(t, nethttp.StatusOK, rec.Code)

		var response dto.ReservationStatusResponse
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
		handler := NewHandler(mockService)

		reservedBy := "Jane Doe"
		reservedAt := time.Now()
		statusOutput := &service.ReservationStatusOutput{
			IsReserved:     true,
			ReservedByName: &reservedBy,
			ReservedAt:     &reservedAt,
			Status:         "active",
		}

		mockService.On("GetReservationStatus", mock.Anything, "public-slug", "item-123").
			Return(statusOutput, nil)

		c, rec := CreateTestContextWithParams(e, nethttp.MethodGet, "/wishlists/:slug/items/:itemId/status", nil,
			[]string{"slug", "itemId"}, []string{"public-slug", "item-123"}, nil)

		err := handler.GetReservationStatus(c)
		require.NoError(t, err)
		assert.Equal(t, nethttp.StatusOK, rec.Code)

		var response dto.ReservationStatusResponse
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
		handler := NewHandler(mockService)

		statusOutput := &service.ReservationStatusOutput{
			IsReserved: true,
			Status:     "purchased",
		}

		mockService.On("GetReservationStatus", mock.Anything, "public-slug", "item-123").
			Return(statusOutput, nil)

		c, rec := CreateTestContextWithParams(e, nethttp.MethodGet, "/wishlists/:slug/items/:itemId/status", nil,
			[]string{"slug", "itemId"}, []string{"public-slug", "item-123"}, nil)

		err := handler.GetReservationStatus(c)
		require.NoError(t, err)
		assert.Equal(t, nethttp.StatusOK, rec.Code)

		var response dto.ReservationStatusResponse
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response.IsReserved)
		assert.Equal(t, "purchased", response.Status)

		mockService.AssertExpectations(t)
	})
}
