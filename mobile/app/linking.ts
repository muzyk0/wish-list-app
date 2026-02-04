/**
 * Deep Linking Configuration
 *
 * This file configures deep linking support for the mobile app,
 * allowing navigation from the web frontend to specific screens
 * in the mobile app using the wishlistapp:// URL scheme.
 *
 * Supported deep links:
 * - wishlistapp://home → Home tab
 * - wishlistapp://auth/login → Login screen
 * - wishlistapp://auth/register → Registration screen
 * - wishlistapp://my/reservations → User reservations
 * - wishlistapp://lists → Lists tab
 * - wishlistapp://lists/[id] → Specific list view
 * - wishlistapp://public/[slug] → Public wishlist view
 */

export const linking = {
  prefixes: ["wishlistapp://", "https://lk.domain.com"],
  config: {
    screens: {
      "(tabs)": {
        screens: {
          index: "home",
          lists: "lists",
          explore: "explore",
          reservations: "my/reservations",
          profile: "profile",
        },
      },
      auth: {
        screens: {
          login: "auth/login",
          register: "auth/register",
        },
      },
      lists: {
        screens: {
          create: "lists/create",
          "[id]/index": "lists/:id",
          "[id]/edit": "lists/:id/edit",
        },
      },
      "gift-items": {
        screens: {
          "[id]/edit": "gift-items/:id/edit",
        },
      },
      public: {
        screens: {
          "[slug]": "public/:slug",
        },
      },
      modal: "modal",
    },
  },
};

// Deep link URL examples and their corresponding routes:
// wishlistapp://home → (tabs)/index
// wishlistapp://auth/login → auth/login
// wishlistapp://auth/register → auth/register
// wishlistapp://my/reservations → (tabs)/reservations
// wishlistapp://lists → (tabs)/lists
// wishlistapp://lists/123 → lists/[id]/index
// wishlistapp://lists/123/edit → lists/[id]/edit
// wishlistapp://gift-items/456/edit → gift-items/[id]/edit
// wishlistapp://public/birthday-2026 → public/[slug]
