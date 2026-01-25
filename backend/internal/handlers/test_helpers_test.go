package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/labstack/echo/v4"
)

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
func CreateTestContext(e *echo.Echo, method, path string, body interface{}, auth *AuthContext) (echo.Context, *httptest.ResponseRecorder) {
	var req *http.Request
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		req = httptest.NewRequest(method, path, bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	} else {
		req = httptest.NewRequest(method, path, http.NoBody)
	}

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if auth != nil {
		SetAuthContext(c, *auth)
	}

	return c, rec
}

// CreateTestContextWithParams creates an Echo context with params and optional auth context
func CreateTestContextWithParams(e *echo.Echo, method, path string, body interface{}, paramNames, paramValues []string, auth *AuthContext) (echo.Context, *httptest.ResponseRecorder) {
	c, rec := CreateTestContext(e, method, path, body, auth)
	c.SetParamNames(paramNames...)
	c.SetParamValues(paramValues...)
	return c, rec
}
