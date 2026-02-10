# Feature Specification: Backend Architecture Migration

**Feature Branch**: `003-backend-arch-migration`
**Created**: 2026-02-10
**Status**: Draft
**Input**: User description: "Migrate Go backend to domain-driven architecture with internal/app, internal/pkg, and internal/domain structure"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Developer Navigates Domain Code Independently (Priority: P1)

A developer working on the wishlist feature needs to find all related code (request handling, business logic, data access, data contracts) in a single domain directory rather than searching across multiple flat directories (handlers/, services/, repositories/). Each domain module is self-contained with its own delivery, service, repository, and model layers.

**Why this priority**: This is the core value proposition of the migration. A self-contained domain structure reduces cognitive load, speeds up onboarding, and prevents accidental coupling between unrelated domains. Without this, the remaining stories have no foundation.

**Independent Test**: Can be verified by confirming that a developer can find, read, and modify all wishlist-related code within a single domain directory without needing to navigate to other directories for core domain logic.

**Acceptance Scenarios**:

1. **Given** a developer looking for wishlist business logic, **When** they navigate to the wishlist domain directory, **Then** they find service, repository, handler, DTO, and model files within that directory tree.
2. **Given** a developer modifying a domain handler, **When** they check imports, **Then** the handler only imports from its own domain module, shared packages, or application packages, never from another domain's internals.
3. **Given** an existing API endpoint for wishlists, **When** the migration is complete, **Then** the endpoint URL, request format, and response format remain identical to pre-migration behavior.

---

### User Story 2 - Application Infrastructure is Centralized (Priority: P1)

A developer or DevOps engineer needs application-level concerns (configuration, database connection, server setup, middleware, documentation initialization) organized in a dedicated application directory separate from domain logic. This separation makes it clear what code is infrastructure vs. business.

**Why this priority**: Application infrastructure is the foundation that all domains depend on. Centralizing it prevents duplication and ensures consistent configuration, database pooling, and server setup across all domains.

**Independent Test**: Can be verified by confirming that the application directory contains all infrastructure code (config, database, server, router) and that it can boot the application without any domain code present.

**Acceptance Scenarios**:

1. **Given** the application starts, **When** the configuration is loaded, **Then** all environment variables are read from a centralized configuration module within the application directory.
2. **Given** the application starts, **When** the database connection is established, **Then** the connection pool is initialized from the database package with proper health checking.
3. **Given** the application starts, **When** routes are registered, **Then** the router delegates to each domain's route registration function.

---

### User Story 3 - Shared Libraries are Reusable Across Domains (Priority: P2)

A developer building a new domain feature needs access to shared utilities (authentication middleware, response helpers, validation, logging, encryption) from a common shared package directory without importing domain-specific code.

**Why this priority**: Shared packages enable consistency and code reuse across domains. Without properly extracted shared libraries, domains either duplicate utility code or create circular dependencies.

**Independent Test**: Can be verified by confirming that shared packages have zero imports from domain packages and can be independently compiled and tested.

**Acceptance Scenarios**:

1. **Given** a domain handler needs to verify authentication, **When** it imports the auth middleware, **Then** the middleware comes from the shared packages and works identically for all domains.
2. **Given** a domain handler needs to return a standardized response, **When** it imports response helpers, **Then** the helpers come from shared packages with consistent formats for success, error, and validation responses.
3. **Given** a developer adds a new shared utility, **When** they create it in the shared packages directory, **Then** it compiles successfully with no dependencies on any domain package.

---

### User Story 4 - Each Domain Has DTOs Separate from Models (Priority: P2)

A developer working on the API layer needs separate Data Transfer Objects (DTOs) for request/response handling, distinct from internal domain models. This ensures that changes to the API contract don't force changes to internal data structures and vice versa.

**Why this priority**: Separation of DTOs from models is a key architectural boundary. It prevents leaking internal data structures through the API and allows independent evolution of the API contract and the data model.

**Independent Test**: Can be verified by confirming that each domain has request and response types that are distinct from the domain's model types, with explicit conversion functions between them.

**Acceptance Scenarios**:

1. **Given** a wishlist create request, **When** the handler receives it, **Then** it binds to a request DTO (not the domain model) and converts it via a conversion method.
2. **Given** a wishlist is fetched from the database, **When** the handler returns it, **Then** it converts the domain model to a response DTO before serialization.
3. **Given** a new field is added to the internal wishlist model, **When** the API response should not expose that field, **Then** the DTO omits it without requiring API versioning.

---

### User Story 5 - Domain Routes are Self-Registered (Priority: P3)

A developer adding a new domain module needs a clear, standardized way to register its routes without modifying the central router file for every endpoint. Each domain declares its own routes, and the application router simply delegates to each domain.

**Why this priority**: Self-registered routes reduce merge conflicts when multiple developers work on different domains simultaneously and make it immediately clear which routes belong to which domain.

**Independent Test**: Can be verified by confirming each domain has a route registration file and the central router only calls each domain's route registration function.

**Acceptance Scenarios**:

1. **Given** a new domain module is created, **When** the developer adds routes, **Then** they only need to create a route registration file in their domain and register it once in the central router.
2. **Given** the existing API has versioned endpoints under `/api/v1/`, **When** routes are migrated to domain-owned registration, **Then** all existing endpoint paths remain unchanged.

---

### User Story 6 - All Existing Tests Continue to Pass (Priority: P1)

After the migration is complete, all existing unit tests, integration tests, and mock-based tests must continue to pass without modification to test logic (only import paths change).

**Why this priority**: Test continuity proves that the migration preserved all business logic correctly. Broken tests would indicate behavioral regressions introduced during the restructuring.

**Independent Test**: Can be verified by running the full test suite and confirming 100% pass rate with no skipped or modified test assertions.

**Acceptance Scenarios**:

1. **Given** the migration is complete, **When** the full test suite is executed, **Then** all pre-existing tests pass.
2. **Given** a test file was moved to a new domain directory, **When** it is executed, **Then** it produces the same pass/fail results as before the migration.
3. **Given** mock repositories exist for testing, **When** they are moved to the new domain structure, **Then** they continue to satisfy their respective interfaces.

---

### Edge Cases

- What happens when two domains need to call each other's services? The application layer injects one domain's service interface into another at startup. Domains define their own interfaces and never import another domain's internal packages directly.
- How does the system handle the partially migrated state during an incremental migration? Both old and new paths must work simultaneously until a domain is fully migrated.
- What happens to background services (account cleanup, handoff code cleanup) that don't fit a single domain? They reside in the application layer (not as domains), since they are infrastructure support rather than HTTP-serving business domains.
- What happens to the existing API documentation annotations during migration? They must be updated to reflect new package paths while preserving the generated documentation output.
- How are shared database access patterns handled when repositories move to domains? The executor interface and database connection remain in shared infrastructure, while domain repositories import them.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST reorganize all domain-specific code (auth, users, wishlists, items, wishlist-items, reservations, health, storage) into self-contained modules under a domain directory. Health and storage are lightweight domains that serve HTTP endpoints. Background jobs (account cleanup) and cross-cutting services (email) MUST reside in the application layer, not as domains.
- **FR-002**: Each domain module MUST contain its own delivery (handler + DTOs + routes), service, repository, and models sub-packages.
- **FR-003**: System MUST extract application infrastructure (config, database, server, router, documentation) into a dedicated application directory.
- **FR-004**: System MUST extract shared libraries (auth middleware, response helpers, validation, encryption, logging) into a shared packages directory.
- **FR-005**: All existing API endpoints MUST maintain their exact URLs, request formats, and response formats after migration.
- **FR-006**: All existing database migrations MUST remain functional and accessible from their designated location.
- **FR-007**: Domain modules MUST NOT import from other domain modules' internal packages; cross-domain communication MUST use interface injection at the application layer, where each domain defines its own service interfaces and the application startup wires dependencies between domains.
- **FR-008**: The application entry point MUST remain minimal, delegating to application infrastructure for initialization.
- **FR-009**: System MUST preserve all existing third-party integrations (cloud storage, encryption key management, caching) in appropriate shared or domain packages.
- **FR-010**: Each domain MUST have a route registration function that the central router calls to mount domain-specific endpoints.
- **FR-011**: System MUST preserve the existing authentication flow (token-based auth, refresh tokens, mobile handoff) without behavioral changes.
- **FR-012**: Background services (account cleanup, handoff code cleanup) MUST continue to function with the same scheduling and lifecycle behavior.

### Constitution Requirements

- **CR-001**: Code Quality - All code MUST meet high standards of quality, maintainability, and readability
- **CR-002**: Test-First - Unit tests MUST be written for all business logic before implementation
- **CR-003**: API Contracts - All API contracts MUST be explicitly defined using OpenAPI/Swagger specifications
- **CR-004**: Data Privacy - No personally identifiable information (PII) MAY be stored without encryption
- **CR-005**: Semantic Versioning - All releases MUST follow semantic versioning (MAJOR.MINOR.PATCH) standards
- **CR-006**: Specification Checkpoints - Features MUST be fully specified before implementation begins

### Key Entities

- **Domain Module**: A self-contained business capability area (auth, user, wishlist, item, wishlist-item, reservation) with its own handler, service, repository, model, and DTO layers.
- **Application Package**: Infrastructure code that supports all domains: configuration loading, database connection management, server setup, route registration, and API documentation.
- **Shared Library**: Reusable utility packages (auth middleware, response formatting, input validation, encryption, logging) that any domain can import without coupling to another domain.
- **Delivery Layer**: The interface of a domain, containing the handler (request processing), DTOs (request/response contracts), and routes (endpoint registration).
- **Service Layer**: Business logic for a domain that orchestrates repository calls, enforces business rules, and returns domain models.
- **Repository Layer**: Data access for a domain that interacts with the database using the executor pattern for transaction support.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of existing automated tests pass after migration with no changes to test assertions (only import path updates allowed).
- **SC-002**: All 8 domain modules (auth, user, wishlist, item, wishlist-item, reservation, health, storage) are fully self-contained, with zero cross-domain internal package imports.
- **SC-003**: A new developer can locate all code for any single business domain within one directory tree in under 30 seconds.
- **SC-004**: All existing API endpoints return identical responses (status codes, response body structure, headers) before and after migration, verifiable by replaying recorded API requests.
- **SC-005**: The application starts and serves requests within the same time tolerance as before migration (no measurable startup performance regression).
- **SC-006**: Shared packages have zero dependencies on any domain package, verifiable by static import analysis.
- **SC-007**: The central router file contains no domain-specific route definitions, only calls to domain route registration functions.

## Clarifications

### Session 2026-02-10

- Q: How should domains communicate when one needs data or operations from another? → A: Interface injection at startup — the application layer wires domain services together by injecting interfaces defined in each domain. No shared contracts package needed.
- Q: Where should cross-cutting components (health, storage, background jobs, email) live? → A: Health and storage are lightweight domains (they serve HTTP endpoints). Background jobs (account cleanup) and cross-cutting services (email) live in the application layer.

## Assumptions

- The migration is a **structural refactoring only** - no new features, no behavior changes, no database schema changes.
- The existing web framework, database access library, and all current dependencies remain unchanged. No library replacements are part of this migration.
- The partially migrated domains (health and storage already restructured) will be moved to align with the new naming convention as lightweight domains with their own handlers.
- OAuth handler code will be grouped with the auth domain since it handles authentication-related flows.
- The application entry point path may be renamed to align with the target convention, but this is optional and can be deferred.
- Database migration files will remain accessible from their current or a clearly standard location.
- The current shared utilities directory will be split between application infrastructure and reusable shared libraries, then removed.
- Incremental migration is acceptable: domains can be migrated one at a time while the application remains functional.
