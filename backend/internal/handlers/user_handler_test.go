package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"wish-list/internal/analytics"
	"wish-list/internal/auth"
	"wish-list/internal/services"
	"wish-list/internal/validation"

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

func (m *MockUserService) Register(ctx context.Context, input services.RegisterUserInput) (*services.UserOutput, error) {
	args := m.Called(ctx, input)
	v := args.Get(0)
	if v != nil {
		if result, ok := v.(*services.UserOutput); ok {
			return result, args.Error(1)
		}
	}
	return nil, args.Error(1)
}

func (m *MockUserService) Login(ctx context.Context, input services.LoginUserInput) (*services.UserOutput, error) {
	args := m.Called(ctx, input)
	v := args.Get(0)
	if v != nil {
		if result, ok := v.(*services.UserOutput); ok {
			return result, args.Error(1)
		}
	}
	return nil, args.Error(1)
}

func (m *MockUserService) GetUser(ctx context.Context, userID string) (*services.UserOutput, error) {
	args := m.Called(ctx, userID)
	v := args.Get(0)
	if v != nil {
		if result, ok := v.(*services.UserOutput); ok {
			return result, args.Error(1)
		}
	}
	return nil, args.Error(1)
}

func (m *MockUserService) UpdateUser(ctx context.Context, userID string, input services.UpdateUserInput) (*services.UserOutput, error) {
	args := m.Called(ctx, userID, input)
	v := args.Get(0)
	if v != nil {
		if result, ok := v.(*services.UserOutput); ok {
			return result, args.Error(1)
		}
	}
	return nil, args.Error(1)
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
	handler := NewUserHandler(mockService, tokenManager, nil, analyticsService)

	// Test input
	reqBody := RegisterRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Expected output
	expectedUser := &services.UserOutput{
		ID:        "123",
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		AvatarUrl: "",
	}

	// Setup expectations
	mockService.On("Register", mock.Anything, services.RegisterUserInput{
		Email:     reqBody.Email,
		Password:  reqBody.Password,
		FirstName: reqBody.FirstName,
		LastName:  reqBody.LastName,
		AvatarUrl: reqBody.AvatarUrl,
	}).Return(expectedUser, nil)

	// Call the handler
	err := handler.Register(c)

	// Assertions
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var response AuthResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, expectedUser, response.User)
	assert.NotEmpty(t, response.Token)

	mockService.AssertExpectations(t)
}

func TestUserHandler_Login(t *testing.T) {
	e := setupTestEcho()

	// Create mock service
	mockService := new(MockUserService)
	tokenManager := auth.NewTokenManager("test-secret")
	analyticsService := analytics.NewAnalyticsService(false)
	handler := NewUserHandler(mockService, tokenManager, nil, analyticsService)

	// Test input
	reqBody := LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Expected output
	expectedUser := &services.UserOutput{
		ID:        "123",
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		AvatarUrl: "",
	}

	// Setup expectations
	mockService.On("Login", mock.Anything, services.LoginUserInput{
		Email:    reqBody.Email,
		Password: reqBody.Password,
	}).Return(expectedUser, nil)

	// Call the handler
	err := handler.Login(c)

	// Assertions
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response AuthResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, expectedUser, response.User)
	assert.NotEmpty(t, response.Token)

	mockService.AssertExpectations(t)
}

func TestUserHandler_Register_BadRequest(t *testing.T) {
	e := setupTestEcho()

	// Create mock service
	mockService := new(MockUserService)
	tokenManager := auth.NewTokenManager("test-secret")
	analyticsService := analytics.NewAnalyticsService(false)
	handler := NewUserHandler(mockService, tokenManager, nil, analyticsService)

	// Invalid input - empty body (validation fails before service call)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", http.NoBody)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Call the handler
	err := handler.Register(c)

	// Assertions
	require.NoError(t, err)
	// Validation should fail for empty email/password, returning 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	// Service should NOT be called because validation fails first
	mockService.AssertNotCalled(t, "Register")
}

func TestUserHandler_Login_BadRequest(t *testing.T) {
	e := setupTestEcho()

	// Create mock service
	mockService := new(MockUserService)
	tokenManager := auth.NewTokenManager("test-secret")
	analyticsService := analytics.NewAnalyticsService(false)
	handler := NewUserHandler(mockService, tokenManager, nil, analyticsService)

	// Invalid input - empty body (validation fails before service call)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", http.NoBody)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Call the handler
	err := handler.Login(c)

	// Assertions
	require.NoError(t, err)
	// Validation should fail for empty email/password, returning 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	// Service should NOT be called because validation fails first
	mockService.AssertNotCalled(t, "Login")
}

func TestUserHandler_Register_Conflict(t *testing.T) {
	e := setupTestEcho()

	// Create mock service
	mockService := new(MockUserService)
	tokenManager := auth.NewTokenManager("test-secret")
	analyticsService := analytics.NewAnalyticsService(false)
	handler := NewUserHandler(mockService, tokenManager, nil, analyticsService)

	// Test input
	reqBody := RegisterRequest{
		Email:     "existing@example.com",
		Password:  "password123",
		FirstName: "Jane",
		LastName:  "Smith",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Setup expectations - return error
	mockService.On("Register", mock.Anything, services.RegisterUserInput{
		Email:     reqBody.Email,
		Password:  reqBody.Password,
		FirstName: reqBody.FirstName,
		LastName:  reqBody.LastName,
		AvatarUrl: reqBody.AvatarUrl,
	}).Return((*services.UserOutput)(nil), assert.AnError)

	// Call the handler
	err := handler.Register(c)

	// Assertions
	require.NoError(t, err)
	assert.Equal(t, http.StatusConflict, rec.Code)

	mockService.AssertExpectations(t)
}

func TestUserHandler_Login_Unauthorized(t *testing.T) {
	e := setupTestEcho()

	// Create mock service
	mockService := new(MockUserService)
	tokenManager := auth.NewTokenManager("test-secret")
	analyticsService := analytics.NewAnalyticsService(false)
	handler := NewUserHandler(mockService, tokenManager, nil, analyticsService)

	// Test input
	reqBody := LoginRequest{
		Email:    "wrong@example.com",
		Password: "wrongpassword",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Setup expectations - return error
	mockService.On("Login", mock.Anything, services.LoginUserInput{
		Email:    reqBody.Email,
		Password: reqBody.Password,
	}).Return((*services.UserOutput)(nil), assert.AnError)

	// Call the handler
	err := handler.Login(c)

	// Assertions
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	mockService.AssertExpectations(t)
}

// T045a: Unit tests for user profile management endpoints

func TestUserHandler_GetProfile(t *testing.T) {
	t.Run("authenticated user retrieves own profile", func(t *testing.T) {
		e := setupTestEcho()
		mockService := new(MockUserService)
		tokenManager := auth.NewTokenManager("test-secret")
		analyticsService := analytics.NewAnalyticsService(false)
	handler := NewUserHandler(mockService, tokenManager, nil, analyticsService)

		authCtx := DefaultAuthContext()
		expectedUser := &services.UserOutput{
			ID:        authCtx.UserID,
			Email:     authCtx.Email,
			FirstName: "John",
			LastName:  "Doe",
		}

		mockService.On("GetUser", mock.Anything, authCtx.UserID).Return(expectedUser, nil)

		c, rec := CreateTestContext(e, http.MethodGet, "/api/users/me", nil, &authCtx)

		err := handler.GetProfile(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response services.UserOutput
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, expectedUser.ID, response.ID)
		assert.Equal(t, expectedUser.Email, response.Email)

		mockService.AssertExpectations(t)
	})

	t.Run("unauthenticated request returns unauthorized", func(t *testing.T) {
		e := setupTestEcho()
		mockService := new(MockUserService)
		tokenManager := auth.NewTokenManager("test-secret")
		analyticsService := analytics.NewAnalyticsService(false)
	handler := NewUserHandler(mockService, tokenManager, nil, analyticsService)

		// No auth context
		c, rec := CreateTestContext(e, http.MethodGet, "/api/users/me", nil, nil)

		err := handler.GetProfile(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		mockService.AssertNotCalled(t, "GetUser")
	})

	t.Run("user not found returns not found", func(t *testing.T) {
		e := setupTestEcho()
		mockService := new(MockUserService)
		tokenManager := auth.NewTokenManager("test-secret")
		analyticsService := analytics.NewAnalyticsService(false)
	handler := NewUserHandler(mockService, tokenManager, nil, analyticsService)

		authCtx := DefaultAuthContext()
		mockService.On("GetUser", mock.Anything, authCtx.UserID).Return((*services.UserOutput)(nil), assert.AnError)

		c, rec := CreateTestContext(e, http.MethodGet, "/api/users/me", nil, &authCtx)

		err := handler.GetProfile(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)

		mockService.AssertExpectations(t)
	})
}

func TestUserHandler_UpdateProfile(t *testing.T) {
	t.Run("update profile with valid data", func(t *testing.T) {
		e := setupTestEcho()
		mockService := new(MockUserService)
		tokenManager := auth.NewTokenManager("test-secret")
		analyticsService := analytics.NewAnalyticsService(false)
	handler := NewUserHandler(mockService, tokenManager, nil, analyticsService)

		authCtx := DefaultAuthContext()
		email := "updated@example.com"
		password := "newpassword123"
		firstName := "Jane"
		lastName := "Smith"
		reqBody := UpdateProfileRequest{
			Email:     &email,
			Password:  &password,
			FirstName: &firstName,
			LastName:  &lastName,
		}

		expectedUser := &services.UserOutput{
			ID:        authCtx.UserID,
			Email:     email,
			FirstName: firstName,
			LastName:  lastName,
		}

		mockService.On("UpdateUser", mock.Anything, authCtx.UserID, mock.MatchedBy(func(input services.UpdateUserInput) bool {
			return input.Email != nil && *input.Email == email &&
				input.Password != nil && *input.Password == password &&
				input.FirstName != nil && *input.FirstName == firstName &&
				input.LastName != nil && *input.LastName == lastName
		})).Return(expectedUser, nil)

		c, rec := CreateTestContext(e, http.MethodPut, "/api/users/me", reqBody, &authCtx)

		err := handler.UpdateProfile(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response services.UserOutput
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, expectedUser.Email, response.Email)
		assert.Equal(t, expectedUser.FirstName, response.FirstName)

		mockService.AssertExpectations(t)
	})

	t.Run("update profile unauthorized", func(t *testing.T) {
		e := setupTestEcho()
		mockService := new(MockUserService)
		tokenManager := auth.NewTokenManager("test-secret")
		analyticsService := analytics.NewAnalyticsService(false)
	handler := NewUserHandler(mockService, tokenManager, nil, analyticsService)

		email := "updated@example.com"
		password := "newpassword123"
		reqBody := UpdateProfileRequest{
			Email:    &email,
			Password: &password,
		}

		// No auth context
		c, rec := CreateTestContext(e, http.MethodPut, "/api/users/me", reqBody, nil)

		err := handler.UpdateProfile(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		mockService.AssertNotCalled(t, "UpdateUser")
	})

	t.Run("update profile with invalid body", func(t *testing.T) {
		e := setupTestEcho()
		mockService := new(MockUserService)
		tokenManager := auth.NewTokenManager("test-secret")
		analyticsService := analytics.NewAnalyticsService(false)
	handler := NewUserHandler(mockService, tokenManager, nil, analyticsService)

		authCtx := DefaultAuthContext()

		// Create request with invalid JSON
		req := httptest.NewRequest(http.MethodPut, "/api/users/me", bytes.NewReader([]byte("invalid json")))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		SetAuthContext(c, authCtx)

		err := handler.UpdateProfile(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		mockService.AssertNotCalled(t, "UpdateUser")
	})

	t.Run("update profile service error", func(t *testing.T) {
		e := setupTestEcho()
		mockService := new(MockUserService)
		tokenManager := auth.NewTokenManager("test-secret")
		analyticsService := analytics.NewAnalyticsService(false)
	handler := NewUserHandler(mockService, tokenManager, nil, analyticsService)

		authCtx := DefaultAuthContext()
		email := "updated@example.com"
		password := "newpassword123"
		firstName := "Jane"
		lastName := "Smith"
		reqBody := UpdateProfileRequest{
			Email:     &email,
			Password:  &password,
			FirstName: &firstName,
			LastName:  &lastName,
		}

		mockService.On("UpdateUser", mock.Anything, authCtx.UserID, mock.AnythingOfType("services.UpdateUserInput")).
			Return((*services.UserOutput)(nil), assert.AnError)

		c, rec := CreateTestContext(e, http.MethodPut, "/api/users/me", reqBody, &authCtx)

		err := handler.UpdateProfile(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

		mockService.AssertExpectations(t)
	})
}
