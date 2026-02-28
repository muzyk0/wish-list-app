# E2E Tests

End-to-end tests for Wish List application using Playwright.

## Directory Structure

```
e2e/
├── frontend/          # Frontend UX tests
│   └── public-wishlist-responsive.spec.ts
├── backend/           # Backend API tests
│   ├── items-api.spec.ts
│   └── README.md
├── mobile/            # Mobile app tests (coming soon)
│   └── README.md
├── shared/            # Shared utilities and helpers
│   └── test-helpers.ts
└── README.md          # This file
```

## Quick Start

```bash
# Run all E2E tests
pnpm test

# Run only backend tests
pnpm test e2e/backend

# Run only mobile tests (when available)
pnpm test e2e/mobile

# Run with UI
pnpm test:ui

# Run in debug mode
pnpm test:debug
```

## Backend Tests

Located in `e2e/backend/`

**Coverage:**
- ✅ Items API (independent items)
- ✅ Many-to-many relationships
- ✅ Soft delete
- ✅ Pagination & filtering
- ✅ Public endpoints
- ✅ Error handling

**Status:** ✅ Complete (30+ tests)

See [backend/README.md](./backend/README.md) for details.

## Frontend Tests

Located in `e2e/frontend/`

**Coverage:**
- ✅ Responsive public wishlist layout across key breakpoints
- ✅ No horizontal overflow on mobile widths
- ✅ Reserved state is visible without exposing reserver identity

## Mobile Tests

Located in `e2e/mobile/`

**Status:** 📋 Planned

Mobile tests will cover:
- Authentication flow
- Wishlist management
- Item management
- Deep linking
- Offline mode

## Shared Utilities

Located in `e2e/shared/`

Reusable test helpers:
- `createTestUser()` - Generate test users
- `registerAndLogin()` - Auth helper
- `createTestWishlist()` - Wishlist factory
- `createTestItem()` - Item factory
- `waitForCondition()` - Async wait helper
- `assertStatus()` - Response assertion

Example usage:

```typescript
import { registerAndLogin, createTestWishlist } from '../shared/test-helpers';

test('my test', async ({ request }) => {
  const { token, userId } = await registerAndLogin(request);
  const wishlist = await createTestWishlist(request, token);
  // ... rest of test
});
```

## Configuration

Tests are configured in `/playwright.config.ts`

**Key settings:**
- Base URL: `http://localhost:8080`
- Auto-start backend if not running
- Parallel execution disabled for API tests
- HTML reporter enabled

## CI/CD Integration

```yaml
# Example GitHub Actions
- name: Run E2E Tests
  run: |
    pnpm install
    pnpm test
```

## Writing New Tests

1. **Choose directory**: `backend/` or `mobile/`
2. **Create test file**: `feature-name.spec.ts`
3. **Use shared helpers**: Import from `../shared/test-helpers`
4. **Follow patterns**: See existing tests for structure

```typescript
import { test, expect } from '@playwright/test';
import { registerAndLogin } from '../shared/test-helpers';

test.describe('Feature Name', () => {
  test('should do something', async ({ request }) => {
    const { token } = await registerAndLogin(request);

    const response = await request.get('/api/endpoint', {
      headers: { Authorization: `Bearer ${token}` },
    });

    expect(response.status()).toBe(200);
  });
});
```

## Best Practices

1. **Use shared helpers** for common operations
2. **Create fresh test data** for each test (don't rely on existing data)
3. **Clean up after tests** (handled automatically by Playwright)
4. **Use descriptive test names** (should read like documentation)
5. **Group related tests** with `test.describe()`
6. **Test happy path AND error cases**
7. **Keep tests independent** (don't depend on other tests)

## Troubleshooting

**Tests fail with "Connection refused"**
- Ensure database is running: `make db-up`
- Check backend can start: `make backend`

**Tests timeout**
- Increase timeout in playwright.config.ts
- Check server logs for errors

**Flaky tests**
- Use `waitForCondition()` for async operations
- Avoid hardcoded delays
- Check for race conditions

**Authentication fails**
- Verify JWT secret is configured
- Check user registration works manually
- Ensure database is clean
