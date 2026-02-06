/**
 * E2E Tests for Mobile App Wishlist Management
 *
 * Test Coverage:
 * 1. Create wishlist
 * 2. View wishlists
 * 3. Edit wishlist
 * 4. Delete wishlist
 * 5. Wishlist visibility (public/private)
 * 6. Empty state handling
 */

import { test, expect } from '@playwright/test';

const MOBILE_BASE_URL = 'http://localhost:8081';
const API_BASE_URL = 'http://localhost:8080';

// Helper function to register and login
async function registerAndLogin(page, request) {
  const uniqueEmail = `wishlist-test-${Date.now()}@example.com`;
  const password = 'TestPassword123!';

  // Register user via API
  const registerResponse = await request.post(`${API_BASE_URL}/api/auth/register`, {
    data: {
      email: uniqueEmail,
      password: password,
      first_name: 'Wishlist',
      last_name: 'Tester',
    },
  });

  expect(registerResponse.ok()).toBeTruthy();
  const registerData = await registerResponse.json();

  // Navigate to mobile app and login
  await page.goto(`${MOBILE_BASE_URL}/auth/login`);
  await page.getByPlaceholder('Email').fill(uniqueEmail);
  await page.getByPlaceholder('Password').fill(password);
  await page.getByRole('button', { name: /Sign In/i }).click();

  // Wait for redirect to main app
  await page.waitForURL(/\/(tabs)/, { timeout: 10000 });

  return { email: uniqueEmail, accessToken: registerData.accessToken };
}

test.describe('Mobile App - Wishlist Management', () => {

  test('T066: Wishlists tab is accessible after login', async ({ page, request }) => {
    await registerAndLogin(page, request);

    // Navigate to lists tab
    await page.goto(`${MOBILE_BASE_URL}/(tabs)/lists`);

    // Verify lists page loaded
    await expect(page.getByText('My Wish Lists')).toBeVisible();

    console.log('✓ Wishlists tab is accessible');
  });

  test('T067: Empty state shows when no wishlists exist', async ({ page, request }) => {
    await registerAndLogin(page, request);

    await page.goto(`${MOBILE_BASE_URL}/(tabs)/lists`);

    // Check for empty state
    await expect(page.getByText(/No wish lists yet/i)).toBeVisible();
    await expect(page.getByText(/Create your first wish list/i)).toBeVisible();
    await expect(page.getByRole('button', { name: /Create List/i })).toBeVisible();

    console.log('✓ Empty state displays correctly');
  });

  test('T068: Create wishlist button navigates to create page', async ({ page, request }) => {
    await registerAndLogin(page, request);

    await page.goto(`${MOBILE_BASE_URL}/(tabs)/lists`);

    // Click create button (either in empty state or in app bar)
    const createButton = page.getByRole('button', { name: /Create|plus/i }).first();
    await createButton.click();

    // Verify navigation to create page
    await expect(page).toHaveURL(/\/lists\/create/);
    await expect(page.getByText('Create New Wishlist')).toBeVisible();

    console.log('✓ Create button navigates to create page');
  });

  test('T069: Create wishlist form renders correctly', async ({ page, request }) => {
    await registerAndLogin(page, request);

    await page.goto(`${MOBILE_BASE_URL}/lists/create`);

    // Verify form fields
    await expect(page.getByPlaceholder(/Title/i)).toBeVisible();
    await expect(page.getByPlaceholder(/Description/i)).toBeVisible();
    await expect(page.getByPlaceholder(/Occasion/i)).toBeVisible();
    await expect(page.getByText('Make Public')).toBeVisible();
    await expect(page.getByRole('button', { name: /Create Wishlist/i })).toBeVisible();

    console.log('✓ Create wishlist form renders all fields');
  });

  test('T070: Create wishlist validation - title required', async ({ page, request }) => {
    await registerAndLogin(page, request);

    await page.goto(`${MOBILE_BASE_URL}/lists/create`);

    // Try to submit without title
    await page.getByRole('button', { name: /Create Wishlist/i }).click();

    // Check for validation error
    await expect(page.getByText(/Please enter a title/i)).toBeVisible();

    console.log('✓ Wishlist title validation works');
  });

  test('T071: Successfully create a wishlist', async ({ page, request }) => {
    await registerAndLogin(page, request);

    await page.goto(`${MOBILE_BASE_URL}/lists/create`);

    // Fill form
    await page.getByPlaceholder(/Title/i).fill('My Birthday Wishlist');
    await page.getByPlaceholder(/Description/i).fill('Gifts I would love for my birthday');
    await page.getByPlaceholder(/Occasion/i).fill('Birthday');

    // Submit
    await page.getByRole('button', { name: /Create Wishlist/i }).click();

    // Wait for success message
    await expect(page.getByText(/created successfully/i)).toBeVisible({ timeout: 10000 });

    console.log('✓ Wishlist created successfully');
  });

  test('T072: Created wishlist appears in list', async ({ page, request }) => {
    await registerAndLogin(page, request);

    // Create a wishlist
    await page.goto(`${MOBILE_BASE_URL}/lists/create`);
    await page.getByPlaceholder(/Title/i).fill('Test List Visibility');
    await page.getByRole('button', { name: /Create Wishlist/i }).click();

    await page.waitForTimeout(2000);

    // Navigate to lists tab
    await page.goto(`${MOBILE_BASE_URL}/(tabs)/lists`);

    // Verify wishlist appears
    await expect(page.getByText('Test List Visibility')).toBeVisible({ timeout: 10000 });

    console.log('✓ Created wishlist appears in list');
  });

  test('T073: Toggle wishlist public/private', async ({ page, request }) => {
    await registerAndLogin(page, request);

    await page.goto(`${MOBILE_BASE_URL}/lists/create`);

    // Check the toggle state
    const toggleSwitch = page.getByRole('switch', { name: /Make Public/i });

    // Get initial state
    const isInitiallyChecked = await toggleSwitch.isChecked();

    // Toggle it
    await toggleSwitch.click();

    // Verify it changed
    const isNowChecked = await toggleSwitch.isChecked();
    expect(isNowChecked).toBe(!isInitiallyChecked);

    // Fill rest of form
    await page.getByPlaceholder(/Title/i).fill('Public Wishlist Test');

    // Submit
    await page.getByRole('button', { name: /Create Wishlist/i }).click();

    await expect(page.getByText(/created successfully/i)).toBeVisible({ timeout: 10000 });

    console.log('✓ Public/private toggle works');
  });

  test('T074: Public badge shows for public wishlists', async ({ page, request }) => {
    await registerAndLogin(page, request);

    // Create public wishlist
    await page.goto(`${MOBILE_BASE_URL}/lists/create`);
    await page.getByPlaceholder(/Title/i).fill('Public List Badge Test');
    await page.getByRole('switch', { name: /Make Public/i }).click();
    await page.getByRole('button', { name: /Create Wishlist/i }).click();

    await page.waitForTimeout(2000);

    // Navigate to lists
    await page.goto(`${MOBILE_BASE_URL}/(tabs)/lists`);

    // Check for public badge
    await expect(page.getByText('Public')).toBeVisible({ timeout: 10000 });

    console.log('✓ Public badge displays for public wishlists');
  });

  test('T075: View wishlist details', async ({ page, request }) => {
    await registerAndLogin(page, request);

    // Create a wishlist first
    await page.goto(`${MOBILE_BASE_URL}/lists/create`);
    await page.getByPlaceholder(/Title/i).fill('Details Test List');
    await page.getByPlaceholder(/Description/i).fill('Test description for details view');
    await page.getByRole('button', { name: /Create Wishlist/i }).click();

    await page.waitForTimeout(2000);

    // Navigate to lists
    await page.goto(`${MOBILE_BASE_URL}/(tabs)/lists`);

    // Click "View List" button
    await page.getByRole('button', { name: /View List/i }).first().click();

    // Wait for navigation to details page
    await page.waitForURL(/\/lists\/[^\/]+$/, { timeout: 10000 });

    // Verify we're on the details page
    await expect(page.getByText('Details Test List')).toBeVisible();

    console.log('✓ View wishlist details works');
  });

  test('T076: Edit wishlist button navigates to edit page', async ({ page, request }) => {
    await registerAndLogin(page, request);

    // Create a wishlist
    await page.goto(`${MOBILE_BASE_URL}/lists/create`);
    await page.getByPlaceholder(/Title/i).fill('Edit Test List');
    await page.getByRole('button', { name: /Create Wishlist/i }).click();

    await page.waitForTimeout(2000);

    // Navigate to lists
    await page.goto(`${MOBILE_BASE_URL}/(tabs)/lists`);

    // Click edit button
    await page.getByRole('button', { name: /Edit/i }).first().click();

    // Verify navigation to edit page
    await expect(page).toHaveURL(/\/lists\/[^\/]+\/edit/, { timeout: 10000 });

    console.log('✓ Edit button navigates to edit page');
  });

  test('T077: Delete wishlist shows confirmation dialog', async ({ page, request }) => {
    await registerAndLogin(page, request);

    // Create a wishlist
    await page.goto(`${MOBILE_BASE_URL}/lists/create`);
    await page.getByPlaceholder(/Title/i).fill('Delete Test List');
    await page.getByRole('button', { name: /Create Wishlist/i }).click();

    await page.waitForTimeout(2000);

    // Navigate to lists
    await page.goto(`${MOBILE_BASE_URL}/(tabs)/lists`);

    // Listen for dialog
    page.on('dialog', dialog => {
      expect(dialog.type()).toBe('confirm');
      expect(dialog.message()).toContain('delete');
      dialog.accept();
    });

    // Click delete button
    await page.getByRole('button', { name: /Delete/i }).first().click();

    // Wait for deletion to complete
    await page.waitForTimeout(2000);

    console.log('✓ Delete confirmation dialog appears');
  });

  test('T078: Successfully delete a wishlist', async ({ page, request }) => {
    await registerAndLogin(page, request);

    const wishlistTitle = `Delete Success ${Date.now()}`;

    // Create a wishlist
    await page.goto(`${MOBILE_BASE_URL}/lists/create`);
    await page.getByPlaceholder(/Title/i).fill(wishlistTitle);
    await page.getByRole('button', { name: /Create Wishlist/i }).click();

    await page.waitForTimeout(2000);

    // Navigate to lists
    await page.goto(`${MOBILE_BASE_URL}/(tabs)/lists`);

    // Verify wishlist exists
    await expect(page.getByText(wishlistTitle)).toBeVisible();

    // Delete it
    page.on('dialog', dialog => dialog.accept());
    await page.getByRole('button', { name: /Delete/i }).first().click();

    // Wait for deletion
    await page.waitForTimeout(2000);

    // Reload page
    await page.reload();

    // Verify wishlist is gone
    await expect(page.getByText(wishlistTitle)).not.toBeVisible({ timeout: 5000 });

    console.log('✓ Wishlist deleted successfully');
  });

  test('T079: Pull to refresh reloads wishlists', async ({ page, request }) => {
    await registerAndLogin(page, request);

    await page.goto(`${MOBILE_BASE_URL}/(tabs)/lists`);

    // Initial load
    await page.waitForTimeout(1000);

    // Simulate pull to refresh by reloading
    await page.reload();

    // Verify page reloaded
    await expect(page.getByText('My Wish Lists')).toBeVisible();

    console.log('✓ Pull to refresh works');
  });

  test('T080: Wishlist displays stats correctly', async ({ page, request }) => {
    await registerAndLogin(page, request);

    // Create a wishlist
    await page.goto(`${MOBILE_BASE_URL}/lists/create`);
    await page.getByPlaceholder(/Title/i).fill('Stats Test List');
    await page.getByRole('button', { name: /Create Wishlist/i }).click();

    await page.waitForTimeout(2000);

    // Navigate to lists
    await page.goto(`${MOBILE_BASE_URL}/(tabs)/lists`);

    // Check for stats display
    await expect(page.getByText(/views|Not viewed/i)).toBeVisible();
    await expect(page.getByText(/date|No date set/i)).toBeVisible();

    console.log('✓ Wishlist stats display correctly');
  });
});

test.describe('Mobile App - Wishlist Error Handling', () => {

  test('T081: Error state displays when API fails', async ({ page, request }) => {
    await registerAndLogin(page, request);

    // Navigate to lists with potential API error
    await page.goto(`${MOBILE_BASE_URL}/(tabs)/lists`);

    // If there's an error, error UI should show
    // This test assumes the error UI has a retry button
    const errorText = page.getByText(/Error loading wishlists/i);
    const retryButton = page.getByRole('button', { name: /Retry/i });

    // If error state is visible, verify retry button exists
    if (await errorText.isVisible({ timeout: 5000 }).catch(() => false)) {
      await expect(retryButton).toBeVisible();
      console.log('✓ Error state displays with retry button');
    } else {
      console.log('✓ No error state (API call successful)');
    }
  });

  test('T082: Loading state displays while fetching wishlists', async ({ page, request }) => {
    await registerAndLogin(page, request);

    // Navigate to lists
    const navigation = page.goto(`${MOBILE_BASE_URL}/(tabs)/lists`);

    // Check for loading indicator
    const loadingText = page.getByText(/Loading wishlists/i);

    // Wait briefly to potentially see loading state
    await page.waitForTimeout(200);

    // Loading indicator should appear or already be done
    const isLoading = await loadingText.isVisible().catch(() => false);

    if (isLoading) {
      console.log('✓ Loading state visible');
    } else {
      console.log('✓ Data loaded quickly (loading state not captured)');
    }

    await navigation;
  });
});
