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
- **Database**: Use `make generate` to refresh database models for sqlx (manual updates needed)
- **API clients**: Generated from OpenAPI specifications in `/contracts/`
- **Type checking**: Run `npm run type-check` to verify TypeScript correctness
- **Linting & formatting**: Use `make format` for consistent code style across all components

### Mobile Development
- **Navigation**: Uses Expo Router with file-based routing in `/mobile/app/`
- **UI components**: Custom components in `/mobile/components/`
- **API integration**: Uses TanStack Query for data fetching and caching
- **Asset management**: Expo Asset system for images and fonts

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
