# Domain-Driven Structure Refactoring Plan

**Date**: 2026-02-09
**Status**: Planning
**Estimated Time**: 8-10 hours
**Impact**: Major architectural refactoring

---

## Executive Summary

This plan transforms the backend from a flat layer-based structure to a domain-driven architecture with hybrid folders. The refactoring addresses:

- **Functionality conflicts** - Clear domain boundaries
- **Codebase growth** - Scalable organization (currently 45 domain files)
- **DTO duplication** - Unified DTOs per domain
- **Developer productivity** - Faster onboarding, easier code discovery

**Key Decision**: Hybrid domain structure with layered folders (handlers, services, repositories, dtos) within each domain.

---

## 📊 Current State Analysis

### Metrics
- **Handlers**: 16 files
- **Services**: 19 files
- **Repositories**: 10 files
- **Test files**: 28 files
- **Total domain files**: 45 files

### Identified Issues

#### 1. DTO Duplication
```go
// wishlist_handler.go
type CreateGiftItemRequest struct {
    Name        string  `json:"name"`
    Description string  `json:"description"`
    Link        string  `json:"link"`
    Price       float64 `json:"price"`
    // ... 8 fields
}

// item_handler.go
type CreateItemRequest struct {
    Title       string  `json:"title"`  // Different field name!
    Description string  `json:"description"`
    Link        string  `json:"link"`
    Price       float64 `json:"price"`
    // ... 7 fields (nearly identical)
}
```

**Impact**: Maintenance overhead, API inconsistency, confusion

#### 2. Functionality Conflicts
- Item operations split between `wishlist_handler.go` and `item_handler.go`
- Auth logic spread across `auth/`, `handlers/auth_handler.go`, `handlers/oauth_handler.go`
- Unclear boundaries lead to "which file do I modify?" confusion

#### 3. Codebase Growth
- Flat structure doesn't scale beyond 50-100 files
- 45 files already causing navigation friction
- Adding new features requires touching multiple distant directories

---

## 🎯 Target Architecture

### Proposed Structure

```
backend/internal/
├── domains/                      # Business domains
│   ├── auth/                     # Authentication & Identity
│   │   ├── handlers/
│   │   │   ├── auth_handler.go
│   │   │   ├── oauth_handler.go
│   │   │   └── user_handler.go
│   │   ├── services/
│   │   │   ├── user_service.go
│   │   │   └── account_cleanup_service.go
│   │   ├── repositories/
│   │   │   └── user_repository.go
│   │   ├── dtos/
│   │   │   ├── requests.go
│   │   │   └── responses.go
│   │   ├── middleware/
│   │   │   └── jwt_middleware.go
│   │   └── auth.go               # Domain exports
│   │
│   ├── wishlists/                # Wishlist Management
│   │   ├── handlers/
│   │   │   ├── wishlist_handler.go
│   │   │   └── wishlist_item_handler.go
│   │   ├── services/
│   │   │   ├── wishlist_service.go
│   │   │   └── wishlist_item_service.go
│   │   ├── repositories/
│   │   │   ├── wishlist_repository.go
│   │   │   ├── wishlistitem_repository.go
│   │   │   └── template_repository.go
│   │   ├── dtos/
│   │   │   ├── requests.go
│   │   │   └── responses.go
│   │   └── wishlists.go
│   │
│   ├── items/                    # Gift Items (Independent Resources)
│   │   ├── handlers/
│   │   │   └── item_handler.go
│   │   ├── services/
│   │   │   └── item_service.go
│   │   ├── repositories/
│   │   │   └── giftitem_repository.go
│   │   ├── dtos/
│   │   │   ├── requests.go      # Unified DTOs!
│   │   │   └── responses.go
│   │   └── items.go
│   │
│   ├── reservations/             # Reservations & Purchases
│   │   ├── handlers/
│   │   │   └── reservation_handler.go
│   │   ├── services/
│   │   │   ├── reservation_service.go
│   │   │   └── email_service.go
│   │   ├── repositories/
│   │   │   └── reservation_repository.go
│   │   ├── dtos/
│   │   │   ├── requests.go
│   │   │   └── responses.go
│   │   └── reservations.go
│   │
│   ├── storage/                  # File Storage (S3)
│   │   ├── handlers/
│   │   │   └── s3_handler.go
│   │   ├── services/
│   │   │   └── s3_service.go
│   │   └── storage.go
│   │
│   └── health/                   # Health Checks
│       ├── handlers/
│       │   └── health_handler.go
│       └── health.go
│
└── shared/                       # Cross-cutting concerns
    ├── middleware/
    │   ├── cors.go
    │   ├── rate_limit.go
    │   └── middleware.go
    ├── config/
    ├── db/
    ├── cache/
    ├── encryption/
    ├── validation/
    ├── analytics/
    └── aws/                      # S3 client
```

---

## 🏗️ Domain Identification

### Domain Boundaries

| Domain | Responsibility | Files | Rationale |
|--------|---------------|-------|-----------|
| **auth** | Authentication, OAuth, User management | 11 files | Security boundary, complex auth flows, potential microservice |
| **wishlists** | Wishlist CRUD, templates, wishlist-item relationships | 10 files | Core domain, high cohesion, many-to-many with items |
| **items** | Gift items as independent resources | 5 files | Independent lifecycle, can exist without wishlists |
| **reservations** | Item reservations, purchase tracking, notifications | 5 files | Complex business logic, separate from items/wishlists |
| **storage** | S3 file uploads, media management | 3 files | Infrastructure concern with domain-specific usage |
| **health** | System monitoring, health checks | 2 files | Operational concern, simplest domain |

### Domain Relationships

```
┌─────────────┐
│    auth     │───┐
└─────────────┘   │
                  │ user_id
                  ↓
┌─────────────┐   ┌─────────────┐
│  wishlists  │←──│    items    │
└─────────────┘   └─────────────┘
       │                 │
       │ wishlist_id     │ item_id
       │                 │
       └────→┌─────────────┐←────┘
             │reservations │
             └─────────────┘
```

**Key Insights**:
- Items are independent (user owns items, not wishlists)
- Many-to-many via `wishlist_items` junction table
- Reservations reference both wishlists and items directly
- Auth domain provides user context to all domains

---

## 🎁 Benefits Analysis

### 1. Faster Onboarding
- **Before**: "Where's wishlist code?" → Search 3 directories
- **After**: `domains/wishlists/` → Everything in one place
- **Impact**: 70% reduction in "file hopping" for new developers

### 2. Easier Code Discovery
- **Before**: Navigate `handlers/` → `services/` → `repositories/`
- **After**: Entire vertical slice in one domain folder
- **Impact**: Improved developer productivity, reduced context switching

### 3. Better Microservices Separation
- **Before**: Extract wishlists → hunt scattered dependencies
- **After**: `domains/wishlists/` is already a clean boundary
- **Impact**: Future microservices extraction becomes trivial

### 4. Clearer Ownership
- **Before**: "Who owns user management?" → Mixed with auth/OAuth
- **After**: Clear team ownership per domain folder
- **Impact**: Better accountability, faster code reviews

### 5. Resolves DTO Duplication
- **Before**: Multiple `CreateItemRequest` variants
- **After**: Single source of truth in `domains/items/dtos/`
- **Impact**: Consistent API, reduced maintenance

### 6. Resolves Functionality Conflicts
- **Before**: Item operations in 2 handlers
- **After**: Clear `domains/items/` vs `domains/wishlists/` separation
- **Impact**: No more "which file?" confusion

### 7. Manages Codebase Growth
- **Before**: Flat list → unmanageable at 100+ files
- **After**: Bounded domains → predictable scaling
- **Impact**: Sustainable growth path

---

## 🚀 Migration Plan

### Overview
- **Total Time**: 8-10 hours
- **Approach**: Incremental, one domain at a time
- **Risk**: Low (reversible at each checkpoint)
- **Rollback Strategy**: Git checkpoints after each phase

### Phase 0: Preparation (1-2 hours)

**Objective**: Create safety checkpoints and verify baseline

```bash
# 1. Create feature branch
git checkout -b refactor/domain-driven-structure

# 2. Ensure all tests pass
go test ./...
# Expected: PASS

# 3. Run linter
golangci-lint run
# Expected: No issues

# 4. Check coverage baseline
go test -cover ./... > coverage-baseline.txt

# 5. Commit current state
git add -A
git commit -m "chore: checkpoint before domain refactoring"

# 6. Document current import structure
grep -r "wish-list/internal" backend/cmd backend/internal > imports-before.txt
```

**Validation**:
- ✅ All tests pass
- ✅ Linter clean
- ✅ Coverage baseline recorded
- ✅ Git checkpoint created

---

### Phase 1: Create Domain Structure (30 mins)

**Objective**: Create new folder hierarchy without moving files

```bash
# Create domain folders
mkdir -p backend/internal/domains/{auth,wishlists,items,reservations,storage,health}

# Create shared folder
mkdir -p backend/internal/shared/{middleware,config,db,cache,encryption,validation,analytics,aws}

# Create layer folders within each domain
for domain in auth wishlists items reservations storage health; do
    mkdir -p backend/internal/domains/$domain/{handlers,services,repositories,dtos}
done

# Verify structure
tree backend/internal/domains -L 2
tree backend/internal/shared -L 1
```

**Validation**:
- ✅ All folders created
- ✅ Structure matches target architecture
- ✅ No files moved yet (low risk)

---

### Phase 2: Move Cross-Cutting Concerns (1 hour)

**Objective**: Move infrastructure packages to `shared/`

```bash
# Move cross-cutting packages
mv backend/internal/middleware backend/internal/shared/
mv backend/internal/config backend/internal/shared/
mv backend/internal/db backend/internal/shared/
mv backend/internal/cache backend/internal/shared/
mv backend/internal/encryption backend/internal/shared/
mv backend/internal/validation backend/internal/shared/
mv backend/internal/analytics backend/internal/shared/
mv backend/internal/aws backend/internal/shared/

# Update imports in moved files
find backend/internal/shared -name "*.go" -type f -exec sed -i '' \
    's|wish-list/internal/middleware|wish-list/internal/shared/middleware|g' {} +
find backend/internal/shared -name "*.go" -type f -exec sed -i '' \
    's|wish-list/internal/config|wish-list/internal/shared/config|g' {} +
# ... repeat for each package

# Update imports in remaining files
find backend/internal/handlers backend/internal/services backend/internal/repositories backend/cmd -name "*.go" -type f -exec sed -i '' \
    's|wish-list/internal/middleware|wish-list/internal/shared/middleware|g' {} +
# ... repeat for all shared packages

# Run tests
go test ./internal/shared/...
```

**Validation**:
- ✅ Shared packages moved successfully
- ✅ All imports updated
- ✅ Tests pass for shared packages
- ✅ `go build ./cmd/server` succeeds

**Commit**:
```bash
git add -A
git commit -m "refactor: move cross-cutting concerns to shared/"
```

---

### Phase 3: Migrate Domains One-by-One (3-4 hours)

**Strategy**: Start with simplest domain (health) to learn process, then tackle complex ones

#### Domain Migration Order

1. **health** (2 files) - Learning/validation
2. **storage** (3 files) - Simple, minimal dependencies
3. **reservations** (5 files) - Moderate complexity
4. **items** (5 files) - DTO consolidation practice
5. **wishlists** (10 files) - Large but straightforward
6. **auth** (11 files) - Most complex, save for last

#### Per-Domain Migration Steps

**Example: Health Domain**

```bash
# Step 1: Move handler files
mv backend/internal/handlers/health_handler.go \
   backend/internal/domains/health/handlers/
mv backend/internal/handlers/health_handler_test.go \
   backend/internal/domains/health/handlers/

# Step 2: Update package in moved files (no change needed - still "package handlers")

# Step 3: Update imports in moved files
# In health_handler.go:
# OLD: "wish-list/internal/config"
# NEW: "wish-list/internal/shared/config"

# Step 4: Create domain export file
cat > backend/internal/domains/health/health.go << 'EOF'
package health

import "wish-list/internal/domains/health/handlers"

// NewHealthHandler creates a new health handler for monitoring endpoints
func NewHealthHandler() *handlers.HealthHandler {
    return handlers.NewHealthHandler()
}
EOF

# Step 5: Update main.go
# In cmd/server/main.go:
# OLD: import "wish-list/internal/handlers"
#      healthHandler := handlers.NewHealthHandler()
# NEW: import healthDomain "wish-list/internal/domains/health"
#      healthHandler := healthDomain.NewHealthHandler()

# Step 6: Run tests
go test ./internal/domains/health/...

# Step 7: Run full test suite
go test ./...

# Step 8: Commit
git add -A
git commit -m "refactor(health): migrate to domain structure"
```

**Repeat for each domain with domain-specific adjustments.**

---

### Phase 4: Consolidate DTOs (2 hours)

**Objective**: Extract DTOs to dedicated `dtos/` folders and eliminate duplication

#### Items Domain (Example)

```bash
# Step 1: Create DTO files
touch backend/internal/domains/items/dtos/requests.go
touch backend/internal/domains/items/dtos/responses.go

# Step 2: Extract DTOs from item_handler.go
# Move CreateItemRequest, UpdateItemRequest to dtos/requests.go
# Move ItemResponse, PaginatedItemsResponse to dtos/responses.go

# Step 3: Merge duplicates
# Compare CreateItemRequest (item_handler) vs CreateGiftItemRequest (wishlist_handler)
# Choose unified field names (prefer item_handler version)
# Create single canonical DTO in items/dtos/requests.go

# Step 4: Update handler imports
# In domains/items/handlers/item_handler.go:
# Add: import "wish-list/internal/domains/items/dtos"
# Replace: type CreateItemRequest → var req dtos.CreateItemRequest

# Step 5: Update cross-domain references
# If wishlists domain uses item DTOs:
# import "wish-list/internal/domains/items/dtos"

# Step 6: Run tests
go test ./internal/domains/items/...

# Step 7: Commit
git add -A
git commit -m "refactor(items): consolidate DTOs to dtos/ package"
```

**Repeat for each domain.**

---

### Phase 5: Update All Imports (1 hour)

**Objective**: Ensure all import paths reference new structure

```bash
# Step 1: Update cmd/server/main.go
# Replace all handler imports with domain imports
# OLD: import "wish-list/internal/handlers"
# NEW: import authDomain "wish-list/internal/domains/auth"
#      import wishlistsDomain "wish-list/internal/domains/wishlists"
#      ...

# Step 2: Update route registration
# OLD: authHandler := handlers.NewAuthHandler(...)
# NEW: authHandler := authDomain.NewAuthHandler(...)

# Step 3: Search for any remaining old imports
grep -r "wish-list/internal/handlers" backend/
grep -r "wish-list/internal/services" backend/
grep -r "wish-list/internal/repositories" backend/
# Should return no results (except in domains/)

# Step 4: Update tests outside domain folders
find backend/cmd -name "*_test.go" -type f -exec sed -i '' \
    's|wish-list/internal/handlers|wish-list/internal/domains|g' {} +

# Step 5: Run all tests
go test ./...

# Step 6: Commit
git add -A
git commit -m "refactor: update all import paths to domain structure"
```

---

### Phase 6: Validation & Cleanup (1 hour)

**Objective**: Verify migration success and clean up

```bash
# Step 1: Run full test suite
go test ./... -v
# Expected: All tests pass

# Step 2: Run linter
golangci-lint run
# Expected: No new issues

# Step 3: Check coverage
go test -cover ./... > coverage-after.txt
diff coverage-baseline.txt coverage-after.txt
# Expected: Similar coverage %

# Step 4: Build application
go build ./cmd/server
# Expected: Success

# Step 5: Run application (smoke test)
./server &
SERVER_PID=$!
sleep 3
curl http://localhost:8080/healthz
# Expected: 200 OK
kill $SERVER_PID

# Step 6: Verify no import cycles
go list -f '{{.ImportPath}}: {{.Imports}}' ./internal/domains/... | grep -i cycle
# Expected: No output

# Step 7: Remove old empty directories
rmdir backend/internal/handlers 2>/dev/null || echo "Directory not empty"
rmdir backend/internal/services 2>/dev/null || echo "Directory not empty"
rmdir backend/internal/repositories 2>/dev/null || echo "Directory not empty"
rmdir backend/internal/auth 2>/dev/null || echo "Directory not empty"
# If not empty, investigate remaining files

# Step 8: Generate updated import report
grep -r "wish-list/internal" backend/cmd backend/internal > imports-after.txt
diff imports-before.txt imports-after.txt

# Step 9: Update Swagger docs (if applicable)
swag init -g cmd/server/main.go -d internal/domains/*/handlers

# Step 10: Final commit
git add -A
git commit -m "refactor: complete domain-driven structure migration

- Migrated 45 domain files to domain folders
- Consolidated DTOs per domain
- Moved cross-cutting concerns to shared/
- All tests passing
- Zero import cycles"
```

**Validation Checklist**:
- ✅ All tests pass (28 tests)
- ✅ Linter clean
- ✅ Coverage unchanged
- ✅ Build succeeds
- ✅ Server runs
- ✅ No import cycles
- ✅ Swagger docs updated
- ✅ Old directories removed

---

## ⚠️ Risk Assessment

### Risk 1: Import Cycles

**Problem**: Circular dependencies between domains
**Likelihood**: Medium
**Impact**: High (build failure)

**Example**:
```
domains/auth imports domains/wishlists
domains/wishlists imports domains/auth
→ Import cycle detected
```

**Mitigation**:
- Keep domains independent - communicate via interfaces
- Shared types go in `shared/models/`
- Use dependency injection at composition root (main.go)
- If domain A needs domain B, consider:
  - Is domain B actually shared infrastructure? → Move to `shared/`
  - Can you invert the dependency? → Use interfaces
  - Should they be one domain? → Merge domains

**Detection**: `go build` fails immediately with "import cycle not allowed"

**Resolution**:
```bash
# Identify cycle
go list -f '{{.ImportPath}}: {{.Imports}}' ./internal/domains/... | grep -C 3 cycle

# Fix by extracting interface to shared/
mkdir -p backend/internal/shared/interfaces
# Move interface definition to shared/interfaces/
# Use interface type in domain A
```

---

### Risk 2: Test Breakage

**Problem**: 28 test files with hardcoded imports
**Likelihood**: High
**Impact**: Medium (fixable with bulk find/replace)

**Mitigation**:
- Update imports one domain at a time
- Run tests after each domain migration
- Use IDE "Find & Replace in Files" for bulk updates
- Keep test files with their corresponding source files

**Rollback Strategy**:
```bash
# If tests fail after domain migration
git log --oneline -10  # Find last good commit
git revert <commit-hash>  # Revert specific domain
# OR
git reset --hard <commit-hash>  # Reset to checkpoint
```

**Bulk Import Update** (use carefully):
```bash
# Update all test files in migrated domain
find backend/internal/domains/items -name "*_test.go" -type f -exec sed -i '' \
    's|wish-list/internal/handlers|wish-list/internal/domains/items/handlers|g' {} +
find backend/internal/domains/items -name "*_test.go" -type f -exec sed -i '' \
    's|wish-list/internal/services|wish-list/internal/domains/items/services|g' {} +
```

---

### Risk 3: CI/CD Pipeline

**Problem**: CI may have path-specific configurations
**Likelihood**: Low
**Impact**: Medium (CI failures, deployment blocked)

**Check These Files**:
```bash
# Search for hardcoded paths
grep -r "internal/handlers" .github/ Makefile docker-compose.yml Dockerfile
grep -r "internal/services" .github/ Makefile docker-compose.yml Dockerfile
```

**Mitigation**:
- Review CI configuration before migration
- Update any path-specific test commands
- Test CI locally with `act` (GitHub Actions locally):
  ```bash
  # Install act: brew install act
  act -l  # List workflows
  act pull_request  # Run PR workflow locally
  ```

**Common CI Updates**:
```yaml
# .github/workflows/test.yml
# OLD
- run: go test ./internal/handlers/...
# NEW
- run: go test ./internal/domains/...
```

---

### Risk 4: Swagger/OpenAPI Generation

**Problem**: Swagger annotations may not find handlers
**Likelihood**: Medium
**Impact**: Low (docs not updated, API still works)

**Check Current Command**:
```bash
# Find current swag command
grep -r "swag init" backend/ Makefile .github/
```

**Update Swagger Generation**:
```bash
# OLD (may be in Makefile)
swag init -g cmd/server/main.go

# NEW (explicit domain handlers)
swag init -g cmd/server/main.go \
    -d internal/domains/auth/handlers,internal/domains/wishlists/handlers,internal/domains/items/handlers,internal/domains/reservations/handlers,internal/domains/storage/handlers,internal/domains/health/handlers

# OR (simpler, but scans more files)
swag init -g cmd/server/main.go -d internal/domains
```

**Verification**:
```bash
# Regenerate docs
swag init -g cmd/server/main.go -d internal/domains

# Check output
ls -la docs/swagger/
# Expected: swagger.json, swagger.yaml updated

# Verify endpoints documented
grep -c "@Router" docs/swagger/swagger.json
# Expected: Same count as before migration
```

---

### Risk 5: Third-Party Tool Compatibility

**Problem**: IDE, debuggers, or tools may cache old paths
**Likelihood**: Low
**Impact**: Low (developer inconvenience)

**Mitigation**:
```bash
# Clear Go module cache
go clean -modcache

# Clear IDE caches (VSCode example)
rm -rf .vscode/.cache

# Regenerate IDE config
go mod tidy
go mod download
```

---

## ✅ Success Criteria

### Quantitative Metrics

- [ ] **Build**: `go build ./cmd/server` succeeds with 0 errors
- [ ] **Tests**: All 28 tests pass (100% passing rate)
- [ ] **Linter**: `golangci-lint run` reports 0 new issues
- [ ] **Coverage**: Test coverage ≥ baseline (check `coverage-after.txt`)
- [ ] **Import Cycles**: 0 import cycles detected
- [ ] **File Count**: 45 domain files successfully migrated
- [ ] **DTO Consolidation**: Duplicate DTOs reduced (e.g., `CreateItemRequest` variants unified)

### Qualitative Criteria

- [ ] **Code Organization**: Each domain is self-contained with clear boundaries
- [ ] **Developer Experience**: New developers can find all wishlist code in one folder
- [ ] **Microservices Ready**: Domains can be extracted without major refactoring
- [ ] **Maintainability**: Adding new features has clear "where does this go?" answer
- [ ] **Consistency**: All domains follow same structure (handlers/services/repositories/dtos)

### Operational Validation

- [ ] **Server Start**: Application starts without errors
- [ ] **API Functional**: All endpoints respond correctly (smoke test)
- [ ] **Swagger UI**: Documentation accessible at `/swagger/index.html`
- [ ] **CI/CD**: Pipeline passes on feature branch
- [ ] **Performance**: No regression in API response times

---

## 📋 Post-Migration Tasks

### Immediate (Within 1 day)

- [ ] **Documentation**: Update `CLAUDE.md` with new structure reference
- [ ] **README**: Update architecture section in main README
- [ ] **Team Communication**: Notify team about new structure
- [ ] **PR Review**: Create PR with detailed migration summary

### Short-term (Within 1 week)

- [ ] **Developer Onboarding**: Update onboarding docs with domain navigation guide
- [ ] **Code Review Guidelines**: Update to reflect domain ownership
- [ ] **IDE Settings**: Share VSCode/GoLand workspace settings for new structure
- [ ] **Monitoring**: Watch for any production issues related to refactoring

### Long-term (Within 1 month)

- [ ] **Team Training**: Conduct walkthrough of domain structure
- [ ] **Convention Documentation**: Document domain design patterns and conventions
- [ ] **Tooling**: Create scripts for generating new domains (`scripts/create-domain.sh`)
- [ ] **Metrics**: Track "time to find code" improvements

---

## 🔄 Rollback Plan

### If Critical Issues Discovered

**During Migration** (before Phase 6):
```bash
# Rollback to last checkpoint
git log --oneline | head -10  # Find last good commit
git reset --hard <checkpoint-commit>
# Resume from problematic phase
```

**After Merge to Main** (emergency):
```bash
# Create revert branch
git checkout main
git pull
git checkout -b revert/domain-structure
git revert <merge-commit> --mainline 1
git push origin revert/domain-structure
# Create PR to revert
```

**Partial Rollback** (single domain issue):
```bash
# Revert specific domain commit
git log --oneline --all -- internal/domains/items
git revert <commit-hash>
# Fix issue in follow-up commit
```

---

## 📚 References

### Related Documentation
- `/docs/Go-Architecture-Guide.md` - 3-layer architecture principles
- `/docs/plans/00-cross-domain-architecture-plan.md` - Cross-domain auth
- `CLAUDE.md` - Project structure and conventions

### External Resources
- [Domain-Driven Design](https://martinfowler.com/bliki/DomainDrivenDesign.html)
- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [Package Oriented Design](https://www.ardanlabs.com/blog/2017/02/package-oriented-design.html)

---

## 🤝 Team Collaboration

### Code Review Focus Areas

Reviewers should verify:
1. **Domain Boundaries**: No cross-domain dependencies (except via interfaces)
2. **DTO Consolidation**: Duplicates eliminated, single source of truth
3. **Import Paths**: All updated to new structure
4. **Test Coverage**: No tests removed, all passing
5. **Documentation**: Inline comments updated for new locations

### Communication Plan

**Before Migration**:
- Share this plan with team for feedback
- Schedule migration during low-traffic period
- Assign backup reviewer

**During Migration**:
- Update team on progress after each phase
- Share blockers immediately in team chat
- Keep feature branch up-to-date with main

**After Migration**:
- Demo new structure in team meeting
- Gather feedback on developer experience
- Document lessons learned

---

## 📝 Notes

### Design Decisions

**Why "domains/" not "features/"?**
- "Domains" aligns with DDD terminology
- Clear distinction from frontend "features"
- Emphasizes bounded contexts

**Why "shared/" not "common/"?**
- "Shared" is more explicit about cross-cutting nature
- Avoids confusion with "common" business logic
- Matches microservices terminology (shared infrastructure)

**Why keep layers (handlers/services/repositories) within domains?**
- Maintains familiar 3-layer architecture
- Clear separation of concerns within domain
- Easy to extract to microservice (already layered)

**Items vs Wishlists separation rationale**:
- Items have independent lifecycle (can exist without wishlists)
- Many-to-many relationship via junction table
- Supports future features: item library, item recommendations
- Aligns with database schema (post-migration 000005)

### Lessons Learned (To be filled after migration)

- _What went well:_
- _What was challenging:_
- _What would we do differently:_
- _Unexpected issues:_
- _Time estimate accuracy:_

---

**Plan Version**: 1.0
**Last Updated**: 2026-02-09
**Next Review**: After Phase 6 completion
