# API Operation Annotations

These annotations document individual API endpoints and are placed above handler functions.

## Basic Operation Info

| Annotation | Description |
|------------|-------------|
| `@summary` | Short summary of the operation |
| `@description` | Detailed explanation of the operation |
| `@description.markdown` | Load description from markdown file (e.g., `details.md`) |
| `@id` | Unique identifier for the operation (must be unique across all operations) |
| `@tags` | Comma-separated list of tags |
| `@deprecated` | Mark endpoint as deprecated |

## Content Types

| Annotation | Description |
|------------|-------------|
| `@accept` | MIME types the operation can consume (POST, PUT, PATCH only) |
| `@produce` | MIME types the operation can produce |

See [General API Info - MIME Types](./03-general-api-info.md#mime-types) for available values.

## Parameters

Format: `@Param name location type required description attributes`

```go
// @Param id path int true "User ID"
// @Param email query string false "Filter by email"
// @Param user body model.CreateUserRequest true "User object"
```

### Parameter Locations (location)

| Location | Description | Example |
|----------|-------------|---------|
| `path` | URL path parameter | `/users/{id}` |
| `query` | URL query parameter | `/users?email=test@example.com` |
| `header` | HTTP header | `Authorization: Bearer token` |
| `body` | Request body | POST/PUT/PATCH payloads |
| `formData` | Form data | Multipart or URL-encoded forms |

### Data Types (type)

| Type | Go Types | Description |
|------|----------|-------------|
| `string` | `string` | String value |
| `integer` | `int`, `uint`, `uint32`, `uint64` | Integer number |
| `number` | `float32`, `float64` | Floating point number |
| `boolean` | `bool` | Boolean value |
| `file` | | File upload |
| Custom | User-defined struct | `model.Account` |
| `array` | Slice | `[]string`, `[]int` |

### Examples

```go
// Path parameters
// @Param id path int true "Account ID"
// @Param slug path string true "Wishlist slug"

// Query parameters
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search query"

// Header parameters
// @Param Authorization header string true "Bearer token"
// @Param X-API-VERSION header string false "API version"

// Body parameters
// @Param user body model.CreateUserRequest true "User data"
// @Param wishlist body model.UpdateWishlistRequest true "Wishlist data"

// Form data
// @Param file formData file true "File to upload"
// @Param name formData string true "File name"
```

## Responses

Format: `@Success|@Failure|@Response code {type} dataType "description"`

```go
// @Success 200 {object} model.User "Success"
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 404 {object} map[string]string "Not Found"
```

### Response Types

| Type | Usage |
|------|-------|
| `{object}` | Single object |
| `{array}` | Array of objects |
| `{string}` | Plain string |
| `{primitive}` | Primitive type (use actual type) |

### Examples

```go
// Single object response
// @Success 200 {object} model.User "User retrieved successfully"

// Array response
// @Success 200 {array} model.User "List of users"

// String response
// @Success 200 {string} string "Operation successful"

// Primitive responses
// @Success 200 {integer} int "Count"
// @Success 200 {boolean} bool "Status"

// Map response
// @Success 200 {object} map[string]interface{} "Dynamic response"

// Multiple status codes
// @Success 200 {object} model.User
// @Success 201 {object} model.User
// @Failure 400 {object} httputil.HTTPError
// @Failure 401 {object} httputil.HTTPError
// @Failure 404 {object} httputil.HTTPError
// @Failure 500 {object} httputil.HTTPError

// Default response
// @Response default {object} httputil.HTTPError "Error"
```

## Response Headers

Format: `@Header statusCode {type} headerName "description"`

```go
// @Header 200 {string} Location "/entity/1"
// @Header 200,400 {string} Token "token"
// @Header all {string} X-Rate-Limit "rate limit"
```

## Router

Format: `@Router path [method]`

```go
// @Router /users [get]
// @Router /users [post]
// @Router /users/{id} [get]
// @Router /users/{id} [put]
// @Router /users/{id} [delete]
```

### Multiple Paths

```go
// @Router /groups/{group_id}/users/{user_id}/address [put]
// @Router /users/{user_id}/address [put]
```

### HTTP Methods

- `[get]`, `[post]`, `[put]`, `[delete]`, `[patch]`, `[head]`, `[options]`

### Deprecated Router

```go
// @deprecatedrouter /old-endpoint [get]
```

## Security

```go
// Single security requirement
// @Security BearerAuth

// OR condition
// @Security ApiKeyAuth
// @Security OAuth2Application[write, admin]

// AND condition
// @Security ApiKeyAuth && Firebase
// @Security OAuth2Application[write, admin] && ApiKeyAuth
```

See [Security](./05-security.md) for security definition details.

## Custom Extensions

```go
// @x-name {"key": "value"}
// @x-codeSample file
```

## Complete Handler Example

```go
package handlers

import (
    "net/http"
    "github.com/labstack/echo/v4"
    "your-app/internal/models"
)

// CreateWishlist godoc
//
// @Summary      Create a new wishlist
// @Description  Create a new wishlist for the authenticated user
// @Tags         Wishlists
// @Accept       json
// @Produce      json
// @Param        wishlist body models.CreateWishlistRequest true "Wishlist data"
// @Success      201 {object} models.WishlistResponse "Wishlist created successfully"
// @Failure      400 {object} map[string]string "Invalid request body"
// @Failure      401 {object} map[string]string "Unauthorized"
// @Failure      500 {object} map[string]string "Internal server error"
// @Security     BearerAuth
// @Router       /wishlists [post]
func (h *WishlistHandler) CreateWishlist(c echo.Context) error {
    // Implementation
    return c.JSON(http.StatusCreated, wishlist)
}

// GetWishlist godoc
//
// @Summary      Get wishlist by ID
// @Description  Get a wishlist by its ID. Public wishlists are accessible to all, private only to owner.
// @Tags         Wishlists
// @Produce      json
// @Param        id path string true "Wishlist ID" format(uuid)
// @Success      200 {object} models.WishlistResponse "Wishlist retrieved successfully"
// @Failure      403 {object} map[string]string "Access denied"
// @Failure      404 {object} map[string]string "Wishlist not found"
// @Security     BearerAuth
// @Router       /wishlists/{id} [get]
func (h *WishlistHandler) GetWishlist(c echo.Context) error {
    // Implementation
    return c.JSON(http.StatusOK, wishlist)
}

// ListWishlists godoc
//
// @Summary      List user's wishlists
// @Description  Get all wishlists owned by the authenticated user
// @Tags         Wishlists
// @Produce      json
// @Param        page query int false "Page number" default(1) minimum(1)
// @Param        limit query int false "Items per page" default(10) minimum(1) maximum(100)
// @Success      200 {array} models.WishlistResponse "List of wishlists"
// @Failure      401 {object} map[string]string "Unauthorized"
// @Failure      500 {object} map[string]string "Internal server error"
// @Security     BearerAuth
// @Router       /wishlists [get]
func (h *WishlistHandler) ListWishlists(c echo.Context) error {
    // Implementation
    return c.JSON(http.StatusOK, wishlists)
}

// UploadImage godoc
//
// @Summary      Upload an image
// @Description  Upload an image file for a gift item
// @Tags         Images
// @Accept       multipart/form-data
// @Produce      json
// @Param        file formData file true "Image file"
// @Param        item_id formData string true "Gift item ID"
// @Success      200 {object} models.ImageResponse "Image uploaded successfully"
// @Failure      400 {object} map[string]string "Invalid file"
// @Failure      401 {object} map[string]string "Unauthorized"
// @Security     BearerAuth
// @Router       /images [post]
func (h *ImageHandler) UploadImage(c echo.Context) error {
    // Implementation
    return c.JSON(http.StatusOK, response)
}
```

## Function-Scoped Struct Declaration

You can declare request/response structs inside a function.

**Naming convention**: `<package>.<function>.<struct>`

```go
// @Param request body main.MyHandler.request true "Request body"
// @Success 200 {object} main.MyHandler.response "Success"
// @Router /test [post]
func MyHandler() {
    type request struct {
        Field string `json:"field"`
    }

    type response struct {
        Result string `json:"result"`
    }
}
```

## Model Composition in Response

Override fields in generic responses:

```go
// Override data field with specific type
// @Success 200 {object} jsonresult.JSONResult{data=proto.Order} "Success"

// Array of objects
// @Success 200 {object} jsonresult.JSONResult{data=[]proto.Order} "Success"

// Primitive types
// @Success 200 {object} jsonresult.JSONResult{data=string} "Success"
// @Success 200 {object} jsonresult.JSONResult{data=[]string} "Success"

// Multiple fields
// @Success 200 {object} jsonresult.JSONResult{data1=string,data2=[]string} "Success"

// Deep nesting
// @Success 200 {object} jsonresult.JSONResult{data=proto.Order{details=proto.OrderDetails}} "Success"
```

```go
type JSONResult struct {
    Code    int          `json:"code"`
    Message string       `json:"message"`
    Data    interface{}  `json:"data"`
}

type Order struct {
    ID   uint        `json:"id"`
    Data interface{} `json:"data"`
}
```
