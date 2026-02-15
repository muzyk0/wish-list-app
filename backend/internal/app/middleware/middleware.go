package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	apperrors "wish-list/internal/pkg/errors"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// CustomHTTPErrorHandler handles HTTP errors with custom formatting.
// It integrates with the centralized errors package for consistent error responses.
func CustomHTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	var appErr *apperrors.HTTPError
	if errors.As(err, &appErr) {
		if appErr.Err != nil {
			c.Logger().Errorf("Application error: %v", appErr.Err)
		}

		sendErrorResponse(c, appErr.StatusCode, appErr.Message)
		return
	}

	var echoErr *echo.HTTPError
	if errors.As(err, &echoErr) {
		code := echoErr.Code
		message := http.StatusText(code)
		if msg, ok := echoErr.Message.(string); ok {
			message = msg
		} else {
			message = fmt.Sprintf("%v", echoErr.Message)
		}

		c.Logger().Errorf("HTTP Error: %d - %s - %s", code, c.Request().URL.Path, message)
		sendErrorResponse(c, code, message)
		return
	}

	c.Logger().Errorf("Unhandled error: %v", err)
	sendErrorResponse(c, http.StatusInternalServerError, "Internal server error")
}

// sendErrorResponse sends an error response based on the client's Accept header
func sendErrorResponse(c echo.Context, code int, message string) {
	accept := c.Request().Header.Get("Accept")

	// Simple check for JSON acceptance
	if strings.Contains(accept, "application/json") ||
		strings.Contains(accept, "*/*") ||
		strings.Contains(accept, "text/json") {
		// Return JSON response
		err := c.JSON(code, map[string]any{
			"error":   true,
			"code":    code,
			"message": message,
		})
		if err != nil {
			c.Logger().Errorf("Failed to send JSON response: %v", err)
		}
	} else {
		// Return plain text response
		err := c.String(code, fmt.Sprintf("Error %d: %s", code, message))
		if err != nil {
			c.Logger().Errorf("Failed to send text response: %v", err)
		}
	}
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() echo.MiddlewareFunc {
	return middleware.RequestID()
}

// LoggerMiddleware adds structured logging for requests
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
			// Use JSON marshaling to prevent log injection
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

// RecoverMiddleware recovers from panics and logs the error
func RecoverMiddleware() echo.MiddlewareFunc {
	return middleware.RecoverWithConfig(middleware.RecoverConfig{
		StackSize: 1 << 10, // 1 KB
		LogLevel:  0,       // Disabled to use custom error handler
	})
}

// TimeoutMiddleware adds a timeout to requests
func TimeoutMiddleware(timeout time.Duration) echo.MiddlewareFunc {
	return middleware.ContextTimeoutWithConfig(middleware.ContextTimeoutConfig{
		Timeout: timeout,
	})
}

// RateLimiterMiddleware limits the number of requests per IP
func RateLimiterMiddleware() echo.MiddlewareFunc {
	return middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStore(20), // Allow 20 requests per second per IP
		Skipper: func(c echo.Context) bool {
			// Skip rate limiting for health checks
			return c.Path() == "/health"
		},
		IdentifierExtractor: func(c echo.Context) (string, error) {
			// Use IP address as identifier
			ip := c.RealIP()
			if ip == "" {
				return "", errors.New("unable to extract IP address")
			}
			return ip, nil
		},
		ErrorHandler: func(c echo.Context, err error) error {
			return c.JSON(http.StatusTooManyRequests, map[string]any{
				"error":   true,
				"code":    http.StatusTooManyRequests,
				"message": "Rate limit exceeded",
			})
		},
	})
}

// SecurityHeadersMiddleware adds security headers to all responses
func SecurityHeadersMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Prevent MIME-sniffing attacks
			c.Response().Header().Set("X-Content-Type-Options", "nosniff")

			// Prevent clickjacking attacks
			c.Response().Header().Set("X-Frame-Options", "DENY")

			// Enable XSS protection in browsers
			c.Response().Header().Set("X-XSS-Protection", "1; mode=block")

			// Force HTTPS for 1 year (only set in production)
			if c.Request().Header.Get("X-Forwarded-Proto") == "https" {
				c.Response().Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}

			// Content Security Policy (restrictive default)
			c.Response().Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' https://*.amazonaws.com; script-src 'self'; style-src 'self' 'unsafe-inline'")

			// Referrer policy for privacy
			c.Response().Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// Permissions policy to restrict features
			c.Response().Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

			return next(c)
		}
	}
}
