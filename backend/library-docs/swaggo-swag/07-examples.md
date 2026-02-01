# Common Examples and Patterns

## Multi-line Descriptions

Add descriptions spanning multiple lines:

```go
// @description This is the first line
// @description This is the second line
// @description And so forth.
```

Works for both general API descriptions and route descriptions.

## User-Defined Structs in Responses

### Single Object

```go
// @Success 200 {object} model.Account
```

```go
package model

type Account struct {
    ID   int    `json:"id" example:"1"`
    Name string `json:"name" example:"John Doe"`
}
```

### Array of Objects

```go
// @Success 200 {array} model.Account
```

## Multiple Path Parameters

```go
// GetUserAddress godoc
// @Summary Get user's address
// @Param group_id path int true "Group ID"
// @Param account_id path int true "Account ID"
// @Router /groups/{group_id}/accounts/{account_id}/address [get]
func GetUserAddress(c *gin.Context) {}
```

## Multiple Routes for One Handler

```go
// UpdateAddress godoc
// @Summary Update address
// @Param group_id path int true "Group ID"
// @Param user_id path int true "User ID"
// @Router /groups/{group_id}/users/{user_id}/address [put]
// @Router /users/{user_id}/address [put]
func UpdateAddress(c *gin.Context) {}
```

## Request and Response Headers

### Request Headers

```go
// @Param Authorization header string true "Bearer token"
// @Param X-API-VERSION header string false "API version" default(1.0)
// @Param X-Request-ID header string false "Request ID for tracing"
```

### Response Headers

```go
// @Success 200 {string} string "ok"
// @Failure 400 {string} string "error"
// @Response default {string} string "other error"
// @Header 200 {string} Location "/entity/1"
// @Header 200,400,default {string} Token "token"
// @Header all {string} X-Rate-Limit "requests remaining"
```

## Struct Examples

### With Example Values

```go
type Account struct {
    ID        int      `json:"id" example:"1"`
    Name      string   `json:"name" example:"account name"`
    Email     string   `json:"email" example:"user@example.com"`
    PhotoUrls []string `json:"photo_urls" example:"http://example.com/1.jpg,http://example.com/2.jpg"`
}
```

### With Description

```go
// Account model info
// @Description User account information
// @Description with user id and username
type Account struct {
    // ID this is userid
    ID   int    `json:"id"`
    Name string `json:"name"` // This is Name
}
```

Generates:
```json
{
  "Account": {
    "type": "object",
    "description": "User account information with user id and username",
    "properties": {
      "id": {
        "type": "integer",
        "description": "ID this is userid"
      },
      "name": {
        "type": "string",
        "description": "This is Name"
      }
    }
  }
}
```

## SchemaExample for Request Body

```go
// @Param email body string true "Email message" SchemaExample(Subject: Test\\r\\n\\r\\nBody Message\\r\\n)
```

## Enum with Descriptions

```go
type StatusFilter struct {
    // Sort order:
    // * asc - Ascending, from A to Z
    // * desc - Descending, from Z to A
    Order string `json:"order" enums:"asc,desc"`
}
```

## File Upload

### Single File

```go
// UploadFile godoc
// @Summary Upload a file
// @Accept multipart/form-data
// @Param file formData file true "File to upload"
// @Success 200 {object} UploadResponse
// @Router /upload [post]
func UploadFile(c *gin.Context) {}
```

### Multiple Files

```go
// UploadFiles godoc
// @Summary Upload multiple files
// @Accept multipart/form-data
// @Param files formData file true "Files to upload"
// @Param description formData string false "Description"
// @Success 200 {object} UploadResponse
// @Router /upload-multiple [post]
func UploadFiles(c *gin.Context) {}
```

## Pagination Pattern

```go
type PaginatedResponse struct {
    Data       []Item         `json:"data"`
    Pagination PaginationInfo `json:"pagination"`
}

type PaginationInfo struct {
    Page       int `json:"page" example:"1"`
    Limit      int `json:"limit" example:"10"`
    Total      int `json:"total" example:"100"`
    TotalPages int `json:"total_pages" example:"10"`
}

// ListItems godoc
// @Summary List items with pagination
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Items per page" default(10) minimum(1) maximum(100)
// @Success 200 {object} PaginatedResponse
// @Router /items [get]
func ListItems(c *gin.Context) {}
```

## Error Responses

### Standard Error Format

```go
type HTTPError struct {
    Code    int    `json:"code" example:"400"`
    Message string `json:"message" example:"Invalid request"`
}

// @Failure 400 {object} HTTPError
// @Failure 401 {object} HTTPError
// @Failure 403 {object} HTTPError
// @Failure 404 {object} HTTPError
// @Failure 500 {object} HTTPError
```

### Detailed Error Format

```go
type DetailedError struct {
    Code    int                    `json:"code" example:"400"`
    Message string                 `json:"message" example:"Validation failed"`
    Errors  map[string][]string    `json:"errors,omitempty"`
}

// @Failure 400 {object} DetailedError
```

## Generic Response Wrapper

```go
type Response struct {
    Success bool        `json:"success" example:"true"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}

// Success with user data
// @Success 200 {object} Response{data=model.User}

// Success with array
// @Success 200 {object} Response{data=[]model.User}

// Error response
// @Failure 400 {object} Response
```

## Search and Filter

```go
// SearchUsers godoc
// @Summary Search users
// @Param q query string false "Search query"
// @Param status query string false "Status filter" enums(active,inactive)
// @Param role query string false "Role filter" enums(admin,user,guest)
// @Param created_after query string false "Created after date" format(date)
// @Param created_before query string false "Created before date" format(date)
// @Param sort_by query string false "Sort field" enums(name,email,created_at) default(created_at)
// @Param sort_order query string false "Sort order" enums(asc,desc) default(desc)
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Page size" default(20) minimum(1) maximum(100)
// @Success 200 {object} PaginatedResponse{data=[]model.User}
// @Router /users/search [get]
func SearchUsers(c *gin.Context) {}
```

## CRUD Operations

```go
// Create
// @Summary Create item
// @Param item body model.CreateItemRequest true "Item data"
// @Success 201 {object} model.Item
// @Failure 400 {object} HTTPError
// @Router /items [post]
func CreateItem(c *gin.Context) {}

// Read (single)
// @Summary Get item by ID
// @Param id path string true "Item ID" format(uuid)
// @Success 200 {object} model.Item
// @Failure 404 {object} HTTPError
// @Router /items/{id} [get]
func GetItem(c *gin.Context) {}

// Read (list)
// @Summary List items
// @Param page query int false "Page" default(1)
// @Param limit query int false "Limit" default(10)
// @Success 200 {array} model.Item
// @Router /items [get]
func ListItems(c *gin.Context) {}

// Update
// @Summary Update item
// @Param id path string true "Item ID" format(uuid)
// @Param item body model.UpdateItemRequest true "Updated data"
// @Success 200 {object} model.Item
// @Failure 400 {object} HTTPError
// @Failure 404 {object} HTTPError
// @Router /items/{id} [put]
func UpdateItem(c *gin.Context) {}

// Delete
// @Summary Delete item
// @Param id path string true "Item ID" format(uuid)
// @Success 204
// @Failure 404 {object} HTTPError
// @Router /items/{id} [delete]
func DeleteItem(c *gin.Context) {}
```

## Authentication Flow

```go
// Register godoc
// @Summary Register new user
// @Tags Authentication
// @Param user body RegisterRequest true "User registration data"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} HTTPError
// @Failure 409 {object} HTTPError "User already exists"
// @Router /auth/register [post]
func Register(c *gin.Context) {}

// Login godoc
// @Summary User login
// @Tags Authentication
// @Param credentials body LoginRequest true "Login credentials"
// @Success 200 {object} AuthResponse
// @Failure 401 {object} HTTPError "Invalid credentials"
// @Router /auth/login [post]
func Login(c *gin.Context) {}

// Logout godoc
// @Summary User logout
// @Tags Authentication
// @Security BearerAuth
// @Success 204
// @Router /auth/logout [post]
func Logout(c *gin.Context) {}

// RefreshToken godoc
// @Summary Refresh access token
// @Tags Authentication
// @Param refresh body RefreshTokenRequest true "Refresh token"
// @Success 200 {object} AuthResponse
// @Failure 401 {object} HTTPError
// @Router /auth/refresh [post]
func RefreshToken(c *gin.Context) {}

type AuthResponse struct {
    User  User   `json:"user"`
    Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}
```

## Webhook Handler

```go
// WebhookHandler godoc
// @Summary Handle webhook
// @Description Process incoming webhook from external service
// @Tags Webhooks
// @Param X-Signature header string true "Webhook signature"
// @Param payload body WebhookPayload true "Webhook data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} HTTPError "Invalid payload"
// @Failure 401 {object} HTTPError "Invalid signature"
// @Router /webhooks/service [post]
func WebhookHandler(c *gin.Context) {}
```

## Batch Operations

```go
type BatchRequest struct {
    IDs []string `json:"ids" example:"[\"id1\",\"id2\",\"id3\"]"`
}

type BatchResponse struct {
    Success []string         `json:"success"`
    Failed  []BatchFailure   `json:"failed"`
}

type BatchFailure struct {
    ID     string `json:"id"`
    Reason string `json:"reason"`
}

// BatchDelete godoc
// @Summary Delete multiple items
// @Param request body BatchRequest true "Item IDs"
// @Success 200 {object} BatchResponse
// @Router /items/batch/delete [post]
func BatchDelete(c *gin.Context) {}
```

## Export/Import

```go
// ExportData godoc
// @Summary Export user data
// @Produce json
// @Produce text/csv
// @Param format query string false "Export format" enums(json,csv) default(json)
// @Success 200 {file} file "Exported data"
// @Security BearerAuth
// @Router /export [get]
func ExportData(c *gin.Context) {}

// ImportData godoc
// @Summary Import data from file
// @Accept multipart/form-data
// @Param file formData file true "Data file (JSON or CSV)"
// @Success 200 {object} ImportResult
// @Security BearerAuth
// @Router /import [post]
func ImportData(c *gin.Context) {}

type ImportResult struct {
    Imported int      `json:"imported" example:"100"`
    Skipped  int      `json:"skipped" example:"5"`
    Errors   []string `json:"errors,omitempty"`
}
```

## Health Check

```go
type HealthResponse struct {
    Status   string            `json:"status" example:"healthy"`
    Version  string            `json:"version" example:"1.0.0"`
    Services map[string]string `json:"services"`
}

// HealthCheck godoc
// @Summary Health check
// @Description Check API and service health
// @Tags System
// @Success 200 {object} HealthResponse
// @Failure 503 {object} HealthResponse
// @Router /health [get]
func HealthCheck(c *gin.Context) {
    response := HealthResponse{
        Status:  "healthy",
        Version: "1.0.0",
        Services: map[string]string{
            "database": "connected",
            "cache":    "connected",
        },
    }
    c.JSON(http.StatusOK, response)
}
```
