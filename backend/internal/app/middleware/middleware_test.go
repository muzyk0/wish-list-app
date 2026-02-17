package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"wish-list/internal/pkg/apperrors"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomHTTPErrorHandler_AppError(t *testing.T) {
	e := echo.New()

	t.Run("simple app error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		appErr := apperrors.NotFound("Wishlist not found")
		CustomHTTPErrorHandler(appErr, c)

		assert.Equal(t, http.StatusNotFound, rec.Code)

		var body map[string]string
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, "Wishlist not found", body["error"])
	})

	t.Run("app error with cause", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		cause := errors.New("sql: no rows")
		appErr := apperrors.NotFound("Item not found").Wrap(cause)
		CustomHTTPErrorHandler(appErr, c)

		assert.Equal(t, http.StatusNotFound, rec.Code)

		var body map[string]string
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, "Item not found", body["error"])
		// Cause must NOT leak to client
		assert.NotContains(t, rec.Body.String(), "sql: no rows")
	})

	t.Run("validation error with details", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		appErr := apperrors.NewValidationError(map[string]string{
			"email":    "must be a valid email address",
			"password": "must be at least 8 characters long",
		})
		CustomHTTPErrorHandler(appErr, c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var body map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, "Validation failed", body["error"])
		details, ok := body["details"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "must be a valid email address", details["email"])
		assert.Equal(t, "must be at least 8 characters long", details["password"])
	})
}

func TestCustomHTTPErrorHandler_EchoError(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	httpErr := echo.NewHTTPError(http.StatusBadRequest, "Bad request")
	CustomHTTPErrorHandler(httpErr, c)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "Bad request", body["error"])
}

func TestCustomHTTPErrorHandler_UnknownError(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	genericErr := errors.New("something unexpected")
	CustomHTTPErrorHandler(genericErr, c)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "Internal server error", body["error"])
	// Internal details must NOT leak
	assert.NotContains(t, rec.Body.String(), "something unexpected")
}

func TestCustomHTTPErrorHandler_CommittedResponse(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Write something to commit the response
	_ = c.String(http.StatusOK, "already sent")

	// Error handler should be a no-op
	CustomHTTPErrorHandler(apperrors.Internal("oops"), c)

	// Response should still be the original 200
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "already sent")
}

func TestRequestIDMiddleware(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := RequestIDMiddleware()
	handler := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	err := handler(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotEmpty(t, rec.Header().Get("X-Request-ID"))
}

func TestLoggerMiddleware(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := LoggerMiddleware()
	handler := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	err := handler(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRecoverMiddleware(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := RecoverMiddleware()
	handler := mw(func(c echo.Context) error {
		panic("test panic")
	})

	err := handler(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestCORSMiddleware(t *testing.T) {
	allowedOrigins := []string{"http://localhost:3000", "http://localhost:19006"}

	t.Run("Allowed origin receives CORS headers", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodOptions, "/", http.NoBody)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "Content-Type")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mw := CORSMiddleware(allowedOrigins)
		handler := mw(func(c echo.Context) error {
			return c.String(http.StatusOK, "OK")
		})

		err := handler(c)
		require.NoError(t, err)

		assert.Equal(t, "http://localhost:3000", rec.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))
		assert.Contains(t, rec.Header().Get("Access-Control-Allow-Methods"), "POST")
		assert.Equal(t, "86400", rec.Header().Get("Access-Control-Max-Age"))
	})

	t.Run("Disallowed origin does not receive CORS headers", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodOptions, "/", http.NoBody)
		req.Header.Set("Origin", "http://malicious-site.com")
		req.Header.Set("Access-Control-Request-Method", "POST")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mw := CORSMiddleware(allowedOrigins)
		handler := mw(func(c echo.Context) error {
			return c.String(http.StatusOK, "OK")
		})

		err := handler(c)
		require.NoError(t, err)

		allowOriginHeader := rec.Header().Get("Access-Control-Allow-Origin")
		assert.NotEqual(t, "http://malicious-site.com", allowOriginHeader)
	})

	t.Run("Multiple allowed origins work correctly", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodOptions, "/", http.NoBody)
		req.Header.Set("Origin", "http://localhost:19006")
		req.Header.Set("Access-Control-Request-Method", "GET")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mw := CORSMiddleware(allowedOrigins)
		handler := mw(func(c echo.Context) error {
			return c.String(http.StatusOK, "OK")
		})

		err := handler(c)
		require.NoError(t, err)

		assert.Equal(t, "http://localhost:19006", rec.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))
	})

	t.Run("Credentials enabled for cross-domain cookies", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", http.NoBody)
		req.Header.Set("Origin", "http://localhost:3000")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mw := CORSMiddleware(allowedOrigins)
		handler := mw(func(c echo.Context) error {
			return c.String(http.StatusOK, "OK")
		})

		err := handler(c)
		require.NoError(t, err)

		assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))
	})
}

func TestTimeoutMiddleware(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	timeout := 5 * time.Second
	mw := TimeoutMiddleware(timeout)
	handler := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	err := handler(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRateLimiterMiddleware(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := RateLimiterMiddleware()
	handler := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	err := handler(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRateLimiterMiddlewareHealthEndpoint(t *testing.T) {
	e := echo.New()
	e.GET("/healthz", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/healthz", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/healthz")

	mw := RateLimiterMiddleware()
	handler := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	err := handler(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}
