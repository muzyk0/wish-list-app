# API Documentation

This directory contains auto-generated OpenAPI (Swagger) documentation for the Wish List API.

## Files

- `docs.go` - Generated Go code for swagger documentation
- `swagger.json` - OpenAPI specification in JSON format
- `swagger.yaml` - OpenAPI specification in YAML format

## Viewing the Documentation

### Swagger UI

When the backend server is running, you can view the interactive API documentation at:

```
http://localhost:8080/swagger/index.html
```

### OpenAPI Spec Files

You can also import the generated `swagger.json` or `swagger.yaml` files into tools like:
- [Swagger Editor](https://editor.swagger.io/)
- [Postman](https://www.postman.com/)
- [Insomnia](https://insomnia.rest/)

## Regenerating Documentation

After making changes to API handlers or adding new endpoints, regenerate the documentation:

```bash
# From the project root
make swagger-generate

# Or directly with swag
swag init -g cmd/server/main.go -d backend -o backend/docs --parseDependency --parseInternal
```

## Adding Documentation to Handlers

Use Go doc comments with Swagger annotations before handler functions:

### Example: Authentication Endpoint

```go
// Register godoc
// @Summary Register a new user
// @Description Create a new user account with email and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param user body RegisterRequest true "User registration information"
// @Success 201 {object} AuthResponse "User created successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Router /auth/register [post]
func (h *UserHandler) Register(c echo.Context) error {
    // handler implementation
}
```

### Example: Protected Endpoint

```go
// GetProfile godoc
// @Summary Get user profile
// @Description Get the authenticated user's profile information
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserOutput "User profile"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /protected/profile [get]
func (h *UserHandler) GetProfile(c echo.Context) error {
    // handler implementation
}
```

## Annotation Reference

- `@Summary` - Short description of the endpoint
- `@Description` - Detailed description
- `@Tags` - Group endpoints by tag (e.g., "Authentication", "User", "Wishlist")
- `@Accept` - Request content type (e.g., "json")
- `@Produce` - Response content type (e.g., "json")
- `@Param` - Request parameters (path, query, body)
- `@Success` - Success response (status code + schema)
- `@Failure` - Error response (status code + schema)
- `@Security` - Security scheme (e.g., "BearerAuth")
- `@Router` - API route path and HTTP method

## Documentation Resources

- [Swaggo Documentation](https://github.com/swaggo/swag)
- [OpenAPI Specification](https://swagger.io/specification/)
- [Echo Swagger Integration](https://github.com/swaggo/echo-swagger)
