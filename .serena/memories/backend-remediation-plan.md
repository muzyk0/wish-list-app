# Backend Remediation Plan

## Task Checklist

### Phase 1: Critical Security (Must Complete First)

#### Task 1.1: Fix SQL Injection in GiftItemRepository
- [x] **File:** `internal/domain/item/repository/giftitem_repository.go`
- [x] **Lines:** 176-200
- [x] **Action:**
  1. Create strict whitelist map for sort fields
  2. Create strict whitelist map for order directions
  3. Validate inputs against whitelists
  4. Add error handling for invalid sort/order parameters
  5. Write tests for SQL injection attempts
- [x] **Test:** Verify malicious input doesn't reach database

#### Task 1.2: Remove Hardcoded JWT Secret
- [x] **File:** `internal/app/config/config.go`
- [x] **Lines:** 48-50
- [x] **Action:**
  1. Import `crypto/rand` and `encoding/base64`
  2. Generate random 32-byte secret for development
  3. Add warning log when using generated secret
  4. Keep production requirement strict
- [x] **Test:** Verify different secret on each dev restart

#### Task 1.3: Fix Information Disclosure in Auth Middleware
- [x] **File:** `internal/pkg/auth/middleware.go`
- [x] **Lines:** 29-31
- [x] **Action:**
  1. Remove `err.Error()` from HTTP response
  2. Log full error internally with c.Logger()
  3. Return generic "Invalid or expired token" message
- [x] **Test:** Verify no internal error details in response

#### Task 1.4: Add Rate Limiting to OAuth Endpoints
- [x] **Files:**
  - `internal/app/middleware/rate_limit.go`
  - `internal/domain/auth/delivery/http/routes.go`
- [x] **Action:**
  1. Add `NewOAuthRateLimiter()` function
  2. Configure appropriate limits (5 req/min)
  3. Apply middleware to OAuth routes
  4. Add rate limit headers
- [x] **Test:** Verify 429 returned after limit exceeded

---

### Phase 2: High Priority Issues

#### Task 2.1: Fix Goroutine Leaks (Rate Limiter)
- [x] **File:** `internal/app/middleware/rate_limit.go`
- [x] **Lines:** 49-60
- [x] **Action:**
  1. Add `context.Context` parameter to constructor
  2. Pass context to cleanupLoop
  3. Handle ctx.Done() in cleanup loop
  4. Update all instantiations
- [x] **Test:** Verify goroutines exit on shutdown

#### Task 2.2: Fix Goroutine Leaks (CodeStore)
- [x] **File:** `internal/pkg/auth/code_store.go`
- [x] **Lines:** 113-132
- [x] **Action:**
  1. Update `StartCleanupRoutine` to accept context
  2. Use select with ctx.Done()
  3. Update app.go initialization
- [x] **Test:** Verify no goroutine leaks

#### Task 2.3: Add OAuth Input Validation
- [x] **File:** `internal/domain/auth/delivery/http/oauth_handler.go`
- [x] **Lines:** 336-371
- [x] **Action:**
  1. Import `net/mail` and `net/url`
  2. Validate email format with mail.ParseAddress
  3. Sanitize names (trim, length check)
  4. Validate avatar URL format
  5. Return appropriate errors
- [x] **Test:** Verify invalid inputs rejected

#### Task 2.4: Optimize CodeStore Lookup
- [ ] **File:** `internal/pkg/auth/code_store.go`
- [ ] **Lines:** 58-90
- [ ] **Action:**
  1. Use direct map lookup instead of iteration
  2. Keep constant-time comparison for security
  3. Maintain thread safety
- [ ] **Test:** Benchmark performance improvement

#### Task 2.5: Fix HTTP Status Codes
- [ ] **File:** `internal/domain/auth/delivery/http/handler.go`
- [ ] **Lines:** 137-141, others
- [ ] **Action:**
  1. Review all error responses
  2. Change client errors from 500 to 400
  3. Document status code conventions
- [ ] **Test:** Verify correct status codes

---

### Phase 3: Code Duplication & Architecture

#### Task 3.1: Create PII Encryption Helper Package
- [ ] **New File:** `internal/pkg/pii/encryption.go`
- [ ] **Action:**
  1. Create `PIIEncryptor` struct
  2. Extract encrypt/decrypt logic from repositories
  3. Add `EncryptField`, `DecryptField` methods
  4. Refactor UserRepository to use helper
  5. Refactor ReservationRepository to use helper
- [ ] **Test:** Verify encryption/decryption still works

#### Task 3.2: Create Centralized Error Handler
- [ ] **New File:** `internal/pkg/errors/handler.go`
- [ ] **Action:**
  1. Define service error types with status codes
  2. Create Echo error handler middleware
  3. Refactor all handlers to use new pattern
  4. Remove duplicate error handling code
- [ ] **Test:** All existing error scenarios still work

#### Task 3.3: Fix Pagination (Move to DB Level)
- [ ] **File:** `internal/domain/wishlist/delivery/http/handler.go`
- [ ] **Lines:** 245-294
- [ ] **Action:**
  1. Add pagination parameters to service interface
  2. Add pagination to repository query
  3. Update handler to use new signature
  4. Remove in-memory pagination
- [ ] **Test:** Verify pagination works with large datasets

#### Task 3.4: Extract Configuration Constants
- [ ] **New File:** `internal/app/config/constants.go`
- [ ] **Action:**
  1. Define all magic numbers as constants
  2. Add comments explaining each value
  3. Replace all hardcoded values
  4. Make DB pool settings configurable
- [ ] **Test:** No hardcoded values remain

#### Task 3.5: Remove Deprecated Create Method
- [ ] **File:** `internal/domain/item/repository/giftitem_repository.go`
- [ ] **Lines:** 99-102
- [ ] **Action:**
  1. Delete `Create` method
  2. Search for all usages
  3. Update all calls to use `CreateWithOwner`
  4. Update interface definition
  5. Regenerate mocks
- [ ] **Test:** All tests pass

#### Task 3.6: Split GiftItemRepository
- [ ] **Files:** 
  - `internal/domain/item/repository/giftitem_repository.go` (refactor)
  - `internal/domain/item/repository/reservation_repository.go` (new)
- [ ] **Action:**
  1. Move reservation methods to new repository
  2. Move purchase methods to new repository
  3. Keep only CRUD in original
  4. Update service layer
  5. Update wire injection
- [ ] **Test:** All functionality preserved

---

### Phase 4: Code Quality & Style

#### Task 4.1: Standardize Naming Conventions
- [ ] **Files:** All Go files
- [ ] **Action:**
  1. Choose convention (PascalCase for exported)
  2. Rename inconsistent identifiers
  3. Update JSON tags if needed
  4. Update tests
- [ ] **Test:** No naming inconsistencies

#### Task 4.2: Clean Up Import Aliases
- [ ] **Files:** Multiple
- [ ] **Action:**
  1. Remove unnecessary aliases
  2. Use direct imports where no conflict
  3. Standardize on `net/http` without alias
- [ ] **Test:** All imports compile

#### Task 4.3: Add Request ID Middleware
- [ ] **New File:** `internal/app/middleware/request_id.go`
- [ ] **Action:**
  1. Create middleware to generate request_id
  2. Add to response headers
  3. Inject into context
  4. Update logger to include request_id
- [ ] **Test:** Each request has unique ID

#### Task 4.4: Replace Panic with Log.Fatal
- [x] **File:** `internal/app/config/config.go`
- [x] **Line:** 44
- [x] **Action:**
  1. Replace `panic` with `log.Fatal`
  2. Ensure proper cleanup before exit
- [x] **Test:** Clean shutdown on missing config

#### Task 4.5: Fix Deprecated Comment Format
- [ ] **Files:** Any remaining deprecated comments
- [ ] **Action:**
  1. Use Go-standard format: `// Deprecated: ...`
  2. Ensure tooling picks up deprecations
- [ ] **Test:** `go vet` recognizes deprecations

#### Task 4.6: Add Context Cancellation Checks
- [ ] **Files:** Repository files
- [ ] **Action:**
  1. Add ctx.Err() checks in long operations
  2. Add early return on cancellation
- [ ] **Test:** Operations cancel properly

#### Task 4.7: Standardize Error Wrapping
- [ ] **Files:** All service and repository files
- [ ] **Action:**
  1. Always use `fmt.Errorf("...: %w", err)`
  2. Add context to all errors
  3. Use sentinel errors for specific cases
- [ ] **Test:** All errors properly wrapped

#### Task 4.8: Remove Unused Sentinel Errors
- [ ] **Files:** Repository files
- [ ] **Action:**
  1. Search for all defined sentinel errors
  2. Verify each is actually used
  3. Remove unused ones
- [ ] **Test:** No dead code

#### Task 4.9: Make HTTP Timeouts Configurable
- [ ] **File:** `internal/domain/auth/delivery/http/oauth_handler.go`
- [ ] **Action:**
  1. Add OAuth timeout to config
  2. Use config value in HTTP client
  3. Set reasonable default
- [ ] **Test:** Timeout configurable

#### Task 4.10: Add Missing GoDoc Comments
- [ ] **Files:** All exported functions
- [ ] **Action:**
  1. Document all exported types
  2. Document all exported functions
  3. Follow GoDoc conventions
- [ ] **Test:** `go doc` shows proper documentation

---

## Implementation Order

### Week 1: Security Critical
```
1. Fix SQL Injection
2. Remove Hardcoded JWT Secret
3. Fix Info Disclosure
4. Add OAuth Rate Limiting
```

### Week 2: High Priority
```
5. Fix Goroutine Leaks (Rate Limiter)
6. Fix Goroutine Leaks (CodeStore)
7. Add OAuth Validation
8. Optimize CodeStore
9. Fix HTTP Status Codes
```

### Week 3: Architecture (Part 1)
```
10. Create PII Helper Package
11. Create Error Handler
12. Fix Pagination
```

### Week 4: Architecture (Part 2)
```
13. Extract Constants
14. Remove Deprecated Code
15. Split Repository
```

### Week 5: Code Quality
```
16. Standardize Naming
17. Clean Imports
18. Add Request ID
19. Replace Panic
20. Fix Comments
21. Add Context Checks
22. Standardize Errors
23. Remove Unused Errors
24. Config Timeouts
25. Add GoDoc
```

### Week 6: Testing & Validation
```
26. Security Tests
27. Load Tests
28. Integration Tests
29. Code Review
30. Documentation Review
```

---

## File Changes Summary

### New Files to Create:
1. `internal/pkg/pii/encryption.go`
2. `internal/pkg/errors/handler.go`
3. `internal/app/middleware/request_id.go`
4. `internal/app/config/constants.go`
5. `internal/domain/item/repository/reservation_repository.go`
6. `internal/domain/item/repository/purchase_repository.go`

### Files to Modify:
1. `internal/domain/item/repository/giftitem_repository.go` (SQL injection, deprecated removal, split)
2. `internal/app/config/config.go` (JWT secret, constants)
3. `internal/pkg/auth/middleware.go` (error disclosure)
4. `internal/app/middleware/rate_limit.go` (goroutine leaks)
5. `internal/pkg/auth/code_store.go` (goroutine leaks, optimization)
6. `internal/domain/auth/delivery/http/handler.go` (status codes)
7. `internal/domain/auth/delivery/http/oauth_handler.go` (validation, timeouts)
8. `internal/domain/auth/delivery/http/routes.go` (rate limiting)
9. `internal/domain/wishlist/delivery/http/handler.go` (pagination)
10. `internal/domain/user/repository/user_repository.go` (use PII helper)
11. `internal/domain/reservation/repository/reservation_repository.go` (use PII helper)
12. All handler files (error handling)
13. All service files (error wrapping)

### Files to Delete:
1. Deprecated methods in repositories (after migration)

---

## Testing Requirements

### Unit Tests Required:
- [x] SQL injection prevention
- [ ] Rate limiting behavior
- [x] OAuth validation
- [ ] CodeStore performance
- [ ] PII encryption/decryption
- [x] Error handling middleware
- [ ] Pagination correctness
- [x] Context cancellation

### Integration Tests Required:
- [ ] OAuth flow end-to-end
- [ ] Authentication flow
- [ ] Wishlist CRUD operations
- [ ] Reservation flow
- [ ] Graceful shutdown

### Load Tests Required:
- [ ] Rate limiter under load
- [ ] Database connection pool
- [ ] Concurrent code exchanges
- [ ] Pagination with large datasets

---

## Migration Guide

### For Deprecated Create Method:
```bash
# Search for usages
grep -r "\.Create(" --include="*.go" | grep -v "CreateWith"

# Replace:
repository.Create(ctx, item)
# With:
repository.CreateWithOwner(ctx, item)
```

### For Configuration Changes:
```bash
# Add to .env.example:
OAUTH_HTTP_TIMEOUT=10s
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=5m
```

### For PII Encryption Migration:
No migration needed - refactoring only, behavior unchanged.

---

## Success Criteria

- [x] All Critical and High issues resolved
- [x] Security tests passing
- [ ] Load tests showing improvement
- [ ] No deprecated code remaining
- [x] Code coverage maintained or improved
- [ ] Documentation updated
- [x] No breaking changes (or documented)
- [ ] Performance benchmarks improved

---

## Notes

- Each task should be committed separately with clear messages
- Run tests after each task
- Update this checklist as tasks are completed
- Document any deviations from plan

---

## Completion Log

**Date: 2026-02-15**

### Phase 2 Completion - High Priority Issues ✅
- Task 2.4: Optimized CodeStore ExchangeCode from O(n) to O(1) map lookup while maintaining constant-time comparison for security
- Task 2.5: Fixed HTTP status code in auth handler (500 -> 400 for invalid UUID format)

### Phase 3 Progress - Architecture Improvements ✅
- Task 3.1: Created `internal/pkg/pii` package with FieldEncryptor for centralized PII encryption/decryption
- Task 3.3: Moved pagination from in-memory to database level for public wishlist gift items
- Task 3.4: Extracted configuration constants to `internal/app/config/constants.go`
- Task 3.5: Removed deprecated `Create` method from GiftItemRepository, updated all usages to `CreateWithOwner`

### Phase 4 Progress - Code Quality ✅
- Task 4.9: Made OAuth HTTP timeout configurable via `OAUTH_HTTP_TIMEOUT` environment variable

---

**Date: 2026-02-14**

### Phase 1: Critical Security - COMPLETED ✅
All 4 tasks completed and committed:
- `cb5ea45` - security(repository): fix SQL injection in GetByOwnerPaginated
- `2908b2b` - security(config): generate random JWT secret for development
- `43b746e` - security(auth): remove error details from JWT validation responses
- `0aade78` - security(auth): add rate limiting to OAuth endpoints

### Phase 2: High Priority - COMPLETED (5/5) ✅
Completed tasks:
- Task 2.1: Goroutine leaks in Rate Limiter - DONE
- Task 2.2: Goroutine leaks in CodeStore - DONE
- Task 2.3: OAuth Input Validation - DONE
- Task 2.4: Optimize CodeStore Lookup (O(1) map lookup) - DONE
- Task 2.5: Fix HTTP Status Codes (500 -> 400 for client errors) - DONE

### Phase 3: Code Duplication & Architecture - PARTIALLY COMPLETED (3/6) ✅
Completed tasks:
- Task 3.1: Create PII Encryption Helper Package - DONE
- Task 3.4: Extract Configuration Constants - DONE
- Task 3.5: Remove Deprecated Create Method - DONE

Pending tasks:
- Task 3.2: Create Centralized Error Handler
- Task 3.3: Fix Pagination (Move to DB Level)
- Task 3.6: Split GiftItemRepository

### Phase 4: Code Quality - PARTIALLY COMPLETED (2/10) ✅
Completed tasks:
- Task 4.4: Replace Panic with Log.Fatal - DONE (part of Phase 1.2)
- Task 4.9: Make HTTP Timeouts Configurable - DONE
