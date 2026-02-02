# Swag - Go Swagger Documentation Generator

Swag converts Go annotations to Swagger Documentation 2.0 for popular Go web frameworks.

## Documentation Index

This documentation is split into focused sections for easier navigation:

1. **[Getting Started](./01-getting-started.md)** - Installation and basic setup
2. **[CLI Reference](./02-cli-reference.md)** - Command-line interface and options
3. **[General API Info](./03-general-api-info.md)** - API-level annotations (title, version, host, etc.)
4. **[API Operations](./04-api-operations.md)** - Endpoint-level annotations (routes, params, responses)
5. **[Security](./05-security.md)** - Authentication and authorization annotations
6. **[Attributes & Validation](./06-attributes-validation.md)** - Field validation and constraints
7. **[Examples](./07-examples.md)** - Common usage patterns and code examples
8. **[Advanced Features](./08-advanced-features.md)** - Generics, custom types, overrides

## Quick Links

- **Supported Frameworks**: gin, echo, buffalo, net/http, gorilla/mux, chi, fiber, hertz
- **GitHub**: https://github.com/swaggo/swag
- **Swagger 2.0 Spec**: https://swagger.io/docs/specification/2-0/basic-structure/

## Basic Workflow

1. Add annotations to your API code
2. Install swag: `go install github.com/swaggo/swag/cmd/swag@latest`
3. Run: `swag init`
4. Import generated docs: `import _ "your-module/docs"`
5. Access Swagger UI at `/swagger/index.html`

## Quick Example

```go
// @title           My API
// @version         1.0
// @description     API description
// @host            localhost:8080
// @BasePath        /api/v1

// ShowAccount godoc
// @Summary      Get account by ID
// @Tags         accounts
// @Param        id   path      int  true  "Account ID"
// @Success      200  {object}  model.Account
// @Router       /accounts/{id} [get]
func ShowAccount(c echo.Context) {
    // handler code
}
```

See individual documentation files for detailed information on each topic.
