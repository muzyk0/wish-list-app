# Quickstart: Public Wishlist with Guest Reservation

**Feature**: 005-public-wishlist-reservation
**Date**: 2026-02-20

## Prerequisites

- Node.js 20+
- pnpm 9+
- Backend server running at `http://localhost:8080` (or set `NEXT_PUBLIC_API_URL`)
- PostgreSQL database with migrations applied (at least through `000005`)

## Setup

```bash
# 1. Switch to feature branch
git checkout 005-public-wishlist-reservation

# 2. Install frontend dependencies
cd frontend && pnpm install

# 3. Start database (if not running)
cd ../database && docker compose up -d

# 4. Start backend (if not running)
cd ../backend && make run

# 5. Start frontend dev server
cd ../frontend && pnpm dev
```

## Test the Feature

### View a Public Wishlist

1. Ensure at least one public wishlist exists in the database (create via mobile app or seed data)
2. Navigate to `http://localhost:3000/public/{slug}` where `{slug}` is the wishlist's `public_slug`
3. Verify: wishlist title, description, occasion, and all gift items display correctly
4. Toggle language via the language switcher (RU/EN)

### Reserve a Gift

1. On a public wishlist page, find an available (unreserved) gift item
2. Click "Reserve Gift" button
3. Fill in name and email in the dialog
4. Submit and verify success toast
5. Verify the item now shows "Reserved" badge
6. Check `localStorage` → `guest_reservations` for stored token

### View My Reservations

1. Navigate to `http://localhost:3000/my/reservations`
2. Verify reserved items appear with correct details
3. Alternative: click "My Reservations" link in the footer

### Cancel a Reservation

1. On the My Reservations page, find an active reservation
2. Click "Cancel" button
3. Confirm cancellation
4. Verify the item returns to "available" on the public wishlist page

## Development Commands

```bash
# Run frontend dev server
cd frontend && pnpm dev

# Run tests
cd frontend && pnpm test

# Run linting
cd frontend && pnpm lint

# Format code
cd frontend && pnpm format

# Type check
cd frontend && pnpm type-check

# Run Storybook
cd frontend && pnpm storybook

# Regenerate API types (if OpenAPI spec changes)
cd frontend && pnpm generate:api
```

## Key Files

| File | Purpose |
|------|---------|
| `src/app/public/[slug]/page.tsx` | Public wishlist page (entry point) |
| `src/components/public-wishlist/` | Independent-styled wishlist components |
| `src/components/guest/GuestReservationDialog.tsx` | Guest reservation form dialog |
| `src/components/wish-list/MyReservations.tsx` | My Reservations component |
| `src/app/my/reservations/page.tsx` | My Reservations page |
| `src/lib/api/client.ts` | API client methods |
| `src/lib/api/types.ts` | TypeScript type definitions |
| `src/i18n/locales/en.json` | English translations |
| `src/i18n/locales/ru.json` | Russian translations |
| `src/widgets/Header.tsx` | Site header (navigation links) |

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `NEXT_PUBLIC_API_URL` | `http://localhost:8080/api` | Backend API base URL |
| `NEXT_PUBLIC_MOBILE_APP_DOMAIN` | `lk.domain.com` | Mobile app domain for redirects |
