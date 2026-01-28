package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	"github.com/jmoiron/sqlx"
)

// Executor interface abstracts database operations for both DB and Tx
// This allows repositories to work with or without transactions
type Executor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

type DB struct {
	*sqlx.DB
}

func New(ctx context.Context, connUrl string) (*DB, error) {
	db, err := sqlx.ConnectContext(ctx, "pgx", connUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// set reasonable connection pool defaults
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{db}, nil
}

// UUIDToString converts pgtype.UUID to string
func UUIDToString(uuid pgtype.UUID) (string, error) {
	if !uuid.Valid {
		return "", errors.New("invalid UUID")
	}
	return uuid.String(), nil
}

// StringToUUID converts string to pgtype.UUID
func StringToUUID(str string) (pgtype.UUID, error) {
	var uuid pgtype.UUID
	err := uuid.Scan(str)
	if err != nil {
		return uuid, fmt.Errorf("invalid UUID string: %w", err)
	}
	return uuid, nil
}

// TextToString converts pgtype.Text to string
func TextToString(text pgtype.Text) string {
	if text.Valid {
		return text.String
	}
	return ""
}

// StringToText converts string to pgtype.Text
func StringToText(str string) pgtype.Text {
	return pgtype.Text{
		String: str,
		Valid:  str != "",
	}
}

// BoolToBool converts pgtype.Bool to bool
func BoolToBool(b pgtype.Bool) bool {
	return b.Bool
}

// BoolToPgBool converts bool to pgtype.Bool
func BoolToPgBool(b bool) pgtype.Bool {
	return pgtype.Bool{
		Bool:  b,
		Valid: true,
	}
}

// NumericToFloat64 converts pgtype.Numeric to float64
func NumericToFloat64(n pgtype.Numeric) float64 {
	if !n.Valid {
		return 0
	}
	f, err := n.Float64Value()
	if err != nil {
		return 0
	}
	return f.Float64
}
