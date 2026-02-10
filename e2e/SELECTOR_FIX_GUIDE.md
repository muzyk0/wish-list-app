# Mobile App Test Selector Fix Guide

## Problem

The original tests were written for standard HTML forms with `placeholder` attributes, but the mobile app uses **React Native Paper** components which render differently in the web version.

## Key Differences

### ❌ What Doesn't Work

```typescript
// DON'T USE - These selectors don't exist in React Native Web
await page.getByPlaceholder('Email')
await page.getByPlaceholder('Password')
await page.getByLabel('Email')
```

### ✅ What Works

```typescript
// USE THESE - React Native Web compatible selectors
await page.locator('input[type="email"]').first()
await page.locator('input[type="password"]').first()
await page.getByRole('button', { name: 'Sign In' })
await page.getByRole('button', { name: /Continue with Google/i })
await page.getByText('Welcome Back')
```

## React Native Paper Input Structure

React Native Paper `TextInput` components render as:
- `<input type="email">` for email fields (no placeholder attribute)
- `<input type="password">` for password fields (no placeholder attribute)
- Labels are separate `<div>` elements with styled text
- No `<label>` tags or `for` attributes

## Selector Strategy

### 1. Input Fields

**Email Input:**
```typescript
const emailInput = page.locator('input[type="email"]').first();
await emailInput.fill('test@example.com');
```

**Password Input:**
```typescript
const passwordInput = page.locator('input[type="password"]').first();
await passwordInput.fill('password123');
```

**Generic Input (when order is known):**
```typescript
const inputs = await page.locator('input').all();
await inputs[0].fill('email@example.com');  // First input
await inputs[1].fill('password');            // Second input
```

### 2. Buttons

**By Exact Text:**
```typescript
await page.getByRole('button', { name: 'Sign In' }).click();
await page.getByRole('button', { name: 'Create Account' }).click();
```

**By Pattern (Case Insensitive):**
```typescript
await page.getByRole('button', { name: /sign in/i }).click();
await page.getByRole('button', { name: /continue with google/i }).click();
```

### 3. Text Content

**Headers:**
```typescript
await expect(page.getByText('Welcome Back')).toBeVisible();
await expect(page.getByText('Create Account')).toBeVisible();
```

**Navigation:**
```typescript
await page.getByText('Lists').click();
await page.getByText('Profile').click();
```

### 4. Navigation Elements

**Tabs (Bottom Navigation):**
```typescript
// Try both role and text selectors
const listsTab = page.getByRole('button', { name: 'Lists' })
  .or(page.getByText('Lists'));
await listsTab.click();
```

**Back Button:**
```typescript
await page.getByRole('button', { name: /Back/i }).first().click();
```

## Common Patterns

### Login Flow

```typescript
test('Login flow', async ({ page }) => {
  await page.goto('http://localhost:8081/auth/login');

  // Fill form
  await page.locator('input[type="email"]').first().fill('user@example.com');
  await page.locator('input[type="password"]').first().fill('password123');

  // Submit
  await page.getByRole('button', { name: 'Sign In' }).click();

  // Wait for navigation
  await page.waitForURL(/\/(tabs)/);
});
```

### Registration Flow

```typescript
test('Registration flow', async ({ page }) => {
  await page.goto('http://localhost:8081/auth/register');

  // Get all inputs
  const inputs = await page.locator('input').all();

  // Fill in order: email, password, firstName, lastName
  if (inputs.length >= 4) {
    await inputs[0].fill('new@example.com');
    await inputs[1].fill('Password123!');
    await inputs[2].fill('John');
    await inputs[3].fill('Doe');
  }

  // Submit
  await page.getByRole('button', { name: /Create Account/i }).click();
});
```

### Form Validation

```typescript
test('Form validation', async ({ page }) => {
  await page.goto('http://localhost:8081/auth/login');

  // Submit without filling
  await page.getByRole('button', { name: 'Sign In' }).click();

  // Check for validation error dialog or message
  // Note: React Native Paper may show Alert dialogs
  await page.waitForTimeout(1000);
});
```

## Debugging Selectors

### Inspect Available Elements

```typescript
// Get all inputs
const inputs = await page.locator('input').all();
console.log('Input count:', inputs.length);

// Get all buttons
const buttons = await page.locator('button, [role="button"]').all();
for (const button of buttons) {
  const text = await button.textContent();
  console.log('Button:', text);
}

// Get all text content
const texts = await page.locator('h1, h2, h3, p, span').allTextContents();
console.log('Texts:', texts);
```

### Use Playwright Inspector

```bash
# Run test in debug mode
pnpm test:debug -- e2e/mobile-auth-fixed.spec.ts
```

### Take Screenshots

```typescript
// Take screenshot before action
await page.screenshot({ path: 'before-click.png' });

// Perform action
await page.getByRole('button', { name: 'Sign In' }).click();

// Take screenshot after
await page.screenshot({ path: 'after-click.png' });
```

## Waiting Strategies

### Network Idle
```typescript
await page.waitForLoadState('networkidle');
```

### Specific Element
```typescript
await expect(page.getByText('Welcome Back')).toBeVisible({ timeout: 10000 });
```

### URL Change
```typescript
await page.waitForURL(/\/(tabs)/, { timeout: 10000 });
```

### Timeout for React Native Animations
```typescript
// React Native Web animations may need time
await page.waitForTimeout(1000);
```

## File-by-File Fix Checklist

### ✅ mobile-auth-fixed.spec.ts
- [x] Uses `input[type="email"]` selector
- [x] Uses `input[type="password"]` selector
- [x] Uses button role selectors
- [x] Uses text content selectors
- [x] Handles React Native Web structure

### 🔧 mobile-auth.spec.ts (Original - Needs Full Rewrite)
- [ ] Replace all `getByPlaceholder()` with `locator('input[type="..."]')`
- [ ] Replace all `getByLabel()` with appropriate selectors
- [ ] Update form filling logic
- [ ] Add proper wait strategies

### 🔧 mobile-wishlists.spec.ts
- [ ] Update wishlist form selectors
- [ ] Fix create/edit form selectors
- [ ] Update list item selectors
- [ ] Add proper wait for API calls

### 🔧 mobile-navigation.spec.ts
- [ ] Update tab navigation selectors
- [ ] Fix deep link testing
- [ ] Update URL assertions for Expo Router

### 🔧 mobile-ui-ux.spec.ts
- [ ] Update accessibility selectors
- [ ] Fix responsive design tests
- [ ] Update touch interaction tests

## Best Practices

1. **Always use `.first()`** when selecting inputs by type:
   ```typescript
   await page.locator('input[type="email"]').first().fill('...');
   ```

2. **Prefer role selectors for buttons**:
   ```typescript
   await page.getByRole('button', { name: 'Sign In' }).click();
   ```

3. **Use regex for flexible matching**:
   ```typescript
   await page.getByRole('button', { name: /sign in/i }).click();
   ```

4. **Add timeouts for React Native animations**:
   ```typescript
   await page.waitForTimeout(1000);
   ```

5. **Use `toBeVisible()` instead of `toExist()`**:
   ```typescript
   await expect(page.getByText('Welcome')).toBeVisible();
   ```

6. **Wait for network idle on page loads**:
   ```typescript
   await page.waitForLoadState('networkidle');
   ```

## Testing the Fixes

Run individual test files to verify:

```bash
# Test fixed auth
pnpm test -- e2e/mobile-auth-fixed.spec.ts --project="Mobile Chrome"

# Debug mode
pnpm test:debug -- e2e/mobile-auth-fixed.spec.ts

# UI mode
pnpm test:ui
```

## Common Errors and Solutions

### Error: "Element not found"
**Solution**: Use Playwright Inspector to find correct selector
```bash
pnpm test:debug -- e2e/your-test.spec.ts
```

### Error: "Element is not visible"
**Solution**: Add wait for visibility
```typescript
await expect(element).toBeVisible({ timeout: 10000 });
```

### Error: "Input value not set"
**Solution**: Ensure input is focused before filling
```typescript
await input.focus();
await input.fill('value');
```

### Error: "Navigation didn't happen"
**Solution**: Increase timeout for navigation
```typescript
await page.waitForURL(/expected-url/, { timeout: 15000 });
```

## Next Steps

1. Run `mobile-auth-fixed.spec.ts` to verify the approach works
2. Create similar fixed versions for other test files
3. Update original test files with correct selectors
4. Document any app-specific selector patterns
5. Create helper functions for common operations

## Example Helper Functions

```typescript
// helpers/mobile-test-helpers.ts

export async function loginUser(page, email: string, password: string) {
  await page.goto('http://localhost:8081/auth/login');
  await page.locator('input[type="email"]').first().fill(email);
  await page.locator('input[type="password"]').first().fill(password);
  await page.getByRole('button', { name: 'Sign In' }).click();
  await page.waitForURL(/\/(tabs)/, { timeout: 10000 });
}

export async function registerUser(page, email: string, password: string, firstName: string, lastName: string) {
  await page.goto('http://localhost:8081/auth/register');
  const inputs = await page.locator('input').all();
  if (inputs.length >= 4) {
    await inputs[0].fill(email);
    await inputs[1].fill(password);
    await inputs[2].fill(firstName);
    await inputs[3].fill(lastName);
  }
  await page.getByRole('button', { name: /Create Account/i }).click();
}

export async function navigateToTab(page, tabName: string) {
  const tab = page.getByRole('button', { name: tabName })
    .or(page.getByText(tabName));
  await tab.click();
  await page.waitForTimeout(500);
}
```
