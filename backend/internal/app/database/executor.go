package database

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// Executor interface abstracts database operations for both DB and Tx
// This allows repositories to work with or without transactions
type Executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowxContext(ctx context.Context, query string, args ...any) *sqlx.Row
	GetContext(ctx context.Context, dest any, query string, args ...any) error
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
}
