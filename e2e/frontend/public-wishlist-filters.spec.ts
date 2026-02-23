import { randomUUID } from 'crypto';
import { expect, test } from '@playwright/test';

const API_BASE = 'http://localhost:8080/api';
const FRONTEND_BASE = 'http://localhost:3000';

interface PublicFixture {
  slug: string;
}

let fixture: PublicFixture;

test.describe('Frontend Public Wishlist Filters', () => {
  test.beforeAll(async ({ request }) => {
    const uniqueId = randomUUID();
    const email = `filters_${uniqueId}@example.com`;

    const registerResponse = await request.post(`${API_BASE}/auth/register`, {
      data: {
        email,
        password: 'Test123456!',
        first_name: 'Filters',
        last_name: 'Tester',
      },
    });
    expect(registerResponse.status()).toBe(201);
    const registerData = await registerResponse.json();
    const token = registerData.accessToken as string;

    const wishlistResponse = await request.post(`${API_BASE}/wishlists`, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
      data: {
        title: `Filters Wishlist ${uniqueId}`,
        description: 'Search, filters and sorting test',
        is_public: true,
      },
    });
    expect(wishlistResponse.status()).toBe(201);
    const wishlistData = await wishlistResponse.json();

    const createItem = async (title: string, price: number, priority: number) => {
      const response = await request.post(
        `${API_BASE}/wishlists/${wishlistData.id}/items/new`,
        {
          headers: {
            Authorization: `Bearer ${token}`,
          },
          data: {
            title,
            description: `${title} description`,
            price,
            priority,
            notes: '',
          },
        },
      );

      expect(response.status()).toBe(201);
      return response.json();
    };

    const appleItem = await createItem('Apple Watch', 300, 9);
    const bookItem = await createItem('Book Reader', 50, 2);
    const cameraItem = await createItem('Camera Lens', 200, 7);

    const reserveResponse = await request.post(
      `${API_BASE}/public/reservations/wishlist/${wishlistData.id}/item/${bookItem.id}`,
      {
        data: {
          guest_name: 'Filter Guest',
        },
      },
    );
    expect(reserveResponse.status()).toBe(200);

    const markPurchasedResponse = await request.post(
      `${API_BASE}/items/${cameraItem.id}/mark-purchased`,
      {
        headers: {
          Authorization: `Bearer ${token}`,
        },
        data: {
          purchased_price: 180,
        },
      },
    );
    expect(markPurchasedResponse.status()).toBe(200);

    fixture = {
      slug: wishlistData.public_slug as string,
    };
  });

  test('searches by gift name', async ({ page }) => {
    await page.goto(`${FRONTEND_BASE}/public/${fixture.slug}`);

    await expect(page.locator('article.wl-item-card')).toHaveCount(3);

    await page
      .locator('[data-testid="wishlist-search-input"]')
      .fill('Camera Lens');

    await expect(page.locator('article.wl-item-card')).toHaveCount(1);
    await expect(
      page.getByRole('heading', { name: 'Camera Lens', exact: true }),
    ).toBeVisible();
  });

  test('filters by reserved status', async ({ page }) => {
    await page.goto(`${FRONTEND_BASE}/public/${fixture.slug}`);

    await page
      .locator('[data-testid="wishlist-status-filter"]')
      .selectOption('reserved');

    await expect(page.locator('article.wl-item-card')).toHaveCount(1);
    await expect(
      page.getByRole('heading', { name: 'Book Reader', exact: true }),
    ).toBeVisible();
    await expect(
      page.getByRole('heading', { name: 'Camera Lens', exact: true }),
    ).toHaveCount(0);
  });

  test('sorts by price descending', async ({ page }) => {
    await page.goto(`${FRONTEND_BASE}/public/${fixture.slug}`);

    await page.locator('[data-testid="wishlist-sort"]').selectOption('price_desc');

    const firstTitle = await page
      .locator('article.wl-item-card h3')
      .first()
      .textContent();
    expect(firstTitle?.trim()).toBe('Apple Watch');
  });
});
