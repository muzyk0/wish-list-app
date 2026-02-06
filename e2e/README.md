# Mobile App E2E Tests

## Overview

Comprehensive E2E test suite for the Wish List mobile application web version using Playwright.

## Test Files Created

### 1. mobile-auth.spec.ts
**Test Coverage: Authentication Flows**
- Registration form rendering and validation
- Login form rendering and validation
- Successful user registration
- Successful user login
- Invalid credentials handling
- Navigation between auth screens
- OAuth button visibility
- Form validation (email format, password strength, required fields)
- Duplicate registration prevention
- Security (XSS prevention, password masking)

**Total Tests**: 17 test cases (T050-T065)

### 2. mobile-wishlists.spec.ts
**Test Coverage: Wishlist Management**
- Wishlist tab accessibility
- Empty state display
- Create wishlist form rendering and validation
- Successfully create wishlist
- Edit wishlist navigation
- Delete wishlist with confirmation
- Public/private wishlist toggle
- Wishlist stats display
- Pull-to-refresh functionality
- Error state handling
- Loading state display

**Total Tests**: 17 test cases (T066-T082)

### 3. mobile-navigation.spec.ts
**Test Coverage: Navigation & Routing**
- Bottom tab navigation
- Tab switching between Lists, Profile, Reservations, Explore, Home
- Deep linking to specific wishlists
- Back navigation
- Protected route authentication
- App bar header display
- Modal navigation
- 404 handling for invalid routes
- URL parameter preservation
- Navigation performance
- Universal links
- Auth redirect flow

**Total Tests**: 20 test cases (T083-T101)

### 4. mobile-ui-ux.spec.ts
**Test Coverage: UI/UX & Accessibility**
- Responsive design (mobile, tablet, desktop viewports)
- Orientation change handling
- Touch interactions (tap, swipe, long press)
- Accessibility (labels, keyboard navigation, focus indicators, color contrast)
- Loading states and spinners
- Error handling UI
- Form UX (autofocus, Enter key submit, input clearing, password toggle)
- Performance (page load time, console errors, image loading)
- Visual regression (screenshot capture)

**Total Tests**: 27 test cases (T102-T128)

## Test Configuration

### Playwright Configuration
Located in `/playwright.config.ts`:

- **Test Directory**: `./e2e`
- **Browsers**: Chromium, Firefox, WebKit
- **Mobile Devices**: Pixel 5, iPhone 12, iPad Pro
- **Base URLs**:
  - Backend API: `http://localhost:8080`
  - Mobile Web: `http://localhost:8081` (Expo web server)

### Web Servers
Playwright automatically starts:
1. Backend Go server on port 8080
2. Mobile Expo web server on port 8081

## Running Tests

### All Mobile Tests
```bash
pnpm test -- e2e/mobile-*.spec.ts
```

### Specific Test File
```bash
pnpm test -- e2e/mobile-auth.spec.ts
```

### Specific Device
```bash
pnpm test -- e2e/mobile-auth.spec.ts --project="Mobile Chrome"
pnpm test -- e2e/mobile-auth.spec.ts --project="Mobile Safari"
pnpm test -- e2e/mobile-auth.spec.ts --project="Tablet"
```

### Debug Mode
```bash
pnpm test:debug -- e2e/mobile-auth.spec.ts
```

### UI Mode
```bash
pnpm test:ui
```

### With Specific Workers
```bash
pnpm test -- e2e/mobile-*.spec.ts --workers=1
```

## Test Strategy

### Helper Functions
Each test file includes a `registerAndLogin` helper function that:
1. Registers a unique test user via API
2. Navigates to login page
3. Fills credentials
4. Submits form
5. Waits for redirect to main app

This ensures tests run independently and don't interfere with each other.

### Test Isolation
- Each test uses unique email addresses with timestamps
- Tests clean up after themselves where possible
- API calls use the request context for authentication

### Assertions
- Uses Playwright's `expect` for assertions
- Includes timeout parameters for async operations
- Validates both positive and negative test cases

## Test Coverage Summary

| Category | Test Count | Coverage |
|----------|-----------|----------|
| Authentication | 17 | Registration, Login, OAuth, Validation, Security |
| Wishlist Management | 17 | CRUD operations, Visibility, Stats, Error handling |
| Navigation | 20 | Tabs, Deep links, Protected routes, Performance |
| UI/UX | 27 | Responsive, Accessibility, Touch, Loading states |
| **Total** | **81** | **Comprehensive mobile web app coverage** |

## Key Features Tested

✅ User registration and login
✅ Wishlist CRUD operations
✅ Public/private wishlist visibility
✅ Multi-device responsive design
✅ Touch interactions
✅ Accessibility (WCAG compliance basics)
✅ Loading and error states
✅ Form validation
✅ Navigation flows
✅ Deep linking
✅ Protected routes
✅ OAuth integration UI
✅ Performance metrics

## Test Maintenance

### Adding New Tests
1. Create a new `.spec.ts` file in `/e2e/`
2. Import Playwright test utilities
3. Define test suites with `test.describe()`
4. Use helper functions for common setups
5. Follow existing naming conventions (T[number])

### Updating Tests
- When mobile app UI changes, update selectors
- When API endpoints change, update helper functions
- Keep test IDs sequential for tracking

### Best Practices
- Use semantic selectors (role, placeholder, text)
- Avoid brittle selectors (classes, IDs)
- Include descriptive test names
- Log success messages for visibility
- Handle async operations with proper timeouts

## Known Limitations

1. **OAuth Testing**: OAuth flows are UI-only (buttons visible), not fully tested due to external provider dependencies
2. **Visual Regression**: Screenshots captured but not compared automatically
3. **Performance**: Basic timing checks only, not comprehensive performance testing
4. **Network Simulation**: No offline mode or slow network testing
5. **Swipe Gestures**: Placeholder tests, complex gestures not fully implemented

## Future Enhancements

- [ ] Visual regression testing with screenshot comparison
- [ ] Network condition simulation (offline, slow 3G)
- [ ] Complex touch gesture testing
- [ ] OAuth flow mocking for full integration testing
- [ ] Accessibility score calculation
- [ ] Performance budget enforcement
- [ ] Cross-browser compatibility matrix
- [ ] CI/CD integration with automated reporting

## Continuous Integration

For CI/CD pipelines:
```bash
# Headless mode with retries
pnpm test -- e2e/mobile-*.spec.ts --retries=2 --workers=4

# Generate HTML report
pnpm test -- e2e/mobile-*.spec.ts --reporter=html
```

## Troubleshooting

### Tests Failing to Start
- Ensure backend server is running on port 8080
- Ensure mobile expo server is running on port 8081
- Check `playwright.config.ts` webServer configuration

### Tests Timing Out
- Increase timeout in test: `{ timeout: 30000 }`
- Check network connectivity
- Verify backend database is accessible

### Flaky Tests
- Use `test.retry(2)` for specific tests
- Increase wait times for dynamic content
- Use `waitForLoadState('networkidle')` for page loads

## Contributing

When adding new features to the mobile app:
1. Write E2E tests first (TDD approach)
2. Follow existing test structure and conventions
3. Ensure tests pass on all device types
4. Update this README with new test coverage

## Contact

For questions or issues with the test suite, refer to the main project documentation or create an issue in the project repository.
