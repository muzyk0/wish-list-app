package helpers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"wish-list/internal/pkg/apperrors"
	"wish-list/internal/pkg/validation"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestRequest struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"gte=0,lte=120"`
}

func TestBindAndValidate(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        string
		expectedStatusCode int
		expectedError      bool
		errorContains      string
	}{
		{
			name:               "valid request",
			requestBody:        `{"name":"John Doe","email":"john@example.com","age":30}`,
			expectedStatusCode: 0, // No error response
			expectedError:      false,
		},
		{
			name:               "invalid JSON",
			requestBody:        `{"name":"John"`,
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      true,
			errorContains:      "Invalid request body",
		},
		{
			name:               "empty body",
			requestBody:        ``,
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      true,
			errorContains:      "required",
		},
		{
			name:               "missing required field",
			requestBody:        `{"name":"John Doe","age":30}`,
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      true,
			errorContains:      "required",
		},
		{
			name:               "invalid email format",
			requestBody:        `{"name":"John Doe","email":"not-an-email","age":30}`,
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      true,
			errorContains:      "email",
		},
		{
			name:               "age out of range (negative)",
			requestBody:        `{"name":"John Doe","email":"john@example.com","age":-5}`,
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      true,
			errorContains:      "gte",
		},
		{
			name:               "age out of range (too high)",
			requestBody:        `{"name":"John Doe","email":"john@example.com","age":150}`,
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      true,
			errorContains:      "lte",
		},
		{
			name:               "malformed JSON (trailing comma)",
			requestBody:        `{"name":"John",}`,
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      true,
			errorContains:      "Invalid request body",
		},
		{
			name:               "null values for required fields",
			requestBody:        `{"name":null,"email":null,"age":0}`,
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      true,
			errorContains:      "required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create Echo instance with validator
			e := echo.New()
			e.Validator = validation.NewValidator()

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Test BindAndValidate
			var testReq TestRequest
			err := BindAndValidate(c, &testReq)

			if tt.expectedError {
				require.NotNil(t, err, "Expected non-nil error for invalid input")
				var appErr *apperrors.AppError
				require.True(t, errors.As(err, &appErr), "Error should be apperrors.AppError")
				assert.Equal(t, tt.expectedStatusCode, appErr.Code, "Status code mismatch")
			} else {
				assert.Nil(t, err, "Expected no error but got: %v", err)
				assert.Equal(t, "John Doe", testReq.Name)
				assert.Equal(t, "john@example.com", testReq.Email)
				assert.Equal(t, 30, testReq.Age)
			}
		})
	}
}

func TestBindAndValidateWithoutValidator(t *testing.T) {
	t.Run("should return error when validator is not set", func(t *testing.T) {
		// Create Echo instance WITHOUT validator
		e := echo.New()

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"John","email":"john@example.com","age":30}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		var testReq TestRequest
		err := BindAndValidate(c, &testReq)

		// When validator is not set, c.Validate() returns error
		require.NotNil(t, err, "Expected non-nil error when validator is not set")
		assert.Contains(t, err.Error(), "validator", "Error should mention validator not being registered")
	})
}

func TestBindAndValidateEdgeCases(t *testing.T) {
	t.Run("very large JSON", func(t *testing.T) {
		e := echo.New()
		e.Validator = validation.NewValidator()

		largeJSON := `{"name":"` + strings.Repeat("a", 10000) + `","email":"test@example.com","age":25}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(largeJSON))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		var testReq TestRequest
		err := BindAndValidate(c, &testReq)

		assert.Nil(t, err, "Should handle large JSON successfully")
	})

	t.Run("special characters in JSON", func(t *testing.T) {
		e := echo.New()
		e.Validator = validation.NewValidator()

		specialJSON := `{"name":"John \"The Rock\" Doe","email":"john@example.com","age":30}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(specialJSON))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		var testReq TestRequest
		err := BindAndValidate(c, &testReq)

		assert.Nil(t, err, "Should handle special characters")
		assert.Equal(t, `John "The Rock" Doe`, testReq.Name)
	})
}
