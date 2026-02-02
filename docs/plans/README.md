# Implementation Plans Index

**Updated**: 2026-02-02
**Architecture**: Cross-Domain (Frontend + Mobile + Backend on different providers)

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        WISH LIST APPLICATION                             │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌───────────────────┐  ┌───────────────────┐  ┌───────────────────┐   │
│  │  Frontend (Web)   │  │  Mobile (App)     │  │  Backend (API)    │   │
│  │  Next.js          │  │  React Native     │  │  Go/Echo          │   │
│  │  Vercel           │  │  Expo + Vercel    │  │  Render           │   │
│  │                   │  │                   │  │                   │   │
│  │  Public features: │  │  Personal Cabinet:│  │  Endpoints:       │   │
│  │  • View wishlists │  │  • Create lists   │  │  • /auth/*        │   │
│  │  • Reserve items  │  │  • Manage items   │  │  • /wishlists/*   │   │
│  │  • Cancel reserve │  │  • View reserves  │  │  • /reservations/*│   │
│  │  • Auth + redirect│  │  • Settings       │  │  • /public/*      │   │
│  └───────────────────┘  └───────────────────┘  └───────────────────┘   │
│           │                      │                      ▲              │
│           └──────────────────────┴──────────────────────┘              │
│                       HTTPS + JWT + CORS                                │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Plans Summary

| Plan | Focus | Priority | Status |
|------|-------|----------|--------|
| [00-cross-domain-architecture-plan.md](./00-cross-domain-architecture-plan.md) | **Auth architecture, CORS, handoff flow** | 🔴 Critical | New |
| [01-frontend-security-and-quality-plan.md](./01-frontend-security-and-quality-plan.md) | Frontend security, token management | 🔴 Critical | Updated |
| [02-mobile-app-completion-plan.md](./02-mobile-app-completion-plan.md) | Mobile auth, features, PR issues | 🟡 High | Updated |
| [03-api-backend-improvements-plan.md](./03-api-backend-improvements-plan.md) | Backend auth, CORS, Render deployment | 🔴 Critical | Updated |

---

## Implementation Order

### Phase 0: Architecture Foundation (Week 1)
**Must complete before other phases**

| Task | Plan | Description |
|------|------|-------------|
| CORS Configuration | 03 | Allow Frontend + Mobile origins |
| Refresh Token Endpoint | 03 | `POST /auth/refresh` |
| Mobile Handoff Endpoints | 03 | `POST /auth/mobile-handoff`, `POST /auth/exchange` |
| Health Check | 03 | `GET /health` for Render |

### Phase 1: Security & Auth (Week 1-2)
**Parallel work: Frontend + Mobile + Backend**

| Component | Tasks |
|-----------|-------|
| Backend | Token refresh, handoff, CORS |
| Frontend | In-memory tokens, refresh flow, mobile redirect |
| Mobile | SecureStore, code exchange, deep link handling |

### Phase 2: Core Features (Week 2-3)

| Component | Tasks |
|-----------|-------|
| Mobile | Gift item CRUD, image upload, reservations |
| Frontend | Guest reservation, my reservations view |
| Backend | Pagination, OpenAPI fixes |

### Phase 3: Polish & Deploy (Week 3-4)

| Component | Tasks |
|-----------|-------|
| Backend | Render deployment, rate limiting |
| Frontend | Vercel optimization |
| Mobile | App Store preparation |

---

## Key Decisions

### Authentication Strategy

| Aspect | Decision | Rationale |
|--------|----------|-----------|
| Token Storage (Web) | In-memory + refresh cookie | Prevents XSS token theft |
| Token Storage (Mobile) | expo-secure-store | Native secure storage |
| Cross-domain Auth | OAuth-style handoff | Secure token transfer between apps |
| Token Lifetime | Access: 15m, Refresh: 7d | Balance security + UX |

### Deployment Strategy

| Component | Provider | Domain |
|-----------|----------|--------|
| Frontend | Vercel | wishlist.com |
| Mobile Web | Vercel | N/A (native app) |
| Backend | Render | api.wishlist.com |
| Database | Render PostgreSQL | Internal |

---

## Quick Start

### Backend
```bash
cd backend

# Development
go run ./cmd/server

# Docker
docker build -t wishlist-api .
docker run -p 8080:8080 --env-file .env wishlist-api
```

### Frontend
```bash
cd frontend

# Development
npm run dev

# Build
npm run build
```

### Mobile
```bash
cd mobile

# Development
npx expo start

# Build for preview
npx eas build --platform ios --profile preview
```

---

## Environment Variables

### Backend (Render)
```bash
DATABASE_URL=postgresql://...
JWT_SECRET=<generated>
JWT_ACCESS_TOKEN_EXPIRY=15m
JWT_REFRESH_TOKEN_EXPIRY=7d
CORS_ALLOWED_ORIGINS=https://wishlist.com
ENV=production
```

### Frontend (Vercel)
```bash
NEXT_PUBLIC_API_URL=https://api.wishlist.com
NEXT_PUBLIC_MOBILE_SCHEME=wishlistapp
NEXT_PUBLIC_MOBILE_UNIVERSAL_LINK=https://wishlist.com/app
```

### Mobile
```bash
EXPO_PUBLIC_API_URL=https://api.wishlist.com
```

---

## Dependencies Between Plans

```
                    ┌─────────────────────────────┐
                    │ 00-cross-domain-architecture │
                    │        (Foundation)          │
                    └──────────────┬──────────────┘
                                   │
           ┌───────────────────────┼───────────────────────┐
           │                       │                       │
           ▼                       ▼                       ▼
┌──────────────────┐    ┌──────────────────┐    ┌──────────────────┐
│  01-frontend     │    │  02-mobile       │    │  03-backend      │
│  (Security)      │    │  (Completion)    │    │  (Improvements)  │
└──────────────────┘    └──────────────────┘    └──────────────────┘
```

---

## Verification Checklist

### Cross-Domain Auth Working
- [ ] Frontend can login and get tokens
- [ ] Token refresh works automatically
- [ ] "Personal Cabinet" redirects to Mobile
- [ ] Mobile receives and exchanges auth code
- [ ] Logout clears tokens on both platforms

### CORS Working
- [ ] Frontend can call Backend API
- [ ] Cookies sent with `credentials: include`
- [ ] Preflight requests succeed

### Guest Flow Working
- [ ] Guest can view public wishlist
- [ ] Guest can reserve item (no auth)
- [ ] Guest receives email with management link
- [ ] Guest can cancel reservation

---

## Notes

- All plans updated for cross-domain architecture (2026-02-02)
- Backend must deploy first to enable Frontend/Mobile development
- Universal Links require Apple App Site Association file
- Android App Links require Digital Asset Links file
