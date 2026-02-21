'use client';

import { Gift } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Card, CardContent } from '@/components/ui/card';

export function WishlistEmptyState() {
  const { t } = useTranslation();

  return (
    <Card>
      <CardContent className="py-12 flex flex-col items-center gap-3 text-center">
        <Gift className="h-12 w-12 text-muted-foreground" aria-hidden />
        <p className="font-medium">{t('publicWishlist.empty.title')}</p>
        <p className="text-sm text-muted-foreground">
          {t('publicWishlist.empty.description')}
        </p>
      </CardContent>
    </Card>
  );
}
