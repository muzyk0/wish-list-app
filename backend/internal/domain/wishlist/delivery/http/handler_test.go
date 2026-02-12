package http

import (
	"bytes"
	"context"
	"encoding/json"
	nethttp "net/http"
	"net/http/httptest"
	"testing"

	"wish-list/internal/domain/wishlist/delivery/http/dto"
	"wish-list/internal/domain/wishlist/service"
	"wish-list/internal/pkg/validation"

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

// DefaultAuthContext returns a default authenticated user context for testing
func DefaultAuthContext() AuthContext {
	return AuthContext{
		UserID:   "123e4567-e89b-12d3-a456-426614174000",
		Email:    "test@example.com",
		UserType: "user",
	}
}

// SetAuthContext sets the authentication context on an Echo context
func SetAuthContext(c echo.Context, auth AuthContext) {
	c.Set("user_id", auth.UserID)
	c.Set("email", auth.Email)
	c.Set("user_type", auth.UserType)
}

// CreateTestContext creates an Echo context with optional auth context
func CreateTestContext(e *echo.Echo, method, path string, body any, auth *AuthContext) (echo.Context, *httptest.ResponseRecorder) {
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

	return c, rec
}

// CreateTestContextWithParams creates an Echo context with params and optional auth context
func CreateTestContextWithParams(e *echo.Echo, method, path string, body any, paramNames, paramValues []string, auth *AuthContext) (echo.Context, *httptest.ResponseRecorder) {
	c, rec := CreateTestContext(e, method, path, body, auth)
	c.SetParamNames(paramNames...)
	c.SetParamValues(paramValues...)
	return c, rec
}

// MockWishListService implements the WishListServiceInterface for testing
type MockWishListService struct {
	mock.Mock
}

func (m *MockWishListService) CreateWishList(ctx context.Context, userID string, input service.CreateWishListInput) (*service.WishListOutput, error) {
	args := m.Called(ctx, userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.WishListOutput), args.Error(1)
}

func (m *MockWishListService) GetWishList(ctx context.Context, wishListID string) (*service.WishListOutput, error) {
	args := m.Called(ctx, wishListID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.WishListOutput), args.Error(1)
}

func (m *MockWishListService) GetWishListByPublicSlug(ctx context.Context, publicSlug string) (*service.WishListOutput, error) {
	args := m.Called(ctx, publicSlug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.WishListOutput), args.Error(1)
}

func (m *MockWishListService) GetWishListsByOwner(ctx context.Context, userID string) ([]*service.WishListOutput, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*service.WishListOutput), args.Error(1)
}

func (m *MockWishListService) UpdateWishList(ctx context.Context, wishListID, userID string, input service.UpdateWishListInput) (*service.WishListOutput, error) {
	args := m.Called(ctx, wishListID, userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.WishListOutput), args.Error(1)
}

func (m *MockWishListService) DeleteWishList(ctx context.Context, wishListID, userID string) error {
	args := m.Called(ctx, wishListID, userID)
	return args.Error(0)
}

func (m *MockWishListService) CreateGiftItem(ctx context.Context, wishListID string, input service.CreateGiftItemInput) (*service.GiftItemOutput, error) {
	args := m.Called(ctx, wishListID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.GiftItemOutput), args.Error(1)
}

func (m *MockWishListService) GetGiftItem(ctx context.Context, giftItemID string) (*service.GiftItemOutput, error) {
	args := m.Called(ctx, giftItemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.GiftItemOutput), args.Error(1)
}

func (m *MockWishListService) GetGiftItemsByWishList(ctx context.Context, wishListID string) ([]*service.GiftItemOutput, error) {
	args := m.Called(ctx, wishListID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*service.GiftItemOutput), args.Error(1)
}

func (m *MockWishListService) UpdateGiftItem(ctx context.Context, giftItemID string, input service.UpdateGiftItemInput) (*service.GiftItemOutput, error) {
	args := m.Called(ctx, giftItemID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.GiftItemOutput), args.Error(1)
}

func (m *MockWishListService) DeleteGiftItem(ctx context.Context, giftItemID string) error {
	args := m.Called(ctx, giftItemID)
	return args.Error(0)
}

func (m *MockWishListService) MarkGiftItemAsPurchased(ctx context.Context, giftItemID, userID string, purchasedPrice float64) (*service.GiftItemOutput, error) {
	args := m.Called(ctx, giftItemID, userID, purchasedPrice)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.GiftItemOutput), args.Error(1)
}

// T029a: Unit tests for public wish list retrieval endpoint
func TestHandler_GetWishListByPublicSlug(t *testing.T) {
	t.Run("valid slug returns wish list", func(t *testing.T) {
		e := echo.New()
		mockService := new(MockWishListService)
		handler := NewHandler(mockService)

		expectedWishList := &service.WishListOutput{
			ID:          "123e4567-e89b-12d3-a456-426614174000",
			OwnerID:     "123e4567-e89b-12d3-a456-426614174001",
			Title:       "Birthday Wish List",
			Description: "My birthday gifts",
			PublicSlug:  "birthday-2026",
			IsPublic:    true,
		}

		mockService.On("GetWishListByPublicSlug", mock.Anything, "birthday-2026").
			Return(expectedWishList, nil)

		req := httptest.NewRequest(nethttp.MethodGet, "/public/wishlists/birthday-2026", nethttp.NoBody)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("slug")
		c.SetParamValues("birthday-2026")

		err := handler.GetWishListByPublicSlug(c)

		require.NoError(t, err)
		assert.Equal(t, nethttp.StatusOK, rec.Code)

		var response dto.WishListResponse
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, expectedWishList.ID, response.ID)
		assert.Equal(t, expectedWishList.Title, response.Title)
		assert.Equal(t, expectedWishList.PublicSlug, response.PublicSlug)

		mockService.AssertExpectations(t)
	})

	t.Run("invalid slug returns not found", func(t *testing.T) {
		e := echo.New()
		mockService := new(MockWishListService)
		handler := NewHandler(mockService)

		mockService.On("GetWishListByPublicSlug", mock.Anything, "non-existent-slug").
			Return((*service.WishListOutput)(nil), assert.AnError)

		req := httptest.NewRequest(nethttp.MethodGet, "/public/wishlists/non-existent-slug", nethttp.NoBody)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("slug")
		c.SetParamValues("non-existent-slug")

		err := handler.GetWishListByPublicSlug(c)

		require.NoError(t, err)
		assert.Equal(t, nethttp.StatusNotFound, rec.Code)

		var response map[string]string
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["error"], "wish list not found")

		mockService.AssertExpectations(t)
	})

	t.Run("deleted list returns not found", func(t *testing.T) {
		e := echo.New()
		mockService := new(MockWishListService)
		handler := NewHandler(mockService)

		mockService.On("GetWishListByPublicSlug", mock.Anything, "deleted-list").
			Return((*service.WishListOutput)(nil), assert.AnError)

		req := httptest.NewRequest(nethttp.MethodGet, "/public/wishlists/deleted-list", nethttp.NoBody)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("slug")
		c.SetParamValues("deleted-list")

		err := handler.GetWishListByPublicSlug(c)

		require.NoError(t, err)
		assert.Equal(t, nethttp.StatusNotFound, rec.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("public wish list with special characters in slug", func(t *testing.T) {
		e := echo.New()
		mockService := new(MockWishListService)
		handler := NewHandler(mockService)

		expectedWishList := &service.WishListOutput{
			ID:         "123e4567-e89b-12d3-a456-426614174000",
			Title:      "Владислав's Birthday",
			PublicSlug: "vladislavs-birthday-2026",
			IsPublic:   true,
		}

		mockService.On("GetWishListByPublicSlug", mock.Anything, "vladislavs-birthday-2026").
			Return(expectedWishList, nil)

		req := httptest.NewRequest(nethttp.MethodGet, "/public/wishlists/vladislavs-birthday-2026", nethttp.NoBody)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("slug")
		c.SetParamValues("vladislavs-birthday-2026")

		err := handler.GetWishListByPublicSlug(c)

		require.NoError(t, err)
		assert.Equal(t, nethttp.StatusOK, rec.Code)

		mockService.AssertExpectations(t)
	})
}

// T048a: Unit tests for wish list update/delete endpoints
func TestHandler_UpdateWishList(t *testing.T) {
	t.Run("owner can update own wishlist", func(t *testing.T) {
		e := setupTestEcho()
		mockService := new(MockWishListService)
		handler := NewHandler(mockService)

		authCtx := DefaultAuthContext()
		wishListID := "123e4567-e89b-12d3-a456-426614174000"

		title := "Updated Birthday List"
		description := "Updated description"
		reqBody := dto.UpdateWishListRequest{
			Title:       &title,
			Description: &description,
		}

		expectedWishList := &service.WishListOutput{
			ID:      wishListID,
			Title:   title,
			OwnerID: authCtx.UserID,
		}

		mockService.On("UpdateWishList", mock.Anything, wishListID, authCtx.UserID, mock.AnythingOfType("service.UpdateWishListInput")).
			Return(expectedWishList, nil)

		c, rec := CreateTestContextWithParams(e, nethttp.MethodPut, "/wishlists/"+wishListID, reqBody,
			[]string{"id"}, []string{wishListID}, &authCtx)

		err := handler.UpdateWishList(c)

		require.NoError(t, err)
		assert.Equal(t, nethttp.StatusOK, rec.Code)

		var response dto.WishListResponse
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, expectedWishList.Title, response.Title)

		mockService.AssertExpectations(t)
	})

	t.Run("unauthorized update returns error", func(t *testing.T) {
		e := setupTestEcho()
		mockService := new(MockWishListService)
		handler := NewHandler(mockService)

		wishListID := "123e4567-e89b-12d3-a456-426614174000"

		title := "Updated Birthday List"
		reqBody := dto.UpdateWishListRequest{
			Title: &title,
		}

		// No auth context
		c, rec := CreateTestContextWithParams(e, nethttp.MethodPut, "/wishlists/"+wishListID, reqBody,
			[]string{"id"}, []string{wishListID}, nil)

		err := handler.UpdateWishList(c)

		require.NoError(t, err)
		assert.Equal(t, nethttp.StatusUnauthorized, rec.Code)

		mockService.AssertNotCalled(t, "UpdateWishList")
	})

	t.Run("update with service error", func(t *testing.T) {
		e := setupTestEcho()
		mockService := new(MockWishListService)
		handler := NewHandler(mockService)

		authCtx := DefaultAuthContext()
		wishListID := "123e4567-e89b-12d3-a456-426614174000"

		title := "Updated Birthday List"
		reqBody := dto.UpdateWishListRequest{
			Title: &title,
		}

		mockService.On("UpdateWishList", mock.Anything, wishListID, authCtx.UserID, mock.AnythingOfType("service.UpdateWishListInput")).
			Return((*service.WishListOutput)(nil), assert.AnError)

		c, rec := CreateTestContextWithParams(e, nethttp.MethodPut, "/wishlists/"+wishListID, reqBody,
			[]string{"id"}, []string{wishListID}, &authCtx)

		err := handler.UpdateWishList(c)

		require.NoError(t, err)
		assert.Equal(t, nethttp.StatusInternalServerError, rec.Code)

		mockService.AssertExpectations(t)
	})
}

func TestHandler_DeleteWishList(t *testing.T) {
	t.Run("owner can delete own wishlist", func(t *testing.T) {
		e := echo.New()
		mockService := new(MockWishListService)
		handler := NewHandler(mockService)

		authCtx := DefaultAuthContext()
		wishListID := "123e4567-e89b-12d3-a456-426614174000"

		mockService.On("DeleteWishList", mock.Anything, wishListID, authCtx.UserID).
			Return(nil)

		c, rec := CreateTestContextWithParams(e, nethttp.MethodDelete, "/wishlists/"+wishListID, nil,
			[]string{"id"}, []string{wishListID}, &authCtx)

		err := handler.DeleteWishList(c)

		require.NoError(t, err)
		assert.Equal(t, nethttp.StatusNoContent, rec.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("unauthorized deletion returns error", func(t *testing.T) {
		e := echo.New()
		mockService := new(MockWishListService)
		handler := NewHandler(mockService)

		wishListID := "123e4567-e89b-12d3-a456-426614174000"

		// No auth context
		c, rec := CreateTestContextWithParams(e, nethttp.MethodDelete, "/wishlists/"+wishListID, nil,
			[]string{"id"}, []string{wishListID}, nil)

		err := handler.DeleteWishList(c)

		require.NoError(t, err)
		assert.Equal(t, nethttp.StatusUnauthorized, rec.Code)

		mockService.AssertNotCalled(t, "DeleteWishList")
	})

	t.Run("delete with service error", func(t *testing.T) {
		e := echo.New()
		mockService := new(MockWishListService)
		handler := NewHandler(mockService)

		authCtx := DefaultAuthContext()
		wishListID := "123e4567-e89b-12d3-a456-426614174000"

		mockService.On("DeleteWishList", mock.Anything, wishListID, authCtx.UserID).
			Return(assert.AnError)

		c, rec := CreateTestContextWithParams(e, nethttp.MethodDelete, "/wishlists/"+wishListID, nil,
			[]string{"id"}, []string{wishListID}, &authCtx)

		err := handler.DeleteWishList(c)

		require.NoError(t, err)
		assert.Equal(t, nethttp.StatusInternalServerError, rec.Code)

		mockService.AssertExpectations(t)
	})
}

// T048a: Additional authorization tests for wish list update/delete endpoints
func TestHandler_UpdateWishList_AuthorizationChecks(t *testing.T) {
	t.Run("update non-existent wishlist returns not found", func(t *testing.T) {
		e := setupTestEcho()
		mockService := new(MockWishListService)
		handler := NewHandler(mockService)

		title := "New Title"
		reqBody := dto.UpdateWishListRequest{
			Title: &title,
		}

		authCtx := DefaultAuthContext()

		mockService.On("UpdateWishList", mock.Anything, "non-existent-id", mock.Anything, mock.AnythingOfType("service.UpdateWishListInput")).
			Return((*service.WishListOutput)(nil), service.ErrWishListNotFound)

		c, rec := CreateTestContextWithParams(e, nethttp.MethodPut, "/wishlists/non-existent-id", reqBody,
			[]string{"id"}, []string{"non-existent-id"}, &authCtx)

		err := handler.UpdateWishList(c)

		require.NoError(t, err)
		assert.Equal(t, nethttp.StatusNotFound, rec.Code)

		mockService.AssertExpectations(t)
	})
}
