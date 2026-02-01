# CLI Reference

## swag init

Create Swagger documentation files.

### Basic Usage
```sh
swag init
```

### Common Options

| Option | Short | Default | Description |
|--------|-------|---------|-------------|
| `--generalInfo` | `-g` | `main.go` | Go file with general API info |
| `--dir` | `-d` | `./` | Directories to parse (comma-separated) |
| `--output` | `-o` | `./docs` | Output directory for generated files |
| `--exclude` | | | Exclude directories/files (comma-separated) |
| `--propertyStrategy` | `-p` | `camelcase` | Property naming: `snakecase`, `camelcase`, `pascalcase` |
| `--outputTypes` | `--ot` | `go,json,yaml` | Output file types (comma-separated) |

### Parsing Options

| Option | Default | Description |
|--------|---------|-------------|
| `--parseVendor` | `false` | Parse go files in vendor folder |
| `--parseDependency` / `--pd` | `false` | Parse go files in dependencies |
| `--parseDependencyLevel` / `--pdl` | `0` | 0=disabled, 1=models only, 2=operations only, 3=all |
| `--parseInternal` | `false` | Parse go files in internal packages |
| `--parseFuncBody` | `false` | Parse API info within function bodies |
| `--parseDepth` | `100` | Dependency parse depth |
| `--parseGoList` | `true` | Parse dependency via 'go list' |

### Documentation Options

| Option | Default | Description |
|--------|---------|-------------|
| `--markdownFiles` / `--md` | | Folder with markdown files for descriptions |
| `--codeExampleFiles` / `--cef` | | Folder with code examples for x-codeSamples |
| `--generatedTime` | `false` | Generate timestamp in docs.go |
| `--instanceName` | | Name different swagger document instances |

### Validation & Formatting

| Option | Default | Description |
|--------|---------|-------------|
| `--requiredByDefault` | `false` | Set all fields as required by default |
| `--collectionFormat` / `--cf` | `csv` | Default collection format |
| `--overridesFile` | `.swaggo` | File for global type overrides |
| `--templateDelims` / `--td` | | Custom Go template delimiters (e.g., `[[,]]`) |

### Filtering & State

| Option | Default | Description |
|--------|---------|-------------|
| `--tags` / `-t` | | Filter APIs by tags (comma-separated, `!tag` to exclude) |
| `--state` | | Initial state for state machine |
| `--quiet` / `-q` | `false` | Make logger quiet |

### Full Command Reference

```sh
swag init [options]
```

### Examples

#### Basic with custom general info location
```sh
swag init -g cmd/api/main.go
```

#### Parse dependencies and internal packages
```sh
swag init --parseDependency --parseInternal
```

#### Generate only JSON and YAML (skip Go)
```sh
swag init --outputTypes json,yaml
```

#### Custom output directory
```sh
swag init -o ./swagger-docs
```

#### Exclude specific directories
```sh
swag init --exclude vendor,tmp,test
```

#### With markdown descriptions
```sh
swag init --md ./api-docs
```

#### Filter by tags
```sh
swag init --tags "users,auth"
```

#### Exclude tags
```sh
swag init --tags "!internal,!deprecated"
```

#### Use snake_case for properties
```sh
swag init -p snakecase
```

#### Custom template delimiters (avoid {{ }} conflicts)
```sh
swag init -td "[[,]]"
```

## swag fmt

Format swagger comments like `go fmt`.

### Basic Usage
```sh
swag fmt
```

### Options

| Option | Short | Default | Description |
|--------|-------|---------|-------------|
| `--dir` | `-d` | `./` | Directories to format (comma-separated) |
| `--exclude` | | | Exclude directories/files (comma-separated) |
| `--generalInfo` | `-g` | `main.go` | Go file with general API info |

### Examples

#### Format all files
```sh
swag fmt
```

#### Exclude specific directories
```sh
swag fmt --exclude ./vendor,./internal
```

#### Format specific directory
```sh
swag fmt -d ./handlers
```

### Important Notes

**Function Comments Required**: When using `swag fmt`, ensure you have a standard doc comment for the function. This is because `swag fmt` indents swag comments with tabs, which is only allowed *after* a standard doc comment.

✅ **Correct**:
```go
// ListAccounts lists all existing accounts
//
//  @Summary      List accounts
//  @Description  get accounts
//  @Tags         accounts
//  @Router       /accounts [get]
func ListAccounts(ctx *gin.Context) {}
```

❌ **Incorrect**:
```go
//  @Summary      List accounts
//  @Description  get accounts
//  @Tags         accounts
//  @Router       /accounts [get]
func ListAccounts(ctx *gin.Context) {}
```

## Workflow Example

```sh
# 1. Format annotations
swag fmt

# 2. Generate documentation
swag init --parseDependency --parseInternal

# 3. Run your app
go run main.go

# 4. Access Swagger UI
# Open http://localhost:8080/swagger/index.html
```

## Troubleshooting

### Issue: "No operations defined"
**Solution**: Make sure you have `@Router` annotations and run `swag init` in the correct directory.

### Issue: "Cannot find package"
**Solution**: Use `--parseDependency` or `--parseInternal` flags.

### Issue: "Template parsing error with {{}}"
**Solution**: Use custom delimiters: `swag init -td "[[,]]"`

### Issue: "Struct not found"
**Solution**: Increase parse depth: `swag init --parseDepth 200`
