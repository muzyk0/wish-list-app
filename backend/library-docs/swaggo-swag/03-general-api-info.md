# General API Info Annotations

These annotations define API-level metadata and are typically placed in `main.go` or your main API file.

## Required Annotations

| Annotation | Description | Example |
|------------|-------------|---------|
| `@title` | **Required.** The title of the application | `// @title Swagger Example API` |
| `@version` | **Required.** Application API version | `// @version 1.0` |

## API Information

| Annotation | Description | Example |
|------------|-------------|---------|
| `@description` | Short description of the application | `// @description This is a sample server.` |
| `@termsOfService` | Terms of Service URL | `// @termsOfService http://swagger.io/terms/` |

## Contact Information

| Annotation | Description | Example |
|------------|-------------|---------|
| `@contact.name` | Contact person/organization name | `// @contact.name API Support` |
| `@contact.url` | Contact URL (must be valid URL format) | `// @contact.url http://www.swagger.io/support` |
| `@contact.email` | Contact email (must be valid email) | `// @contact.email support@swagger.io` |

## License

| Annotation | Description | Example |
|------------|-------------|---------|
| `@license.name` | **Required** if using license. License name | `// @license.name Apache 2.0` |
| `@license.url` | License URL (must be valid URL format) | `// @license.url http://www.apache.org/licenses/LICENSE-2.0.html` |

## Server Configuration

| Annotation | Description | Example |
|------------|-------------|---------|
| `@host` | Host (name or IP) serving the API | `// @host localhost:8080` |
| `@BasePath` | Base path for the API | `// @BasePath /api/v1` |
| `@schemes` | Transfer protocols (space-separated) | `// @schemes http https` |

## Content Types

| Annotation | Description | Example |
|------------|-------------|---------|
| `@accept` | MIME types the API consumes | `// @accept json` |
| `@produce` | MIME types the API produces | `// @produce json` |

**Note**: `@accept` only affects operations with request body (POST, PUT, PATCH).

See [MIME Types](#mime-types) section for available values.

## Query Parameters

| Annotation | Description | Example |
|------------|-------------|---------|
| `@query.collection.format` | Default array param format | `// @query.collection.format multi` |

**Formats**: `csv`, `multi`, `pipes`, `tsv`, `ssv` (default: `csv`)

## Tags

| Annotation | Description | Example |
|------------|-------------|---------|
| `@tag.name` | Name of a tag | `// @tag.name Accounts` |
| `@tag.description` | Description of the tag | `// @tag.description Account management endpoints` |
| `@tag.docs.url` | External documentation URL | `// @tag.docs.url https://example.com/docs` |
| `@tag.docs.description` | External docs description | `// @tag.docs.description Full documentation` |

## External Documentation

| Annotation | Description | Example |
|------------|-------------|---------|
| `@externalDocs.description` | External document description | `// @externalDocs.description OpenAPI` |
| `@externalDocs.url` | External document URL | `// @externalDocs.url https://swagger.io/resources/open-api/` |

## Custom Extensions

| Annotation | Description | Example |
|------------|-------------|---------|
| `@x-name` | Custom extension (must start with `x-`, JSON value) | `// @x-example-key {"key": "value"}` |

## MIME Types

### Aliases

| Alias | MIME Type |
|-------|-----------|
| `json` | `application/json` |
| `xml` | `text/xml` |
| `plain` | `text/plain` |
| `html` | `text/html` |
| `mpfd` | `multipart/form-data` |
| `x-www-form-urlencoded` | `application/x-www-form-urlencoded` |
| `json-api` | `application/vnd.api+json` |
| `json-stream` | `application/x-json-stream` |
| `octet-stream` | `application/octet-stream` |
| `png` | `image/png` |
| `jpeg` | `image/jpeg` |
| `gif` | `image/gif` |
| `event-stream` | `text/event-stream` |

You can also use any valid MIME type matching `*/*` format.

## Complete Example

```go
package main

// @title           Wish List API
// @version         1.0
// @description     A wishlist management API with gift reservation system
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @accept   json
// @produce  json

// @schemes  http https

// @tag.name Users
// @tag.description User management and authentication endpoints

// @tag.name Wishlists
// @tag.description Wishlist CRUD operations

// @tag.name Gift Items
// @tag.description Gift item management within wishlists

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/specification/

func main() {
    r := gin.Default()
    // ... setup routes
}
```

## Markdown Descriptions

For longer descriptions with formatting, images, and code examples:

| Annotation | Description | Example |
|------------|-------------|---------|
| `@description.markdown` | Load description from markdown file | `// @description.markdown` |
| `@tag.description.markdown` | Load tag description from `{tagname}.md` | `// @tag.description.markdown` |

**Usage**:
```sh
swag init --md ./api-docs
```

The markdown files should be in the specified directory:
- `api.md` - Main API description
- `users.md` - Users tag description
- `wishlists.md` - Wishlists tag description

## Dynamic Configuration Example

Set API info programmatically:

```go
package main

import (
    "github.com/gin-gonic/gin"
    echoSwagger "github.com/swaggo/echo-swagger"

    "./docs"
)

func main() {
    // Dynamic configuration
    docs.SwaggerInfo.Title = "My Dynamic API"
    docs.SwaggerInfo.Description = "Description set at runtime"
    docs.SwaggerInfo.Version = "2.0"
    docs.SwaggerInfo.Host = os.Getenv("API_HOST")
    docs.SwaggerInfo.BasePath = "/api/v2"
    docs.SwaggerInfo.Schemes = []string{"https"}

    e := echo.New()
    e.GET("/swagger/*", echoSwagger.WrapHandler)
    e.Start(":8080")
}
```
