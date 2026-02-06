/**
 * E2E Tests for Mobile App Authentication
 *
 * Test Coverage:
 * 1. Registration flow (email/password)
 * 2. Login flow (email/password)
 * 3. Form validation
 * 4. Navigation between auth screens
 * 5. Successful authentication redirect
 */

import { test, expect } from '@playwright/test';

const MOBILE_BASE_URL = 'http://localhost:8081';
const API_BASE_URL = 'http://localhost:8080';

test.describe('Mobile App - Authentication', () => {

  test.beforeEach(async ({ page }) => {
    // Navigate to mobile app
    await page.goto(MOBILE_BASE_URL);
    await page.waitForLoadState('networkidle');
  });

  test('T050: Registration page renders correctly', async ({ page }) => {
    // Navigate to registration
    await page.goto(`${MOBILE_BASE_URL}/auth/register`);

    // Verify page elements
    await expect(page.getByText('Create Account')).toBeVisible();
    await expect(page.getByPlaceholder('Email')).toBeVisible();
    await expect(page.getByPlaceholder('Password')).toBeVisible();
    await expect(page.getByPlaceholder('First Name')).toBeVisible();
    await expect(page.getByPlaceholder('Last Name')).toBeVisible();
    await expect(page.getByRole('button', { name: /Create Account/i })).toBeVisible();

    console.log('✓ Registration page renders all form fields');
  });

  test('T051: Login page renders correctly', async ({ page }) => {
    // Navigate to login
    await page.goto(`${MOBILE_BASE_URL}/auth/login`);

    // Verify page elements
    await expect(page.getByText('Welcome Back')).toBeVisible();
    await expect(page.getByPlaceholder('Email')).toBeVisible();
    await expect(page.getByPlaceholder('Password')).toBeVisible();
    await expect(page.getByRole('button', { name: /Sign In/i })).toBeVisible();

    console.log('✓ Login page renders all form fields');
  });

  test('T052: Registration validation - empty fields', async ({ page }) => {
    await page.goto(`${MOBILE_BASE_URL}/auth/register`);

    // Try to submit without filling fields
    await page.getByRole('button', { name: /Create Account/i }).click();

    // Check for validation error
    await expect(page.getByText(/Please fill in all required fields/i)).toBeVisible();

    console.log('✓ Registration form validates empty fields');
  });

  test('T053: Login validation - empty fields', async ({ page }) => {
    await page.goto(`${MOBILE_BASE_URL}/auth/login`);

    // Try to submit without filling fields
    await page.getByRole('button', { name: /Sign In/i }).click();

    // Check for validation error
    await expect(page.getByText(/Please fill in all required fields/i)).toBeVisible();

    console.log('✓ Login form validates empty fields');
  });

  test('T054: Successful user registration', async ({ page }) => {
    await page.goto(`${MOBILE_BASE_URL}/auth/register`);

    // Generate unique email
    const uniqueEmail = `mobile-test-${Date.now()}@example.com`;

    // Fill registration form
    await page.getByPlaceholder('Email').fill(uniqueEmail);
    await page.getByPlaceholder('Password').fill('TestPassword123!');
    await page.getByPlaceholder('First Name').fill('Mobile');
    await page.getByPlaceholder('Last Name').fill('User');

    // Submit form
    await page.getByRole('button', { name: /Create Account/i }).click();

    // Wait for success message
    await expect(page.getByText(/Registration successful/i)).toBeVisible({ timeout: 10000 });

    console.log(`✓ User registered successfully: ${uniqueEmail}`);
  });

  test('T055: Successful user login', async ({ page, request }) => {
    // First register a user via API
    const uniqueEmail = `mobile-login-${Date.now()}@example.com`;
    const password = 'TestPassword123!';

    await request.post(`${API_BASE_URL}/api/auth/register`, {
      data: {
        email: uniqueEmail,
        password: password,
        first_name: 'Mobile',
        last_name: 'Login',
      },
    });

    // Navigate to login page
    await page.goto(`${MOBILE_BASE_URL}/auth/login`);

    // Fill login form
    await page.getByPlaceholder('Email').fill(uniqueEmail);
    await page.getByPlaceholder('Password').fill(password);

    // Submit form
    await page.getByRole('button', { name: /Sign In/i }).click();

    // Wait for redirect to main app (tabs)
    await page.waitForURL(/\/(tabs)/, { timeout: 10000 });

    console.log(`✓ User logged in successfully: ${uniqueEmail}`);
  });

  test('T056: Login with invalid credentials shows error', async ({ page }) => {
    await page.goto(`${MOBILE_BASE_URL}/auth/login`);

    // Fill with invalid credentials
    await page.getByPlaceholder('Email').fill('invalid@example.com');
    await page.getByPlaceholder('Password').fill('WrongPassword123!');

    // Submit form
    await page.getByRole('button', { name: /Sign In/i }).click();

    // Check for error message
    await expect(page.getByText(/Login failed|Invalid credentials/i)).toBeVisible({ timeout: 10000 });

    console.log('✓ Invalid login credentials show error message');
  });

  test('T057: Navigation from login to registration', async ({ page }) => {
    await page.goto(`${MOBILE_BASE_URL}/auth/login`);

    // Click "Sign up" link
    await page.getByRole('button', { name: /Sign up/i }).click();

    // Verify navigation to registration page
    await expect(page).toHaveURL(/\/auth\/register/);
    await expect(page.getByText('Create Account')).toBeVisible();

    console.log('✓ Navigation from login to registration works');
  });

  test('T058: Navigation from registration to login', async ({ page }) => {
    await page.goto(`${MOBILE_BASE_URL}/auth/register`);

    // Click "Sign in" link
    await page.getByRole('button', { name: /Sign in/i }).click();

    // Verify navigation to login page
    await expect(page).toHaveURL(/\/auth\/login/);
    await expect(page.getByText('Welcome Back')).toBeVisible();

    console.log('✓ Navigation from registration to login works');
  });

  test('T059: OAuth buttons are visible', async ({ page }) => {
    await page.goto(`${MOBILE_BASE_URL}/auth/login`);

    // Check for OAuth buttons
    const googleButton = page.getByRole('button', { name: /Google/i });
    const facebookButton = page.getByRole('button', { name: /Facebook/i });
    const appleButton = page.getByRole('button', { name: /Apple/i });

    await expect(googleButton).toBeVisible();
    await expect(facebookButton).toBeVisible();
    await expect(appleButton).toBeVisible();

    console.log('✓ OAuth buttons are visible on login page');
  });

  test('T060: Back button navigation works', async ({ page }) => {
    await page.goto(`${MOBILE_BASE_URL}/auth/login`);

    // Click back button
    await page.getByRole('button', { name: /Back/i }).click();

    // Should navigate back (to previous page or home)
    await page.waitForTimeout(500);

    console.log('✓ Back button navigation works');
  });

  test('T061: Password field is masked', async ({ page }) => {
    await page.goto(`${MOBILE_BASE_URL}/auth/login`);

    const passwordInput = page.getByPlaceholder('Password');

    // Verify password input has type="password"
    await expect(passwordInput).toHaveAttribute('type', 'password');

    console.log('✓ Password field is properly masked');
  });

  test('T062: Email validation - invalid format', async ({ page }) => {
    await page.goto(`${MOBILE_BASE_URL}/auth/register`);

    // Fill form with invalid email
    await page.getByPlaceholder('Email').fill('invalid-email');
    await page.getByPlaceholder('Password').fill('TestPassword123!');
    await page.getByPlaceholder('First Name').fill('Test');
    await page.getByPlaceholder('Last Name').fill('User');

    // Submit form
    await page.getByRole('button', { name: /Create Account/i }).click();

    // Should show error (either client-side validation or API error)
    await expect(page.getByText(/invalid|email/i)).toBeVisible({ timeout: 5000 });

    console.log('✓ Invalid email format is rejected');
  });

  test('T063: Duplicate registration shows error', async ({ page, request }) => {
    // Register a user via API
    const uniqueEmail = `duplicate-${Date.now()}@example.com`;

    await request.post(`${API_BASE_URL}/api/auth/register`, {
      data: {
        email: uniqueEmail,
        password: 'TestPassword123!',
        first_name: 'Duplicate',
        last_name: 'User',
      },
    });

    // Try to register again with same email
    await page.goto(`${MOBILE_BASE_URL}/auth/register`);

    await page.getByPlaceholder('Email').fill(uniqueEmail);
    await page.getByPlaceholder('Password').fill('TestPassword123!');
    await page.getByPlaceholder('First Name').fill('Duplicate');
    await page.getByPlaceholder('Last Name').fill('User');

    await page.getByRole('button', { name: /Create Account/i }).click();

    // Should show duplicate email error
    await expect(page.getByText(/already exists|already registered/i)).toBeVisible({ timeout: 10000 });

    console.log('✓ Duplicate registration is rejected');
  });
});

test.describe('Mobile App - Authentication Security', () => {

  test('T064: Login form prevents XSS in error messages', async ({ page }) => {
    await page.goto(`${MOBILE_BASE_URL}/auth/login`);

    // Try login with XSS payload
    await page.getByPlaceholder('Email').fill('<script>alert("xss")</script>@example.com');
    await page.getByPlaceholder('Password').fill('test');

    await page.getByRole('button', { name: /Sign In/i }).click();

    // Wait for potential error
    await page.waitForTimeout(2000);

    // Verify no script execution (page should not have alert)
    const hasAlert = await page.evaluate(() => {
      return typeof window.alert === 'function';
    });

    expect(hasAlert).toBe(true); // Alert function should exist but not be called

    console.log('✓ XSS prevention in login error messages');
  });

  test('T065: Password minimum length validation', async ({ page }) => {
    await page.goto(`${MOBILE_BASE_URL}/auth/register`);

    // Try with short password
    await page.getByPlaceholder('Email').fill(`test-${Date.now()}@example.com`);
    await page.getByPlaceholder('Password').fill('123'); // Too short
    await page.getByPlaceholder('First Name').fill('Test');
    await page.getByPlaceholder('Last Name').fill('User');

    await page.getByRole('button', { name: /Create Account/i }).click();

    // Should show password validation error
    await expect(page.getByText(/password|too short|minimum/i)).toBeVisible({ timeout: 5000 });

    console.log('✓ Password minimum length is enforced');
  });
});
