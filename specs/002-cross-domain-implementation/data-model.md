# Data Model: Cross-Domain Authentication

**Feature**: 002-cross-domain-implementation
**Date**: 2026-02-02

## Overview

This document defines the data model for cross-domain authentication entities. Most entities are transient (JWT tokens, handoff codes) and not persisted to the database. The existing `users` table is leveraged for user authentication.

---

## Entities

### 1. AccessToken (Transient - JWT)

Short-lived token for API authentication. Not stored in database.

| Field | Type | Description | Constraints |
|-------|------|-------------|-------------|
| user_id | UUID | Reference to user | Required |
| email | string | User's email | Required, from user record |
| user_type | string | Type of user | "user" or "guest" |
| iat | timestamp | Issued at time | Auto-set on creation |
| exp | timestamp | Expiration time | 15 minutes from iat |
| iss | string | Token issuer | "wish-list-app" |

**JWT Structure**:
```json
{
  "user_id": "uuid",
  "email": "user@example.com",
  "user_type": "user",
  "iat": 1706889600,
  "exp": 1706890500,
  "iss": "wish-list-app"
}
```

**State Transitions**:
```
[Created] → (15 min) → [Expired]
                         ↓
                    [Refresh] → [New Token Created]
```

---

### 2. RefreshToken (Transient - JWT)

Long-lived token for obtaining new access tokens. Stored in httpOnly cookie (Frontend) or SecureStore (Mobile).

| Field | Type | Description | Constraints |
|-------|------|-------------|-------------|
| user_id | UUID | Reference to user | Required |
| email | string | User's email | Required |
| user_type | string | Type of user | "user" or "guest" |
| token_id | UUID | Unique token identifier | For future blacklisting |
| iat | timestamp | Issued at time | Auto-set on creation |
| exp | timestamp | Expiration time | 7 days from iat |
| iss | string | Token issuer | "wish-list-app" |

**JWT Structure**:
```json
{
  "user_id": "uuid",
  "email": "user@example.com",
  "user_type": "user",
  "token_id": "uuid",
  "iat": 1706889600,
  "exp": 1707494400,
  "iss": "wish-list-app"
}
```

**State Transitions**:
```
[Created] → (used for refresh) → [Rotated - New Token Issued, Old Invalidated]
     ↓
(7 days) → [Expired]
     ↓
(logout) → [Invalidated]
```

---

### 3. HandoffCode (In-Memory)

One-time code for Frontend → Mobile authentication transfer.

| Field | Type | Description | Constraints |
|-------|------|-------------|-------------|
| code | string | Random secure code | 32 bytes, base64url encoded |
| user_id | UUID | Reference to user | Required |
| expires_at | timestamp | Expiration time | 60 seconds from creation |
| created_at | timestamp | Creation time | Auto-set |

**In-Memory Structure**:
```go
type CodeStore struct {
    mu    sync.RWMutex
    codes map[string]codeEntry // key is the code string
}

type codeEntry struct {
    UserID    uuid.UUID
    ExpiresAt time.Time
}
```

**State Transitions**:
```
[Created] → (60 sec) → [Expired & Cleaned]
     ↓
[Exchanged] → [Deleted Immediately]
```

**Validation Rules**:
- Code must be cryptographically random (crypto/rand)
- Code must be deleted after single use
- Code comparison must be constant-time to prevent timing attacks
- Background cleanup job removes expired codes every 30 seconds

---

### 4. GuestToken (Transient - JWT)

Token for unauthenticated users to manage their reservations.

| Field | Type | Description | Constraints |
|-------|------|-------------|-------------|
| guest_id | UUID | Generated guest identifier | Auto-generated |
| email | string | Guest's email | Required, from reservation |
| user_type | string | Type of user | Always "guest" |
| reservation_id | UUID | Associated reservation | Optional, for scoped tokens |
| iat | timestamp | Issued at time | Auto-set |
| exp | timestamp | Expiration time | 24 hours from iat |
| iss | string | Token issuer | "wish-list-app" |

**JWT Structure**:
```json
{
  "user_id": "guest-uuid",
  "email": "guest@example.com",
  "user_type": "guest",
  "iat": 1706889600,
  "exp": 1706976000,
  "iss": "wish-list-app"
}
```

---

## Existing Entities (Reference)

### User (Database - Already Exists)

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Primary key |
| email | string | Unique email address |
| password_hash | string | Bcrypt hashed password |
| first_name | string | User's first name |
| last_name | string | User's last name |
| avatar_url | string | Profile image URL |
| created_at | timestamp | Account creation time |
| updated_at | timestamp | Last update time |

---

## Relationships

```
┌─────────────────────────────────────────────────────────────┐
│                     USER (Database)                          │
│  id, email, password_hash, first_name, last_name            │
└─────────────────────────┬───────────────────────────────────┘
                          │
          ┌───────────────┼───────────────┬──────────────┐
          │               │               │              │
          ▼               ▼               ▼              ▼
    ┌──────────┐   ┌──────────┐   ┌────────────┐  ┌──────────┐
    │ Access   │   │ Refresh  │   │ Handoff    │  │ Guest    │
    │ Token    │   │ Token    │   │ Code       │  │ Token    │
    │ (JWT)    │   │ (JWT)    │   │ (Memory)   │  │ (JWT)    │
    └──────────┘   └──────────┘   └────────────┘  └──────────┘

    Lifetime:       Lifetime:      Lifetime:       Lifetime:
    15 minutes      7 days         60 seconds      24 hours

    Storage:        Storage:       Storage:        Storage:
    Memory (FE)     Cookie (FE)    In-Memory (BE)  URL param
    SecureStore     SecureStore
    (Mobile)        (Mobile)
```

---

## Storage Strategy

| Entity | Storage Location | Encryption | Cleanup |
|--------|------------------|------------|---------|
| User | PostgreSQL | Password bcrypt hashed | Manual deletion |
| AccessToken | Memory (FE), SecureStore (Mobile) | HTTPS transit | Automatic expiry |
| RefreshToken | httpOnly Cookie (FE), SecureStore (Mobile) | HTTPS transit, platform encryption | Rotation, logout |
| HandoffCode | Backend memory | N/A (short-lived) | 30s background job |
| GuestToken | URL parameter | HTTPS transit | Automatic expiry |

---

## Validation Rules Summary

### Token Creation
- Access tokens: Only after successful authentication
- Refresh tokens: Only after successful authentication, rotated on refresh
- Handoff codes: Only for authenticated users via `POST /auth/mobile-handoff`
- Guest tokens: Only when creating guest reservations

### Token Validation
- Check signature with HMAC-SHA256
- Verify issuer is "wish-list-app"
- Check expiration time
- Verify user_id exists in database (for critical operations)

### Security Constraints
- No tokens in localStorage/sessionStorage (Frontend)
- All tokens via HTTPS only
- Handoff codes one-time use only
- Refresh token rotation on every use
