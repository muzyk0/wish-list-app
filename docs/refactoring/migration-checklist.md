# Domain-Driven Structure Migration Checklist

**Date Started**: _____________________
**Completed**: _____________________
**Total Time**: _____ hours

---

## Phase 0: Preparation (1-2 hours)

**Objective**: Create safety checkpoints and verify baseline

### Setup
- [ ] Create feature branch: `git checkout -b refactor/domain-driven-structure`
- [ ] Verify all tests pass: `go test ./...`
- [ ] Run linter: `golangci-lint run`
- [ ] Record coverage baseline: `go test -cover ./... > coverage-baseline.txt`
- [ ] Create checkpoint: `git commit -m "chore: checkpoint before domain refactoring"`
- [ ] Document imports: `grep -r "wish-list/internal" backend/cmd backend/internal > imports-before.txt`

### Validation
- [ ] ✅ All tests passing
- [ ] ✅ Linter clean (0 issues)
- [ ] ✅ Coverage baseline recorded
- [ ] ✅ Git checkpoint created
- [ ] ✅ Team notified about upcoming refactoring

**Time Spent**: _____ hours
**Issues**: _______________________________________________

---

## Phase 1: Create Domain Structure (30 mins)

**Objective**: Create new folder hierarchy without moving files

### Create Folders
```bash
mkdir -p backend/internal/domains/{auth,wishlists,items,reservations,storage,health}
mkdir -p backend/internal/shared/{middleware,config,db,cache,encryption,validation,analytics,aws}

for domain in auth wishlists items reservations storage health; do
    mkdir -p backend/internal/domains/$domain/{handlers,services,repositories,dtos}
done
```

### Checklist
- [ ] Created `domains/` folder
- [ ] Created 6 domain subfolders: auth, wishlists, items, reservations, storage, health
- [ ] Created `shared/` folder
- [ ] Created 8 shared subfolders: middleware, config, db, cache, encryption, validation, analytics, aws
- [ ] Created layer folders (handlers, services, repositories, dtos) in each domain
- [ ] Verified structure: `tree backend/internal/domains -L 2`
- [ ] Verified structure: `tree backend/internal/shared -L 1`

### Validation
- [ ] ✅ All folders created successfully
- [ ] ✅ Structure matches target architecture
- [ ] ✅ No files moved yet (zero risk)
- [ ] ✅ Build still succeeds: `go build ./cmd/server`

**Time Spent**: _____ mins
**Issues**: _______________________________________________

---

## Phase 2: Move Cross-Cutting Concerns (1 hour)

**Objective**: Move infrastructure packages to `shared/`

### Move Packages
- [ ] Move `middleware/` to `shared/`: `mv backend/internal/middleware backend/internal/shared/`
- [ ] Move `config/` to `shared/`: `mv backend/internal/config backend/internal/shared/`
- [ ] Move `db/` to `shared/`: `mv backend/internal/db backend/internal/shared/`
- [ ] Move `cache/` to `shared/`: `mv backend/internal/cache backend/internal/shared/`
- [ ] Move `encryption/` to `shared/`: `mv backend/internal/encryption backend/internal/shared/`
- [ ] Move `validation/` to `shared/`: `mv backend/internal/validation backend/internal/shared/`
- [ ] Move `analytics/` to `shared/`: `mv backend/internal/analytics backend/internal/shared/`
- [ ] Move `aws/` to `shared/`: `mv backend/internal/aws backend/internal/shared/`

### Update Imports in Moved Files
```bash
find backend/internal/shared -name "*.go" -type f -exec sed -i '' \
    's|wish-list/internal/middleware|wish-list/internal/shared/middleware|g' {} +
find backend/internal/shared -name "*.go" -type f -exec sed -i '' \
    's|wish-list/internal/config|wish-list/internal/shared/config|g' {} +
find backend/internal/shared -name "*.go" -type f -exec sed -i '' \
    's|wish-list/internal/db|wish-list/internal/shared/db|g' {} +
find backend/internal/shared -name "*.go" -type f -exec sed -i '' \
    's|wish-list/internal/cache|wish-list/internal/shared/cache|g' {} +
find backend/internal/shared -name "*.go" -type f -exec sed -i '' \
    's|wish-list/internal/encryption|wish-list/internal/shared/encryption|g' {} +
find backend/internal/shared -name "*.go" -type f -exec sed -i '' \
    's|wish-list/internal/validation|wish-list/internal/shared/validation|g' {} +
find backend/internal/shared -name "*.go" -type f -exec sed -i '' \
    's|wish-list/internal/analytics|wish-list/internal/shared/analytics|g' {} +
find backend/internal/shared -name "*.go" -type f -exec sed -i '' \
    's|wish-list/internal/aws|wish-list/internal/shared/aws|g' {} +
```

- [ ] Updated imports in moved files

### Update Imports in Remaining Files
```bash
for pkg in middleware config db cache encryption validation analytics aws; do
    find backend/internal/handlers backend/internal/services backend/internal/repositories backend/internal/auth backend/cmd -name "*.go" -type f -exec sed -i '' \
        "s|wish-list/internal/$pkg|wish-list/internal/shared/$pkg|g" {} +
done
```

- [ ] Updated imports in handlers/
- [ ] Updated imports in services/
- [ ] Updated imports in repositories/
- [ ] Updated imports in auth/
- [ ] Updated imports in cmd/

### Validation
- [ ] ✅ Shared packages moved successfully
- [ ] ✅ All imports updated
- [ ] ✅ Tests pass: `go test ./internal/shared/...`
- [ ] ✅ Build succeeds: `go build ./cmd/server`
- [ ] ✅ No import errors

### Commit
- [ ] Committed: `git add -A && git commit -m "refactor: move cross-cutting concerns to shared/"`

**Time Spent**: _____ hours
**Issues**: _______________________________________________

---

## Phase 3: Migrate Domains (3-4 hours)

### Domain 1: Health (Simplest - Learning)

#### Move Files
- [ ] Move `health_handler.go` to `domains/health/handlers/`
- [ ] Move `health_handler_test.go` to `domains/health/handlers/`

#### Update Files
- [ ] Update imports in `health_handler.go`
- [ ] Create `domains/health/health.go` domain export file

#### Update Main
- [ ] Update `cmd/server/main.go` import
- [ ] Update `cmd/server/main.go` handler instantiation

#### Validation
- [ ] ✅ Domain tests pass: `go test ./internal/domains/health/...`
- [ ] ✅ All tests pass: `go test ./...`
- [ ] ✅ Build succeeds

#### Commit
- [ ] Committed: `git commit -m "refactor(health): migrate to domain structure"`

**Time**: _____ mins
**Issues**: _______________________________________________

---

### Domain 2: Storage

#### Move Files
- [ ] Move `s3_handler.go` to `domains/storage/handlers/`
- [ ] Move `s3_handler_test.go` to `domains/storage/handlers/`
- [ ] Move relevant AWS service code to `domains/storage/services/`

#### Update Files
- [ ] Update imports in moved files
- [ ] Create `domains/storage/storage.go` domain export file

#### Update Main
- [ ] Update `cmd/server/main.go` imports
- [ ] Update handler instantiation

#### Validation
- [ ] ✅ Domain tests pass: `go test ./internal/domains/storage/...`
- [ ] ✅ All tests pass: `go test ./...`
- [ ] ✅ Build succeeds

#### Commit
- [ ] Committed: `git commit -m "refactor(storage): migrate to domain structure"`

**Time**: _____ mins
**Issues**: _______________________________________________

---

### Domain 3: Reservations

#### Move Files
- [ ] Move `reservation_handler.go` to `domains/reservations/handlers/`
- [ ] Move `reservation_handler_test.go` to `domains/reservations/handlers/`
- [ ] Move `reservation_service.go` to `domains/reservations/services/`
- [ ] Move `reservation_service_test.go` to `domains/reservations/services/`
- [ ] Move `reservation_repository.go` to `domains/reservations/repositories/`
- [ ] Move `reservation_repository_test.go` to `domains/reservations/repositories/`
- [ ] Move `email_service.go` to `domains/reservations/services/` (domain-specific)

#### Update Files
- [ ] Update imports in all moved files
- [ ] Create `domains/reservations/reservations.go` domain export file

#### Update Main
- [ ] Update `cmd/server/main.go` imports
- [ ] Update handler/service instantiation

#### Validation
- [ ] ✅ Domain tests pass: `go test ./internal/domains/reservations/...`
- [ ] ✅ All tests pass: `go test ./...`
- [ ] ✅ Build succeeds

#### Commit
- [ ] Committed: `git commit -m "refactor(reservations): migrate to domain structure"`

**Time**: _____ mins
**Issues**: _______________________________________________

---

### Domain 4: Items

#### Move Files
- [ ] Move `item_handler.go` to `domains/items/handlers/`
- [ ] Move `item_service.go` to `domains/items/services/`
- [ ] Move `item_service_test.go` to `domains/items/services/`
- [ ] Move `giftitem_repository.go` to `domains/items/repositories/`
- [ ] Move `giftitem_repository_test.go` to `domains/items/repositories/`

#### Update Files
- [ ] Update imports in all moved files
- [ ] Create `domains/items/items.go` domain export file

#### Update Main
- [ ] Update `cmd/server/main.go` imports
- [ ] Update handler/service instantiation

#### Validation
- [ ] ✅ Domain tests pass: `go test ./internal/domains/items/...`
- [ ] ✅ All tests pass: `go test ./...`
- [ ] ✅ Build succeeds

#### Commit
- [ ] Committed: `git commit -m "refactor(items): migrate to domain structure"`

**Time**: _____ mins
**Issues**: _______________________________________________

---

### Domain 5: Wishlists

#### Move Files
- [ ] Move `wishlist_handler.go` to `domains/wishlists/handlers/`
- [ ] Move `wishlist_handler_test.go` to `domains/wishlists/handlers/`
- [ ] Move `wishlist_item_handler.go` to `domains/wishlists/handlers/`
- [ ] Move `wishlist_service.go` to `domains/wishlists/services/`
- [ ] Move `wishlist_service_test.go` to `domains/wishlists/services/`
- [ ] Move `wishlist_service_template_methods.go` to `domains/wishlists/services/`
- [ ] Move `wishlist_item_service.go` to `domains/wishlists/services/`
- [ ] Move `wishlist_item_service_test.go` to `domains/wishlists/services/`
- [ ] Move `wishlist_repository.go` to `domains/wishlists/repositories/`
- [ ] Move `wishlist_repository_test.go` to `domains/wishlists/repositories/`
- [ ] Move `wishlistitem_repository.go` to `domains/wishlists/repositories/`
- [ ] Move `template_repository.go` to `domains/wishlists/repositories/`

#### Update Files
- [ ] Update imports in all moved files
- [ ] Create `domains/wishlists/wishlists.go` domain export file

#### Update Main
- [ ] Update `cmd/server/main.go` imports
- [ ] Update handler/service instantiation

#### Validation
- [ ] ✅ Domain tests pass: `go test ./internal/domains/wishlists/...`
- [ ] ✅ All tests pass: `go test ./...`
- [ ] ✅ Build succeeds

#### Commit
- [ ] Committed: `git commit -m "refactor(wishlists): migrate to domain structure"`

**Time**: _____ mins
**Issues**: _______________________________________________

---

### Domain 6: Auth (Most Complex)

#### Move Files
- [ ] Move `auth_handler.go` to `domains/auth/handlers/`
- [ ] Move `oauth_handler.go` to `domains/auth/handlers/`
- [ ] Move `user_handler.go` to `domains/auth/handlers/`
- [ ] Move `user_handler_test.go` to `domains/auth/handlers/`
- [ ] Move `user_service.go` to `domains/auth/services/`
- [ ] Move `user_service_test.go` to `domains/auth/services/`
- [ ] Move `account_cleanup_service.go` to `domains/auth/services/`
- [ ] Move `user_repository.go` to `domains/auth/repositories/`
- [ ] Move `user_repository_test.go` to `domains/auth/repositories/`
- [ ] Move `auth/middleware.go` to `domains/auth/middleware/`
- [ ] Move `auth/middleware_test.go` to `domains/auth/middleware/`
- [ ] Move `auth/token_manager.go` to `domains/auth/services/` OR keep in `domains/auth/`
- [ ] Move `auth/code_store.go` to `domains/auth/services/` OR keep in `domains/auth/`

#### Update Files
- [ ] Update imports in all moved files
- [ ] Create `domains/auth/auth.go` domain export file

#### Update Main
- [ ] Update `cmd/server/main.go` imports
- [ ] Update handler/service/middleware instantiation

#### Validation
- [ ] ✅ Domain tests pass: `go test ./internal/domains/auth/...`
- [ ] ✅ All tests pass: `go test ./...`
- [ ] ✅ Build succeeds
- [ ] ✅ Auth middleware still works

#### Commit
- [ ] Committed: `git commit -m "refactor(auth): migrate to domain structure"`

**Time**: _____ mins
**Issues**: _______________________________________________

---

## Phase 4: Consolidate DTOs (2 hours)

### Items Domain DTOs

#### Extract DTOs
- [ ] Create `domains/items/dtos/requests.go`
- [ ] Create `domains/items/dtos/responses.go`
- [ ] Move `CreateItemRequest` to `dtos/requests.go`
- [ ] Move `UpdateItemRequest` to `dtos/requests.go`
- [ ] Move `MarkPurchasedRequest` to `dtos/requests.go`
- [ ] Move `ItemResponse` to `dtos/responses.go`
- [ ] Move `PaginatedItemsResponse` to `dtos/responses.go`

#### Merge Duplicates
- [ ] Compare `CreateItemRequest` vs `CreateGiftItemRequest` (from wishlist_handler)
- [ ] Choose canonical field names (prefer `CreateItemRequest` version)
- [ ] Remove duplicate `CreateGiftItemRequest` from wishlists domain
- [ ] Update wishlists handlers to import `items/dtos`

#### Update Imports
- [ ] Update `domains/items/handlers/item_handler.go` imports
- [ ] Update cross-domain references (if any)

#### Validation
- [ ] ✅ Domain tests pass: `go test ./internal/domains/items/...`
- [ ] ✅ All tests pass: `go test ./...`
- [ ] ✅ Build succeeds

#### Commit
- [ ] Committed: `git commit -m "refactor(items): consolidate DTOs"`

**Time**: _____ mins
**Issues**: _______________________________________________

---

### Wishlists Domain DTOs

#### Extract DTOs
- [ ] Create `domains/wishlists/dtos/requests.go`
- [ ] Create `domains/wishlists/dtos/responses.go`
- [ ] Move `CreateWishListRequest` to `dtos/requests.go`
- [ ] Move `UpdateWishListRequest` to `dtos/requests.go`
- [ ] Move wishlist-specific item requests to `dtos/requests.go`
- [ ] Move `WishListResponse` to `dtos/responses.go`
- [ ] Remove duplicate item DTOs (now in items/dtos)

#### Update Imports
- [ ] Update `domains/wishlists/handlers/` imports
- [ ] Add imports to `domains/items/dtos` where needed

#### Validation
- [ ] ✅ Domain tests pass: `go test ./internal/domains/wishlists/...`
- [ ] ✅ All tests pass: `go test ./...`
- [ ] ✅ Build succeeds

#### Commit
- [ ] Committed: `git commit -m "refactor(wishlists): consolidate DTOs"`

**Time**: _____ mins
**Issues**: _______________________________________________

---

### Reservations Domain DTOs

#### Extract DTOs
- [ ] Create `domains/reservations/dtos/requests.go`
- [ ] Create `domains/reservations/dtos/responses.go`
- [ ] Move reservation request DTOs
- [ ] Move reservation response DTOs

#### Update Imports
- [ ] Update `domains/reservations/handlers/` imports

#### Validation
- [ ] ✅ Tests pass
- [ ] ✅ Build succeeds

#### Commit
- [ ] Committed: `git commit -m "refactor(reservations): consolidate DTOs"`

**Time**: _____ mins

---

### Auth Domain DTOs

#### Extract DTOs
- [ ] Create `domains/auth/dtos/requests.go`
- [ ] Create `domains/auth/dtos/responses.go`
- [ ] Move auth/OAuth request DTOs
- [ ] Move user request/response DTOs

#### Update Imports
- [ ] Update all auth handlers imports

#### Validation
- [ ] ✅ Tests pass
- [ ] ✅ Build succeeds

#### Commit
- [ ] Committed: `git commit -m "refactor(auth): consolidate DTOs"`

**Time**: _____ mins

---

## Phase 5: Update All Imports (1 hour)

### Main Application
- [ ] Update `cmd/server/main.go` - replace all handler imports with domain imports
- [ ] Update route registration to use domain exports
- [ ] Update middleware initialization

### Search for Old Imports
```bash
grep -r "wish-list/internal/handlers\"" backend/
grep -r "wish-list/internal/services\"" backend/
grep -r "wish-list/internal/repositories\"" backend/
```

- [ ] No results from grep (except within domains/)

### Update Tests Outside Domains
- [ ] Update `cmd/server/main_test.go` (if exists)
- [ ] Update integration tests (if any)

### Validation
- [ ] ✅ All tests pass: `go test ./...`
- [ ] ✅ Build succeeds: `go build ./cmd/server`
- [ ] ✅ No old import paths remain

### Commit
- [ ] Committed: `git commit -m "refactor: update all import paths to domain structure"`

**Time Spent**: _____ mins
**Issues**: _______________________________________________

---

## Phase 6: Validation & Cleanup (1 hour)

### Run All Tests
- [ ] Run full test suite: `go test ./... -v`
- [ ] All 28 tests passing
- [ ] No new test failures

### Run Linter
- [ ] Run linter: `golangci-lint run`
- [ ] 0 new issues introduced

### Check Coverage
- [ ] Generate coverage report: `go test -cover ./... > coverage-after.txt`
- [ ] Compare with baseline: `diff coverage-baseline.txt coverage-after.txt`
- [ ] Coverage % similar or improved

### Build & Run
- [ ] Build application: `go build ./cmd/server`
- [ ] Start server: `./server &`
- [ ] Test health endpoint: `curl http://localhost:8080/healthz`
- [ ] Stop server: `kill $SERVER_PID`

### Check Import Cycles
- [ ] Run: `go list -f '{{.ImportPath}}: {{.Imports}}' ./internal/domains/... | grep -i cycle`
- [ ] 0 import cycles detected

### Remove Old Directories
- [ ] Check for remaining files: `ls backend/internal/handlers/`
- [ ] Check for remaining files: `ls backend/internal/services/`
- [ ] Check for remaining files: `ls backend/internal/repositories/`
- [ ] Check for remaining files: `ls backend/internal/auth/`
- [ ] Remove empty directories:
  ```bash
  rmdir backend/internal/handlers 2>/dev/null || echo "Not empty"
  rmdir backend/internal/services 2>/dev/null || echo "Not empty"
  rmdir backend/internal/repositories 2>/dev/null || echo "Not empty"
  rmdir backend/internal/auth 2>/dev/null || echo "Not empty"
  ```

### Update Documentation
- [ ] Update imports report: `grep -r "wish-list/internal" backend/cmd backend/internal > imports-after.txt`
- [ ] Compare before/after: `diff imports-before.txt imports-after.txt`

### Swagger Docs
- [ ] Regenerate Swagger: `swag init -g cmd/server/main.go -d internal/domains`
- [ ] Verify docs updated: `ls -la docs/swagger/`
- [ ] Check endpoint count: `grep -c "@Router" docs/swagger/swagger.json`

### Final Commit
```bash
git add -A
git commit -m "refactor: complete domain-driven structure migration

- Migrated 45 domain files to domain folders
- Consolidated DTOs per domain
- Moved cross-cutting concerns to shared/
- All tests passing (28/28)
- Zero import cycles
- Coverage maintained"
```

- [ ] Final commit created

### Validation Checklist
- [ ] ✅ All 28 tests pass
- [ ] ✅ Linter clean (0 new issues)
- [ ] ✅ Coverage ≥ baseline
- [ ] ✅ Build succeeds
- [ ] ✅ Server runs successfully
- [ ] ✅ No import cycles
- [ ] ✅ Swagger docs updated
- [ ] ✅ Old directories removed
- [ ] ✅ Documentation updated

**Time Spent**: _____ hours
**Issues**: _______________________________________________

---

## Post-Migration

### Create Pull Request
- [ ] Push branch: `git push origin refactor/domain-driven-structure`
- [ ] Create PR with detailed summary
- [ ] Link to plan: `/docs/refactoring/domain-driven-structure-plan.md`
- [ ] Request reviews from: _______________________

### PR Description Template
```markdown
## Domain-Driven Structure Migration

Refactors backend from flat layer-based structure to domain-driven architecture.

### Summary
- ✅ Migrated 45 domain files to 6 domains
- ✅ Consolidated duplicate DTOs
- ✅ Moved cross-cutting concerns to shared/
- ✅ All 28 tests passing
- ✅ Zero import cycles

### Domains
- `auth/` - Authentication & Identity (11 files)
- `wishlists/` - Wishlist Management (10 files)
- `items/` - Gift Items (5 files)
- `reservations/` - Reservations & Purchases (5 files)
- `storage/` - File Storage (3 files)
- `health/` - Health Checks (2 files)

### Benefits
- 70% reduction in "file hopping"
- Clear domain boundaries for microservices
- Faster onboarding for new developers
- Resolved DTO duplication

### Testing
- All tests passing: `go test ./...`
- No regressions in coverage
- Linter clean: `golangci-lint run`

### Documentation
- Plan: `/docs/refactoring/domain-driven-structure-plan.md`
- Checklist: `/docs/refactoring/migration-checklist.md`

### Migration Time
- Estimated: 8-10 hours
- Actual: _____ hours
```

- [ ] PR created with above description

### Update Project Documentation
- [ ] Update `CLAUDE.md` backend structure section
- [ ] Update main `README.md` architecture section
- [ ] Update onboarding docs (if applicable)

### Team Communication
- [ ] Notify team in Slack/chat about structure change
- [ ] Schedule walkthrough meeting (optional)
- [ ] Share migration learnings

### Monitor Post-Merge
- [ ] CI/CD pipeline passes
- [ ] No production issues reported (first 24 hours)
- [ ] Team feedback collected

---

## Summary

**Total Time**: _____ hours
**Estimated**: 8-10 hours
**Variance**: _____ hours (under/over)

**Domains Migrated**: _____ / 6

**Issues Encountered**:
1. _________________________________________________
2. _________________________________________________
3. _________________________________________________

**Lessons Learned**:
1. _________________________________________________
2. _________________________________________________
3. _________________________________________________

**Would Do Differently**:
1. _________________________________________________
2. _________________________________________________

**Success Metrics**:
- ✅ All tests passing: _____ / 28
- ✅ Import cycles: 0
- ✅ Build successful: Yes/No
- ✅ Coverage maintained: Yes/No
- ✅ Linter clean: Yes/No

---

**Completed By**: _____________________
**Date**: _____________________
**Reviewed By**: _____________________
