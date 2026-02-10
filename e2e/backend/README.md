# Backend API E2E Tests

Comprehensive E2E tests for the Wish List Backend API.

## Test Files

### `items-api.spec.ts`
Tests for new Items API architecture:
- ✅ Independent items CRUD (`/api/items`)
- ✅ Many-to-many relationships (`/api/wishlists/:id/items`)
- ✅ Soft delete functionality
- ✅ Pagination and filtering
- ✅ Public endpoints
- ✅ Error handling

**Test Coverage**: 30+ tests across 10 test suites

## Running Tests

```bash
# From project root
cd /Users/vladislav/Web/wish-list-app

# Run all backend API tests
pnpm test e2e/backend

# Run specific test file
pnpm test e2e/backend/items-api

# Run with UI
pnpm test:ui e2e/backend

# Run in debug mode
pnpm test:debug e2e/backend/items-api
```

## Test Structure

Each test file follows this pattern:

1. **Setup** - Register user, create test data
2. **Feature Tests** - Test specific functionality
3. **Cleanup** - Handled automatically by Playwright

## Adding New Tests

```typescript
import { test, expect } from '@playwright/test';

test.describe('Feature Name', () => {
  test('should do something', async ({ request }) => {
    const response = await request.get('/api/endpoint');
    expect(response.status()).toBe(200);
  });
});
```

## API Base URL

Tests use `http://localhost:8080/api` as base URL.

Backend is automatically started by Playwright config if not running.

## Prerequisites

- Backend server must be able to start
- PostgreSQL database must be running
- Environment variables configured in `.env`
