# API & Backend Improvements Plan

**Generated**: 2026-02-02 (Updated)
**Source**: Cross-Domain Architecture, PR Issues
**Status**: Ready for Implementation
**Host**: Render

## Architecture Context

```
Backend (Go/Echo) on Render
├── API: api.wishlist.com
├── Auth: JWT (access + refresh tokens)
├── CORS: Allow Frontend + Mobile (Expo Web)
└── Storage: PostgreSQL + S3
```

---

## Part 1: Cross-Domain Auth Endpoints (CRITICAL)

**Priority**: Must be implemented before Frontend/Mobile auth works

### Task 1.1: Add Refresh Token Support
**Files**: `backend/internal/handlers/auth_handler.go`, `backend/internal/services/auth_service.go`
**Priority**: Critical
**Effort**: 2 hours

**Login Response Update**:
```go
// handlers/auth_handler.go

type LoginResponse struct {
    AccessToken  string      `json:"accessToken"`
    RefreshToken string      `json:"refreshToken,omitempty"` // Only for mobile
    User         UserOutput  `json:"user"`
}

// @Summary      Login user
// @Description  Authenticate user and return tokens
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "Login credentials"
// @Success      200 {object} LoginResponse
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c echo.Context) error {
    // ... existing validation

    user, err := h.authService.Authenticate(ctx, req.Email, req.Password)
    if err != nil {
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
    }

    accessToken, err := h.authService.GenerateAccessToken(user.ID)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
    }

    refreshToken, err := h.authService.GenerateRefreshToken(user.ID)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
    }

    // Set refresh token as httpOnly cookie for web clients
    c.SetCookie(&http.Cookie{
        Name:     "refreshToken",
        Value:    refreshToken,
        Path:     "/",
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteNoneMode, // Required for cross-domain
        MaxAge:   7 * 24 * 60 * 60,      // 7 days
    })

    return c.JSON(http.StatusOK, LoginResponse{
        AccessToken:  accessToken,
        RefreshToken: refreshToken, // Also in body for mobile
        User:         toUserOutput(user),
    })
}
```

---

### Task 1.2: Create Refresh Token Endpoint
**File**: `backend/internal/handlers/auth_handler.go`
**Priority**: Critical
**Effort**: 1 hour

```go
// @Summary      Refresh access token
// @Description  Exchange refresh token for new access token
// @Tags         Auth
// @Produce      json
// @Success      200 {object} map[string]string "accessToken"
// @Failure      401 {object} map[string]string "Unauthorized"
// @Router       /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c echo.Context) error {
    // Try cookie first (web clients)
    refreshToken := ""
    cookie, err := c.Cookie("refreshToken")
    if err == nil {
        refreshToken = cookie.Value
    }

    // Fall back to Authorization header (mobile clients)
    if refreshToken == "" {
        auth := c.Request().Header.Get("Authorization")
        if strings.HasPrefix(auth, "Bearer ") {
            refreshToken = strings.TrimPrefix(auth, "Bearer ")
        }
    }

    if refreshToken == "" {
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "No refresh token"})
    }

    // Validate refresh token
    userID, err := h.authService.ValidateRefreshToken(refreshToken)
    if err != nil {
        // Clear invalid cookie
        c.SetCookie(&http.Cookie{
            Name:     "refreshToken",
            Value:    "",
            Path:     "/",
            HttpOnly: true,
            Secure:   true,
            MaxAge:   -1,
        })
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid refresh token"})
    }

    // Generate new tokens
    accessToken, err := h.authService.GenerateAccessToken(userID)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
    }

    newRefreshToken, err := h.authService.GenerateRefreshToken(userID)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
    }

    // Update cookie
    c.SetCookie(&http.Cookie{
        Name:     "refreshToken",
        Value:    newRefreshToken,
        Path:     "/",
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteNoneMode,
        MaxAge:   7 * 24 * 60 * 60,
    })

    return c.JSON(http.StatusOK, map[string]interface{}{
        "accessToken":  accessToken,
        "refreshToken": newRefreshToken,
    })
}
```

---

### Task 1.3: Create Mobile Handoff Endpoint
**File**: `backend/internal/handlers/auth_handler.go`
**Priority**: Critical
**Effort**: 1 hour

```go
// In-memory code store (use Redis in production)
type CodeStore struct {
    mu    sync.RWMutex
    codes map[string]codeEntry
}

type codeEntry struct {
    UserID    uuid.UUID
    ExpiresAt time.Time
}

var codeStore = &CodeStore{
    codes: make(map[string]codeEntry),
}

// @Summary      Generate mobile handoff code
// @Description  Generate short-lived code for Frontend → Mobile auth transfer
// @Tags         Auth
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} map[string]interface{} "code and expiresIn"
// @Router       /auth/mobile-handoff [post]
func (h *AuthHandler) MobileHandoff(c echo.Context) error {
    userID := getUserIDFromContext(c)

    // Generate random code
    code := generateSecureCode(32)

    // Store with 60 second expiry
    codeStore.Set(code, userID, 60*time.Second)

    return c.JSON(http.StatusOK, map[string]interface{}{
        "code":      code,
        "expiresIn": 60,
    })
}

// @Summary      Exchange handoff code for tokens
// @Description  Exchange short-lived code for access and refresh tokens
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body map[string]string true "code"
// @Success      200 {object} LoginResponse
// @Failure      401 {object} map[string]string "Invalid or expired code"
// @Router       /auth/exchange [post]
func (h *AuthHandler) ExchangeCode(c echo.Context) error {
    var req struct {
        Code string `json:"code" validate:"required"`
    }

    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
    }

    // Get and delete code (one-time use)
    userID, ok := codeStore.GetAndDelete(req.Code)
    if !ok {
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid or expired code"})
    }

    // Get user
    user, err := h.userService.GetByID(c.Request().Context(), userID)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "User not found"})
    }

    // Generate tokens
    accessToken, err := h.authService.GenerateAccessToken(userID)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
    }

    refreshToken, err := h.authService.GenerateRefreshToken(userID)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
    }

    return c.JSON(http.StatusOK, LoginResponse{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        User:         toUserOutput(user),
    })
}
```

---

### Task 1.4: Configure CORS for Multiple Domains
**File**: `backend/internal/middleware/cors.go`, `backend/cmd/server/main.go`
**Priority**: Critical
**Effort**: 30 minutes

```go
// middleware/cors.go
package middleware

import (
    "net/http"
    "os"
    "strings"

    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
)

func CORSConfig() middleware.CORSConfig {
    // Get allowed origins from env
    originsStr := os.Getenv("CORS_ALLOWED_ORIGINS")
    origins := strings.Split(originsStr, ",")

    // Add default development origins
    if os.Getenv("ENV") != "production" {
        origins = append(origins,
            "http://localhost:3000",
            "http://localhost:8081",
            "http://localhost:19006", // Expo web
        )
    }

    return middleware.CORSConfig{
        AllowOrigins:     origins,
        AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
        AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
        AllowCredentials: true, // Required for cookies
        MaxAge:           86400,
    }
}
```

```go
// main.go
func main() {
    e := echo.New()

    // CORS - must be before routes
    e.Use(middleware.CORSWithConfig(customMiddleware.CORSConfig()))

    // ... rest of setup
}
```

**Environment Variables (Render)**:
```
CORS_ALLOWED_ORIGINS=https://wishlist.com,https://www.wishlist.com
ENV=production
```

---

### Task 1.5: Add Logout Endpoint
**File**: `backend/internal/handlers/auth_handler.go`
**Priority**: High
**Effort**: 30 minutes

```go
// @Summary      Logout user
// @Description  Clear refresh token cookie and invalidate token
// @Tags         Auth
// @Security     BearerAuth
// @Success      200 {object} map[string]string "message"
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c echo.Context) error {
    // Clear refresh token cookie
    c.SetCookie(&http.Cookie{
        Name:     "refreshToken",
        Value:    "",
        Path:     "/",
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteNoneMode,
        MaxAge:   -1, // Delete cookie
    })

    // Optionally: Add token to blacklist if using Redis

    return c.JSON(http.StatusOK, map[string]string{"message": "Logged out successfully"})
}
```

---

## Part 2: OpenAPI Schema Fixes

### Task 2.1: Fix Scheme-Relative URLs (#14)
**Files**: `api/openapi3.yaml`, `api/split/openapi.yaml`
**Priority**: Medium
**Effort**: 10 minutes

```yaml
servers:
  - url: http://localhost:8080/api
    description: Development server
  - url: https://api.wishlist.com
    description: Production server (Render)
```

---

### Task 2.2: Fix Empty Pagination Schema (#18)
**File**: `api/split/components/schemas/internal_handlers.UserReservationsResponse.yaml`
**Priority**: Medium
**Effort**: 15 minutes

```yaml
pagination:
  type: object
  properties:
    page:
      type: integer
      example: 1
    limit:
      type: integer
      example: 20
    total:
      type: integer
      example: 100
    totalPages:
      type: integer
      example: 5
  required:
    - page
    - limit
    - total
    - totalPages
```

---

### Task 2.3: Add Missing 401/404 Responses (#19, #20)
**Files**: `api/split/paths/*.yaml`
**Priority**: Medium
**Effort**: 20 minutes

```yaml
responses:
  '401':
    description: Authentication required
    content:
      application/json:
        schema:
          $ref: '#/components/schemas/ErrorResponse'
  '404':
    description: Resource not found
    content:
      application/json:
        schema:
          $ref: '#/components/schemas/ErrorResponse'
```

---

### Task 2.4: Fix Security Requirements (#21, #22)
**Files**: Reservation and gift-items paths
**Priority**: High
**Effort**: 15 minutes

```yaml
# Allow both authenticated and guest access
security:
  - BearerAuth: []
  - {} # Anonymous access for public resources
```

---

## Part 3: Render Deployment Configuration

### Task 3.1: Create render.yaml
**File**: `render.yaml`
**Priority**: High
**Effort**: 30 minutes

```yaml
services:
  - type: web
    name: wishlist-api
    env: docker
    region: frankfurt
    plan: starter
    healthCheckPath: /health
    envVars:
      - key: DATABASE_URL
        fromDatabase:
          name: wishlist-db
          property: connectionString
      - key: JWT_SECRET
        generateValue: true
      - key: JWT_ACCESS_TOKEN_EXPIRY
        value: 15m
      - key: JWT_REFRESH_TOKEN_EXPIRY
        value: 7d
      - key: CORS_ALLOWED_ORIGINS
        value: https://wishlist.com,https://www.wishlist.com
      - key: ENV
        value: production
      - key: AWS_S3_BUCKET
        sync: false
      - key: AWS_ACCESS_KEY_ID
        sync: false
      - key: AWS_SECRET_ACCESS_KEY
        sync: false
      - key: AWS_REGION
        value: eu-central-1

databases:
  - name: wishlist-db
    region: frankfurt
    plan: starter
    databaseName: wishlist
    user: wishlist
```

---

### Task 3.2: Update Dockerfile for Render
**File**: `backend/Dockerfile`
**Priority**: High
**Effort**: 15 minutes

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /server ./cmd/server

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install CA certificates for HTTPS
RUN apk add --no-cache ca-certificates tzdata

# Copy binary
COPY --from=builder /server /app/server
COPY --from=builder /app/internal/db/migrations /app/migrations

# Create non-root user
RUN adduser -D -g '' appuser
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["/app/server"]
```

---

### Task 3.3: Add Health Check Endpoint
**File**: `backend/cmd/server/main.go`
**Priority**: High
**Effort**: 10 minutes

```go
// Health check endpoint for Render
e.GET("/health", func(c echo.Context) error {
    // Check database connection
    if err := db.PingContext(c.Request().Context()); err != nil {
        return c.JSON(http.StatusServiceUnavailable, map[string]string{
            "status": "unhealthy",
            "error":  "database connection failed",
        })
    }

    return c.JSON(http.StatusOK, map[string]string{
        "status": "healthy",
    })
})
```

---

## Part 4: REST API Best Practices

### Task 4.1: Add Pagination Support (#77)
**Files**: List handlers
**Priority**: Medium
**Effort**: 4 hours

```go
type PaginatedResponse struct {
    Data       interface{} `json:"data"`
    Pagination Pagination  `json:"pagination"`
}

type Pagination struct {
    Page       int `json:"page"`
    Limit      int `json:"limit"`
    Total      int `json:"total"`
    TotalPages int `json:"totalPages"`
}

func (h *WishListHandler) ListWishLists(c echo.Context) error {
    page, _ := strconv.Atoi(c.QueryParam("page"))
    limit, _ := strconv.Atoi(c.QueryParam("limit"))

    if page < 1 {
        page = 1
    }
    if limit < 1 || limit > 100 {
        limit = 20
    }

    offset := (page - 1) * limit

    wishlists, total, err := h.service.ListWithPagination(ctx, userID, limit, offset)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }

    return c.JSON(http.StatusOK, PaginatedResponse{
        Data: wishlists,
        Pagination: Pagination{
            Page:       page,
            Limit:      limit,
            Total:      total,
            TotalPages: (total + limit - 1) / limit,
        },
    })
}
```

---

### Task 4.2: Add Rate Limiting
**File**: `backend/internal/middleware/rate_limit.go`
**Priority**: High
**Effort**: 1 hour

```go
package middleware

import (
    "net/http"
    "sync"
    "time"

    "github.com/labstack/echo/v4"
    "golang.org/x/time/rate"
)

type RateLimiter struct {
    visitors map[string]*rate.Limiter
    mu       sync.RWMutex
    r        rate.Limit
    b        int
}

func NewRateLimiter(rps float64, burst int) *RateLimiter {
    return &RateLimiter{
        visitors: make(map[string]*rate.Limiter),
        r:        rate.Limit(rps),
        b:        burst,
    }
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    limiter, exists := rl.visitors[ip]
    if !exists {
        limiter = rate.NewLimiter(rl.r, rl.b)
        rl.visitors[ip] = limiter
    }

    return limiter
}

func (rl *RateLimiter) Middleware() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            ip := c.RealIP()
            limiter := rl.getLimiter(ip)

            if !limiter.Allow() {
                return c.JSON(http.StatusTooManyRequests, map[string]string{
                    "error": "Rate limit exceeded",
                })
            }

            return next(c)
        }
    }
}

// Usage in main.go:
// authGroup.Use(NewRateLimiter(5, 10).Middleware()) // 5 req/sec, burst 10
```

---

## Implementation Priority

### Week 1: Critical Auth & CORS
- [ ] Task 1.1: Refresh token support
- [ ] Task 1.2: Refresh endpoint
- [ ] Task 1.3: Mobile handoff endpoints
- [ ] Task 1.4: CORS configuration
- [ ] Task 1.5: Logout endpoint
- [ ] Task 3.3: Health check

### Week 2: Deployment & OpenAPI
- [ ] Task 3.1: render.yaml
- [ ] Task 3.2: Dockerfile update
- [ ] Task 2.1-2.4: OpenAPI fixes

### Week 3: API Improvements
- [ ] Task 4.1: Pagination
- [ ] Task 4.2: Rate limiting

---

## Verification

```bash
cd backend

# 1. Regenerate Swagger
swag init

# 2. Run tests
go test ./...

# 3. Build Docker image
docker build -t wishlist-api .

# 4. Test locally
docker run -p 8080:8080 --env-file .env wishlist-api

# 5. Test CORS
curl -I -X OPTIONS http://localhost:8080/api/auth/login \
  -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: POST"

# 6. Test refresh token
curl -X POST http://localhost:8080/api/auth/refresh \
  -H "Cookie: refreshToken=<token>"
```

---

## Environment Variables for Render

| Variable | Description | Example |
|----------|-------------|---------|
| DATABASE_URL | PostgreSQL connection string | postgres://... |
| JWT_SECRET | Secret for JWT signing | (auto-generated) |
| JWT_ACCESS_TOKEN_EXPIRY | Access token TTL | 15m |
| JWT_REFRESH_TOKEN_EXPIRY | Refresh token TTL | 7d |
| CORS_ALLOWED_ORIGINS | Allowed origins, comma-separated | https://wishlist.com |
| ENV | Environment name | production |
| AWS_S3_BUCKET | S3 bucket for uploads | wishlist-uploads |
| AWS_REGION | AWS region | eu-central-1 |
