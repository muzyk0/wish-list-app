# Advanced Features

## Generics Support

Swag supports Go generics in responses.

### Basic Generic Response

```go
type GenericResponse[T any] struct {
    Success bool   `json:"success"`
    Data    T      `json:"data"`
    Message string `json:"message"`
}

// @Success 200 {object} GenericResponse[model.User]
// @Router /user [get]
func GetUser(w http.ResponseWriter, r *http.Request) {}
```

### Nested Generics

```go
type GenericNestedResponse[T any, U any] struct {
    Primary   T `json:"primary"`
    Secondary U `json:"secondary"`
}

// @Success 200 {object} GenericNestedResponse[model.Post, model.Author]
// @Router /post [get]
func GetPost(w http.ResponseWriter, r *http.Request) {}
```

### Generic in Generic

```go
type GenericInnerType[T any] struct {
    Value T `json:"value"`
}

type GenericOuterType[T any] struct {
    Data T `json:"data"`
}

// @Success 200 {object} GenericOuterType[GenericInnerType[model.Post]]
// @Router /nested [get]
func GetNested(w http.ResponseWriter, r *http.Request) {}
```

### Generic Array

```go
// @Success 200 {object} GenericResponse[[]model.User]
// @Router /users [get]
func ListUsers(w http.ResponseWriter, r *http.Request) {}
```

## Custom Type Support

### Using swaggertype Tag

Override how custom types are represented in Swagger.

#### Override Primitive Type

```go
type Account struct {
    ID sql.NullInt64 `json:"id" swaggertype:"integer"`
}
```

#### Override Struct to Primitive

```go
type TimestampTime struct {
    time.Time
}

type Account struct {
    RegisterTime TimestampTime `json:"register_time" swaggertype:"primitive,integer"`
}
```

#### Override Array Type

```go
type Account struct {
    Coeffs []big.Float `json:"coeffs" swaggertype:"array,number"`
}
```

#### Base64 Encoding Example

```go
type CertificateKeyPair struct {
    Cert []byte `json:"cert" swaggertype:"string" format:"base64" example:"U3dhZ2dlciByb2Nrcw=="`
    Key  []byte `json:"key" swaggertype:"string" format:"base64" example:"U3dhZ2dlciByb2Nrcw=="`
}
```

Generates:
```json
{
  "CertificateKeyPair": {
    "type": "object",
    "properties": {
      "cert": {
        "type": "string",
        "format": "base64",
        "example": "U3dhZ2dlciByb2Nrcw=="
      },
      "key": {
        "type": "string",
        "format": "base64",
        "example": "U3dhZ2dlciByb2Nrcw=="
      }
    }
  }
}
```

## Global Type Overrides

For generated files where you can't add tags, use a `.swaggo` file.

### Create .swaggo File

```
// Replace all NullInt64 with int
replace database/sql.NullInt64 int

// Don't include NullString fields
skip database/sql.NullString

// Replace custom time type
replace github.com/yourapp/pkg.CustomTime string
```

### Go Code

```go
type MyStruct struct {
    ID   sql.NullInt64  `json:"id"`     // Will be: integer
    Name sql.NullString `json:"name"`   // Will be: skipped
}
```

### Directives

| Directive | Syntax | Description |
|-----------|--------|-------------|
| Comment | `//` | Comment line |
| Replace | `replace path/to/type path/to/replacement` | Replace type |
| Skip | `skip path/to/type` | Exclude type |

**Important**: Use full paths to prevent conflicts when multiple packages define the same type name.

### Usage

```sh
swag init --overridesFile .swaggo
```

Or specify a custom file:
```sh
swag init --overridesFile custom-overrides.txt
```

## Ignore Fields

### Using swaggerignore Tag

```go
type Account struct {
    ID       string `json:"id"`
    Name     string `json:"name"`
    Internal int    `swaggerignore:"true"`  // Not in swagger docs
}
```

### Using Global Overrides

In `.swaggo`:
```
skip github.com/yourapp/model.InternalField
```

## Field Extensions

Add custom x-* extensions to struct fields.

```go
type Account struct {
    ID string `json:"id" extensions:"x-nullable,x-abc=def,!x-omitempty"`
}
```

Generates:
```json
{
  "Account": {
    "type": "object",
    "properties": {
      "id": {
        "type": "string",
        "x-nullable": true,
        "x-abc": "def",
        "x-omitempty": false
      }
    }
  }
}
```

**Rules**:
- Extensions must start with `x-`
- Use `!` prefix for boolean false: `!x-omitempty`
- Use `=` for string values: `x-abc=def`
- No `=` means boolean true: `x-nullable`

## Rename Model Display Name

```go
type Resp struct {
    Code int `json:"code"`
}//@name Response
```

The model will appear as "Response" in Swagger docs instead of "Resp".

## Custom Template Delimiters

If your code contains `{{` or `}}`, change template delimiters:

```sh
swag init -td "[[,]]"
```

Now use `[[` and `]]` instead of default `{{` and `}}`.

## Parse Internal and Dependency Packages

### Parse Dependencies

For structs defined in external packages:

```sh
swag init --parseDependency
```

With level control:
```sh
swag init --parseDependencyLevel 1  # Models only
swag init --parseDependencyLevel 2  # Operations only
swag init --parseDependencyLevel 3  # All
```

### Parse Internal Packages

For structs in your internal packages:

```sh
swag init --parseInternal
```

### Both

```sh
swag init --parseDependency --parseInternal
```

## Output Type Selection

Generate only specific file types:

```sh
# Only Go file
swag init --outputTypes go

# Only JSON
swag init --outputTypes json

# Only YAML
swag init --outputTypes yaml

# JSON and YAML (no Go)
swag init --outputTypes json,yaml

# All (default)
swag init --outputTypes go,json,yaml
```

## Markdown Documentation

### Setup

```sh
swag init --md ./api-docs
```

### API Description

**In main.go**:
```go
// @description.markdown
```

**File**: `api-docs/api.md`
```markdown
# My API

This API provides comprehensive wishlist management.

## Features

- User authentication
- Wishlist CRUD operations
- Gift item management
```

### Tag Descriptions

**In main.go**:
```go
// @tag.name Users
// @tag.description.markdown
```

**File**: `api-docs/users.md`
```markdown
# User Management

Endpoints for user registration, authentication, and profile management.
```

### Operation Descriptions

```go
// @description.markdown user-details
```

**File**: `api-docs/user-details.md`
```markdown
# Get User Details

Returns comprehensive user information including:
- Profile data
- Preferences
- Statistics
```

## Code Examples

### Setup

```sh
swag init --codeExampleFiles ./examples
```

### Add to Operation

```go
// @x-codeSample file
// @Summary Get user
// @Router /users/{id} [get]
```

**File**: `examples/get_user.md` (filename matches summary)

## State Machine

For advanced scenarios with state-based docs:

```sh
swag init --state production
```

In code:
```go
// @State production
// @Router /admin [get]
func AdminEndpoint() {}  // Only in production state
```

## Filter by Tags

Generate docs for specific tags only:

```sh
# Include only users and auth tags
swag init --tags "users,auth"

# Exclude internal and deprecated tags
swag init --tags "!internal,!deprecated"
```

## Property Naming Strategies

Control JSON property naming:

```sh
# camelCase (default)
swag init --propertyStrategy camelcase

# snake_case
swag init --propertyStrategy snakecase

# PascalCase
swag init --propertyStrategy pascalcase
```

## Parse Function Bodies

Parse API info from inside function bodies:

```sh
swag init --parseFuncBody
```

```go
func MyHandler() {
    type request struct {
        Field string `json:"field"`
    }

    type response struct {
        Result string `json:"result"`
    }
}

// Use in annotations:
// @Param req body MyHandler.request true "Request"
// @Success 200 {object} MyHandler.response
```

## Collection Format Default

Set default array parameter format:

```sh
swag init --collectionFormat multi
```

Options: `csv`, `ssv`, `tsv`, `pipes`, `multi`

## Required by Default

Make all fields required by default:

```sh
swag init --requiredByDefault
```

Then use `omitempty` to mark optional:
```go
type User struct {
    ID    string `json:"id"`              // Required
    Email string `json:"email,omitempty"` // Optional
}
```

## Multiple Swagger Instances

Generate multiple Swagger docs in one project:

```sh
swag init --instanceName public -g public/main.go
swag init --instanceName admin -g admin/main.go
```

Use in code:
```go
import publicDocs "app/docs/public"
import adminDocs "app/docs/admin"

publicDocs.SwaggerInfo.Title = "Public API"
adminDocs.SwaggerInfo.Title = "Admin API"
```

## Complete Advanced Example

```go
package main

// @title Advanced API
// @version 2.0
// @description.markdown

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {}

type GenericResponse[T any] struct {
    Success bool   `json:"success"`
    Data    T      `json:"data"`
    Meta    Meta   `json:"meta"`
}//@name ApiResponse

type Meta struct {
    Timestamp int64  `json:"timestamp" swaggertype:"integer" format:"int64"`
    RequestID string `json:"request_id" extensions:"x-internal"`
}

type User struct {
    ID        string         `json:"id" format:"uuid"`
    Email     string         `json:"email" format:"email"`
    CreatedAt sql.NullTime   `json:"created_at" swaggertype:"string" format:"date-time"`
    Internal  string         `swaggerignore:"true"`
}//@name UserModel

// GetUser godoc
// @Summary Get user by ID
// @Description.markdown user-details
// @Tags users
// @Param id path string true "User ID" format(uuid)
// @Success 200 {object} GenericResponse[UserModel]
// @Security BearerAuth
// @Router /users/{id} [get]
func GetUser(c echo.Context) error {
    return nil
}
```

Run with:
```sh
swag init \
  --parseDependency \
  --parseInternal \
  --md ./docs/markdown \
  --cef ./docs/examples \
  --propertyStrategy camelcase \
  --outputTypes go,json
```
