'use client';

import Link from 'next/link';
import { useTranslation } from 'react-i18next';
import { Button } from '@/components/ui/button';

export function WishlistNotFound() {
  const { t } = useTranslation();

  return (
    <div className="min-h-screen flex items-center justify-center p-6">
      <div className="text-center max-w-xs">
        <div
          className="w-20 h-20 rounded-full mx-auto mb-6 flex items-center justify-center"
          style={{ background: 'var(--wl-accent-light)' }}
        >
          <span className="text-3xl" role="img" aria-label="Gift">
            🎁
          </span>
        </div>

        <h1
          className="wl-display text-2xl font-bold mb-2"
          style={{ color: 'var(--wl-text)' }}
        >
          {t('publicWishlist.notFound.title')}
        </h1>

        <p className="text-sm mb-6" style={{ color: 'var(--wl-muted)' }}>
          {t('publicWishlist.notFound.description')}
        </p>

        <Button asChild variant="outline">
          <Link href="/">{t('publicWishlist.notFound.action')}</Link>
        </Button>
      </div>
    </div>
  );
}
