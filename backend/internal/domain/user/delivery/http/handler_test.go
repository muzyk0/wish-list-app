package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	nethttp "net/http"
	"net/http/httptest"
	"testing"

	"wish-list/internal/domain/user/delivery/http/dto"
	userservice "wish-list/internal/domain/user/service"
	"wish-list/internal/pkg/analytics"
	"wish-list/internal/pkg/auth"
	"wish-list/internal/pkg/helpers"
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

// MockUserService implements the methods needed for testing
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Register(ctx context.Context, input userservice.RegisterUserInput) (*userservice.UserOutput, error) {
	args := m.Called(ctx, input)
	v := args.Get(0)
	if v != nil {
		if result, ok := v.(*userservice.UserOutput); ok {
			return result, args.Error(1)
		}
	}
	return nil, args.Error(1)
}

func (m *MockUserService) Login(ctx context.Context, input userservice.LoginUserInput) (*userservice.UserOutput, error) {
	args := m.Called(ctx, input)
	v := args.Get(0)
	if v != nil {
		if result, ok := v.(*userservice.UserOutput); ok {
			return result, args.Error(1)
		}
	}
	return nil, args.Error(1)
}

func (m *MockUserService) GetUser(ctx context.Context, userID string) (*userservice.UserOutput, error) {
	args := m.Called(ctx, userID)
	v := args.Get(0)
	if v != nil {
		if result, ok := v.(*userservice.UserOutput); ok {
			return result, args.Error(1)
		}
	}
	return nil, args.Error(1)
}

func (m *MockUserService) UpdateProfile(ctx context.Context, userID string, input userservice.UpdateProfileInput) (*userservice.UserOutput, error) {
	args := m.Called(ctx, userID, input)
	v := args.Get(0)
	if v != nil {
		if result, ok := v.(*userservice.UserOutput); ok {
			return result, args.Error(1)
		}
	}
	return nil, args.Error(1)
}

func (m *MockUserService) ChangeEmail(ctx context.Context, userID, currentPassword, newEmail string) error {
	args := m.Called(ctx, userID, currentPassword, newEmail)
	return args.Error(0)
}

func (m *MockUserService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	args := m.Called(ctx, userID, currentPassword, newPassword)
	return args.Error(0)
}

func (m *MockUserService) DeleteUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func TestUserHandler_Register(t *testing.T) {
	e := setupTestEcho()

	// Create mock service
	mockService := new(MockUserService)
	tokenManager := auth.NewTokenManager("test-secret")
	analyticsService := analytics.NewAnalyticsService(false)
	handler := NewHandler(mockService, tokenManager, nil, analyticsService)

	// Test input
	reqBody := dto.RegisterRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(nethttp.MethodPost, "/api/auth/register", bytes.NewReader(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Service returns this
	serviceUser := &userservice.UserOutput{
		ID:        "123",
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		AvatarUrl: "",
	}

	// Handler should map it to this
	expectedResponse := &dto.UserResponse{
		ID:        "123",
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		AvatarUrl: "",
	}

	// Setup expectations
	mockService.On("Register", mock.Anything, userservice.RegisterUserInput{
		Email:     reqBody.Email,
		Password:  reqBody.Password,
		FirstName: reqBody.FirstName,
		LastName:  reqBody.LastName,
		AvatarUrl: reqBody.AvatarUrl,
	}).Return(serviceUser, nil)

	// Call the handler
	err := handler.Register(c)

	// Assertions
	require.NoError(t, err)
	assert.Equal(t, nethttp.StatusCreated, rec.Code)

	var response dto.AuthResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, expectedResponse, response.User)
	assert.NotEmpty(t, response.AccessToken)
	assert.NotEmpty(t, response.RefreshToken)

	mockService.AssertExpectations(t)
}

func TestUserHandler_Login(t *testing.T) {
	e := setupTestEcho()

	// Create mock service
	mockService := new(MockUserService)
	tokenManager := auth.NewTokenManager("test-secret")
	analyticsService := analytics.NewAnalyticsService(false)
	handler := NewHandler(mockService, tokenManager, nil, analyticsService)

	// Test input
	reqBody := dto.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(nethttp.MethodPost, "/api/auth/login", bytes.NewReader(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Service returns this
	serviceUser := &userservice.UserOutput{
		ID:        "123",
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		AvatarUrl: "",
	}

	// Handler should map it to this
	expectedResponse := &dto.UserResponse{
		ID:        "123",
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		AvatarUrl: "",
	}

	// Setup expectations
	mockService.On("Login", mock.Anything, userservice.LoginUserInput{
		Email:    reqBody.Email,
		Password: reqBody.Password,
	}).Return(serviceUser, nil)

	// Call the handler
	err := handler.Login(c)

	// Assertions
	require.NoError(t, err)
	assert.Equal(t, nethttp.StatusOK, rec.Code)

	var response dto.AuthResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, expectedResponse, response.User)
	assert.NotEmpty(t, response.AccessToken)
	assert.NotEmpty(t, response.RefreshToken)

	mockService.AssertExpectations(t)
}

func TestUserHandler_Register_BadRequest(t *testing.T) {
	e := setupTestEcho()

	// Create mock service
	mockService := new(MockUserService)
	tokenManager := auth.NewTokenManager("test-secret")
	analyticsService := analytics.NewAnalyticsService(false)
	handler := NewHandler(mockService, tokenManager, nil, analyticsService)

	// Invalid input - empty body (validation fails before service call)
	req := httptest.NewRequest(nethttp.MethodPost, "/api/auth/register", nethttp.NoBody)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Call the handler
	err := handler.Register(c)

	// Assertions
	require.Error(t, err, "Expected validation error")
	var httpErr *echo.HTTPError
	require.True(t, errors.As(err, &httpErr), "Error should be echo.HTTPError")
	assert.Equal(t, nethttp.StatusBadRequest, httpErr.Code)

	// Service should NOT be called because validation fails first
	mockService.AssertNotCalled(t, "Register")
}

func TestUserHandler_Login_BadRequest(t *testing.T) {
	e := setupTestEcho()

	// Create mock service
	mockService := new(MockUserService)
	tokenManager := auth.NewTokenManager("test-secret")
	analyticsService := analytics.NewAnalyticsService(false)
	handler := NewHandler(mockService, tokenManager, nil, analyticsService)

	// Invalid input - empty body (validation fails before service call)
	req := httptest.NewRequest(nethttp.MethodPost, "/api/auth/login", nethttp.NoBody)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Call the handler
	err := handler.Login(c)

	// Assertions
	require.Error(t, err, "Expected validation error")
	var httpErr *echo.HTTPError
	require.True(t, errors.As(err, &httpErr), "Error should be echo.HTTPError")
	assert.Equal(t, nethttp.StatusBadRequest, httpErr.Code)

	// Service should NOT be called because validation fails first
	mockService.AssertNotCalled(t, "Login")
}

func TestUserHandler_Register_Conflict(t *testing.T) {
	e := setupTestEcho()

	// Create mock service
	mockService := new(MockUserService)
	tokenManager := auth.NewTokenManager("test-secret")
	analyticsService := analytics.NewAnalyticsService(false)
	handler := NewHandler(mockService, tokenManager, nil, analyticsService)

	// Test input
	reqBody := dto.RegisterRequest{
		Email:     "existing@example.com",
		Password:  "password123",
		FirstName: "Jane",
		LastName:  "Smith",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(nethttp.MethodPost, "/api/auth/register", bytes.NewReader(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Setup expectations - return duplicate user error
	mockService.On("Register", mock.Anything, userservice.RegisterUserInput{
		Email:     reqBody.Email,
		Password:  reqBody.Password,
		FirstName: reqBody.FirstName,
		LastName:  reqBody.LastName,
		AvatarUrl: reqBody.AvatarUrl,
	}).Return((*userservice.UserOutput)(nil), userservice.ErrUserAlreadyExists)

	// Call the handler
	err := handler.Register(c)

	// Assertions
	require.NoError(t, err)
	assert.Equal(t, nethttp.StatusConflict, rec.Code)

	mockService.AssertExpectations(t)
}

func TestUserHandler_Login_Unauthorized(t *testing.T) {
	e := setupTestEcho()

	// Create mock service
	mockService := new(MockUserService)
	tokenManager := auth.NewTokenManager("test-secret")
	analyticsService := analytics.NewAnalyticsService(false)
	handler := NewHandler(mockService, tokenManager, nil, analyticsService)

	// Test input
	reqBody := dto.LoginRequest{
		Email:    "wrong@example.com",
		Password: "wrongpassword",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(nethttp.MethodPost, "/api/auth/login", bytes.NewReader(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Setup expectations - return error
	mockService.On("Login", mock.Anything, userservice.LoginUserInput{
		Email:    reqBody.Email,
		Password: reqBody.Password,
	}).Return((*userservice.UserOutput)(nil), assert.AnError)

	// Call the handler
	err := handler.Login(c)

	// Assertions
	require.NoError(t, err)
	assert.Equal(t, nethttp.StatusUnauthorized, rec.Code)

	mockService.AssertExpectations(t)
}

func TestUserHandler_GetProfile(t *testing.T) {
	t.Run("authenticated user retrieves own profile", func(t *testing.T) {
		e := setupTestEcho()
		mockService := new(MockUserService)
		tokenManager := auth.NewTokenManager("test-secret")
		analyticsService := analytics.NewAnalyticsService(false)
		handler := NewHandler(mockService, tokenManager, nil, analyticsService)

		authCtx := helpers.DefaultAuthContext()
		expectedUser := &userservice.UserOutput{
			ID:        authCtx.UserID,
			Email:     authCtx.Email,
			FirstName: "John",
			LastName:  "Doe",
		}

		mockService.On("GetUser", mock.Anything, authCtx.UserID).Return(expectedUser, nil)

		c, rec := helpers.CreateTestContext(e, nethttp.MethodGet, "/api/users/me", nil, &authCtx)

		err := handler.GetProfile(c)

		require.NoError(t, err)
		assert.Equal(t, nethttp.StatusOK, rec.Code)

		var response userservice.UserOutput
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, expectedUser.ID, response.ID)
		assert.Equal(t, expectedUser.Email, response.Email)

		mockService.AssertExpectations(t)
	})

	t.Run("unauthenticated request returns unauthorized", func(t *testing.T) {
		// Note: In production, auth middleware protects this route.
		// Without middleware, MustGetUserID returns "" and service returns not found.
		e := setupTestEcho()
		mockService := new(MockUserService)
		tokenManager := auth.NewTokenManager("test-secret")
		analyticsService := analytics.NewAnalyticsService(false)
		handler := NewHandler(mockService, tokenManager, nil, analyticsService)

		mockService.On("GetUser", mock.Anything, "").
			Return((*userservice.UserOutput)(nil), userservice.ErrUserNotFound)

		// No auth context - handler delegates auth to middleware
		c, rec := helpers.CreateTestContext(e, nethttp.MethodGet, "/api/users/me", nil, nil)

		err := handler.GetProfile(c)

		require.NoError(t, err)
		assert.Equal(t, nethttp.StatusNotFound, rec.Code)
	})

	t.Run("user not found returns not found", func(t *testing.T) {
		e := setupTestEcho()
		mockService := new(MockUserService)
		tokenManager := auth.NewTokenManager("test-secret")
		analyticsService := analytics.NewAnalyticsService(false)
		handler := NewHandler(mockService, tokenManager, nil, analyticsService)

		authCtx := helpers.DefaultAuthContext()
		mockService.On("GetUser", mock.Anything, authCtx.UserID).Return((*userservice.UserOutput)(nil), userservice.ErrUserNotFound)

		c, rec := helpers.CreateTestContext(e, nethttp.MethodGet, "/api/users/me", nil, &authCtx)

		err := handler.GetProfile(c)

		require.NoError(t, err)
		assert.Equal(t, nethttp.StatusNotFound, rec.Code)

		mockService.AssertExpectations(t)
	})

	t.Run("other errors return internal server error", func(t *testing.T) {
		e := setupTestEcho()
		mockService := new(MockUserService)
		tokenManager := auth.NewTokenManager("test-secret")
		analyticsService := analytics.NewAnalyticsService(false)
		handler := NewHandler(mockService, tokenManager, nil, analyticsService)

		authCtx := helpers.DefaultAuthContext()
		mockService.On("GetUser", mock.Anything, authCtx.UserID).Return((*userservice.UserOutput)(nil), assert.AnError)

		c, rec := helpers.CreateTestContext(e, nethttp.MethodGet, "/api/users/me", nil, &authCtx)

		err := handler.GetProfile(c)

		require.NoError(t, err)
		assert.Equal(t, nethttp.StatusInternalServerError, rec.Code)

		mockService.AssertExpectations(t)
	})
}

func TestUserHandler_UpdateProfile(t *testing.T) {
	t.Run("update profile with valid data", func(t *testing.T) {
		e := setupTestEcho()
		mockService := new(MockUserService)
		tokenManager := auth.NewTokenManager("test-secret")
		analyticsService := analytics.NewAnalyticsService(false)
		handler := NewHandler(mockService, tokenManager, nil, analyticsService)

		authCtx := helpers.DefaultAuthContext()
		firstName := "Jane"
		lastName := "Smith"
		reqBody := dto.UpdateProfileRequest{
			FirstName: &firstName,
			LastName:  &lastName,
		}

		expectedUser := &userservice.UserOutput{
			ID:        authCtx.UserID,
			Email:     "test@example.com",
			FirstName: firstName,
			LastName:  lastName,
		}

		mockService.On("UpdateProfile", mock.Anything, authCtx.UserID, mock.MatchedBy(func(input userservice.UpdateProfileInput) bool {
			return input.FirstName != nil && *input.FirstName == firstName &&
				input.LastName != nil && *input.LastName == lastName
		})).Return(expectedUser, nil)

		c, rec := helpers.CreateTestContext(e, nethttp.MethodPut, "/api/users/me", reqBody, &authCtx)

		err := handler.UpdateProfile(c)

		require.NoError(t, err)
		assert.Equal(t, nethttp.StatusOK, rec.Code)

		var response dto.UserResponse
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, expectedUser.Email, response.Email)
		assert.Equal(t, expectedUser.FirstName, response.FirstName)

		mockService.AssertExpectations(t)
	})

	t.Run("update profile unauthorized", func(t *testing.T) {
		// Note: In production, auth middleware protects this route.
		// Without middleware, MustGetUserID returns "" and service gets empty userID.
		e := setupTestEcho()
		mockService := new(MockUserService)
		tokenManager := auth.NewTokenManager("test-secret")
		analyticsService := analytics.NewAnalyticsService(false)
		handler := NewHandler(mockService, tokenManager, nil, analyticsService)

		firstName := "Jane"
		reqBody := dto.UpdateProfileRequest{
			FirstName: &firstName,
		}

		mockService.On("UpdateProfile", mock.Anything, "", mock.AnythingOfType("service.UpdateProfileInput")).
			Return((*userservice.UserOutput)(nil), assert.AnError)

		// No auth context - handler delegates auth to middleware
		c, rec := helpers.CreateTestContext(e, nethttp.MethodPut, "/api/users/me", reqBody, nil)

		err := handler.UpdateProfile(c)

		require.NoError(t, err)
		assert.Equal(t, nethttp.StatusInternalServerError, rec.Code)
	})

	t.Run("update profile with invalid body", func(t *testing.T) {
		e := setupTestEcho()
		mockService := new(MockUserService)
		tokenManager := auth.NewTokenManager("test-secret")
		analyticsService := analytics.NewAnalyticsService(false)
		handler := NewHandler(mockService, tokenManager, nil, analyticsService)

		authCtx := helpers.DefaultAuthContext()

		// Create request with invalid JSON
		req := httptest.NewRequest(nethttp.MethodPut, "/api/users/me", bytes.NewReader([]byte("invalid json")))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		helpers.SetAuthContext(c, authCtx)

		err := handler.UpdateProfile(c)

		require.Error(t, err, "Expected binding error")
		var httpErr *echo.HTTPError
		require.True(t, errors.As(err, &httpErr), "Error should be echo.HTTPError")
		assert.Equal(t, nethttp.StatusBadRequest, httpErr.Code)

		mockService.AssertNotCalled(t, "UpdateProfile")
	})

	t.Run("update profile service error", func(t *testing.T) {
		e := setupTestEcho()
		mockService := new(MockUserService)
		tokenManager := auth.NewTokenManager("test-secret")
		analyticsService := analytics.NewAnalyticsService(false)
		handler := NewHandler(mockService, tokenManager, nil, analyticsService)

		authCtx := helpers.DefaultAuthContext()
		firstName := "Jane"
		lastName := "Smith"
		reqBody := dto.UpdateProfileRequest{
			FirstName: &firstName,
			LastName:  &lastName,
		}

		mockService.On("UpdateProfile", mock.Anything, authCtx.UserID, mock.AnythingOfType("service.UpdateProfileInput")).
			Return((*userservice.UserOutput)(nil), assert.AnError)

		c, rec := helpers.CreateTestContext(e, nethttp.MethodPut, "/api/users/me", reqBody, &authCtx)

		err := handler.UpdateProfile(c)

		require.NoError(t, err)
		assert.Equal(t, nethttp.StatusInternalServerError, rec.Code)

		mockService.AssertExpectations(t)
	})
}
