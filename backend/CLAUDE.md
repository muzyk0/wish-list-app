# Backend CLAUDE.md

Backend-specific patterns, best practices, and conventions for the Go/Echo/PostgreSQL API.

> For general project overview, deployment architecture, and cross-component information, see the root [`/CLAUDE.md`](../CLAUDE.md).

## Backend Structure (Domain-Driven)

- Main entry point: `backend/cmd/server/main.go`
- Application wiring: `backend/internal/app/app.go`
- Server & routing: `backend/internal/app/server/`
- Configuration: `backend/internal/app/config/`
- Database connection: `backend/internal/app/database/postgres.go`
- Database migrations: `backend/internal/app/database/migrations/`
- Middleware: `backend/internal/app/middleware/`
- Background jobs: `backend/internal/app/jobs/`
- Swagger docs: `backend/internal/app/swagger/`
- Shared libraries: `backend/internal/pkg/` (analytics, apperrors, auth, aws, cache, encryption, helpers, logger, pii, validation)
- Domain modules: `backend/internal/domain/{name}/` — each domain contains:
  - `delivery/http/handler.go` — HTTP request handling
  - `delivery/http/dto/` — Request/response DTOs
  - `delivery/http/routes.go` — Route registration
  - `service/` — Business logic
  - `repository/` — Database access (sqlx)
  - `models/` — Domain entity structs
- Domains: auth, user, wishlist, item, wishlist_item, reservation, health, storage

**Import rules**: `pkg/` has no imports from `domain/` or `app/`. `app/` wires all domains together.

**Architecture Guide**: For comprehensive backend architecture documentation, see `/docs/Go-Architecture-Guide.md`. This guide covers:
- Domain-driven 3-layer architecture (Handler-Service-Repository) per domain module
- The ONE non-negotiable rule: JSON serialization ONLY in handlers
- Complete code examples with good patterns and anti-patterns
- Data flow, validation strategy, testing approach
- Security considerations and when to evolve the architecture

## API Documentation (Swagger/OpenAPI)

The backend uses **swaggo/swag** to generate Swagger/OpenAPI documentation from Go annotations.

### Documentation Structure

Library documentation is organized in modular, AI-agent-friendly files:

**Location**: `/backend/library-docs/swaggo-swag/`

**Files**:
1. **README.md** - Index and quick reference
2. **01-getting-started.md** - Installation, setup, and basic workflow
3. **02-cli-reference.md** - `swag init` and `swag fmt` commands with all options
4. **03-general-api-info.md** - API-level annotations (`@title`, `@version`, `@host`, `@BasePath`, etc.)
5. **04-api-operations.md** - Endpoint annotations (`@Summary`, `@Param`, `@Success`, `@Router`, etc.)
6. **05-security.md** - Authentication schemes (`@securityDefinitions`, `@Security`)
7. **06-attributes-validation.md** - Field validation and constraints (enums, min/max, format, etc.)
8. **07-examples.md** - Common patterns (CRUD, pagination, file upload, error responses)
9. **08-advanced-features.md** - Generics, custom types, global overrides

### Quick Reference

**Generate Swagger docs**:
```bash
make docs                           # Complete API documentation generation (recommended)
                                    # - Generates Swagger 2.0 from Go annotations
                                    # - Converts to OpenAPI 3.0
                                    # - Splits into organized files
                                    # - Regenerates frontend/mobile client schemas

# Or manually:
swag init                           # Generate docs only
swag init --parseDependency         # Include external packages
swag init --parseInternal           # Include internal packages
swag fmt                            # Format annotations
```

**Common annotations**:
```go
// General API info (in main.go)
// @title           Wish List API
// @version         1.0
// @description     API description
// @host            localhost:8080
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

// Endpoint annotations (in handlers)
// @Summary      Create wishlist
// @Description  Create a new wishlist for authenticated user
// @Tags         Wishlists
// @Accept       json
// @Produce      json
// @Param        wishlist body CreateWishlistRequest true "Wishlist data"
// @Success      201 {object} WishlistResponse "Success"
// @Failure      400 {object} map[string]string "Bad request"
// @Security     BearerAuth
// @Router       /wishlists [post]
```

### Important Notes

- **Regenerate after changes**: Always run `make docs` after modifying handler DTOs to update OpenAPI specs and client schemas
- **Handler DTOs Required**: Always use handler-specific response types (not service types) in `@Success` and `@Failure` annotations
- **Validation Tags**: Use `validate:"required"` in response DTOs for OpenAPI schema generation
- **Format Tags**: Use `format:"uuid"`, `format:"email"`, etc. for proper schema types
- **Parse Dependencies**: When structs are in external packages, use `swag init --parseDependency`
- **Swagger UI**: Access at `http://localhost:8080/swagger/index.html` after running backend

### Best Practices

1. **Document as you code**: Add Swagger annotations when creating handlers
2. **Use handler DTOs**: Never expose service types directly in Swagger docs
3. **Regenerate documentation**: Run `make docs` after any DTO changes to update all schemas
4. **Validate annotations**: Run `make docs` to catch annotation errors early
5. **Keep examples**: Add `example` tags to struct fields for better API docs
6. **Security first**: Always add `@Security` annotations to protected endpoints

## Backend Best Practices & Patterns

### Error Handling (Updated Feb 2026)

**Unified Error System**: All handlers use `internal/pkg/apperrors` package for consistent error responses.

**Handler Error Pattern** (Required for all handlers):
1. Create `errors.go` in each handler package with error mapping function:
   ```go
   func mapXxxServiceError(err error) error {
       switch {
       case errors.Is(err, service.ErrXxxNotFound):
           return apperrors.NotFound("Xxx not found")
       case errors.Is(err, service.ErrXxxForbidden):
           return apperrors.Forbidden("Access denied")
       default:
           return apperrors.Internal("Failed to process request").Wrap(err)
       }
   }
   ```
2. In handlers, return mapped errors: `return mapXxxServiceError(err)`
3. **Never use inline `c.JSON(status, map[string]string{...})`** - always use apperrors

**Error Types**:
- `apperrors.BadRequest(msg)` - 400
- `apperrors.Unauthorized(msg)` - 401
- `apperrors.Forbidden(msg)` - 403
- `apperrors.NotFound(msg)` - 404
- `apperrors.Conflict(msg)` - 409
- `apperrors.Internal(msg).Wrap(err)` - 500 with wrapped cause
- `apperrors.BadGateway(msg)` - 502

**Middleware**: `middleware.CustomHTTPErrorHandler` converts all `*apperrors.AppError` to JSON `{"error": "msg", "details": {...}}`

**Test Setup**: Always register error handler in test echo instances:
```go
func setupTestEcho() *echo.Echo {
    e := echo.New()
    e.Validator = validation.NewValidator()
    e.HTTPErrorHandler = middleware.CustomHTTPErrorHandler  // Required!
    return e
}
```

### Sentinel Errors (Original Pattern)
- **Sentinel Errors**: Use sentinel errors for type-safe error handling instead of string matching
  ```go
  var (
      ErrWishListNotFound  = errors.New("wishlist not found")
      ErrWishListForbidden = errors.New("not authorized to access this wishlist")
  )

  // Check with errors.Is()
  if errors.Is(err, services.ErrWishListNotFound) {
      return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
  }
  ```
- **Never use string matching** like `strings.Contains(err.Error(), "not found")` - it's brittle and error-prone
- **Wrap errors** with `fmt.Errorf("%w", err)` to preserve error types for `errors.Is()` checks

### HTTP Status Codes
- **401 Unauthorized**: Authentication required (missing or invalid token)
- **403 Forbidden**: Authenticated but not authorized (ownership/permission check failed)
- **404 Not Found**: Resource doesn't exist
- **500 Internal Server Error**: Unexpected server errors only
- **Important**: Authorization failures should return 403, NOT 500

### Context Hierarchy & Lifecycle Management
- **Use parent context hierarchy**: Pass application context to services instead of creating separate contexts
  ```go
  // In main.go
  appCtx, appCancel := context.WithCancel(context.Background())
  defer appCancel()

  // Pass to services
  accountCleanupService.StartScheduledCleanup(appCtx)
  ```
- **Single source of truth**: Application context controls all background goroutines
- **Graceful shutdown**: Cancel parent context to stop all child goroutines automatically

### Graceful Shutdown Pattern
1. Create application context at startup
2. Pass context to all background services
3. Use `select` in goroutines to monitor context cancellation:
   ```go
   select {
   case <-ticker.C:
       // Do work
   case <-ctx.Done():
       log.Println("Shutting down...")
       return
   }
   ```
4. On shutdown signal, cancel context and stop tickers

### Docker & Security
- **Never hardcode credentials** in docker-compose.yml
- **Use environment variable interpolation**:
  ```yaml
  DATABASE_URL: ${DATABASE_URL:-postgresql://user:password@postgres:5432/db}
  ```
- **Multi-stage builds**: Use builder stage for compilation, minimal runtime stage
- **Health checks**: Implement proper health checks for all services
- **Non-root users**: Always run containers as non-root user for security

### Database Testing
- **Use sqlmock** for testing database interactions without real database:
  ```go
  mockDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
  sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
  ```
- **Test both success and failure cases** (connection success, connection failure, timeouts)
- **Verify expectations**: Always call `mock.ExpectationsWereMet()` at end of tests

### Type Conversions & NULL Handling
- **pgtype.Numeric to float64**: Always check `Valid` field and handle conversion errors
  ```go
  var price float64
  if item.Price.Valid {
      priceValue, err := item.Price.Float64Value()
      if err == nil && priceValue.Valid {
          price = priceValue.Float64
      }
  }
  ```
- **pgtype.Text**: Check `Valid` field before accessing `String`
  ```go
  // WRONG - crashes if FirstName.Valid is false
  userName := user.FirstName.String

  // CORRECT - safe NULL handling
  var userName string
  if user.FirstName.Valid {
      userName = user.FirstName.String
  }
  if user.LastName.Valid {
      if userName != "" {
          userName += " "
      }
      userName += user.LastName.String
  }
  ```
- **pgtype.UUID**: Use `Scan()` method for parsing string UUIDs
- **pgtype.Date**: Parse RFC3339 strings using `time.Parse(time.RFC3339, dateString)`
- **Safe pattern for nullable fields**: Always check `.Valid` before accessing `.String`, `.Int32`, etc.

### Transaction Safety & Atomicity
- **Wrap related operations in transactions**: Use database transactions to ensure atomicity for multi-step operations
  ```go
  tx, err := s.db.BeginTxx(ctx, nil)
  if err != nil {
      return fmt.Errorf("failed to start transaction: %w", err)
  }
  defer tx.Rollback() // Auto-rollback on panic or early return

  // Perform all operations within transaction
  if err := repo.DeleteWithExecutor(ctx, tx, id); err != nil {
      return err // Rollback happens automatically
  }

  // Commit only after all operations succeed
  if err := tx.Commit(); err != nil {
      return fmt.Errorf("failed to commit: %w", err)
  }
  ```
- **Send notifications after commit**: Never send emails or external notifications inside a transaction
  - Collect notification data during transaction
  - Send notifications only after successful commit
  - If notifications fail, don't rollback the transaction
- **Return errors immediately**: Don't log and continue - return errors so transaction can rollback
- **Use repository methods within transactions**: Pass transaction executor to repository methods

### Repository Pattern & Architecture
- **Never bypass repositories**: All database operations must go through repository layer
  ```go
  // WRONG - Service layer using raw SQL
  _, err = tx.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)

  // CORRECT - Service layer using repository
  if err := s.userRepo.DeleteWithExecutor(ctx, tx, id); err != nil {
      return err
  }
  ```
- **Executor Pattern for transactions**: Repositories accept `db.Executor` interface to work with both DB and Tx
  ```go
  // Executor interface in db package
  type Executor interface {
      ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
      QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row
      GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
      SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
  }

  // Repository implementation
  func (r *UserRepository) Delete(ctx context.Context, id pgtype.UUID) error {
      return r.DeleteWithExecutor(ctx, r.db, id)
  }

  func (r *UserRepository) DeleteWithExecutor(ctx context.Context, executor db.Executor, id pgtype.UUID) error {
      query := "DELETE FROM users WHERE id = $1"
      result, err := executor.ExecContext(ctx, query, id)
      // ... error handling
  }
  ```
- **Benefits of Executor pattern**:
  - Maintains clean separation of concerns
  - Fully testable with mocks
  - Works with or without transactions
  - Single source of truth for database logic
  - No layer violations (service → repository → database)

### Repository Constructor Best Practices
- **Interface-based Dependency Injection**: Constructors must return interfaces, not concrete types
  ```go
  // WRONG - returns concrete type
  func NewUserRepository(database *db.DB) *UserRepository {
      return &UserRepository{db: database}
  }

  // CORRECT - returns interface
  func NewUserRepository(database *db.DB) UserRepositoryInterface {
      return &UserRepository{db: database}
  }
  ```
- **Why this matters**: Interface return types enable proper dependency injection and make mocking easier in tests
- **Apply consistently**: All repository constructors should follow this pattern across the codebase
- **Multiple constructors**: If you have variants (e.g., `NewUserRepositoryWithEncryption`), all must return the interface

### Schema Evolution and Migration Impact
- **Migration 000005 Example**: When `gift_items.wishlist_id` was removed and replaced with many-to-many via `wishlist_items`:
  - **All SQL queries must be updated**: Any JOIN from `gift_items` to `wishlists` must go through the junction table
  - **Reservation queries affected**: `reservations` gained its own `wishlist_id` column (NOT NULL)
  - **Insert statements need updates**: New columns must be added to INSERT and SELECT/RETURNING clauses

- **Common migration pitfalls**:
  ```go
  // BROKEN - gi.wishlist_id no longer exists after migration 000005
  LEFT JOIN wishlists w ON gi.wishlist_id = w.id

  // CORRECT - use junction table
  LEFT JOIN wishlist_items wi ON wi.gift_item_id = gi.id
  LEFT JOIN wishlists w ON wi.wishlist_id = w.id
  ```

- **Verification checklist after schema migrations**:
  1. Search codebase for old column names (e.g., `git grep "gi.wishlist_id"`)
  2. Update all JOIN clauses in repositories
  3. Add new columns to INSERT statements
  4. Add new columns to SELECT and RETURNING clauses
  5. Update struct field mappings in service layer
  6. Run all tests to catch compilation errors
  7. Check handler tests for obsolete field references

### Test Maintenance After Refactoring
- **Remove obsolete tests promptly**: When handlers/methods are moved or deleted, remove their tests immediately
- **Dead code causes compile errors in Go**:
  - Unused helper functions (`stringPtr`, `float64Ptr`) must be deleted if no callers exist
  - Go compiler will fail on unused functions in test files
- **Mock methods can stay**: Even if no tests call them, mock methods satisfying interfaces won't cause errors
- **Field name changes propagate**: Schema changes (e.g., `WishlistID` → `OwnerID`) break tests referencing old fields
- **Test cleanup pattern**:
  1. Identify tests for moved/deleted functionality
  2. Add comment explaining why tests were removed
  3. Remove entire test functions, not just mark as skipped
  4. Remove unused helper functions and types
  5. Verify build and tests pass

### Logging Best Practices
- **Structured Logging with slog**: Always use `internal/pkg/logger` package for structured JSON logging
  ```go
  import "wish-list/internal/pkg/logger"

  // CORRECT - structured logging with context
  logger.Info("user logged in", "user_id", userID, "session_id", sessionID)
  logger.Error("failed to create item", "error", err, "item_id", itemID)
  logger.Warn("retry attempt", "attempt", retryCount, "max_retries", maxRetries)

  // WRONG - using fmt.Printf or log.Printf
  fmt.Printf("User %s logged in\n", userID) // ❌ Never use in production code
  log.Printf("Error: %v", err)               // ❌ Use logger.Error instead
  ```
- **Conditional success logging**: Only log success when operations actually succeed
  ```go
  // WRONG - logs success even on failure
  if err := sendEmail(); err != nil {
      logger.Error("failed to send email", "error", err)
  }
  logger.Info("email sent successfully") // Always executes!

  // CORRECT - success only logged when err == nil
  if err := sendEmail(); err != nil {
      logger.Error("failed to send email", "error", err)
  } else {
      logger.Info("email sent successfully", "recipient", recipientEmail)
  }
  ```
- **Error context**: Always include relevant IDs and context in error logs for debugging
  ```go
  logger.Error("database query failed", "error", err, "user_id", userID, "query", "GetByID")
  ```
- **Log levels**:
  - `logger.Debug()` - Development debugging (disabled in production)
  - `logger.Info()` - Important operational events (user actions, system state)
  - `logger.Warn()` - Recoverable errors, degraded functionality
  - `logger.Error()` - Errors requiring attention, failed operations
- **Logger initialization**: Automatically initialized in `app.New()` based on `SERVER_ENV`
  - `development` - Debug level, verbose output
  - `production` - Info level, JSON formatting
  - `test` - Warn level, minimal output
- **No PII in logs**: Never log emails, names, or other PII in plaintext (Constitution Requirement CR-004)
  ```go
  // WRONG - exposes PII
  logger.Info("user registered", "email", email, "name", name)

  // CORRECT - use IDs or redact PII
  logger.Info("user registered", "user_id", userID)
  ```

### Production Code Quality
- **Remove debug statements**: Never leave `fmt.Printf` or `log.Printf` debug statements in production code
- **Use structured logging**: Always use `internal/pkg/logger` for all logging (see Logging Best Practices)
  ```go
  // ❌ WRONG - debug statements in production
  fmt.Printf("Debug: user_id = %s\n", userID)
  log.Printf("Processing request: %v", req)

  // ✅ CORRECT - structured logging with logger package
  logger.Debug("processing request", "user_id", userID, "request", req)
  logger.Info("request completed", "duration_ms", duration.Milliseconds())
  ```
- **Clean code**: Remove commented-out code, TODOs, and temporary hacks before committing
