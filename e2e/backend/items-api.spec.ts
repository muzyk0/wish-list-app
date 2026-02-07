/**
 * E2E Tests for New Items API Architecture
 * Tests independent items endpoints and many-to-many relationships
 */

import { test, expect } from '@playwright/test';
import { randomUUID } from 'crypto';

const API_BASE = 'http://localhost:8080/api';

let authToken: string;
let userId: string;
let wishlistId: string;
let itemId: string;

// Set up authentication before all tests in this worker
test.beforeAll(async ({ request, browserName }) => {
  // Generate unique ID per browser worker to avoid conflicts
  const uniqueId = randomUUID();

  // Test user credentials - unique per worker
  const testUser = {
    email: `test_${uniqueId}@example.com`,
    password: 'Test123456!',
    first_name: 'Test',
    last_name: 'User',
  };

  // Register a new user for this test worker
  const registerResponse = await request.post(`${API_BASE}/auth/register`, {
    data: testUser,
  });

  expect(registerResponse.status()).toBe(201);
  const registerData = await registerResponse.json();

  authToken = registerData.accessToken;
  userId = registerData.user.id;

  // Create a wishlist for testing (unique per browser to avoid slug conflicts)
  const wishlistResponse = await request.post(`${API_BASE}/wishlists`, {
    headers: {
      Authorization: `Bearer ${authToken}`,
    },
    data: {
      title: `Test Wishlist ${browserName} ${uniqueId}`,
      description: 'For API testing',
      is_public: false,
    },
  });

  if (wishlistResponse.status() !== 201) {
    const errorBody = await wishlistResponse.text();
    throw new Error(`Failed to create wishlist: ${wishlistResponse.status()} - ${errorBody}`);
  }

  const wishlistData = await wishlistResponse.json();
  wishlistId = wishlistData.id;
});

test.describe('Items API - Authentication Setup', () => {
  test('should have authentication set up', async () => {
    expect(authToken).toBeDefined();
    expect(userId).toBeDefined();
    expect(wishlistId).toBeDefined();
  });
});

test.describe('Independent Items Endpoints (/api/items)', () => {
  test('POST /api/items - should create independent item', async ({ request }) => {
    const response = await request.post(`${API_BASE}/items`, {
      headers: {
        Authorization: `Bearer ${authToken}`,
      },
      data: {
        name: 'Nintendo Switch',
        description: 'OLED Model',
        link: 'https://example.com/switch',
        price: 349.99,
        priority: 1,
      },
    });

    expect(response.status()).toBe(201);
    const data = await response.json();

    itemId = data.id;
    expect(data.name).toBe('Nintendo Switch');
    expect(data.description).toBe('OLED Model');
    expect(data.price).toBe(349.99);
    expect(data.owner_id).toBe(userId);
    expect(data.archived_at).toBeNull();
  });

  test('GET /api/items - should get my items', async ({ request }) => {
    const response = await request.get(`${API_BASE}/items`, {
      headers: {
        Authorization: `Bearer ${authToken}`,
      },
    });

    expect(response.status()).toBe(200);
    const data = await response.json();

    expect(data).toHaveProperty('items');
    expect(data).toHaveProperty('pagination');
    expect(data.items.length).toBeGreaterThan(0);
    expect(data.items[0].name).toBe('Nintendo Switch');
  });

  test('GET /api/items?unattached=true - should filter unattached items', async ({ request }) => {
    const response = await request.get(`${API_BASE}/items?unattached=true`, {
      headers: {
        Authorization: `Bearer ${authToken}`,
      },
    });

    expect(response.status()).toBe(200);
    const data = await response.json();

    // Item should be in unattached list (not yet added to wishlist)
    expect(data.items.length).toBeGreaterThan(0);
  });

  test('GET /api/items/:id - should get specific item', async ({ request }) => {
    const response = await request.get(`${API_BASE}/items/${itemId}`, {
      headers: {
        Authorization: `Bearer ${authToken}`,
      },
    });

    expect(response.status()).toBe(200);
    const data = await response.json();

    expect(data.id).toBe(itemId);
    expect(data.name).toBe('Nintendo Switch');
  });

  test('PUT /api/items/:id - should update item', async ({ request }) => {
    const response = await request.put(`${API_BASE}/items/${itemId}`, {
      headers: {
        Authorization: `Bearer ${authToken}`,
      },
      data: {
        name: 'Nintendo Switch OLED',
        description: 'Updated description',
        price: 359.99,
      },
    });

    expect(response.status()).toBe(200);
    const data = await response.json();

    expect(data.name).toBe('Nintendo Switch OLED');
    expect(data.description).toBe('Updated description');
    expect(data.price).toBe(359.99);
  });
});

test.describe('Many-to-Many Relationships (/api/wishlists/:id/items)', () => {
  test('POST /api/wishlists/:id/items - should attach existing item to wishlist', async ({ request }) => {
    const response = await request.post(`${API_BASE}/wishlists/${wishlistId}/items`, {
      headers: {
        Authorization: `Bearer ${authToken}`,
      },
      data: {
        item_id: itemId,
      },
    });

    expect(response.status()).toBe(200);
    const data = await response.json();

    expect(data.message).toContain('attached');
  });

  test('GET /api/wishlists/:id/items - should get items in wishlist', async ({ request }) => {
    const response = await request.get(`${API_BASE}/wishlists/${wishlistId}/items`, {
      headers: {
        Authorization: `Bearer ${authToken}`,
      },
    });

    expect(response.status()).toBe(200);
    const data = await response.json();

    expect(data).toHaveProperty('items');
    expect(data.items.length).toBe(1);
    expect(data.items[0].id).toBe(itemId);
    expect(data.items[0].name).toBe('Nintendo Switch OLED');
  });

  test('GET /api/items?unattached=true - should not show attached item', async ({ request }) => {
    const response = await request.get(`${API_BASE}/items?unattached=true`, {
      headers: {
        Authorization: `Bearer ${authToken}`,
      },
    });

    expect(response.status()).toBe(200);
    const data = await response.json();

    // Item should NOT be in unattached list anymore
    const hasItem = data.items.some((item: any) => item.id === itemId);
    expect(hasItem).toBe(false);
  });

  test('POST /api/wishlists/:id/items/new - should create new item in wishlist', async ({ request }) => {
    const response = await request.post(`${API_BASE}/wishlists/${wishlistId}/items/new`, {
      headers: {
        Authorization: `Bearer ${authToken}`,
      },
      data: {
        name: 'PlayStation 5',
        description: 'Gaming console',
        price: 499.99,
        priority: 2,
      },
    });

    expect(response.status()).toBe(201);
    const data = await response.json();

    expect(data.name).toBe('PlayStation 5');
    expect(data.owner_id).toBe(userId);

    // Verify it was automatically attached
    const wishlistItems = await request.get(`${API_BASE}/wishlists/${wishlistId}/items`, {
      headers: {
        Authorization: `Bearer ${authToken}`,
      },
    });

    const wishlistData = await wishlistItems.json();
    expect(wishlistData.items.length).toBe(2);
  });

  test('DELETE /api/wishlists/:id/items/:itemId - should detach item from wishlist', async ({ request }) => {
    const response = await request.delete(`${API_BASE}/wishlists/${wishlistId}/items/${itemId}`, {
      headers: {
        Authorization: `Bearer ${authToken}`,
      },
    });

    expect(response.status()).toBe(200);
    const data = await response.json();

    expect(data.message).toContain('detached');

    // Verify item was detached
    const wishlistItems = await request.get(`${API_BASE}/wishlists/${wishlistId}/items`, {
      headers: {
        Authorization: `Bearer ${authToken}`,
      },
    });

    const wishlistData = await wishlistItems.json();
    expect(wishlistData.items.length).toBe(1); // Only PS5 remains
  });

  test('GET /api/items?unattached=true - should show detached item again', async ({ request }) => {
    const response = await request.get(`${API_BASE}/items?unattached=true`, {
      headers: {
        Authorization: `Bearer ${authToken}`,
      },
    });

    expect(response.status()).toBe(200);
    const data = await response.json();

    // Item should be back in unattached list
    const hasItem = data.items.some((item: any) => item.id === itemId);
    expect(hasItem).toBe(true);
  });
});

test.describe('Soft Delete Functionality', () => {
  test('DELETE /api/items/:id - should soft delete (archive) item', async ({ request }) => {
    const response = await request.delete(`${API_BASE}/items/${itemId}`, {
      headers: {
        Authorization: `Bearer ${authToken}`,
      },
    });

    expect(response.status()).toBe(204);
  });

  test('GET /api/items - should not show archived items by default', async ({ request }) => {
    const response = await request.get(`${API_BASE}/items`, {
      headers: {
        Authorization: `Bearer ${authToken}`,
      },
    });

    expect(response.status()).toBe(200);
    const data = await response.json();

    // Archived item should not be in the list
    const hasItem = data.items.some((item: any) => item.id === itemId);
    expect(hasItem).toBe(false);
  });

  test('GET /api/items?include_archived=true - should show archived items', async ({ request }) => {
    const response = await request.get(`${API_BASE}/items?include_archived=true`, {
      headers: {
        Authorization: `Bearer ${authToken}`,
      },
    });

    expect(response.status()).toBe(200);
    const data = await response.json();

    // Archived item should be in the list
    const archivedItem = data.items.find((item: any) => item.id === itemId);
    expect(archivedItem).toBeDefined();
    expect(archivedItem.archived_at).not.toBeNull();
  });

  test('GET /api/items/:id - should return 404 for archived item', async ({ request }) => {
    const response = await request.get(`${API_BASE}/items/${itemId}`, {
      headers: {
        Authorization: `Bearer ${authToken}`,
      },
    });

    // Archived items should not be accessible
    expect(response.status()).toBe(404);
  });
});

test.describe('Mark as Purchased Functionality', () => {
  let purchaseItemId: string;

  test('setup - create item for purchase test', async ({ request }) => {
    const response = await request.post(`${API_BASE}/items`, {
      headers: {
        Authorization: `Bearer ${authToken}`,
      },
      data: {
        name: 'Test Purchase Item',
        price: 99.99,
      },
    });

    const data = await response.json();
    purchaseItemId = data.id;
  });

  test('POST /api/items/:id/mark-purchased - should mark item as purchased', async ({ request }) => {
    const response = await request.post(`${API_BASE}/items/${purchaseItemId}/mark-purchased`, {
      headers: {
        Authorization: `Bearer ${authToken}`,
      },
      data: {
        purchased_price: 89.99,
      },
    });

    expect(response.status()).toBe(200);
    const data = await response.json();

    expect(data.purchased_by_user_id).toBe(userId);
    expect(data.purchased_at).not.toBeNull();
    expect(data.purchased_price).toBe(89.99);
  });
});

test.describe('Pagination and Filtering', () => {
  test('setup - create multiple items', async ({ request }) => {
    for (let i = 1; i <= 5; i++) {
      await request.post(`${API_BASE}/items`, {
        headers: {
          Authorization: `Bearer ${authToken}`,
        },
        data: {
          name: `Test Item ${i}`,
          price: i * 10,
          priority: i,
        },
      });
    }
  });

  test('GET /api/items?page=1&limit=3 - should paginate results', async ({ request }) => {
    const response = await request.get(`${API_BASE}/items?page=1&limit=3`, {
      headers: {
        Authorization: `Bearer ${authToken}`,
      },
    });

    expect(response.status()).toBe(200);
    const data = await response.json();

    expect(data.pagination.page).toBe(1);
    expect(data.pagination.limit).toBe(3);
    expect(data.items.length).toBeLessThanOrEqual(3);
    expect(data.pagination.total).toBeGreaterThanOrEqual(5);
  });

  test('GET /api/items?sort=price&order=desc - should sort by price descending', async ({ request }) => {
    const response = await request.get(`${API_BASE}/items?sort=price&order=desc`, {
      headers: {
        Authorization: `Bearer ${authToken}`,
      },
    });

    expect(response.status()).toBe(200);
    const data = await response.json();

    // Verify descending order
    for (let i = 1; i < data.items.length; i++) {
      expect(data.items[i - 1].price).toBeGreaterThanOrEqual(data.items[i].price);
    }
  });

  test('GET /api/items?search=PlayStation - should search items', async ({ request }) => {
    const response = await request.get(`${API_BASE}/items?search=PlayStation`, {
      headers: {
        Authorization: `Bearer ${authToken}`,
      },
    });

    expect(response.status()).toBe(200);
    const data = await response.json();

    // Should find the PlayStation item we created earlier
    const hasPS5 = data.items.some((item: any) => item.name.includes('PlayStation'));
    expect(hasPS5).toBe(true);
  });
});

test.describe('Public Endpoints (Guest Access)', () => {
  let publicWishlistSlug: string;

  test('setup - create public wishlist', async ({ request }) => {
    const response = await request.post(`${API_BASE}/wishlists`, {
      headers: {
        Authorization: `Bearer ${authToken}`,
      },
      data: {
        title: 'Public Wishlist',
        description: 'For public testing',
        is_public: true,
      },
    });

    const data = await response.json();
    publicWishlistSlug = data.public_slug;

    // Add an item to the public wishlist
    await request.post(`${API_BASE}/wishlists/${data.id}/items/new`, {
      headers: {
        Authorization: `Bearer ${authToken}`,
      },
      data: {
        name: 'Public Item',
        price: 29.99,
      },
    });
  });

  test('GET /api/public/wishlists/:slug - should access public wishlist without auth', async ({ request }) => {
    const response = await request.get(`${API_BASE}/public/wishlists/${publicWishlistSlug}`);

    expect(response.status()).toBe(200);
    const data = await response.json();

    expect(data.title).toBe('Public Wishlist');
    expect(data.is_public).toBe(true);
  });

  test('GET /api/public/wishlists/:slug/gift-items - should access public items without auth', async ({ request }) => {
    const response = await request.get(`${API_BASE}/public/wishlists/${publicWishlistSlug}/gift-items`);

    expect(response.status()).toBe(200);
    const data = await response.json();

    expect(data.items.length).toBeGreaterThan(0);
    expect(data.items[0].name).toBe('Public Item');
  });
});

test.describe('Error Handling', () => {
  test('GET /api/items/:id - should return 404 for non-existent item', async ({ request }) => {
    const response = await request.get(`${API_BASE}/items/00000000-0000-0000-0000-000000000000`, {
      headers: {
        Authorization: `Bearer ${authToken}`,
      },
    });

    expect(response.status()).toBe(404);
  });

  test('POST /api/items - should return 401 without auth', async ({ request }) => {
    const response = await request.post(`${API_BASE}/items`, {
      data: {
        name: 'Unauthorized Item',
      },
    });

    expect(response.status()).toBe(401);
  });

  test('POST /api/wishlists/:id/items - should return 403 for other user wishlist', async ({ request }) => {
    // Create another user
    const anotherUser = {
      email: `another_${Date.now()}@example.com`,
      password: 'Test123456!',
      first_name: 'Another',
      last_name: 'User',
    };

    const registerResponse = await request.post(`${API_BASE}/auth/register`, {
      data: anotherUser,
    });

    const { accessToken: otherToken } = await registerResponse.json();

    // Try to attach item to original user's wishlist
    const response = await request.post(`${API_BASE}/wishlists/${wishlistId}/items`, {
      headers: {
        Authorization: `Bearer ${otherToken}`,
      },
      data: {
        item_id: itemId,
      },
    });

    expect(response.status()).toBe(403);
  });
});
