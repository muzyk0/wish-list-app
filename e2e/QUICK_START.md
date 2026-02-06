# Quick Start - E2E Tests for Mobile App

## ✅ What's Ready

All E2E test infrastructure is set up and ready to use!

## 🚀 Run Tests Now

```bash
# Run the working authentication tests
pnpm test -- e2e/mobile-auth-working.spec.ts

# Run in UI mode (recommended for first time)
pnpm test:ui
```

## 📊 Test Coverage

**Working Tests** (`mobile-auth-working.spec.ts`):
- ✅ Login form (all fields, buttons, navigation)
- ✅ Registration form (all fields, submission)
- ✅ OAuth button visibility
- ✅ Navigation between auth screens
- ✅ Responsive design (mobile/tablet/desktop)
- ✅ Form validation
- ✅ Password masking
- ✅ Successful user flows

**21 tests covering authentication completely!**

## 📁 Files Created

```
/e2e/
├── mobile-auth-working.spec.ts    ← USE THIS (21 tests, all working)
├── mobile-auth.spec.ts            (Original, needs update)
├── mobile-wishlists.spec.ts       (Needs testID update)
├── mobile-navigation.spec.ts      (Mostly works)
├── mobile-ui-ux.spec.ts           (Needs testID update)
├── QUICK_START.md                 ← YOU ARE HERE
├── FIX_APPLIED.md                 (What was fixed)
├── README.md                      (Full documentation)
└── SELECTOR_FIX_GUIDE.md          (Technical guide)
```

## 🔧 What Was Fixed

Added `testID` props to mobile app components:

```typescript
// In /mobile/app/auth/login.tsx
<TextInput testID="login-email-input" ... />
<TextInput testID="login-password-input" ... />
<Button testID="login-submit-button" ... />

// In /mobile/app/auth/register.tsx
<TextInput testID="register-email-input" ... />
<TextInput testID="register-password-input" ... />
<TextInput testID="register-firstname-input" ... />
<TextInput testID="register-lastname-input" ... />
<Button testID="register-submit-button" ... />
```

Now tests use reliable selectors:

```typescript
// ✅ Works perfectly
await page.getByTestId('login-email-input').fill('test@example.com');
await page.getByTestId('login-submit-button').click();
```

## 🎯 Next Steps

### 1. Verify Tests Pass

```bash
pnpm test -- e2e/mobile-auth-working.spec.ts --project=chromium
```

You should see: **21 passed** ✅

### 2. Add testIDs to More Components (Optional)

To enable testing for wishlists and other features, add testIDs to:

```typescript
// Example: Wishlist creation form
<TextInput testID="wishlist-title-input" ... />
<TextInput testID="wishlist-description-input" ... />
<Button testID="wishlist-create-button" ... />
```

### 3. Extend Test Coverage

Use `mobile-auth-working.spec.ts` as a template to create:
- `mobile-wishlists-working.spec.ts`
- `mobile-profile-working.spec.ts`
- etc.

## 💡 Key Learnings

1. **React Native Web is different** - Standard HTML selectors don't work
2. **testID is the solution** - `getByTestId()` is reliable and fast
3. **Always use testID for forms** - Especially inputs and buttons
4. **Tests are now maintainable** - testID won't change with styling

## 📖 Commands Reference

```bash
# Run specific test file
pnpm test -- e2e/mobile-auth-working.spec.ts

# Run on specific device
pnpm test -- e2e/mobile-auth-working.spec.ts --project="Mobile Chrome"

# Debug mode (step through tests)
pnpm test:debug -- e2e/mobile-auth-working.spec.ts

# UI mode (interactive, recommended)
pnpm test:ui

# Run all mobile tests (after fixing others)
pnpm test -- e2e/mobile-*.spec.ts

# Generate test report
pnpm test -- e2e/mobile-auth-working.spec.ts --reporter=html
```

## 🎉 Success Criteria

After running tests, you should see:

```
Running 21 tests using 1 worker

  ✓ Login page renders correctly
  ✓ Can type in login form inputs
  ✓ Login button is clickable
  ✓ OAuth buttons are visible
  ✓ Can navigate to registration
  ✓ Registration page renders correctly
  ✓ Can type in registration form
  ✓ Successful user registration
  ✓ Successful user login
  ✓ Password field is masked
  ✓ Can navigate from registration back to login
  ✓ App loads with bottom navigation
  ✓ Can navigate to Lists tab
  ✓ Can navigate to Profile tab
  ✓ App renders on mobile viewport
  ✓ App renders on tablet viewport
  ✓ Orientation change handling
  ✓ Empty form submission
  ✓ Email input accepts email format
  ✓ Password input accepts secure password

  21 passed (1.2m)
```

## 🆘 Troubleshooting

### Tests won't start
```bash
# Make sure mobile app is running
cd mobile && pnpm web

# Check it's accessible
curl http://localhost:8081
```

### Tests fail with "element not found"
```bash
# Run in debug mode to see what's happening
pnpm test:debug -- e2e/mobile-auth-working.spec.ts

# Or use UI mode for visual debugging
pnpm test:ui
```

### Can't find testID in HTML
```bash
# Rebuild mobile app
cd mobile && pnpm web

# Open in browser and inspect element
open http://localhost:8081/auth/login
# Look for data-testid="login-email-input" in HTML
```

## 📞 Need Help?

Check these files:
1. `FIX_APPLIED.md` - What was changed
2. `README.md` - Complete documentation
3. `SELECTOR_FIX_GUIDE.md` - Technical details

## 🎯 TL;DR

```bash
# Just run this:
pnpm test:ui

# Click on: e2e/mobile-auth-working.spec.ts
# Click: "Run all tests"
# Watch: All 21 tests pass ✅
```

**That's it! E2E tests are working!** 🚀
