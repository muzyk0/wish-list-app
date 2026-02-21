'use client';

import { ExternalLink, Image as ImageIcon } from 'lucide-react';
import Image from 'next/image';
import { useTranslation } from 'react-i18next';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent } from '@/components/ui/card';
import type { GiftItem } from '@/lib/api/types';

interface GiftItemCardProps {
  item: GiftItem;
  /** Slot for the reserve action button/dialog */
  reserveAction?: React.ReactNode;
}

export function GiftItemCard({ item, reserveAction }: GiftItemCardProps) {
  const { t } = useTranslation();

  const isReserved = !!item.reserved_by_user_id;
  const isPurchased = !!item.purchased_by_user_id;

  return (
    <Card className="overflow-hidden">
      <CardContent className="p-4">
        <div className="flex flex-col sm:flex-row gap-4">
          {/* Image */}
          <div className="flex-shrink-0">
            {item.image_url ? (
              <Image
                src={item.image_url}
                alt={item.name}
                width={80}
                height={80}
                className="w-20 h-20 object-cover rounded-md"
              />
            ) : (
              <div className="flex items-center justify-center w-20 h-20 bg-muted rounded-md">
                <ImageIcon
                  className="h-8 w-8 text-muted-foreground"
                  aria-hidden
                />
                <span className="sr-only">
                  {t('publicWishlist.item.noImage')}
                </span>
              </div>
            )}
          </div>

          {/* Content */}
          <div className="flex-1 min-w-0">
            {/* Title and status badges */}
            <div className="flex items-start justify-between gap-2 flex-wrap">
              <h3 className="text-base font-semibold leading-snug">
                {item.name}
              </h3>
              <div className="flex flex-wrap gap-1 shrink-0">
                {isPurchased && (
                  <Badge variant="destructive">
                    {t('publicWishlist.item.purchased')}
                  </Badge>
                )}
                {isReserved && !isPurchased && (
                  <Badge variant="secondary">
                    {t('publicWishlist.item.reserved')}
                  </Badge>
                )}
                {!isReserved && !isPurchased && (
                  <Badge
                    variant="outline"
                    className="text-green-700 border-green-300"
                  >
                    {t('publicWishlist.item.available')}
                  </Badge>
                )}
              </div>
            </div>

            {/* Description */}
            {item.description && (
              <p className="text-sm text-muted-foreground mt-1 line-clamp-2">
                {item.description}
              </p>
            )}

            {/* Price */}
            {item.price !== undefined && item.price !== null && (
              <p className="text-base font-bold mt-2">${item.price}</p>
            )}

            {/* Product link */}
            {item.link && (
              <a
                href={item.link}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-1 text-sm text-blue-600 hover:underline mt-1"
              >
                {t('publicWishlist.item.viewProduct')}
                <ExternalLink className="h-3.5 w-3.5" aria-hidden />
              </a>
            )}

            {/* Footer: priority + reserve action */}
            <div className="mt-3 flex items-center justify-between flex-wrap gap-2">
              <div>
                {item.priority !== undefined &&
                  item.priority !== null &&
                  item.priority > 0 && (
                    <span className="text-xs bg-muted px-2 py-1 rounded">
                      {t('publicWishlist.item.priority')}: {item.priority}/10
                    </span>
                  )}
              </div>
              {reserveAction}
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
