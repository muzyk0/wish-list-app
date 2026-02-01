# Getting Started with Swag

## Installation

### Option 1: Go Install (Recommended)
```sh
go install github.com/swaggo/swag/cmd/swag@latest
```

Requires Go 1.19 or newer.

### Option 2: Docker
```sh
docker run --rm -v $(pwd):/code ghcr.io/swaggo/swag:latest
```

### Option 3: Pre-compiled Binary
Download from [release page](https://github.com/swaggo/swag/releases).

## Basic Setup Steps

### 1. Add Annotations to Your Code

Add general API info in `main.go`:

```go
// @title           Swagger Example API
// @version         1.0
// @description     This is a sample server.
// @host            localhost:8080
// @BasePath        /api/v1

func main() {
    r := gin.Default()
    // ... setup routes
}
```

Add operation annotations in handlers:

```go
// ShowAccount godoc
// @Summary      Show an account
// @Description  get account by ID
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Account ID"
// @Success      200  {object}  model.Account
// @Failure      404  {object}  httputil.HTTPError
// @Router       /accounts/{id} [get]
func ShowAccount(ctx *gin.Context) {
    // handler implementation
}
```

### 2. Generate Documentation

Run in your project root (where `main.go` is):

```sh
swag init
```

This generates:
- `docs/docs.go`
- `docs/swagger.json`
- `docs/swagger.yaml`

If your general API info is not in `main.go`:
```sh
swag init -g http/api.go
```

### 3. Import Generated Docs

```go
import _ "your-module-name/docs"
```

### 4. Add Swagger UI Route

For **Gin**:
```go
import (
    "github.com/swaggo/gin-swagger"
    "github.com/swaggo/files"
)

r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
```

For **Echo**:
```go
import (
    echoSwagger "github.com/swaggo/echo-swagger"
)

e.GET("/swagger/*", echoSwagger.WrapHandler)
```

### 5. Access Swagger UI

Run your app and navigate to:
```
http://localhost:8080/swagger/index.html
```

## Format Annotations (Optional)

Format your swagger comments like `go fmt`:

```sh
swag fmt
```

Exclude folders:
```sh
swag fmt -d ./ --exclude ./internal
```

## Supported Web Frameworks

- [gin](http://github.com/swaggo/gin-swagger)
- [echo](http://github.com/swaggo/echo-swagger)
- [buffalo](https://github.com/swaggo/buffalo-swagger)
- [net/http](https://github.com/swaggo/http-swagger)
- [gorilla/mux](https://github.com/swaggo/http-swagger)
- [go-chi/chi](https://github.com/swaggo/http-swagger)
- [fiber](https://github.com/gofiber/swagger)
- [hertz](https://github.com/hertz-contrib/swagger)

## Dynamic Configuration

You can set API info programmatically:

```go
import "./docs"

func main() {
    docs.SwaggerInfo.Title = "Swagger Example API"
    docs.SwaggerInfo.Description = "This is a sample server."
    docs.SwaggerInfo.Version = "1.0"
    docs.SwaggerInfo.Host = "petstore.swagger.io"
    docs.SwaggerInfo.BasePath = "/v2"
    docs.SwaggerInfo.Schemes = []string{"http", "https"}

    r := gin.New()
    r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
    r.Run()
}
```

## Common Issues

### Import Path
Make sure to use your actual module name:
```go
import _ "github.com/your-username/your-project/docs"
```

### General Info Location
If `swag init` can't find your API info, use `-g`:
```sh
swag init -g cmd/api/main.go
```

### Parse Dependencies
If structs are in external packages:
```sh
swag init --parseDependency
```

If structs are in internal packages:
```sh
swag init --parseInternal
```

Both:
```sh
swag init --parseDependency --parseInternal
```
