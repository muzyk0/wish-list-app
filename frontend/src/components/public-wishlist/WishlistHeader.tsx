'use client';

import { CalendarDays, Eye, Gift, Sparkles } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import type { WishList } from '@/lib/api/types';

interface WishlistHeaderProps {
  wishlist: WishList;
  reservedCount?: number;
}

export function WishlistHeader({
  wishlist,
  reservedCount = 0,
}: WishlistHeaderProps) {
  const { t } = useTranslation();

  const totalItems = wishlist.item_count ?? 0;
  const reservedPct =
    totalItems > 0 ? Math.round((reservedCount / totalItems) * 100) : 0;

  const occasionDate = wishlist.occasion_date
    ? new Date(wishlist.occasion_date).toLocaleDateString(undefined, {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
      })
    : null;

  const viewCount =
    wishlist.view_count && wishlist.view_count !== '0'
      ? Number(wishlist.view_count)
      : null;

  return (
    <header className="mb-10">
      {/* Occasion label */}
      {wishlist.occasion && (
        <div className="flex items-center gap-2 mb-3 flex-wrap">
          <Sparkles
            className="h-3.5 w-3.5 flex-shrink-0"
            style={{ color: 'var(--wl-accent)' }}
            aria-hidden
          />
          <span
            className="text-xs font-semibold tracking-[0.18em] uppercase"
            style={{ color: 'var(--wl-accent)' }}
          >
            {wishlist.occasion}
          </span>
          {occasionDate && (
            <>
              <span style={{ color: 'var(--wl-muted)' }} aria-hidden>
                ·
              </span>
              <span
                className="text-sm flex items-center gap-1"
                style={{ color: 'var(--wl-muted)' }}
              >
                <CalendarDays className="h-3.5 w-3.5" aria-hidden />
                <time>{occasionDate}</time>
              </span>
            </>
          )}
        </div>
      )}

      {/* Title */}
      <h1
        className="wl-display text-4xl sm:text-5xl font-bold leading-tight mb-4"
        style={{ color: 'var(--wl-text)' }}
      >
        {wishlist.title}
      </h1>

      {/* Description */}
      {wishlist.description && (
        <p
          className="text-base sm:text-lg leading-relaxed mb-6"
          style={{ color: 'var(--wl-muted)', maxWidth: '58ch' }}
        >
          {wishlist.description}
        </p>
      )}

      {/* Gold gradient separator */}
      <div
        className="h-px mb-6"
        style={{
          background:
            'linear-gradient(to right, var(--wl-accent), var(--wl-card-border) 55%, transparent)',
        }}
      />

      {/* Stats row */}
      {(totalItems > 0 || viewCount !== null) && (
        <div className="flex flex-wrap items-center gap-5 mb-6">
          {totalItems > 0 && (
            <div className="flex items-center gap-1.5">
              <Gift
                className="h-4 w-4 flex-shrink-0"
                style={{ color: 'var(--wl-accent)' }}
                aria-hidden
              />
              <span className="text-sm" style={{ color: 'var(--wl-muted)' }}>
                <strong style={{ color: 'var(--wl-text)' }}>
                  {totalItems}
                </strong>{' '}
                {t('publicWishlist.items', { count: totalItems })}
              </span>
            </div>
          )}
          {viewCount !== null && (
            <div className="flex items-center gap-1.5">
              <Eye
                className="h-4 w-4 flex-shrink-0"
                style={{ color: 'var(--wl-accent)' }}
                aria-hidden
              />
              <span className="text-sm" style={{ color: 'var(--wl-muted)' }}>
                <strong style={{ color: 'var(--wl-text)' }}>
                  {viewCount.toLocaleString()}
                </strong>{' '}
                {t('publicWishlist.views', { count: viewCount })}
              </span>
            </div>
          )}
        </div>
      )}

      {/* Reservation progress */}
      {totalItems > 0 && reservedCount > 0 && (
        <div>
          <div className="flex justify-between items-center mb-2">
            <span
              className="text-xs uppercase tracking-[0.12em] font-medium"
              style={{ color: 'var(--wl-muted)' }}
            >
              Gifts reserved
            </span>
            <span
              className="text-xs font-bold tabular-nums"
              style={{ color: 'var(--wl-accent)' }}
            >
              {reservedCount} / {totalItems}
            </span>
          </div>
          <div
            className="h-1 rounded-full overflow-hidden"
            style={{ background: 'var(--wl-accent-light)' }}
          >
            <div
              className="h-full rounded-full wl-progress-bar"
              style={{
                width: `${reservedPct}%`,
                background:
                  'linear-gradient(to right, var(--wl-accent), oklch(0.72 0.07 70))',
              }}
            />
          </div>
        </div>
      )}
    </header>
  );
}
