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
      <div className="max-w-3xl mx-auto px-4 py-12">
        <div className="space-y-3 mb-10">
          <div className="h-4 w-28 rounded-full bg-muted animate-pulse" />
          <div className="h-12 w-2/3 rounded-lg bg-muted animate-pulse" />
          <div className="h-4 w-full max-w-sm rounded bg-muted animate-pulse" />
        </div>
        <div className="space-y-3">
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

  const reservedCount = sortedItems.filter(
    (item) => item.reserved_by_user_id || item.purchased_by_user_id,
  ).length;

  // Promo block appears after all items are done animating
  const promoDelay = Math.min(sortedItems.length, 9) + 2;

  return (
    <div className="max-w-3xl mx-auto px-4 py-12">
      {/* Wishlist header with staggered entrance */}
      <div className="wl-fade-up wl-delay-0">
        <WishlistHeader wishlist={wishList} reservedCount={reservedCount} />
      </div>

      {/* Gift items */}
      <div className="space-y-3">
        {sortedItems.length === 0 ? (
          <WishlistEmptyState />
        ) : (
          sortedItems.map((item, index) => (
            <div
              key={item.id}
              className={`wl-fade-up wl-delay-${Math.min(index + 1, 9)}`}
            >
              <GiftItemCard
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
            </div>
          ))
        )}
      </div>

      {/* Mobile app promo */}
      <div
        className="wl-fade-up mt-12 rounded-2xl p-6"
        style={{
          background: 'var(--wl-accent-light)',
          border: '1px solid var(--wl-card-border)',
          animationDelay: `${promoDelay * 70}ms`,
        }}
      >
        <h3
          className="wl-display font-semibold text-lg mb-1"
          style={{ color: 'var(--wl-text)' }}
        >
          {t('publicWishlist.mobilePromo.title')}
        </h3>
        <p className="text-sm mb-4" style={{ color: 'var(--wl-muted)' }}>
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
          style={{ borderColor: 'var(--wl-accent)', color: 'var(--wl-accent)' }}
        >
          {t('publicWishlist.mobilePromo.action')}
        </Button>
      </div>
    </div>
  );
}
