package auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTMiddleware(t *testing.T) {
	e := echo.New()
	tm := NewTokenManager("test-secret")

	// Create a valid token
	tokenString, err := tm.GenerateToken("user-123", "test@example.com", "user", 1)
	require.NoError(t, err)

	// Create a request with valid token
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Apply middleware
	middleware := JWTMiddleware(tm)
	handler := middleware(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	err = handler(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Check that user info was added to context
	assert.Equal(t, "user-123", c.Get("user_id"))
	assert.Equal(t, "test@example.com", c.Get("email"))
	assert.Equal(t, "user", c.Get("user_type"))
}

func TestJWTMiddlewareMissingHeader(t *testing.T) {
	e := echo.New()
	tm := NewTokenManager("test-secret")

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := JWTMiddleware(tm)
	handler := middleware(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	err := handler(c)
	require.Error(t, err)
	var httpErr *echo.HTTPError
	ok := errors.As(err, &httpErr)
	assert.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
}

func TestJWTMiddlewareInvalidFormat(t *testing.T) {
	e := echo.New()
	tm := NewTokenManager("test-secret")

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.Header.Set("Authorization", "InvalidFormat")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := JWTMiddleware(tm)
	handler := middleware(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	err := handler(c)
	require.Error(t, err)
	var httpErr *echo.HTTPError
	ok := errors.As(err, &httpErr)
	assert.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
}

func TestJWTMiddlewareInvalidToken(t *testing.T) {
	e := echo.New()
	tm := NewTokenManager("test-secret")

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := JWTMiddleware(tm)
	handler := middleware(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	err := handler(c)
	require.Error(t, err)
	var httpErr *echo.HTTPError
	ok := errors.As(err, &httpErr)
	assert.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
}

func TestOptionalJWTMiddleware(t *testing.T) {
	e := echo.New()
	tm := NewTokenManager("test-secret")

	// Test with no token (should continue without user context)
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := OptionalJWTMiddleware(tm)
	handler := middleware(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	err := handler(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Check that no user info was added to context
	assert.Nil(t, c.Get("user_id"))
	assert.Nil(t, c.Get("email"))
	assert.Nil(t, c.Get("user_type"))

	// Test with valid token
	tokenString, err := tm.GenerateToken("user-123", "test@example.com", "user", 1)
	require.NoError(t, err)

	req = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	err = handler(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Check that user info was added to context
	assert.Equal(t, "user-123", c.Get("user_id"))
	assert.Equal(t, "test@example.com", c.Get("email"))
	assert.Equal(t, "user", c.Get("user_type"))
}

func TestRequireAuth(t *testing.T) {
	e := echo.New()

	// Test without user context
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := RequireAuth()
	handler := middleware(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	err := handler(c)
	require.Error(t, err)
	var httpErr *echo.HTTPError
	ok := errors.As(err, &httpErr)
	assert.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)

	// Test with user context
	req = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set("user_id", "user-123")

	err = handler(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireUserType(t *testing.T) {
	e := echo.New()

	// Test without user context
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := RequireUserType("admin")
	handler := middleware(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	err := handler(c)
	require.Error(t, err)
	var httpErr *echo.HTTPError
	ok := errors.As(err, &httpErr)
	assert.True(t, ok)
	assert.Equal(t, http.StatusForbidden, httpErr.Code)

	// Test with wrong user type
	req = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set("user_type", "user")

	err = handler(c)
	require.Error(t, err)
	var httpErr2 *echo.HTTPError
	ok = errors.As(err, &httpErr2)
	assert.True(t, ok)
	assert.Equal(t, http.StatusForbidden, httpErr2.Code)

	// Test with correct user type
	req = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set("user_type", "admin")

	err = handler(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetUserFromContext(t *testing.T) {
	e := echo.New()

	// Test without user context
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	userID, email, userType, err := GetUserFromContext(c)
	assert.Empty(t, userID)
	assert.Empty(t, email)
	assert.Empty(t, userType)
	require.Error(t, err)

	// Test with user context
	c.Set("user_id", "user-123")
	c.Set("email", "test@example.com")
	c.Set("user_type", "user")

	userID, email, userType, err = GetUserFromContext(c)
	assert.Equal(t, "user-123", userID)
	assert.Equal(t, "test@example.com", email)
	assert.Equal(t, "user", userType)
	require.NoError(t, err)

	// Test with user context but no user_type (should default to "user")
	c2 := e.NewContext(req, rec)
	c2.Set("user_id", "user-123")
	c2.Set("email", "test@example.com")

	userID, email, userType, err = GetUserFromContext(c2)
	assert.Equal(t, "user-123", userID)
	assert.Equal(t, "test@example.com", email)
	assert.Equal(t, "user", userType) // Should default to "user"
	require.NoError(t, err)
}
