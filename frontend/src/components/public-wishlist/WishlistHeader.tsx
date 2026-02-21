'use client';

import { useTranslation } from 'react-i18next';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import type { WishList } from '@/lib/api/types';

interface WishlistHeaderProps {
  wishlist: WishList;
}

export function WishlistHeader({ wishlist }: WishlistHeaderProps) {
  const { t } = useTranslation();

  const occasionDate = wishlist.occasion_date
    ? new Date(wishlist.occasion_date).toLocaleDateString(undefined, {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
      })
    : null;

  return (
    <Card className="mb-6 border-[var(--wishlist-accent)] bg-[var(--wishlist-bg)]">
      <CardHeader>
        <CardTitle className="text-2xl font-bold text-[var(--wishlist-primary)]">
          {wishlist.title}
        </CardTitle>
        {wishlist.occasion && (
          <p className="text-lg text-muted-foreground">
            {t('publicWishlist.occasion')}: {wishlist.occasion}
            {occasionDate && (
              <span className="ml-2 text-sm">({occasionDate})</span>
            )}
          </p>
        )}
      </CardHeader>
      <CardContent>
        {wishlist.description && (
          <p className="text-muted-foreground mb-4">{wishlist.description}</p>
        )}
        <div className="flex flex-wrap items-center gap-2">
          <Badge variant="secondary">{t('publicWishlist.publicBadge')}</Badge>
          {wishlist.item_count !== undefined && wishlist.item_count > 0 && (
            <Badge variant="outline">
              {t('publicWishlist.items', { count: wishlist.item_count })}
            </Badge>
          )}
          {wishlist.view_count && wishlist.view_count !== '0' && (
            <Badge variant="outline">
              {t('publicWishlist.views', {
                count: Number(wishlist.view_count),
              })}
            </Badge>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
