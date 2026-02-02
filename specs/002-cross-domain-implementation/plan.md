# Implementation Plan: Cross-Domain Architecture

**Branch**: `002-cross-domain-implementation` | **Date**: 2026-02-02 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/002-cross-domain-implementation/spec.md`

## Summary

Implement cross-domain authentication architecture enabling secure token handling across Frontend (Next.js/Vercel), Mobile (Expo/React Native), and Backend (Go/Echo/Render) deployed on different domains. Key components: OAuth-style handoff flow for Frontend→Mobile auth transfer, dual token strategy (access + refresh), secure storage patterns per platform, CORS configuration, and rate limiting.

## Technical Context

**Language/Version**:
- Backend: Go 1.21+
- Frontend: TypeScript 5.x, Next.js 16
- Mobile: TypeScript 5.x, React Native 0.81, Expo 54

**Primary Dependencies**:
- Backend: Echo v4, golang-jwt/jwt/v5, golang.org/x/time/rate
- Frontend: Next.js, TanStack Query
- Mobile: expo-secure-store, expo-linking, expo-router

**Storage**: PostgreSQL (users, tokens metadata), In-memory (handoff codes)

**Testing**:
- Backend: go test
- Frontend: Jest, Playwright
- Mobile: Jest

**Target Platform**:
- Backend: Linux containers (Render)
- Frontend: Vercel Edge/Serverless
- Mobile: iOS 15+, Android 10+

**Project Type**: Web + Mobile (multi-platform)

**Performance Goals**:
- Token refresh: <200ms p95
- Handoff code exchange: <500ms p95
- Auth endpoints: handle 1000 req/s

**Constraints**:
- Different domains = no shared httpOnly cookies between Frontend/Mobile
- Handoff codes must be one-time-use, 60s expiry
- Access tokens: 15 min, Refresh tokens: 7 days

**Scale/Scope**:
- 10,000 concurrent authenticated users
- 3 platforms (Frontend, Mobile, Backend)

## Constitution Check

*GATE: PASSED - All principles addressed in design*

| Principle | Status | Implementation |
|-----------|--------|----------------|
| Code Quality | ✅ | Follow existing project patterns, clean architecture |
| Test-First Approach | ✅ | Unit tests for token logic, integration tests for auth flows |
| API Contract Integrity | ✅ | OpenAPI specs in contracts/, Swagger annotations |
| Data Privacy Protection | ✅ | Tokens encrypted in transit (HTTPS), secure storage |
| Semantic Versioning | ✅ | N/A for this feature (internal implementation) |
| Specification Checkpoints | ✅ | Spec complete, plan reviewed |

## Project Structure

### Documentation (this feature)

```text
specs/002-cross-domain-implementation/
├── plan.md              # This file
├── spec.md              # Feature specification
├── research.md          # Technical research
├── data-model.md        # Data model for auth entities
├── quickstart.md        # Quick start guide
├── contracts/           # API contracts
│   └── auth-api.yaml    # Auth endpoints OpenAPI spec
└── tasks.md             # Implementation tasks (created by /speckit.tasks)
```

### Source Code (repository root)

```text
backend/
├── internal/
│   ├── auth/
│   │   ├── token_manager.go          # Existing - extend for refresh tokens
│   │   ├── code_store.go             # NEW - handoff code storage
│   │   └── middleware.go             # Existing - update for refresh flow
│   ├── handlers/
│   │   ├── auth_handler.go           # NEW - dedicated auth handler
│   │   └── user_handler.go           # Existing - login/register
│   ├── middleware/
│   │   ├── cors.go                   # NEW - CORS configuration
│   │   └── rate_limit.go             # NEW - rate limiting
│   └── services/
│       └── auth_service.go           # NEW - auth business logic
└── cmd/server/main.go                # Update for new middleware

frontend/
├── src/
│   ├── lib/
│   │   ├── auth.ts                   # NEW - auth manager (in-memory tokens)
│   │   ├── api.ts                    # MODIFY - remove localStorage, add refresh
│   │   └── mobile-handoff.ts         # NEW - handoff redirect logic
│   └── hooks/
│       └── useAuth.ts                # NEW - auth hook with refresh

mobile/
├── lib/
│   └── api/
│       ├── auth.ts                   # NEW - SecureStore token management
│       └── api.ts                    # MODIFY - use SecureStore, add refresh
└── app/
    └── _layout.tsx                   # MODIFY - deep link handling for auth
```

**Structure Decision**: Extends existing web + mobile + backend structure. New files concentrated in auth modules across all three platforms. Minimal changes to existing code.

## Complexity Tracking

No constitution violations. Architecture follows existing patterns with focused additions for cross-domain auth.

## Implementation Phases

### Phase 0: Research (Complete)
See [research.md](./research.md) for technical decisions.

### Phase 1: Design & Contracts (Complete)
- [data-model.md](./data-model.md) - Auth entity definitions
- [contracts/auth-api.yaml](./contracts/auth-api.yaml) - OpenAPI specification
- [quickstart.md](./quickstart.md) - Implementation guide

### Phase 2: Backend Auth Endpoints
Priority order:
1. CORS middleware configuration
2. Token refresh endpoint (`POST /auth/refresh`)
3. Handoff code store + endpoints (`POST /auth/mobile-handoff`, `POST /auth/exchange`)
4. Health check (`GET /health`)
5. Logout endpoint (`POST /auth/logout`)
6. Rate limiting middleware

### Phase 3: Frontend Token Management
Priority order:
1. In-memory auth manager (replace localStorage)
2. Automatic token refresh on 401
3. Mobile handoff redirect function
4. Update API client for credentials

### Phase 4: Mobile Auth Integration
Priority order:
1. SecureStore token management
2. Deep link handling for auth codes
3. Code exchange flow
4. Token refresh logic

### Phase 5: Testing & Verification
- CORS validation
- Cross-platform auth flow testing
- Rate limit testing
- Security audit (XSS, token storage)

## Dependencies

```
Phase 0 (Research) ─────┐
                        ▼
Phase 1 (Design) ───────┐
                        ▼
Phase 2 (Backend) ──────┬──────────────────┐
                        │                  │
                        ▼                  ▼
Phase 3 (Frontend)    Phase 4 (Mobile)
                        │                  │
                        └────────┬─────────┘
                                 ▼
                        Phase 5 (Testing)
```

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| CORS misconfiguration | Medium | High | Test with actual domains before deploy |
| Token storage security | Low | Critical | Follow established patterns, security review |
| Handoff code timing attacks | Low | Medium | Cryptographically secure random, constant-time compare |
| Mobile deep link failures | Medium | Medium | Fallback to app store, error handling |

## Success Metrics

Aligned with spec Success Criteria:
- SC-001: Handoff < 5 seconds
- SC-002: 100% automatic refresh
- SC-003: Zero tokens in localStorage
- SC-007: 100% CORS block rate for unauthorized origins
