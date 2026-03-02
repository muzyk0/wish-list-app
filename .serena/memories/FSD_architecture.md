# FSD Architecture (Corrected)

## Layer Structure (Final)
```
src/
├── app/           # Next.js routing (thin wrappers)
├── widgets/       # Reusable page-level compositions (Header, Footer)
├── screens/       # Page-specific compositions (PublicWishlistPage, MyReservationsPage)
├── features/      # User interaction slices
├── entities/      # Domain models + display UI
└── shared/        # Cross-cutting concerns
```

## Key Distinction
- **widgets/**: Reusable across multiple pages (Header, Footer, etc.)
- **screens/**: Page-specific content bound to routes (not reusable)
- **pages/ was renamed to screens/** to avoid Next.js Pages Router conflict

## Dependency Rules
```
app/ → widgets/ ↓
       screens/ → features/ → entities/ → shared/
```

## tsconfig.json Paths
```json
{
  "@/shared/*": "./src/shared/*",
  "@/entities/*": "./src/entities/*",
  "@/features/*": "./src/features/*",
  "@/widgets/*": "./src/widgets/*",
  "@/screens/*": "./src/screens/*",
  "@/app/*": "./src/app/*"
}
```

## Migration Complete
- ✅ PublicWishlistPage moved to screens/public-wishlist/
- ✅ MyReservationsPage moved to screens/my-reservations/
- ✅ App imports updated to @/screens/*
- ✅ Widgets simplified to only Header and Footer
- ✅ Type checking and linting passed
- ✅ Next.js build successful
