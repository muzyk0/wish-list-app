package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomHTTPErrorHandler(t *testing.T) {
	e := echo.New()

	// Create a request and response recorder
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Test with a standard HTTP error
	httpErr := echo.NewHTTPError(http.StatusBadRequest, "Bad request")
	CustomHTTPErrorHandler(httpErr, c)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Bad request")

	// Reset recorder
	rec = httptest.NewRecorder()

	// Test with a generic error
	genericErr := errors.New("generic error")
	CustomHTTPErrorHandler(genericErr, e.NewContext(req, rec))

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "generic error")
}

func TestExtractErrorInfo(t *testing.T) {
	// Test with HTTP error
	httpErr := echo.NewHTTPError(http.StatusBadRequest, "Bad request")
	code, message := extractErrorInfo(httpErr)
	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "Bad request", message)

	// Test with HTTP error with numeric code
	httpErrWithCode := echo.NewHTTPError(422, map[string]any{"field": "invalid"})
	code, message = extractErrorInfo(httpErrWithCode)
	assert.Equal(t, 422, code)
	assert.Contains(t, message, "map[field:invalid]")

	// Test with generic error
	genericErr := errors.New("generic error")
	code, message = extractErrorInfo(genericErr)
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.Equal(t, "generic error", message)

	// Test with nil error
	code, message = extractErrorInfo(nil)
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.Equal(t, "Internal Server Error", message)
}

func TestSendErrorResponse(t *testing.T) {
	e := echo.New()

	// Test JSON response
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	sendErrorResponse(c, http.StatusBadRequest, "Bad request")

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Bad request")
	assert.Contains(t, rec.Body.String(), "\"error\"")

	// Test plain text response
	req = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	sendErrorResponse(c, http.StatusInternalServerError, "Internal error")

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, "Error 500: Internal error", rec.Body.String())
}

func TestRequestIDMiddleware(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := RequestIDMiddleware()
	handler := middleware(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	err := handler(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Check that a request ID was added to the response
	requestID := rec.Header().Get("X-Request-ID")
	assert.NotEmpty(t, requestID)
}

func TestLoggerMiddleware(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := LoggerMiddleware()
	handler := middleware(func(c echo.Context) error {
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

	middleware := RecoverMiddleware()
	handler := middleware(func(c echo.Context) error {
		panic("test panic")
	})

	// The recover middleware should handle the panic and return a response
	err := handler(c)
	// The error should be handled by the middleware and not propagated
	require.NoError(t, err)
	// Check that the response has the expected error status
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestCORSMiddleware(t *testing.T) {
	e := echo.New()

	allowedOrigins := []string{"http://localhost:3000", "http://localhost:19006"}

	req := httptest.NewRequest(http.MethodOptions, "/", http.NoBody)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := CORSMiddleware(allowedOrigins)
	handler := middleware(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	err := handler(c)
	require.NoError(t, err)

	// Check CORS headers
	assert.Equal(t, "http://localhost:3000", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Methods"), "POST")
}

func TestTimeoutMiddleware(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	timeout := 5 * time.Second
	middleware := TimeoutMiddleware(timeout)
	handler := middleware(func(c echo.Context) error {
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

	middleware := RateLimiterMiddleware()
	handler := middleware(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	err := handler(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRateLimiterMiddlewareHealthEndpoint(t *testing.T) {
	e := echo.New()
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// The rate limiter should skip the /health endpoint
	req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/health") // Set path so skipper sees the correct path

	middleware := RateLimiterMiddleware()
	handler := middleware(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	err := handler(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}
