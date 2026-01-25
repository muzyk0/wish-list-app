package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// CustomHTTPErrorHandler handles HTTP errors with custom formatting
func CustomHTTPErrorHandler(err error, c echo.Context) {
	code, message := extractErrorInfo(err)

	// Log the error with context
	c.Logger().Errorf("HTTP Error: %d - %s - %s", code, c.Request().URL.Path, message)

	// Send error response if not already committed
	if c.Response().Committed {
		return
	}

	sendErrorResponse(c, code, message)
}

// extractErrorInfo extracts the HTTP status code and message from an error
func extractErrorInfo(err error) (int, string) {
	code := http.StatusInternalServerError
	message := "Internal Server Error"

	var he *echo.HTTPError
	if errors.As(err, &he) {
		code = he.Code
		if msg, ok := he.Message.(string); ok {
			message = msg
		} else {
			message = fmt.Sprintf("%v", he.Message)
		}
	} else if err != nil {
		// For non-HTTP errors, preserve the original error for logging
		message = err.Error()
	}

	return code, message
}

// sendErrorResponse sends the appropriate error response based on content type
func sendErrorResponse(c echo.Context, code int, message string) {
	if c.Request().Header.Get("Content-Type") == "application/json" {
		if err := c.JSON(code, map[string]any{
			"error":   true,
			"code":    code,
			"message": message,
		}); err != nil {
			c.Logger().Error(err)
		}
	} else {
		if err := c.String(code, message); err != nil {
			c.Logger().Error(err)
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

// CORSMiddleware sets up CORS headers
func CORSMiddleware(allowedOrigins []string) echo.MiddlewareFunc {
	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:  allowedOrigins,
		AllowMethods:  []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders:  []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		ExposeHeaders: []string{echo.HeaderAuthorization},
		MaxAge:        3600, // 1 hour
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
