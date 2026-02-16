package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"time"

	"database/sql"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	var (
		action = flag.String("action", "up", "Migration action: up, down, version")
		steps  = flag.Int("steps", 0, "Number of steps for migration (used with 'down')")
	)
	flag.Parse()

	// Get database URL from environment or use default
	dbURL := os.Getenv("DATABASE_URL")

	if dbURL == "" {
		dbURL = "postgres://user:password@localhost:5432/wishlist_db?sslmode=disable"
	}

	// Connect to database using standard database/sql with pq driver
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Verify database connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	// Create a postgres driver
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal("Failed to create postgres driver:", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/app/database/migrations",
		"postgres", driver)
	if err != nil {
		log.Fatal("Failed to create migrate instance:", err)
	}
	defer m.Close()

	// Execute migration based on action
	switch *action {
	case "up":
		if *steps == 0 {
			err = m.Up()
		} else {
			err = m.Steps(*steps)
		}
	case "down":
		if *steps == 0 {
			err = m.Down()
		} else {
			err = m.Steps(-*steps)
		}
	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			log.Printf("Version error: %v\n", err)
		} else {
			log.Printf("Version: %d, Dirty: %t\n", version, dirty)
		}
	default:
		log.Fatalf("Unknown action: %s", *action)
	}

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatal("Migration failed:", err)
	}

	log.Println("Migration completed successfully")
}
