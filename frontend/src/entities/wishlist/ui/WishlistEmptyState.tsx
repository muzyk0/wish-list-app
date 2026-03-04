'use client';

import { Gift } from 'lucide-react';
import { useTranslation } from 'react-i18next';

export function WishlistEmptyState() {
  const { t } = useTranslation();

  return (
    <div
      className="flex flex-col items-center gap-5 py-20 text-center rounded-2xl"
      style={{
        background: 'var(--wl-card)',
        border: '1px dashed var(--wl-card-border)',
      }}
    >
      <div
        className="w-16 h-16 rounded-full flex items-center justify-center"
        style={{ background: 'var(--wl-accent-light)' }}
      >
        <Gift
          className="h-7 w-7"
          style={{ color: 'var(--wl-accent)' }}
          aria-hidden
        />
      </div>
      <div>
        <p
          className="wl-display text-xl font-semibold mb-1.5"
          style={{ color: 'var(--wl-text)' }}
        >
          {t('publicWishlist.empty.title')}
        </p>
        <p className="text-sm" style={{ color: 'var(--wl-muted)' }}>
          {t('publicWishlist.empty.description')}
        </p>
      </div>
    </div>
  );
}
