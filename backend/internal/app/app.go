package app

import (
	"context"
	"log"

	"wish-list/internal/app/config"
	"wish-list/internal/app/database"
)

// App is the main application struct that wires all dependencies together
type App struct {
	cfg *config.Config
	db  *database.DB
}

// New creates a new App instance with the given configuration and database
func New(cfg *config.Config, db *database.DB) *App {
	return &App{
		cfg: cfg,
		db:  db,
	}
}

// Run starts the application (to be completed in Phase 5)
func (a *App) Run(ctx context.Context) error {
	log.Println("Application starting...")
	// Full wiring will be implemented in Phase 5 (T101-T103)
	return nil
}

// Shutdown gracefully shuts down the application
func (a *App) Shutdown(ctx context.Context) error {
	log.Println("Application shutting down...")
	return nil
}
