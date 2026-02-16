# Logger Package

Structured JSON logging using Go's standard `log/slog` package.

## Quick Start

```go
import "wish-list/internal/pkg/logger"

// Logger is automatically initialized in app.New()
// based on SERVER_ENV configuration

// Basic logging
logger.Info("user logged in", "user_id", userID, "session_id", sessionID)
logger.Error("failed to create item", "error", err, "item_id", itemID)
logger.Warn("retry attempt", "attempt", retryCount, "max_retries", maxRetries)
logger.Debug("processing data", "data_size", len(data))
```

## Log Levels

Controlled by `SERVER_ENV` environment variable:

| Environment | Log Level | Use Case |
|-------------|-----------|----------|
| `development`, `dev` | Debug | Verbose debugging output |
| `production`, `prod` | Info | Standard operational logging |
| `test` | Warn | Minimal test output |

## API

### Basic Logging

```go
logger.Debug(msg string, args ...any)
logger.Info(msg string, args ...any)
logger.Warn(msg string, args ...any)
logger.Error(msg string, args ...any)
```

### Context-Aware Logging

```go
logger.InfoContext(ctx context.Context, msg string, args ...any)
logger.ErrorContext(ctx context.Context, msg string, args ...any)
// ... DebugContext, WarnContext
```

### Creating Contextual Loggers

```go
// Create logger with persistent fields
userLogger := logger.With("user_id", userID, "session_id", sessionID)
userLogger.Info("processing request")
userLogger.Error("request failed", "error", err)
```

### Direct Access

```go
// Get underlying slog.Logger for advanced use
slogLogger := logger.GetLogger()
```

## Output Format

All logs are output as JSON:

```json
{
  "time": "2026-02-16T22:13:41.512009+03:00",
  "level": "INFO",
  "message": "user logged in",
  "user_id": "123",
  "session_id": "abc-def"
}
```

## Best Practices

### ✅ Do

```go
// Use structured fields
logger.Info("user action", "user_id", userID, "action", "login")

// Include error objects
logger.Error("database query failed", "error", err, "query", "GetByID")

// Use appropriate log levels
logger.Debug("cache hit", "key", cacheKey)  // Development only
logger.Info("user logged in", "user_id", userID)  // Important events
logger.Warn("retry attempt", "attempt", 3)  // Recoverable issues
logger.Error("failed to send email", "error", err)  // Errors
```

### ❌ Don't

```go
// Don't use fmt.Printf or log.Printf in production code
fmt.Printf("User %s logged in\n", userID)  // ❌
log.Printf("Error: %v", err)  // ❌

// Don't log PII (emails, names, etc.)
logger.Info("user registered", "email", email, "name", name)  // ❌

// Don't concatenate log messages
logger.Info(fmt.Sprintf("User %s logged in", userID))  // ❌

// Use this instead:
logger.Info("user logged in", "user_id", userID)  // ✅
```

## Privacy (CR-004 Compliance)

Never log personally identifiable information (PII) in plaintext:

```go
// ❌ Wrong - exposes PII
logger.Info("user registered", "email", email, "name", name)

// ✅ Correct - use IDs only
logger.Info("user registered", "user_id", userID)

// ✅ Correct - redacted PII for debugging
logger.Debug("email validation", "email_domain", extractDomain(email))
```

## Initialization

Logger is automatically initialized in `app.New()`:

```go
// internal/app/app.go
func New(cfg *config.Config) (*App, error) {
    logger.Initialize(cfg.ServerEnv)
    logger.Info("initializing application", "env", cfg.ServerEnv)
    // ...
}
```

## Testing

```go
import "wish-list/internal/pkg/logger"

func TestSomething(t *testing.T) {
    // Initialize logger for tests
    logger.Initialize("test")

    // Your test code
    logger.Info("test event", "test_id", t.Name())
}
```

## Environment Variables

- `SERVER_ENV` - Controls log level (`development`, `production`, `test`)

No additional configuration required.
