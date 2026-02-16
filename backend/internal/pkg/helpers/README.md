# Backend Helper Functions

Helper functions to reduce code duplication in HTTP handlers.

## 📦 Available Helpers

### 1. Authentication Helpers (`auth/helpers.go`)

#### `auth.MustGetUserID(c echo.Context) string`
Extracts user ID from context in **protected routes only**.

**Before** (4 lines):
```go
userID, _, _, err := auth.GetUserFromContext(c)
if err != nil || userID == "" {
    return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Not authenticated"})
}
```

**After** (1 line):
```go
userID := auth.MustGetUserID(c)
```

#### `auth.MustGetUserInfo(c echo.Context) (userID, email, userType string)`
Extracts all user info from context in **protected routes only**.

```go
userID, email, userType := auth.MustGetUserInfo(c)
```

**⚠️ Important**: Only use in handlers where `authMiddleware` is applied in `routes.go`!

---

### 2. Pagination Helper (`helpers/pagination.go`)

#### `helpers.ParsePagination(c echo.Context) PaginationParams`
Parses `?page=X&limit=Y` query parameters with validation.

**Defaults**: `page=1`, `limit=10`
**Constraints**: `page >= 1`, `1 <= limit <= 100`

**Before** (12 lines):
```go
page := 1
if pageStr := c.QueryParam("page"); pageStr != "" {
    if parsedPage, err := strconv.Atoi(pageStr); err == nil && parsedPage > 0 {
        page = parsedPage
    }
}

limit := 10
if limitStr := c.QueryParam("limit"); limitStr != "" {
    if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
        limit = parsedLimit
    }
}
```

**After** (1 line):
```go
pagination := helpers.ParsePagination(c)
// pagination.Page, pagination.Limit, pagination.Offset
```

---

### 3. Request Validation Helper (`helpers/request.go`)

#### `helpers.BindAndValidate(c echo.Context, req interface{}) error`
Binds request body to struct and validates it.

**Before** (10 lines):
```go
var req dto.CreateItemRequest
if err := c.Bind(&req); err != nil {
    return c.JSON(http.StatusBadRequest, map[string]string{
        "error": "Invalid request body",
    })
}
if err := c.Validate(&req); err != nil {
    return c.JSON(http.StatusBadRequest, map[string]string{
        "error": err.Error(),
    })
}
```

**After** (4 lines):
```go
var req dto.CreateItemRequest
if err := helpers.BindAndValidate(c, &req); err != nil {
    return err
}
```

---

### 4. UUID Parsing Helper (`helpers/uuid.go`)

#### `helpers.ParseUUID(c echo.Context, uuidStr string) (pgtype.UUID, error)`
Parses string to `pgtype.UUID` with error response.

**Before** (5 lines):
```go
userID := pgtype.UUID{}
if err := userID.Scan(userIDStr); err != nil {
    return c.JSON(http.StatusBadRequest, map[string]string{
        "error": "Invalid UUID format",
    })
}
```

**After** (4 lines):
```go
userID, err := helpers.ParseUUID(c, userIDStr)
if err != nil {
    return err
}
```

#### `helpers.MustParseUUID(uuidStr string) pgtype.UUID`
Parses UUID without HTTP error (returns `Valid=false` on failure).

```go
userID := helpers.MustParseUUID(userIDStr)
if !userID.Valid {
    // Handle error manually
}
```

---

### 5. Cookie Helpers (`auth/cookie.go`)

#### `auth.NewRefreshTokenCookie(value string) *http.Cookie`
Creates secure refresh token cookie (httpOnly, secure, SameSite=None, 7 days).

**Before** (8 lines):
```go
c.SetCookie(&http.Cookie{
    Name:     "refreshToken",
    Value:    refreshToken,
    Path:     "/",
    HttpOnly: true,
    Secure:   true,
    SameSite: http.SameSiteNoneMode,
    MaxAge:   7 * 24 * 60 * 60,
})
```

**After** (1 line):
```go
c.SetCookie(auth.NewRefreshTokenCookie(refreshToken))
```

#### `auth.ClearRefreshTokenCookie() *http.Cookie`
Clears refresh token cookie (for logout).

**Before** (9 lines):
```go
c.SetCookie(&http.Cookie{
    Name:     "refreshToken",
    Value:    "",
    Path:     "/",
    HttpOnly: true,
    Secure:   true,
    SameSite: http.SameSiteNoneMode,
    MaxAge:   -1,
    Expires:  time.Unix(0, 0),
})
```

**After** (1 line):
```go
c.SetCookie(auth.ClearRefreshTokenCookie())
```

---

## 📊 Impact Summary

| Helper | Saves | Usage Count | Total Saved |
|--------|-------|-------------|-------------|
| `MustGetUserID` | ~4 lines | 21+ handlers | ~84 lines |
| `ParsePagination` | ~12 lines | 4 handlers | ~48 lines |
| `BindAndValidate` | ~6 lines | 15+ handlers | ~90 lines |
| `ParseUUID` | ~4 lines | 5 handlers | ~20 lines |
| `NewRefreshTokenCookie` | ~7 lines | 3 handlers | ~21 lines |
| **TOTAL** | | | **~263 lines** |

## 🎯 Next Steps

To apply these helpers to existing handlers, run:

```bash
# See which handlers can be refactored
grep -r "auth.GetUserFromContext" backend/internal/domain/*/delivery/http/handler.go
grep -r "c.QueryParam(\"page\")" backend/internal/domain/*/delivery/http/handler.go
grep -r "c.Bind(&req)" backend/internal/domain/*/delivery/http/handler.go
```

Then refactor handlers one domain at a time:
1. `item/delivery/http/handler.go` (highest duplication)
2. `wishlist_item/delivery/http/handler.go`
3. `wishlist/delivery/http/handler.go`
4. `user/delivery/http/handler.go`
5. `auth/delivery/http/handler.go`
6. `reservation/delivery/http/handler.go`
