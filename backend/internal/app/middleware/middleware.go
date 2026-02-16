package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"wish-list/internal/pkg/apperrors"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// CustomHTTPErrorHandler handles all errors returned from handlers and middleware.
// It produces a unified JSON error response:
//
//	{"error": "message"}                              — standard errors
//	{"error": "Validation failed", "details": {...}}  — validation errors
//
// Priority: AppError > echo.HTTPError > unknown (500).
func CustomHTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	// 1. Application errors (apperrors.AppError)
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		if appErr.Err != nil {
			c.Logger().Errorf("Application error: %v", appErr.Err)
		}

		sendAppErrorResponse(c, appErr)
		return
	}

	// 2. Echo framework errors (echo.HTTPError)
	var echoErr *echo.HTTPError
	if errors.As(err, &echoErr) {
		code := echoErr.Code
		message := http.StatusText(code)
		if msg, ok := echoErr.Message.(string); ok {
			message = msg
		}

		c.Logger().Errorf("HTTP error: %d - %s - %s", code, c.Request().URL.Path, message)
		_ = c.JSON(code, map[string]string{"error": message})
		return
	}

	// 3. Unknown errors — log and return generic 500
	c.Logger().Errorf("Unhandled error: %v", err)
	_ = c.JSON(http.StatusInternalServerError, map[string]string{
		"error": "Internal server error",
	})
}

// sendAppErrorResponse writes the AppError as JSON.
// Validation errors include a "details" field.
func sendAppErrorResponse(c echo.Context, appErr *apperrors.AppError) {
	if len(appErr.Details) > 0 {
		_ = c.JSON(appErr.Code, map[string]any{
			"error":   appErr.Message,
			"details": appErr.Details,
		})
		return
	}

	_ = c.JSON(appErr.Code, map[string]string{"error": appErr.Message})
}

// RequestIDMiddleware adds a unique request ID to each request.
func RequestIDMiddleware() echo.MiddlewareFunc {
	return middleware.RequestID()
}

// LoggerMiddleware adds structured logging for requests.
func LoggerMiddleware() echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:       true,
		LogMethod:    true,
		LogStatus:    true,
		LogLatency:   true,
		LogRemoteIP:  true,
		LogUserAgent: true,
		LogRequestID: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			logEntry := map[string]any{
				"time":       time.Now().Format(time.RFC3339),
				"method":     v.Method,
				"uri":        v.URI,
				"status":     v.Status,
				"latency":    v.Latency.String(),
				"ip":         v.RemoteIP,
				"user_agent": v.UserAgent,
				"request_id": v.RequestID,
			}
			if data, err := json.Marshal(logEntry); err == nil {
				fmt.Println(string(data))
			}
			return nil
		},
	})
}

// RecoverMiddleware recovers from panics and logs the error.
func RecoverMiddleware() echo.MiddlewareFunc {
	return middleware.RecoverWithConfig(middleware.RecoverConfig{
		StackSize: 1 << 10, // 1 KB
		LogLevel:  0,       // Disabled to use custom error handler
	})
}

// TimeoutMiddleware adds a timeout to requests.
func TimeoutMiddleware(timeout time.Duration) echo.MiddlewareFunc {
	return middleware.ContextTimeoutWithConfig(middleware.ContextTimeoutConfig{
		Timeout: timeout,
	})
}

// RateLimiterMiddleware limits the number of requests per IP.
func RateLimiterMiddleware() echo.MiddlewareFunc {
	return middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStore(20),
		Skipper: func(c echo.Context) bool {
			return c.Path() == "/health"
		},
		IdentifierExtractor: func(c echo.Context) (string, error) {
			ip := c.RealIP()
			if ip == "" {
				return "", errors.New("unable to extract IP address")
			}
			return ip, nil
		},
		ErrorHandler: func(c echo.Context, err error) error {
			return c.JSON(http.StatusTooManyRequests, map[string]string{
				"error": "Rate limit exceeded",
			})
		},
	})
}

// SecurityHeadersMiddleware adds security headers to all responses.
func SecurityHeadersMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("X-Content-Type-Options", "nosniff")
			c.Response().Header().Set("X-Frame-Options", "DENY")
			c.Response().Header().Set("X-XSS-Protection", "1; mode=block")

			if c.Request().Header.Get("X-Forwarded-Proto") == "https" {
				c.Response().Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}

			c.Response().Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' https://*.amazonaws.com; script-src 'self'; style-src 'self' 'unsafe-inline'")
			c.Response().Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			c.Response().Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

			return next(c)
		}
	}
}
