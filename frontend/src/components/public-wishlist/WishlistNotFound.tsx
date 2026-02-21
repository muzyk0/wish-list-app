'use client';

import Link from 'next/link';
import { useTranslation } from 'react-i18next';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

export function WishlistNotFound() {
  const { t } = useTranslation();

  return (
    <div className="container mx-auto py-16 px-4 max-w-md text-center">
      <Card>
        <CardHeader>
          <CardTitle className="text-xl">
            {t('publicWishlist.notFound.title')}
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-muted-foreground">
            {t('publicWishlist.notFound.description')}
          </p>
          <Button asChild variant="outline">
            <Link href="/">{t('publicWishlist.notFound.action')}</Link>
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
