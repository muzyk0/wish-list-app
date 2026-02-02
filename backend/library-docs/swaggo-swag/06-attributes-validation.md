# Attributes & Validation

## Parameter Attributes

Attributes are added to `@Param` annotations to define validation rules and constraints.

### Basic Syntax

```go
// @Param name location type required description attribute(value)
// @Param id path int true "User ID" minimum(1)
// @Param email query string false "Email filter" format(email)
```

## Validation Attributes

### String Validation

| Attribute | Description | Example |
|-----------|-------------|---------|
| `minlength(n)` | Minimum string length | `minlength(5)` |
| `maxlength(n)` | Maximum string length | `maxlength(100)` |
| `format(type)` | String format | `format(email)`, `format(uuid)` |
| `pattern(regex)` | RegEx pattern validation | `pattern(^[A-Z]+$)` |

```go
// @Param username query string false "Username" minlength(3) maxlength(50)
// @Param email query string false "Email" format(email)
// @Param code query string false "Code" pattern(^[0-9]{6}$)
```

### Numeric Validation

| Attribute | Description | Example |
|-----------|-------------|---------|
| `minimum(n)` | Minimum value | `minimum(1)` |
| `maximum(n)` | Maximum value | `maximum(100)` |
| `multipleOf(n)` | Must be multiple of n | `multipleOf(5)` |

```go
// @Param age query int false "Age" minimum(18) maximum(120)
// @Param price query number false "Price" minimum(0) maximum(99999.99)
// @Param quantity query int false "Quantity" multipleOf(10)
```

### Enums

| Attribute | Description | Example |
|-----------|-------------|---------|
| `enums(v1,v2,...)` | Allowed values | `enums(active,inactive,pending)` |

```go
// @Param status query string false "Status" enums(active,inactive,pending)
// @Param priority query int false "Priority" enums(1,2,3,4,5)
// @Param rate query number false "Rate" enums(1.0,1.5,2.0,2.5)
```

### Default Values

| Attribute | Description | Example |
|-----------|-------------|---------|
| `default(value)` | Default value if not provided | `default(10)` |

```go
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param sort query string false "Sort order" default(asc)
```

### Examples

| Attribute | Description | Example |
|-----------|-------------|---------|
| `example(value)` | Example value | `example(john@example.com)` |

```go
// @Param email query string false "Email" example(user@example.com)
// @Param name query string false "Name" example(John Doe)
```

### Array Parameters

| Attribute | Description | Example |
|-----------|-------------|---------|
| `collectionFormat(format)` | Array format | `collectionFormat(multi)` |

**Formats**:
- `csv` - Comma separated: `foo,bar` (default)
- `ssv` - Space separated: `foo bar`
- `tsv` - Tab separated: `foo\tbar`
- `pipes` - Pipe separated: `foo|bar`
- `multi` - Multiple parameters: `foo=bar&foo=baz`

```go
// @Param tags query []string false "Filter tags" collectionFormat(multi)
// @Param ids query []int false "IDs" collectionFormat(csv)
```

### Extensions

| Attribute | Description | Example |
|-----------|-------------|---------|
| `extensions(x-name=value,...)` | Custom extensions | `extensions(x-nullable,x-example=test)` |

```go
// @Param field query string false "Field" extensions(x-nullable,x-internal)
```

## Struct Field Attributes

Attributes can also be defined in struct tags.

### JSON Tag

```go
type User struct {
    ID    int    `json:"id"`
    Email string `json:"email,omitempty"`  // omitempty marks as optional
    Name  string `json:"name"`
}
```

**Note**: `omitempty` option marks the field as not required in the schema.

### Validation Tags

```go
type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
    Age      int    `json:"age" validate:"min=18,max=120"`
}
```

### Struct Tag Attributes

```go
type Product struct {
    Name     string   `json:"name" minLength:"1" maxLength:"100" example:"Product Name"`
    Price    float64  `json:"price" minimum:"0" maximum:"99999.99" example:"29.99"`
    Quantity int      `json:"quantity" minimum:"0" default:"1"`
    Tags     []string `json:"tags" enums:"new,sale,featured"`
    Status   string   `json:"status" enums:"active,inactive" default:"active"`
}
```

## Available Attributes Reference

### Common Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `validate` | `string` | Validation rules: `required`, `optional` |
| `json` | `string` | JSON serialization options |
| `default` | `*` | Default value |
| `example` | `*` | Example value |

### String Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `minLength` | `integer` | Minimum length |
| `maxLength` | `integer` | Maximum length |
| `pattern` | `string` | RegEx pattern |
| `format` | `string` | Format (email, uuid, date, etc.) |

### Numeric Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `minimum` | `number` | Minimum value |
| `maximum` | `number` | Maximum value |
| `multipleOf` | `number` | Must be multiple of this value |

### Array Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `minItems` | `integer` | Minimum array items *(future)* |
| `maxItems` | `integer` | Maximum array items *(future)* |
| `uniqueItems` | `boolean` | Items must be unique *(future)* |
| `collectionFormat` | `string` | Format for array parameters |

### Other Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `enums` | `[*]` | Allowed values |
| `extensions` | `string` | Custom extensions |

## Complete Examples

### Example 1: User Registration

```go
type RegisterRequest struct {
    Email     string `json:"email" validate:"required,email" example:"user@example.com"`
    Password  string `json:"password" validate:"required,min=8,max=72" minLength:"8" maxLength:"72"`
    FirstName string `json:"first_name" validate:"required" minLength:"1" maxLength:"50"`
    LastName  string `json:"last_name" validate:"required" minLength:"1" maxLength:"50"`
    Age       int    `json:"age" validate:"required,min=18" minimum:"18" maximum:"120"`
}

// @Param user body RegisterRequest true "User registration data"
```

### Example 2: Search with Filters

```go
// GetUsers godoc
// @Param search query string false "Search term" minlength(3)
// @Param status query string false "Status filter" enums(active,inactive,pending) default(active)
// @Param page query int false "Page number" minimum(1) default(1)
// @Param limit query int false "Items per page" minimum(1) maximum(100) default(10)
// @Param sort query string false "Sort field" enums(name,email,created_at) default(created_at)
// @Param order query string false "Sort order" enums(asc,desc) default(desc)
// @Router /users [get]
func GetUsers(c echo.Context) error {}
```

### Example 3: Product with Validation

```go
type Product struct {
    Name        string   `json:"name" validate:"required" minLength:"1" maxLength:"200" example:"Laptop"`
    Description string   `json:"description" maxLength:"1000"`
    Price       float64  `json:"price" validate:"required" minimum:"0" example:"999.99"`
    Quantity    int      `json:"quantity" validate:"required" minimum:"0" default:"1" example:"10"`
    Tags        []string `json:"tags" example:"electronics,computers"`
    Status      string   `json:"status" enums:"draft,active,archived" default:"draft"`
    SKU         string   `json:"sku" pattern:"^[A-Z]{3}-[0-9]{6}$" example:"LAP-123456"`
}
```

### Example 4: File Upload

```go
// UploadImage godoc
// @Summary Upload image
// @Accept multipart/form-data
// @Param file formData file true "Image file"
// @Param alt_text formData string false "Alt text" maxlength(200)
// @Param tags formData []string false "Image tags" collectionFormat(multi)
// @Success 200 {object} ImageResponse
// @Router /images [post]
func UploadImage(c echo.Context) error {}
```

## Validation with go-playground/validator

If using `github.com/go-playground/validator/v10`:

```go
type User struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8,max=72"`
    Age      int    `json:"age" validate:"required,gte=18,lte=120"`
    Phone    string `json:"phone" validate:"omitempty,e164"`
}
```

Swag will recognize common validator tags and generate appropriate schema constraints.

## Enum from Constants

Generate enums from Go constants:

```go
type Status string

const (
    StatusActive   Status = "active"   // Active status
    StatusInactive Status = "inactive" // @name InactiveStatus
    StatusPending  Status = "pending"  // @name PendingStatus Waiting approval
)

type User struct {
    Status Status `json:"status"` // Enum auto-generated from constants
}
```

No need for `enums:""` tag - it's generated from the const declarations.

## Format Types

Common format values:

| Format | Description | Example |
|--------|-------------|---------|
| `email` | Email address | `user@example.com` |
| `uuid` | UUID string | `123e4567-e89b-12d3-a456-426614174000` |
| `uri` | URI string | `https://example.com` |
| `date` | Date (RFC3339) | `2023-01-15` |
| `date-time` | DateTime (RFC3339) | `2023-01-15T10:30:00Z` |
| `password` | Password (hidden in UI) | `********` |
| `byte` | Base64 encoded | `U3dhZ2dlciByb2Nrcw==` |
| `binary` | Binary data | |
| `int32` | 32-bit integer | |
| `int64` | 64-bit integer | |
| `float` | Float number | |
| `double` | Double number | |

```go
type User struct {
    ID        string `json:"id" format:"uuid"`
    Email     string `json:"email" format:"email"`
    Website   string `json:"website" format:"uri"`
    BirthDate string `json:"birth_date" format:"date"`
    CreatedAt string `json:"created_at" format:"date-time"`
}
```
