'use client';

import { useQuery } from '@tanstack/react-query';
import { useParams } from 'next/navigation';
import { useTranslation } from 'react-i18next';
import { GuestReservationDialog } from '@/components/guest/GuestReservationDialog';
import { GiftItemCard } from '@/components/public-wishlist/GiftItemCard';
import { GiftItemSkeleton } from '@/components/public-wishlist/GiftItemSkeleton';
import { WishlistEmptyState } from '@/components/public-wishlist/WishlistEmptyState';
import { WishlistHeader } from '@/components/public-wishlist/WishlistHeader';
import { WishlistNotFound } from '@/components/public-wishlist/WishlistNotFound';
import { Button } from '@/components/ui/button';
import {
  DOMAIN_CONSTANTS,
  MOBILE_APP_REDIRECT_PATHS,
} from '@/constants/domains';
import { apiClient } from '@/lib/api/client';

export default function PublicWishListPage() {
  const { slug } = useParams<{ slug: string }>();
  const { t } = useTranslation();

  const {
    data: wishList,
    isLoading: isLoadingWishList,
    isError: isErrorWishList,
  } = useQuery({
    queryKey: ['public-wishlist', slug],
    queryFn: () => apiClient.getPublicWishList(slug),
    enabled: !!slug,
    retry: 1,
  });

  const {
    data: giftItemsData,
    isLoading: isLoadingGiftItems,
    isError: isErrorGiftItems,
  } = useQuery({
    queryKey: ['public-gift-items', slug],
    queryFn: () => apiClient.getPublicGiftItems(slug, 1, 100),
    enabled: !!slug && !!wishList,
    retry: 1,
  });

  const isLoading = isLoadingWishList || isLoadingGiftItems;
  const isError = isErrorWishList || isErrorGiftItems;
  const giftItems = giftItemsData?.items || [];

  // Loading state
  if (isLoading) {
    return (
      <div className="container mx-auto py-8 px-4 max-w-3xl">
        <div className="mb-6 space-y-3">
          <div className="h-8 w-1/2 bg-muted rounded animate-pulse" />
          <div className="h-5 w-1/3 bg-muted rounded animate-pulse" />
          <div className="h-4 w-2/3 bg-muted rounded animate-pulse" />
        </div>
        <div className="space-y-4">
          {Array.from({ length: 4 }).map((_, i) => (
            // biome-ignore lint/suspicious/noArrayIndexKey: skeleton list
            <GiftItemSkeleton key={i} />
          ))}
        </div>
      </div>
    );
  }

  // Not found / error state
  if (isError || !wishList) {
    return <WishlistNotFound />;
  }

  const sortedItems = [...giftItems].sort(
    (a, b) => (a.position ?? 0) - (b.position ?? 0),
  );

  return (
    <div className="container mx-auto py-8 px-4 max-w-3xl">
      {/* Wishlist header */}
      <WishlistHeader wishlist={wishList} />

      {/* Gift items */}
      <div className="space-y-4">
        {sortedItems.length === 0 ? (
          <WishlistEmptyState />
        ) : (
          sortedItems.map((item) => (
            <GiftItemCard
              key={item.id}
              item={item}
              reserveAction={
                <GuestReservationDialog
                  wishlistSlug={slug}
                  wishlistId={wishList.id}
                  itemId={item.id}
                  itemName={item.name}
                  isReserved={!!item.reserved_by_user_id}
                  isPurchased={!!item.purchased_by_user_id}
                />
              }
            />
          ))
        )}
      </div>

      {/* Mobile app promo */}
      <div className="mt-10 p-4 bg-blue-50 dark:bg-blue-950 rounded-lg border border-blue-200 dark:border-blue-800">
        <h3 className="font-semibold text-blue-900 dark:text-blue-100 mb-1">
          {t('publicWishlist.mobilePromo.title')}
        </h3>
        <p className="text-blue-700 dark:text-blue-300 text-sm mb-3">
          {t('publicWishlist.mobilePromo.description')}
        </p>
        <Button
          variant="outline"
          size="sm"
          onClick={() => {
            const appScheme = `wishlistapp://${MOBILE_APP_REDIRECT_PATHS.HOME}`;
            window.location.href = appScheme;
            setTimeout(() => {
              if (!document.hidden && document.visibilityState === 'visible') {
                window.location.href = DOMAIN_CONSTANTS.MOBILE_APP_BASE_URL;
              }
            }, 1500);
          }}
        >
          {t('publicWishlist.mobilePromo.action')}
        </Button>
      </div>
    </div>
  );
}
