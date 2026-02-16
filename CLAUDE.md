# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

> For backend-specific patterns and best practices (Go/Echo/PostgreSQL), see [`backend/CLAUDE.md`](backend/CLAUDE.md).

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
- `/docs/plans`: Implementation plans for cross-domain architecture

## Deployment Architecture (Cross-Domain)

The application is deployed across multiple providers with different domains:

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        WISH LIST APPLICATION                             │
├─────────────────────────────────────────────────────────────────────────┤
│  ┌───────────────────┐  ┌───────────────────┐  ┌───────────────────┐   │
│  │  Frontend (Web)   │  │  Mobile (App)     │  │  Backend (API)    │   │
│  │  Next.js          │  │  React Native     │  │  Go/Echo          │   │
│  │  Vercel           │  │  Expo + Vercel    │  │  Render           │   │
│  │                   │  │                   │  │                   │   │
│  │  Features:        │  │  Personal Cabinet:│  │  Endpoints:       │   │
│  │  • View wishlists │  │  • Create lists   │  │  • /auth/*        │   │
│  │  • Reserve items  │  │  • Manage items   │  │  • /wishlists/*   │   │
│  │  • My reservations│  │  • View reserves  │  │  • /reservations/*│   │
│  │  • Redirect to LC │  │  • Settings       │  │  • /public/*      │   │
│  └───────────────────┘  └───────────────────┘  └───────────────────┘   │
│           │                      │                      ▲              │
│           └──────────────────────┴──────────────────────┘              │
│                       HTTPS + JWT + CORS                                │
└─────────────────────────────────────────────────────────────────────────┘
```

| Component | Provider | Purpose |
|-----------|----------|---------|
| Frontend | Vercel | Public pages, guest reservations, auth redirect to Mobile |
| Mobile | Vercel/App Stores | Personal cabinet, create wishlists, manage items |
| Backend | Render | REST API, PostgreSQL, S3 storage |

### Cross-Domain Authentication

Since components are on different domains, **httpOnly cookies cannot be shared**. The authentication strategy:

**Token Storage**:
- **Frontend (Web)**: Access token in memory, refresh token via API call
- **Mobile**: Both tokens in `expo-secure-store`

**Token Lifecycle**:
- Access token: 15 minutes
- Refresh token: 7 days

**Frontend → Mobile Handoff** (OAuth-style):
```
1. User clicks "Personal Cabinet" on Frontend
2. Frontend calls POST /auth/mobile-handoff → receives short-lived code (60s)
3. Frontend redirects to Mobile via Universal Link: wishlistapp://auth?code=xxx
4. Mobile exchanges code for tokens: POST /auth/exchange
5. Mobile stores tokens in SecureStore
```

**Key Backend Endpoints**:
- `POST /auth/login` - Returns accessToken + refreshToken
- `POST /auth/refresh` - Exchange refresh token for new access token
- `POST /auth/mobile-handoff` - Generate code for Frontend→Mobile redirect
- `POST /auth/exchange` - Exchange handoff code for tokens

**CORS Configuration** (Backend):
```go
AllowOrigins: ["https://wishlist.com", "https://www.wishlist.com"]
AllowCredentials: true
```

For detailed implementation, see `/docs/plans/00-cross-domain-architecture-plan.md`.

## Important Development Aspects

### UI Component Management
- **shadcn/ui**: Use `pnpm dlx shadcn@latest add [component]` to add new components (e.g., button, card, input, skeleton)
- **Component location**: UI components are in `frontend/src/components/ui/`
- **Custom components**: Business-specific components are in `frontend/src/components/[domain]/`

### Code Generation & Type Safety
- **API clients**: Generated from OpenAPI specifications in `/contracts/`
- **Type checking**: Run `npm run type-check` to verify TypeScript correctness
- **Linting & formatting**: Use `make format` for consistent code style across all components

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
- Domain-driven 3-layer architecture: Handler → Service → Repository
- Domain modules in `backend/internal/domain/{name}/`
- Shared libraries in `backend/internal/pkg/`
- Domains: auth, user, wishlist, item, wishlist_item, reservation, health, storage
- For detailed structure and patterns, see [`backend/CLAUDE.md`](backend/CLAUDE.md)

### Mobile Structure
- Routes defined in `mobile/app` using Expo Router
- Components in `mobile/components`
- Hooks in `mobile/hooks`
- API clients generated from OpenAPI specs

### Database Schema
- Managed with Docker Compose in `/database/docker-compose.yml`
- Migrations stored in `backend/internal/app/database/migrations/`
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
- **Cross-domain architecture**: Frontend (Vercel), Mobile (Vercel/App Stores), Backend (Render) - see `/docs/plans/`
- **No httpOnly cookies for auth**: Different domains require token-based auth with refresh flow
- **Frontend → Mobile redirect**: Uses OAuth-style handoff with short-lived codes
- S3 integration is available for image uploads in the backend
- Database migrations are managed with golang-migrate
- All components share the same OpenAPI specification for API contracts
- Storybook is configured for frontend component development and testing
- Manual database operations with sqlx are used for database access in the backend
- Specification-driven development requires following the documented tasks and updating progress
- The project enforces test-first approach (Constitution Requirement CR-002)
- API contracts must be explicitly defined (Constitution Requirement CR-003)
- Data privacy is enforced with encryption for PII (Constitution Requirement CR-004)

## Implementation Plans

Implementation plans for the cross-domain architecture are in `/docs/plans/`:

| Plan | Focus |
|------|-------|
| `00-cross-domain-architecture-plan.md` | Auth flow, CORS, handoff - **implement first** |
| `01-frontend-security-and-quality-plan.md` | Token management, Vercel deployment |
| `02-mobile-app-completion-plan.md` | SecureStore, deep links, features |
| `03-api-backend-improvements-plan.md` | Auth endpoints, Render deployment |

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

## Active Technologies
- PostgreSQL (users, tokens metadata), In-memory (handoff codes) (002-cross-domain-implementation)
- Go 1.25.5 + Echo v4.15.0, sqlx v1.4.0, pgx/v5 v5.8.0, golang-jwt/v5 v5.3.1, AWS SDK v2 (003-backend-arch-migration)
- PostgreSQL (via pgx/sqlx), Redis (caching), AWS S3 (file uploads), AWS KMS (encryption) (003-backend-arch-migration)

## Recent Changes
- 002-cross-domain-implementation: Added PostgreSQL (users, tokens metadata), In-memory (handoff codes)
