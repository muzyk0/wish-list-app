# CLAUDE.md - Mobile (Expo/React Native)

This file provides guidance for working with the mobile application.

## Project Overview

Expo Router-based React Native application for iOS and Android, part of the Wish List full-stack application.

**Tech Stack:**
- Expo 54 with Expo Router (file-based routing)
- React Native 0.81.5, React 19.1.0
- TanStack Query for data fetching
- React Native Paper for UI components
- Zustand for state management
- React Hook Form + Zod for form validation
- openapi-fetch for type-safe API calls
- Biome for linting and formatting

## Key Commands

```bash
pnpm start              # Start Expo development server
pnpm ios                # Start iOS simulator
pnpm android            # Start Android emulator
pnpm web                # Start web version
pnpm lint               # Run Biome linter
pnpm format             # Format code with Biome
pnpm type-check         # TypeScript type checking
pnpm test               # Run tests with Vitest
pnpm generate:api       # Regenerate API types from OpenAPI spec
```

## Directory Structure

```
mobile/
├── app/                    # Expo Router file-based routes
│   ├── _layout.tsx         # Root layout (providers, theming)
│   ├── providers.tsx       # QueryClient, PaperProvider, GlobalDialogs
│   ├── (tabs)/             # Tab navigation group
│   │   ├── _layout.tsx     # Tab bar configuration
│   │   ├── index.tsx       # Home tab
│   │   ├── lists.tsx       # Lists tab
│   │   ├── gifts.tsx       # Gifts tab
│   │   ├── reservations.tsx
│   │   └── profile.tsx
│   ├── auth/               # Authentication screens
│   │   ├── _layout.tsx
│   │   ├── login.tsx
│   │   └── register.tsx
│   ├── profile/            # Profile editing screens
│   │   ├── edit.tsx
│   │   ├── change-email.tsx
│   │   └── change-password.tsx
│   ├── lists/              # Wishlist screens
│   │   ├── create.tsx
│   │   └── [id]/           # Dynamic route for list details
│   │       ├── index.tsx
│   │       ├── edit.tsx
│   │       ├── attach-items.tsx
│   │       └── gifts/create.tsx
│   ├── gifts/create.tsx
│   └── onboarding/
├── components/
│   ├── auth/               # Auth-related components
│   │   ├── AuthLayout.tsx
│   │   ├── AuthInput.tsx
│   │   ├── AuthGradientButton.tsx
│   │   ├── AuthDivider.tsx
│   │   └── AuthFooter.tsx
│   ├── wish-list/          # Wishlist domain components
│   │   ├── GiftItemForm.tsx
│   │   ├── GiftItemDisplay.tsx
│   │   ├── ImageUpload.tsx
│   │   ├── ReservationButton.tsx
│   │   └── ReservationItem.tsx
│   ├── home/               # Home screen components
│   ├── ui/                 # Reusable UI components
│   │   ├── Badge.tsx
│   │   └── icon-symbol.tsx
│   ├── GlobalDialogs.tsx   # Global dialog renderer
│   ├── TabsLayout.tsx      # Tab screen wrapper
│   └── OAuthButton.tsx
├── hooks/
│   └── useOAuthHandler.ts  # OAuth flow hook
├── stores/
│   └── dialogStore.ts      # Zustand dialog state
├── contexts/
│   └── ThemeContext.tsx    # Theme provider
├── lib/
│   └── api/                # API client layer
│       ├── client.ts       # Base openapi-fetch client
│       ├── auth.ts         # Auth operations (uses baseClient)
│       ├── api.ts          # ApiClient with middleware
│       ├── schema.ts       # Auto-generated OpenAPI types
│       └── types.ts        # Additional TypeScript types
└── theme/                  # Theme configuration
```

## Routing Patterns (Expo Router)

### Navigation Methods

```typescript
import { Link, router, useLocalSearchParams } from 'expo-router';

// Declarative with Link
<Link href="/lists/123">View List</Link>

// Declarative with typed params
<Link href={{ pathname: '/lists/[id]', params: { id: '123' } }}>
  View List
</Link>

// Imperative navigation
router.push('/lists/123');
router.replace('/(tabs)');  // No history
router.back();

// Access route params
const { id } = useLocalSearchParams<{ id: string }>();
```

### Dynamic Routes

- `[id]` folders create dynamic segments: `lists/[id]/index.tsx`
- Use `useLocalSearchParams()` for type-safe parameter access

### Layout Files

- `_layout.tsx` defines navigation structure for a route group
- `(tabs)` group uses parentheses to hide from URL path
- Root `app/_layout.tsx` sets up providers

## State Management

### Zustand (Global State)

```typescript
// stores/dialogStore.ts
import { dialog } from '@/stores/dialogStore';

// Usage anywhere in app
dialog.success('Operation completed!');
dialog.error('Something went wrong');
dialog.confirm({
  title: 'Confirm',
  message: 'Are you sure?',
  onConfirm: () => doSomething(),
});
dialog.confirmDelete('this item', () => deleteItem());
```

### React Context (Theme)

```typescript
import { useThemeContext } from '@/contexts/ThemeContext';

const { isDark, toggleTheme, theme } = useThemeContext();
```

## API Client Architecture

**Two Clients Pattern** (prevents infinite recursion):

| Client | Location | Middleware | Purpose |
|--------|----------|------------|---------|
| `baseClient` | lib/api/client.ts | No | Auth operations |
| `apiClient` | lib/api/api.ts | Yes | Protected endpoints |

```typescript
// Auth operations (no middleware)
import { loginUser, registerUser } from '@/lib/api';

// Protected endpoints (automatic auth + token refresh)
import { apiClient } from '@/lib/api';
const profile = await apiClient.getProfile();
const lists = await apiClient.getWishLists();
```

**Automatic Token Refresh**: Middleware handles 401 errors transparently.

## Form Handling Pattern

```typescript
import { zodResolver } from '@hookform/resolvers/zod';
import { Controller, useForm } from 'react-hook-form';
import { z } from 'zod';

const schema = z.object({
  email: z.string().email(),
  password: z.string().min(6),
});

type FormData = z.infer<typeof schema>;

const { control, handleSubmit, formState: { errors } } = useForm<FormData>({
  resolver: zodResolver(schema),
  defaultValues: { email: '', password: '' },
});

// In JSX
<Controller
  control={control}
  name="email"
  render={({ field: { onChange, value } }) => (
    <TextInput value={value} onChangeText={onChange} />
  )}
/>

// Submit
<Button onPress={handleSubmit((data) => mutation.mutate(data))}>
  Submit
</Button>
```

## UI Patterns

### Styling

- React Native Paper components with custom theming
- LinearGradient for decorative elements
- BlurView for glass-morphism effects
- Dark theme primary colors: `#FFD700` (gold), `#2d1b4e` (purple)

### Screen Layout

```typescript
// Standard tab screen wrapper
<TabsLayout title="Lists" subtitle="Your wishlists" refreshing={refreshing} onRefresh={onRefresh}>
  {/* Content */}
</TabsLayout>
```

### Dialogs

Use the global `dialog` helper - never use `Alert.alert`:

```typescript
import { dialog } from '@/stores/dialogStore';

// All Alert.alert calls replaced with:
dialog.success('Success message');
dialog.error('Error message');
dialog.confirm({ title, message, onConfirm });
dialog.confirmDelete('item name', onDelete);
dialog.comingSoon();
```

## Authentication

### Token Storage

- **Native**: Expo SecureStore (iOS Keychain, Android Keystore)
- Access token: 15 minutes
- Refresh token: 7 days

### OAuth Flow

```typescript
import { useOAuthHandler } from '@/hooks/useOAuthHandler';

const { oauthLoading, handleOAuth } = useOAuthHandler();

// Usage
<OAuthButtonGroup
  onGooglePress={() => handleOAuth('google')}
  onApplePress={() => handleOAuth('apple')}
  loadingProvider={oauthLoading}
/>
```

### Deep Links

- Custom scheme: `wishlistapp://`
- Universal Links (iOS): `applinks:lk.domain.com`
- App Links (Android): `https://lk.domain.com`

## Conventional Commits

```
<type>[optional scope]: <description>

feat(auth): add biometric login
fix(lists): resolve item duplication
refactor(api): simplify token refresh logic
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`

## Code Style

- **Formatter**: Biome (2 spaces, single quotes)
- **Path alias**: `@/*` maps to project root
- **TypeScript**: Strict mode enabled

Always run `pnpm format` after making changes.

## Cross-Domain Architecture

The mobile app is part of a cross-domain deployment:

| Component | Provider | Domain |
|-----------|----------|--------|
| Frontend | Vercel | wishlist.com |
| Mobile | Vercel/App Stores | lk.domain.com |
| Backend | Render | api.domain.com |

**Auth Handoff**: Frontend → Mobile uses OAuth-style code exchange for token transfer.

## Common Tasks

### Add New Screen

1. Create file in `app/` following routing structure
2. Export default component
3. Use appropriate layout wrapper (`TabsLayout`, `AuthLayout`)

### Add New API Endpoint

1. Update OpenAPI spec in `/api/openapi3.yaml`
2. Run `pnpm generate:api`
3. Add method to `apiClient` in `lib/api/api.ts`

### Add New Component

1. Create in appropriate `components/` subfolder
2. Follow existing styling patterns
3. Use `dialog` for alerts, not `Alert.alert`

## Important Notes

- Never use `Alert.alert` - use `dialog` from `stores/dialogStore`
- API client handles token refresh automatically
- File-based routing: folder structure = URL structure
- Use `expo-secure-store` for sensitive data, not `AsyncStorage`
- Run `pnpm format` before committing
