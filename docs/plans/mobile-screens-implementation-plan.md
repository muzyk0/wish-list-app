# Mobile Screens Implementation Plan

> **Based on**: Design references from `/drafts/screens/`
> **Platform**: React Native / Expo
> **Date**: 2026-02-05

---

## Executive Summary

This plan outlines the implementation of essential mobile app screens based on UX best practices. The screens follow a progressive user flow from first launch to full engagement.

### Screens Implemented

| # | Screen | Status | Files |
|---|--------|--------|-------|
| 1 | Splash/Welcome Screen | ✅ Complete | `mobile/app/splash.tsx` |
| 2 | Onboarding Flow (3 slides) | ✅ Complete | `mobile/app/onboarding/` |
| 3 | Login Screen | ✅ Enhanced | `mobile/app/auth/login.tsx` |
| 4 | Home Screen Dashboard | ✅ Complete | `mobile/app/(tabs)/index.tsx` |
| 5 | User Profile | ✅ Exists | `mobile/app/(tabs)/profile.tsx` |
| 6 | Settings | ✅ In Profile | Integrated |
| 7 | Register Screen | ✅ Enhanced | `mobile/app/auth/register.tsx` |

---

## 1. Splash/Welcome Screen

### Design Requirements
- **Display Duration**: 1-3 seconds with loading indicator
- **Content**: App logo + name/slogan, centered
- **Style**: Minimal, clean, branded

### Technical Approach
```
Location: mobile/app/index.tsx (root)
Type: Initial screen before navigation
```

### Implementation Details

**File**: `mobile/app/splash.tsx`

```tsx
// Key elements:
- Centered logo with app name "Wish List"
- Animated loading indicator
- Auto-navigation after 2 seconds OR auth check completion
- Brand colors: primary theme color
```

**Features**:
- [ ] Centered app logo (create or use existing asset)
- [ ] App name with tagline: "Share your wishes"
- [ ] Loading indicator (ActivityIndicator or animated)
- [ ] Auth state check during splash
- [ ] Navigation logic:
  - If authenticated → Home (`/(tabs)`)
  - If first launch → Onboarding
  - If returning unauthenticated → Login

**Dependencies**:
- `expo-splash-screen` (native splash handling)
- `@react-native-async-storage/async-storage` (first launch check)

---

## 2. Onboarding Flow

### Design Requirements
- **Screens**: 3-4 slides with progress indicator
- **Content**: Core value propositions
- **Style**: Clean with illustrations, minimal friction

### Technical Approach
```
Location: mobile/app/onboarding/
Type: Multi-step flow with swipeable screens
```

### Screen Content for Wish List App

**Screen 1 - Welcome** (1/3)
- Title: "Create Your Wish Lists"
- Description: "Organize your wishes into beautiful lists for any occasion"
- Illustration: Gift boxes / list icon

**Screen 2 - Share** (2/3)
- Title: "Share With Friends & Family"
- Description: "Let loved ones know exactly what you want"
- Illustration: Share/people icon

**Screen 3 - Reserve** (3/3)
- Title: "Reserve Gifts Secretly"
- Description: "Coordinate gifts without spoiling the surprise"
- Illustration: Lock/secret icon

### Implementation Details

**Files**:
```
mobile/app/onboarding/
├── _layout.tsx          # Stack layout for onboarding
├── index.tsx            # Main onboarding carousel
└── components/
    └── OnboardingSlide.tsx
```

**Features**:
- [ ] Swipeable carousel (react-native-pager-view or custom)
- [ ] Progress indicator (dots or progress bar like design: "1/6")
- [ ] Skip button (top-right)
- [ ] Next/Continue button
- [ ] Final "Get Started" button → Login/Register

**Component Structure**:
```tsx
interface OnboardingSlide {
  title: string;
  description: string;
  icon: string; // Material icon name
  backgroundColor?: string;
}
```

**Navigation Flow**:
```
Splash → Onboarding → Login/Register → Home
         (if first launch)
```

**Persistence**:
```tsx
// Mark onboarding as complete
await AsyncStorage.setItem('hasSeenOnboarding', 'true');
```

---

## 3. Login Screen Enhancement

### Current State
- ✅ Basic login form exists (`mobile/app/auth/login.tsx`)
- ✅ Email/password fields
- ⚠️ OAuth buttons exist but may need styling

### Design Requirements (from reference)
- Clean, centered layout
- Logo/branding at top
- Email/password inputs
- "Login with Email" primary button
- Social login options (Google, Apple, Facebook)
- "Don't have an account? Create account" link

### Enhancements Needed

**Visual Updates**:
- [ ] Add app logo at top
- [ ] Add "Let's Get Started!" or welcome text
- [ ] Style social buttons with brand icons
- [ ] Add divider "or" between email and social

**Functional Updates**:
- [ ] Biometric login option (Face ID / Touch ID)
- [ ] Remember me toggle (optional)
- [ ] Forgot password flow

**Files to Modify**:
```
mobile/app/auth/login.tsx
mobile/app/auth/register.tsx
```

---

## 4. Home Screen (Full Implementation)

### Current State
- ⚠️ Stub implementation: `<Text>HomeScreen</Text>`

### Design Requirements
- **Empty State**: Clear CTA guiding to primary action
- **Content State**: Personalized dashboard with quick access

### Technical Approach
```
Location: mobile/app/(tabs)/index.tsx
Type: Dashboard with conditional rendering
```

### Implementation Details

**Empty State** (new users):
```tsx
// Components:
- Welcome message with user name
- Large CTA: "Create Your First Wish List"
- Optional: Quick tips or feature highlights
- Illustration for empty state
```

**Content State** (returning users):
```tsx
// Components:
- Header with greeting: "Hello, {name}!"
- Quick stats (optional): lists count, items count
- Recent/featured wish lists (horizontal scroll)
- Quick actions: Create List, My Reservations
- Activity feed or notifications (optional)
```

**Features**:
- [ ] User greeting with name
- [ ] Empty state with CTA
- [ ] Recent wish lists preview
- [ ] Quick action buttons
- [ ] Pull-to-refresh
- [ ] Loading skeleton

**API Integration**:
```tsx
// Queries needed:
- GET /api/v1/wishlists (user's lists)
- GET /api/v1/profile (user info for greeting)
```

---

## 5. User Profile (Already Implemented)

### Current State
- ✅ Full implementation at `mobile/app/(tabs)/profile.tsx`
- ✅ Avatar display
- ✅ Profile information editing
- ✅ Email change
- ✅ Password change
- ✅ Theme toggle (dark mode)
- ✅ Logout
- ✅ Account deletion

### Minor Enhancements (Optional)
- [ ] Profile photo upload (camera/gallery)
- [ ] Notification preferences section
- [ ] Privacy settings section

---

## 6. Settings (Integrated in Profile)

### Current State
- ✅ Already integrated within Profile screen
- ✅ Appearance (dark mode toggle)
- ✅ Account actions

### Future Considerations (Out of Scope)
- Notification settings
- Privacy settings
- Language selection
- Help & Support
- About / Version info

---

## Implementation Phases

### Phase 1: Foundation (Priority: High)
**Duration**: 1-2 days

1. **Splash Screen**
   - Create splash screen component
   - Implement auth state check
   - Add first-launch detection
   - Configure expo-splash-screen

2. **Onboarding Flow**
   - Create onboarding layout
   - Implement slide carousel
   - Design 3 slides for Wish List app
   - Add skip/complete logic

### Phase 2: Core Screens (Priority: High)
**Duration**: 2-3 days

3. **Home Screen Full Implementation**
   - Design empty state UI
   - Build content state with dashboard
   - Implement API integration
   - Add pull-to-refresh

4. **Login Enhancement**
   - Add branding/logo
   - Style social buttons
   - Add visual polish

### Phase 3: Polish (Priority: Medium)
**Duration**: 1-2 days

5. **Animations & Transitions**
   - Splash → Onboarding transition
   - Onboarding slide animations
   - Home screen loading states

6. **Accessibility**
   - Screen reader support
   - Font scaling
   - Color contrast verification

---

## File Structure (New Files)

```
mobile/
├── app/
│   ├── splash.tsx                    # New: Splash screen
│   ├── onboarding/
│   │   ├── _layout.tsx               # New: Onboarding layout
│   │   └── index.tsx                 # New: Onboarding carousel
│   ├── (tabs)/
│   │   └── index.tsx                 # Modified: Full home screen
│   └── auth/
│       ├── login.tsx                 # Modified: Enhanced login
│       └── register.tsx              # Modified: Enhanced register
├── components/
│   ├── onboarding/
│   │   ├── OnboardingSlide.tsx       # New: Reusable slide
│   │   └── ProgressIndicator.tsx     # New: Dots/progress bar
│   └── home/
│       ├── EmptyState.tsx            # New: Empty state component
│       ├── WishListCard.tsx          # New: List preview card
│       └── QuickActions.tsx          # New: Action buttons
└── assets/
    └── images/
        ├── logo.png                  # App logo (if not exists)
        └── onboarding/               # Onboarding illustrations
```

---

## Dependencies to Add

```bash
# For onboarding carousel
npx expo install react-native-pager-view

# For animations (optional, may already exist)
npx expo install react-native-reanimated

# For first-launch detection
npx expo install @react-native-async-storage/async-storage
```

---

## Design Tokens / Theme

Based on existing app, use consistent styling:

```tsx
// Colors (from react-native-paper theme)
const colors = {
  primary: '#6200ee',       // Purple primary
  secondary: '#03DAC6',     // Teal accent
  background: '#FFFFFF',    // Light mode
  surface: '#FFFFFF',
  error: '#B00020',
  onSurface: '#000000',
};

// Typography
const typography = {
  headline: { fontSize: 24, fontWeight: 'bold' },
  title: { fontSize: 20, fontWeight: '600' },
  body: { fontSize: 16 },
  caption: { fontSize: 12 },
};

// Spacing
const spacing = {
  xs: 4,
  sm: 8,
  md: 16,
  lg: 24,
  xl: 32,
};
```

---

## Navigation Flow Diagram

```
App Launch
    │
    ▼
┌─────────┐
│ Splash  │ (1-3 sec, check auth + first launch)
└────┬────┘
     │
     ├── First Launch? ─────────────────┐
     │                                  ▼
     │                          ┌─────────────┐
     │                          │ Onboarding  │
     │                          │ (3 slides)  │
     │                          └──────┬──────┘
     │                                 │
     │                                 ▼
     ├── Not Authenticated? ──► ┌─────────────┐
     │                          │   Login /   │
     │                          │  Register   │
     │                          └──────┬──────┘
     │                                 │
     ▼                                 ▼
┌─────────┐                    ┌─────────────┐
│  Home   │ ◄──────────────────│   Home      │
│ (tabs)  │                    │  (tabs)     │
└─────────┘                    └─────────────┘
```

---

## Success Criteria

- [ ] Splash screen displays for 2 seconds with branding
- [ ] First-time users see onboarding flow
- [ ] Onboarding can be skipped or completed
- [ ] Onboarding state persists (not shown again)
- [ ] Home screen shows appropriate state (empty vs content)
- [ ] Login screen has improved visual design
- [ ] All screens follow consistent design language
- [ ] Navigation flows work correctly
- [ ] No regressions in existing functionality

---

## Questions for Clarification

Before implementation, please clarify:

1. **Branding**: Do you have a specific logo/icon to use, or should I create a placeholder?

2. **Onboarding Content**: Are the proposed 3 slides (Create, Share, Reserve) appropriate for your app's value proposition?

3. **Home Screen Priority**: For the content state, what should be the primary focus?
   - Recent wish lists
   - Quick actions
   - Activity feed
   - Statistics

4. **Social Login**: Which OAuth providers should be prominently displayed?
   - Google
   - Apple
   - Facebook
   - All three

5. **Animations**: What level of animation polish is desired?
   - Minimal (faster implementation)
   - Standard (recommended)
   - Rich (more development time)

---

## References

- Design inspiration: `/drafts/screens/` (Adrian K / DESIGNME.AGENCY)
- Existing codebase: `/mobile/app/`
- Design system: React Native Paper
