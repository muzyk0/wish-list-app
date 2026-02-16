package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"wish-list/internal/app/config"
	"wish-list/internal/app/middleware"

	"github.com/labstack/echo/v4"
)

// Server wraps the Echo instance with lifecycle management
type Server struct {
	Echo *echo.Echo
	cfg  *config.Config
}

// New creates a new Server instance with middleware pipeline configured
func New(cfg *config.Config, validator echo.Validator) *Server {
	e := echo.New()

	// Set custom validator
	if validator != nil {
		e.Validator = validator
	}

	// Set custom error handler
	e.HTTPErrorHandler = middleware.CustomHTTPErrorHandler

	// Apply middleware in order
	e.Use(middleware.SecurityHeadersMiddleware())
	e.Use(middleware.RequestIDMiddleware())
	e.Use(middleware.LoggerMiddleware())
	e.Use(middleware.RecoverMiddleware())
	e.Use(middleware.CORSMiddleware(cfg.CorsAllowedOrigins))
	e.Use(middleware.TimeoutMiddleware(30 * time.Second))
	e.Use(middleware.RateLimiterMiddleware())

	return &Server{
		Echo: e,
		cfg:  cfg,
	}
}

// Start starts the HTTP server and blocks until shutdown
func (s *Server) Start() error {
	port := fmt.Sprintf(":%d", s.cfg.ServerPort)
	log.Printf("Server is starting on port %s", port)

	serverErrors := make(chan error, 1)
	go func() {
		if err := s.Echo.Start(port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server failed to start: %w", err)
	default:
		return nil
	}
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Starting graceful shutdown...")

	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := s.Echo.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		if closeErr := s.Echo.Close(); closeErr != nil {
			log.Printf("Error closing server: %v", closeErr)
		}
		return err
	}

	return nil
}
