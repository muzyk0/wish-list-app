# Security Annotations

## Security Definitions (General API Level)

Define authentication schemes in your main API file (usually `main.go`).

### Basic Authentication

```go
// @securityDefinitions.basic BasicAuth
```

**Usage in operations**:
```go
// @Security BasicAuth
```

### API Key Authentication

Format: `@securityDefinitions.apikey name`

**Required parameters**:
- `@in` - Where the API key is located (`header`, `query`)
- `@name` - Name of the header or query parameter
- `@description` - Optional description

**Example**:
```go
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
// @description API key for authentication
```

**Bearer Token (JWT) Example**:
```go
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
```

**Query Parameter Example**:
```go
// @securityDefinitions.apikey ApiKeyQuery
// @in query
// @name api_key
// @description API key passed as query parameter
```

### OAuth2 Application Flow

Format: `@securitydefinitions.oauth2.application name`

**Required parameters**:
- `@tokenUrl` - Token endpoint URL
- `@scope.{name}` - Scope definitions
- `@description` - Optional description

**Example**:
```go
// @securitydefinitions.oauth2.application OAuth2Application
// @tokenUrl https://example.com/oauth/token
// @scope.write Grants write access
// @scope.admin Grants read and write access to administrative information
// @description OAuth2 Application Flow
```

### OAuth2 Implicit Flow

Format: `@securitydefinitions.oauth2.implicit name`

**Required parameters**:
- `@authorizationUrl` - Authorization endpoint URL
- `@scope.{name}` - Scope definitions
- `@description` - Optional description

**Example**:
```go
// @securitydefinitions.oauth2.implicit OAuth2Implicit
// @authorizationUrl https://example.com/oauth/authorize
// @scope.write Grants write access
// @scope.admin Grants administrative access
// @description OAuth2 Implicit Flow
```

### OAuth2 Password Flow

Format: `@securitydefinitions.oauth2.password name`

**Required parameters**:
- `@tokenUrl` - Token endpoint URL
- `@scope.{name}` - Scope definitions
- `@description` - Optional description

**Example**:
```go
// @securitydefinitions.oauth2.password OAuth2Password
// @tokenUrl https://example.com/oauth/token
// @scope.read Grants read access
// @scope.write Grants write access
// @description OAuth2 Password Flow
```

### OAuth2 Access Code Flow

Format: `@securitydefinitions.oauth2.accessCode name`

**Required parameters**:
- `@tokenUrl` - Token endpoint URL
- `@authorizationUrl` - Authorization endpoint URL
- `@scope.{name}` - Scope definitions
- `@description` - Optional description

**Example**:
```go
// @securitydefinitions.oauth2.accessCode OAuth2AccessCode
// @tokenUrl https://example.com/oauth/token
// @authorizationUrl https://example.com/oauth/authorize
// @scope.read Grants read access
// @scope.write Grants write access
// @scope.admin Grants administrative access
// @description OAuth2 Authorization Code Flow
```

## Using Security in Operations

### Single Security Requirement

```go
// @Security BearerAuth
```

### OR Condition (Any of)

User must satisfy **any one** of the security requirements:

```go
// @Security ApiKeyAuth
// @Security OAuth2Application[write, admin]
```

This means: ApiKeyAuth **OR** OAuth2 with write+admin scopes

### AND Condition (All of)

User must satisfy **all** security requirements:

```go
// @Security ApiKeyAuth && Firebase
```

This means: ApiKeyAuth **AND** Firebase authentication

### Complex Conditions

```go
// Option 1: API Key AND Firebase
// Option 2: OAuth2 with scopes AND API Key
// @Security ApiKeyAuth && Firebase
// @Security OAuth2Application[write, admin] && ApiKeyAuth
```

## Complete Security Examples

### Example 1: JWT Bearer Authentication

**In main.go**:
```go
// @title           Wishlist API
// @version         1.0
// @description     A wishlist management API

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
    // ...
}
```

**In handlers**:
```go
// CreateWishlist godoc
// @Summary      Create wishlist
// @Security     BearerAuth
// @Router       /wishlists [post]
func CreateWishlist(c echo.Context) error {
    // Only accessible with valid JWT token
}

// GetPublicWishlist godoc
// @Summary      Get public wishlist (no auth required)
// @Router       /public/wishlists/{slug} [get]
func GetPublicWishlist(c echo.Context) error {
    // Accessible without authentication
}
```

### Example 2: Multiple Authentication Methods

**In main.go**:
```go
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key

// @securitydefinitions.oauth2.accessCode OAuth2
// @tokenUrl https://example.com/oauth/token
// @authorizationUrl https://example.com/oauth/authorize
// @scope.read Read access
// @scope.write Write access
// @scope.admin Admin access
```

**In handlers**:
```go
// Option 1: API Key OR OAuth2
// @Security ApiKeyAuth
// @Security OAuth2[read]
// @Router /users [get]
func GetUsers(c echo.Context) error {}

// Option 2: Both required (AND)
// @Security ApiKeyAuth && OAuth2[admin]
// @Router /admin/settings [put]
func UpdateSettings(c echo.Context) error {}
```

### Example 3: OAuth2 with Different Scopes

**In main.go**:
```go
// @securitydefinitions.oauth2.application OAuth2Application
// @tokenUrl https://example.com/oauth/token
// @scope.read Read access to resources
// @scope.write Write access to resources
// @scope.delete Delete resources
// @scope.admin Full administrative access
```

**In handlers**:
```go
// Read-only access
// @Security OAuth2Application[read]
// @Router /items [get]
func ListItems(c echo.Context) error {}

// Write access required
// @Security OAuth2Application[write]
// @Router /items [post]
func CreateItem(c echo.Context) error {}

// Admin access required
// @Security OAuth2Application[admin]
// @Router /users [delete]
func DeleteUser(c echo.Context) error {}

// Multiple scopes required
// @Security OAuth2Application[read, write]
// @Router /items/{id} [put]
func UpdateItem(c echo.Context) error {}
```

### Example 4: API Key in Query Parameter

**In main.go**:
```go
// @securityDefinitions.apikey ApiKeyQuery
// @in query
// @name apikey
// @description API key passed in URL query string
```

**In handlers**:
```go
// @Security ApiKeyQuery
// @Router /webhook [post]
func HandleWebhook(c echo.Context) error {}
```

## Security Parameter Reference

| Parameter | Description | Example |
|-----------|-------------|---------|
| `@in` | Location of security parameter | `header`, `query` |
| `@name` | Parameter name | `Authorization`, `X-API-Key`, `api_key` |
| `@tokenUrl` | OAuth2 token endpoint | `https://example.com/oauth/token` |
| `@authorizationUrl` | OAuth2 authorization endpoint | `https://example.com/oauth/authorize` |
| `@scope.{name}` | OAuth2 scope definition | `@scope.write Grants write access` |
| `@description` | Security scheme description | `API key authentication` |

## Common Patterns

### Pattern 1: Public + Protected Endpoints

```go
// No security - public endpoint
// @Router /public/wishlists/{slug} [get]
func GetPublicWishlist() {}

// Protected endpoint
// @Security BearerAuth
// @Router /wishlists [post]
func CreateWishlist() {}
```

### Pattern 2: Admin vs User Permissions

```go
// Regular user access
// @Security BearerAuth
// @Router /profile [get]
func GetProfile() {}

// Admin-only access
// @Security BearerAuth && AdminRole
// @Router /admin/users [get]
func ListAllUsers() {}
```

### Pattern 3: Flexible Authentication

```go
// Accept either API key or OAuth2
// @Security ApiKeyAuth
// @Security OAuth2[read]
// @Router /data [get]
func GetData() {}
```

## Testing Security

In Swagger UI:
1. Click the "Authorize" button
2. Enter your credentials/token
3. Click "Authorize"
4. Try protected endpoints

For Bearer tokens:
```
Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

For API keys:
```
your-api-key-here
```
