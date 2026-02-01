# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

The Wish List application is a full-stack application consisting of three main components:
- **Backend**: Go-based REST API using Echo framework with PostgreSQL database
- **Frontend**: Next.js 16 application with React 19 and TypeScript
- **Mobile**: Expo/React Native application for iOS and Android

This project uses a specification-driven development approach with the Specify system to manage feature development.

## Key Development Commands

### Component Installation
- **shadcn/ui components**: Use `pnpm dlx shadcn@latest add [component-name]` to install components (e.g., `pnpm dlx shadcn@latest add button card input`)
- **Expo modules**: Use `npx expo install [package-name]` for Expo-specific packages
- **Regular packages**: Use `pnpm add [package-name]` for general packages

## Architecture Structure

The application follows a microservices architecture with shared components:

- `/backend`: Go-based REST API with JWT authentication, AWS S3 integration, and PostgreSQL database
- `/frontend`: Next.js 16 application using Radix UI components, TanStack Query, and Zod for validation
- `/mobile`: Expo Router-based mobile application with React Navigation
- `/database`: Docker Compose configuration for PostgreSQL database
- `/api`: OpenAPI specifications
- `/specs`: Feature specifications using the Specify system
- `/docs`: Documentation files

## Important Development Aspects

### UI Component Management
- **shadcn/ui**: Use `pnpm dlx shadcn@latest add [component]` to add new components (e.g., button, card, input, skeleton)
- **Component location**: UI components are in `frontend/src/components/ui/`
- **Custom components**: Business-specific components are in `frontend/src/components/[domain]/`

### Code Generation & Type Safety
- **API clients**: Generated from OpenAPI specifications in `/contracts/`
- **Type checking**: Run `npm run type-check` to verify TypeScript correctness
- **Linting & formatting**: Use `make format` for consistent code style across all components

### API Documentation (Swagger/OpenAPI)

The backend uses **swaggo/swag** to generate Swagger/OpenAPI documentation from Go annotations.

#### Documentation Structure

Library documentation is organized in modular, AI-agent-friendly files:

**Location**: `/backend/library-docs/swaggo-swag/`

**Files**:
1. **README.md** - Index and quick reference
2. **01-getting-started.md** - Installation, setup, and basic workflow
3. **02-cli-reference.md** - `swag init` and `swag fmt` commands with all options
4. **03-general-api-info.md** - API-level annotations (`@title`, `@version`, `@host`, `@BasePath`, etc.)
5. **04-api-operations.md** - Endpoint annotations (`@Summary`, `@Param`, `@Success`, `@Router`, etc.)
6. **05-security.md** - Authentication schemes (`@securityDefinitions`, `@Security`)
7. **06-attributes-validation.md** - Field validation and constraints (enums, min/max, format, etc.)
8. **07-examples.md** - Common patterns (CRUD, pagination, file upload, error responses)
9. **08-advanced-features.md** - Generics, custom types, global overrides

#### Quick Reference

**Generate Swagger docs**:
```bash
swag init                           # Generate docs
swag init --parseDependency         # Include external packages
swag init --parseInternal           # Include internal packages
swag fmt                            # Format annotations
```

**Common annotations**:
```go
// General API info (in main.go)
// @title           Wish List API
// @version         1.0
// @description     API description
// @host            localhost:8080
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

// Endpoint annotations (in handlers)
// @Summary      Create wishlist
// @Description  Create a new wishlist for authenticated user
// @Tags         Wishlists
// @Accept       json
// @Produce      json
// @Param        wishlist body CreateWishlistRequest true "Wishlist data"
// @Success      201 {object} WishlistResponse "Success"
// @Failure      400 {object} map[string]string "Bad request"
// @Security     BearerAuth
// @Router       /wishlists [post]
```

#### Important Notes

- **Handler DTOs Required**: Always use handler-specific response types (not service types) in `@Success` and `@Failure` annotations
- **Validation Tags**: Use `validate:"required"` in response DTOs for OpenAPI schema generation
- **Format Tags**: Use `format:"uuid"`, `format:"email"`, etc. for proper schema types
- **Parse Dependencies**: When structs are in external packages, use `swag init --parseDependency`
- **Swagger UI**: Access at `http://localhost:8080/swagger/index.html` after running backend

#### Best Practices

1. **Document as you code**: Add Swagger annotations when creating handlers
2. **Use handler DTOs**: Never expose service types directly in Swagger docs
3. **Validate annotations**: Run `swag init` to catch annotation errors early
4. **Keep examples**: Add `example` tags to struct fields for better API docs
5. **Security first**: Always add `@Security` annotations to protected endpoints

### Mobile Development
- **Navigation**: Uses Expo Router with file-based routing in `/mobile/app/`
- **UI components**: Custom components in `/mobile/components/`
- **API integration**: Uses TanStack Query for data fetching and caching
- **Asset management**: Expo Asset system for images and fonts
- **Deep linking**: Custom URL scheme `wishlistapp://` with support for dynamic routes

#### Expo Router Best Practices

**Dynamic Routes**:
```typescript
// File structure: app/lists/[id]/index.tsx

// Access route parameters
import { useLocalSearchParams } from 'expo-router';

export default function ListDetails() {
  const { id } = useLocalSearchParams(); // Type-safe parameter access
  return <Text>List ID: {id}</Text>;
}
```

**Navigation Methods**:
```typescript
import { Link, router } from 'expo-router';

// Method 1: Declarative with Link component (inline ID)
<Link href="/lists/123">View List</Link>

// Method 2: Declarative with typed params
<Link
  href={{
    pathname: '/lists/[id]',
    params: { id: '123' }
  }}
>
  View List
</Link>

// Method 3: Imperative navigation
router.navigate({
  pathname: '/lists/[id]',
  params: { id: '123' }
});

// Method 4: Simple push
router.push('/lists/123');
```

**Deep Link Handling** (in `_layout.tsx`):
- Use regex matching for parameter extraction (not `split()`)
- Validate parameters before navigation
- Handle both cold start (`Linking.getInitialURL()`) and warm start (`Linking.addEventListener()`)
- Example:
  ```typescript
  const match = path.match(/^lists\/([^\/]+)/);
  if (match && match[1]) {
    router.navigate({
      pathname: '/lists/[id]',
      params: { id: match[1] }
    });
  }
  ```

**OAuth and Authentication**:
- Use `AuthSession.AuthRequest` for OAuth flows (not `WebBrowser.openAuthSessionAsync`)
- Enable PKCE with `usePKCE: true`
- Define discovery endpoints as plain objects typed as `AuthSession.DiscoveryDocument`
- Use `expo-secure-store` for token persistence (not `localStorage`)
- Example:
  ```typescript
  const discovery: AuthSession.DiscoveryDocument = {
    authorizationEndpoint: 'https://accounts.google.com/o/oauth2/v2/auth',
    tokenEndpoint: 'https://oauth2.googleapis.com/token',
  };

  const request = new AuthSession.AuthRequest({
    clientId,
    redirectUri,
    scopes: ['openid', 'profile', 'email'],
    usePKCE: true,
  });

  const result = await request.promptAsync(discovery);
  ```

**Deep Linking Configuration** (in `app.json`):
```json
{
  "expo": {
    "scheme": "wishlistapp",
    "ios": {
      "associatedDomains": ["applinks:lk.domain.com"]
    },
    "android": {
      "intentFilters": [
        {
          "action": "VIEW",
          "autoVerify": true,
          "data": [{ "scheme": "https", "host": "lk.domain.com" }],
          "category": ["BROWSABLE", "DEFAULT"]
        }
      ]
    }
  }
}
```

**Testing Deep Links**:
```bash
# iOS Simulator
xcrun simctl openurl booted wishlistapp://lists/123

# Android Emulator
adb shell am start -W -a android.intent.action.VIEW -d "wishlistapp://lists/123"
```

For detailed deep linking documentation, see `/docs/DEEP_LINKING.md`.

### Frontend Development
- **Routing**: Next.js App Router in `frontend/src/app/`
- **Styling**: Tailwind CSS with Radix UI primitives
- **State management**: TanStack Query for server state, React hooks for local state
- **Forms**: React Hook Form with Zod validation

### Formatting Workflow
- **Automatic formatting**: After making changes, always run `make format` or `npm run format` to ensure consistent code style
- **Frontend formatting**: Run `cd frontend && npm run format` for frontend-specific formatting
- **Mobile formatting**: Run `cd mobile && npm run format` for mobile-specific formatting
- **Pre-commit hook**: Consider setting up a pre-commit hook to automatically format code before committing
- **CI/CD integration**: Formatting checks should be part of the CI pipeline to maintain consistency

## Key Technologies & Dependencies

### Backend
- Go 1.25.5 with Echo framework
- PostgreSQL database with sqlx driver
- JWT authentication system
- AWS S3 for image uploads
- Database migrations with golang-migrate
- Manual database operations with sqlx
- Configuration via environment variables

### Frontend
- Next.js 16.1.1 with React 19.2.3
- TypeScript with strict typing
- Shadcn / Radix UI primitives for accessible components
- Tailwind CSS for styling
- TanStack Query for data fetching
- Zod for schema validation
- Storybook for component development
- Biome for linting and formatting
- openapi-fetch for API client generation

### Mobile
- Expo 54 with Expo Router
- React Navigation for routing
- React Native 0.81.5
- TanStack Query for data fetching
- Biome for linting and formatting
- openapi-fetch for API client generation

## Specification-Driven Development

This project uses the Specify system for specification-driven development:

- `/specs/001-wish-list-app/`: Main feature specification directory
  - `spec.md`: Feature specification with user stories and requirements
  - `plan.md`: Implementation plan with technical architecture
  - `tasks.md`: Detailed implementation tasks organized by phase
  - `data-model.md`: Database schema and entity definitions
  - `research.md`: Technical research and decisions
  - `quickstart.md`: Quick start guide
  - `contracts/`: API contract specifications

### Specification Workflow
1. Features are fully specified in `/specs/[feature-id]/spec.md` before implementation
2. Implementation plan is generated in `/specs/[feature-id]/plan.md`
3. Detailed tasks are created in `/specs/[feature-id]/tasks.md`
4. Development follows the task list with progress tracked in the markdown file

## Development Commands

### Setup & Environment
```bash
make setup                    # Set up the development environment
```

### Running Applications
```bash
make backend                  # Start the backend server
make frontend                 # Start the frontend server
make mobile                   # Start the mobile development server
make db-up                    # Start the database with Docker
```

### Database Operations
```bash
make db-up                    # Start database container
make db-down                  # Stop database container
make migrate-up               # Run database migrations
make migrate-down             # Rollback database migrations
make migrate-create           # Create a new migration
```

### Testing
```bash
make test                     # Run tests for all components
make test-backend             # Run backend tests
make test-frontend            # Run frontend tests
make test-mobile              # Run mobile tests
```

### Linting & Formatting
```bash
make lint                     # Run lint for all components
make format                   # Format all components with Biome
make lint-backend             # Run golangci-lint on backend
make lint-frontend            # Run lint on frontend
make lint-mobile              # Run lint on mobile
```

### Building
```bash
make build                    # Build all components
make build-backend            # Build backend only
make build-frontend           # Build frontend only
```

### Additional Commands
```bash
make help                     # Show all available commands
make clean                    # Clean build artifacts
```

## Project-Specific Information

### Frontend Structure
- Components are located in `frontend/src/components`
- App routes defined in `frontend/src/app` using Next.js App Router
- Storybook configuration in `frontend/.storybook`
- Component stories in `frontend/src/stories`
- API clients generated from OpenAPI specs

### Backend Structure
- Main entry point: `backend/cmd/server/main.go`
- Database layer: `backend/internal/db/models` (using sqlx instead of sqlc)
- Authentication: `backend/internal/auth`
- Middleware: `backend/internal/middleware`
- AWS integration: `backend/internal/aws`
- Configuration: `backend/internal/config`
- Handlers: `backend/internal/handlers`
- Repositories: `backend/internal/repositories` (using sqlx)
- Services: `backend/internal/services`
- Old generated code directory removed (migrated from sqlc to manual sqlx operations)

**Architecture Guide**: For comprehensive backend architecture documentation, see `/docs/Go-Architecture-Guide.md`. This guide covers:
- 3-layer architecture (Handler-Service-Repository)
- The ONE non-negotiable rule: JSON serialization ONLY in handlers
- Complete code examples with good patterns and anti-patterns
- Data flow, validation strategy, testing approach
- Security considerations and when to evolve the architecture

### Mobile Structure
- Routes defined in `mobile/app` using Expo Router
- Components in `mobile/components`
- Hooks in `mobile/hooks`
- API clients generated from OpenAPI specs

### Database Schema
- Managed with Docker Compose in `/database/docker-compose.yml`
- Migrations stored in `backend/internal/db/migrations`
- SQL queries in `backend/internal/db/queries/`
- Schema defined in `/specs/001-wish-list-app/data-model.md`

### API Contracts
- OpenAPI specifications in `/contracts/`
- Generated API clients for frontend and mobile
- Shared contracts ensure consistency across all components

## Development Workflow

1. Use `make setup` to initialize the environment
2. Review specifications in `/specs/001-wish-list-app/` to understand requirements
3. Follow the task list in `/specs/001-wish-list-app/tasks.md` for implementation
4. Start services individually with `make db-up`, `make backend`, `make frontend`, `make mobile`
5. Use Biome for consistent code formatting (`make format`)
6. Run tests with `make test` to ensure code quality
7. Use the Makefile for all common operations to maintain consistency
8. Update task status in `/specs/001-wish-list-app/tasks.md` as you complete items

## Backend Best Practices & Patterns

### Error Handling
- **Sentinel Errors**: Use sentinel errors for type-safe error handling instead of string matching
  ```go
  var (
      ErrWishListNotFound  = errors.New("wishlist not found")
      ErrWishListForbidden = errors.New("not authorized to access this wishlist")
  )

  // Check with errors.Is()
  if errors.Is(err, services.ErrWishListNotFound) {
      return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
  }
  ```
- **Never use string matching** like `strings.Contains(err.Error(), "not found")` - it's brittle and error-prone
- **Wrap errors** with `fmt.Errorf("%w", err)` to preserve error types for `errors.Is()` checks

### HTTP Status Codes
- **401 Unauthorized**: Authentication required (missing or invalid token)
- **403 Forbidden**: Authenticated but not authorized (ownership/permission check failed)
- **404 Not Found**: Resource doesn't exist
- **500 Internal Server Error**: Unexpected server errors only
- **Important**: Authorization failures should return 403, NOT 500

### Context Hierarchy & Lifecycle Management
- **Use parent context hierarchy**: Pass application context to services instead of creating separate contexts
  ```go
  // In main.go
  appCtx, appCancel := context.WithCancel(context.Background())
  defer appCancel()

  // Pass to services
  accountCleanupService.StartScheduledCleanup(appCtx)
  ```
- **Single source of truth**: Application context controls all background goroutines
- **Graceful shutdown**: Cancel parent context to stop all child goroutines automatically

### Graceful Shutdown Pattern
1. Create application context at startup
2. Pass context to all background services
3. Use `select` in goroutines to monitor context cancellation:
   ```go
   select {
   case <-ticker.C:
       // Do work
   case <-ctx.Done():
       log.Println("Shutting down...")
       return
   }
   ```
4. On shutdown signal, cancel context and stop tickers

### Docker & Security
- **Never hardcode credentials** in docker-compose.yml
- **Use environment variable interpolation**:
  ```yaml
  DATABASE_URL: ${DATABASE_URL:-postgresql://user:password@postgres:5432/db}
  ```
- **Multi-stage builds**: Use builder stage for compilation, minimal runtime stage
- **Health checks**: Implement proper health checks for all services
- **Non-root users**: Always run containers as non-root user for security

### Database Testing
- **Use sqlmock** for testing database interactions without real database:
  ```go
  mockDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
  sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
  ```
- **Test both success and failure cases** (connection success, connection failure, timeouts)
- **Verify expectations**: Always call `mock.ExpectationsWereMet()` at end of tests

### Type Conversions & NULL Handling
- **pgtype.Numeric to float64**: Always check `Valid` field and handle conversion errors
  ```go
  var price float64
  if item.Price.Valid {
      priceValue, err := item.Price.Float64Value()
      if err == nil && priceValue.Valid {
          price = priceValue.Float64
      }
  }
  ```
- **pgtype.Text**: Check `Valid` field before accessing `String`
  ```go
  // WRONG - crashes if FirstName.Valid is false
  userName := user.FirstName.String

  // CORRECT - safe NULL handling
  var userName string
  if user.FirstName.Valid {
      userName = user.FirstName.String
  }
  if user.LastName.Valid {
      if userName != "" {
          userName += " "
      }
      userName += user.LastName.String
  }
  ```
- **pgtype.UUID**: Use `Scan()` method for parsing string UUIDs
- **pgtype.Date**: Parse RFC3339 strings using `time.Parse(time.RFC3339, dateString)`
- **Safe pattern for nullable fields**: Always check `.Valid` before accessing `.String`, `.Int32`, etc.

### Transaction Safety & Atomicity
- **Wrap related operations in transactions**: Use database transactions to ensure atomicity for multi-step operations
  ```go
  tx, err := s.db.BeginTxx(ctx, nil)
  if err != nil {
      return fmt.Errorf("failed to start transaction: %w", err)
  }
  defer tx.Rollback() // Auto-rollback on panic or early return

  // Perform all operations within transaction
  if err := repo.DeleteWithExecutor(ctx, tx, id); err != nil {
      return err // Rollback happens automatically
  }

  // Commit only after all operations succeed
  if err := tx.Commit(); err != nil {
      return fmt.Errorf("failed to commit: %w", err)
  }
  ```
- **Send notifications after commit**: Never send emails or external notifications inside a transaction
  - Collect notification data during transaction
  - Send notifications only after successful commit
  - If notifications fail, don't rollback the transaction
- **Return errors immediately**: Don't log and continue - return errors so transaction can rollback
- **Use repository methods within transactions**: Pass transaction executor to repository methods

### Repository Pattern & Architecture
- **Never bypass repositories**: All database operations must go through repository layer
  ```go
  // WRONG - Service layer using raw SQL
  _, err = tx.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)

  // CORRECT - Service layer using repository
  if err := s.userRepo.DeleteWithExecutor(ctx, tx, id); err != nil {
      return err
  }
  ```
- **Executor Pattern for transactions**: Repositories accept `db.Executor` interface to work with both DB and Tx
  ```go
  // Executor interface in db package
  type Executor interface {
      ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
      QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row
      GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
      SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
  }

  // Repository implementation
  func (r *UserRepository) Delete(ctx context.Context, id pgtype.UUID) error {
      return r.DeleteWithExecutor(ctx, r.db, id)
  }

  func (r *UserRepository) DeleteWithExecutor(ctx context.Context, executor db.Executor, id pgtype.UUID) error {
      query := "DELETE FROM users WHERE id = $1"
      result, err := executor.ExecContext(ctx, query, id)
      // ... error handling
  }
  ```
- **Benefits of Executor pattern**:
  - Maintains clean separation of concerns
  - Fully testable with mocks
  - Works with or without transactions
  - Single source of truth for database logic
  - No layer violations (service → repository → database)

### Logging Best Practices
- **Conditional success logging**: Only log success when operations actually succeed
  ```go
  // WRONG - logs success even on failure
  if err := sendEmail(); err != nil {
      log.Printf("Failed: %v", err)
  }
  log.Printf("Success!") // Always executes!

  // CORRECT - success only logged when err == nil
  if err := sendEmail(); err != nil {
      log.Printf("Failed: %v", err)
  } else {
      log.Printf("Success!")
  }
  ```
- **Error context**: Include relevant IDs and context in error logs for debugging
- **No PII in logs**: Never log emails, names, or other PII in plaintext (Constitution Requirement CR-004)

### Production Code Quality
- **Remove debug statements**: Never leave `fmt.Printf` debug statements in production code
- **Use structured logging**: Use proper logging library instead of fmt.Printf
- **Clean code**: Remove commented-out code, TODOs, and temporary hacks before committing

## Important Notes

- The application uses JWT-based authentication across all components
- S3 integration is available for image uploads in the backend
- Database migrations are managed with golang-migrate
- All components share the same OpenAPI specification for API contracts
- Storybook is configured for frontend component development and testing
- Manual database operations with sqlx are used for database access in the backend
- Specification-driven development requires following the documented tasks and updating progress
- The project enforces test-first approach (Constitution Requirement CR-002)
- API contracts must be explicitly defined (Constitution Requirement CR-003)
- Data privacy is enforced with encryption for PII (Constitution Requirement CR-004)

## Conventional Commits

This project follows the Conventional Commits specification for commit messages. This ensures consistent and readable commit history that can be used for automated changelog generation and semantic versioning.

### Format

Commit messages MUST follow this format:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Types

Common types include:
- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation only changes
- `style`: Changes that do not affect the meaning of the code (white-space, formatting, missing semi-colons, etc)
- `refactor`: A code change that neither fixes a bug nor adds a feature
- `perf`: A code change that improves performance
- `test`: Adding missing tests or correcting existing tests
- `build`: Changes that affect the build system or external dependencies
- `ci`: Changes to CI configuration files and scripts
- `chore`: Other changes that don't modify src or test files

### Scope

The scope is an optional part that provides additional contextual information about the change. It should be a noun describing a section of the codebase surrounded by parentheses:

```
feat(auth): add JWT refresh token rotation
fix(api): resolve CORS issues in wishlist endpoints
docs(readme): update installation instructions
```

### Breaking Changes

Breaking changes MUST be indicated with an exclamation mark after the type/scope and optionally with a `BREAKING CHANGE` footer:

```
feat(api)!: change authentication header format

BREAKING CHANGE: The Authorization header now expects "Bearer " prefix
instead of "JWT ".
```
