import { randomUUID } from 'crypto';
import { expect, test } from '@playwright/test';
import { API_BASE } from '../shared/test-helpers';

const FRONTEND_BASE = 'http://localhost:3000';

interface PublicFixture {
  slug: string;
  wishlistId: string;
  itemId: string;
}

let fixture: PublicFixture;

test.describe('Frontend Public Wishlist Responsive', () => {
  test.beforeAll(async ({ request }) => {
    const uniqueId = randomUUID();
    const email = `responsive_${uniqueId}@example.com`;

    const registerResponse = await request.post(`${API_BASE}/auth/register`, {
      data: {
        email,
        password: 'Test123456!',
        first_name: 'Responsive',
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
        title: `Responsive Wishlist ${uniqueId}`,
        description: 'Viewport coverage test',
        is_public: true,
      },
    });
    expect(wishlistResponse.status()).toBe(201);
    const wishlistData = await wishlistResponse.json();

    const itemResponse = await request.post(
      `${API_BASE}/wishlists/${wishlistData.id}/items/new`,
      {
        headers: {
          Authorization: `Bearer ${token}`,
        },
        data: {
          title: 'A very long responsive test gift item name to exercise wrapping',
          description:
            'Long description used by e2e to verify layout does not overflow on mobile widths.',
          price: 123.45,
          priority: 5,
          notes: '',
        },
      },
    );
    expect(itemResponse.status()).toBe(201);
    const itemData = await itemResponse.json();

    const reserveResponse = await request.post(
      `${API_BASE}/public/reservations/wishlist/${wishlistData.id}/item/${itemData.id}`,
      {
        data: {
          guest_name: 'Viewport Guest',
        },
      },
    );
    expect(reserveResponse.status()).toBe(200);

    fixture = {
      slug: wishlistData.public_slug as string,
      wishlistId: wishlistData.id as string,
      itemId: itemData.id as string,
    };
  });

  const viewports = [
    { width: 320, height: 740 },
    { width: 375, height: 812 },
    { width: 390, height: 844 },
    { width: 768, height: 1024 },
    { width: 1024, height: 768 },
  ];

  for (const viewport of viewports) {
    test(`public wishlist remains responsive at ${viewport.width}px`, async ({
      page,
    }) => {
      await page.setViewportSize(viewport);
      await page.goto(`${FRONTEND_BASE}/public/${fixture.slug}`);

      const card = page.locator('article.wl-item-card').first();
      await expect(card).toBeVisible();

      // Ensure no horizontal overflow at this viewport width.
      const hasHorizontalOverflow = await page.evaluate(() => {
        const doc = document.documentElement;
        return doc.scrollWidth > doc.clientWidth + 1;
      });
      expect(hasHorizontalOverflow).toBeFalsy();

      // Reserved item should not be reservable again.
      const reserveButton = card.getByRole('button', { name: /reserved/i });
      await expect(reserveButton).toBeDisabled();

      // Public page must not reveal who reserved the gift.
      await expect(page.getByText(/reserved by/i)).toHaveCount(0);
    });
  }
});
