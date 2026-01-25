# Backend Architecture Migration: From sqlc to sqlx

## Overview
This document describes the migration of the backend database layer from sqlc (SQL code generation) to sqlx (direct database operations).

## Previous Architecture (sqlc)
- SQL queries were defined in `.sql` files in `internal/db/queries/`
- Code generation tool (sqlc) generated Go structs and methods based on SQL queries
- Compile-time type safety for database operations
- Generated types lived in `internal/db/models/`
- Required code generation step before compilation

## Current Architecture (sqlx)
- Direct database operations using sqlx with manual struct scanning
- Queries are embedded directly in repository methods
- Runtime query validation with compile-time type safety for struct fields
- Manual error handling and validation
- Simplified build process without code generation

## Key Changes

### 1. Repository Layer
- All repositories now use `*sqlx.DB` directly
- Queries are constructed as string literals within methods
- Struct scanning used instead of generated methods
- Manual error handling with proper error wrapping

### 2. Database Models
- Moved from generated models to manually defined models in `internal/db/models/models.go`
- Maintains the same field structure but with sqlx compatibility
- Explicit field mapping using struct tags

### 3. Error Handling
- All database errors are now properly wrapped using `fmt.Errorf` with `%w` verb
- Consistent error messages with context
- Improved debugging capabilities

## Benefits of Migration

1. **Simplified Build Process**: No need for code generation steps
2. **Greater Flexibility**: Direct control over queries and operations
3. **Reduced Dependencies**: Fewer external dependencies on sqlc
4. **Easier Debugging**: Direct correlation between code and queries
5. **More Control**: Ability to optimize queries manually

## Files Updated

- `internal/db/models/db.go` - Database connection layer
- `internal/db/models/models.go` - Database model definitions
- `internal/repositories/*` - All repository implementations
- `internal/services/*` - Service layer adjustments
- `cmd/server/main.go` - Initialization changes
- `Makefile` - Build process updates

## Migration Impact

- **Performance**: Same or better performance with manual query optimization
- **Type Safety**: Maintained through struct definitions and Go types
- **Maintainability**: Increased flexibility for complex queries
- **Build Time**: Reduced build time without code generation step