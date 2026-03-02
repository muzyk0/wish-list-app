'use client';

import {
  keepPreviousData,
  useInfiniteQuery,
  useQuery,
} from '@tanstack/react-query';
import { useParams } from 'next/navigation';
import { useCallback, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { GiftItemCard, GiftItemSkeleton } from '@/entities/gift-item';
import {
  WishlistEmptyState,
  WishlistHeader,
  WishlistNotFound,
} from '@/entities/wishlist';
import { GuestReservationDialog } from '@/features/guest-reservation';
import { apiClient } from '@/shared/api/client';
import {
  DOMAIN_CONSTANTS,
  MOBILE_APP_REDIRECT_PATHS,
} from '@/shared/config/domains';
import { useDebounce, useIntersectionObserver } from '@/shared/lib/hooks';
import { Button } from '@/shared/ui/button';
import { WishlistFilters } from './WishlistFilters';

type StatusFilter = 'all' | 'available' | 'reserved' | 'purchased';
type SortOption =
  | 'position'
  | 'name_asc'
  | 'name_desc'
  | 'price_asc'
  | 'price_desc'
  | 'priority_desc';

const PAGE_SIZE = 12;

export function PublicWishlistPage() {
  const { slug } = useParams<{ slug: string }>();
  const { t } = useTranslation();
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all');
  const [sortBy, setSortBy] = useState<SortOption>('position');

  const debouncedSearch = useDebounce(searchQuery, 300);

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
    data: itemsData,
    isLoading: isLoadingGiftItems,
    isError: isErrorGiftItems,
    isFetching: isFetchingGiftItems,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
  } = useInfiniteQuery({
    queryKey: [
      'public-gift-items',
      slug,
      debouncedSearch,
      statusFilter,
      sortBy,
    ],
    queryFn: ({ pageParam }) =>
      apiClient.getPublicGiftItems(
        slug,
        pageParam,
        PAGE_SIZE,
        debouncedSearch || undefined,
        statusFilter !== 'all' ? statusFilter : undefined,
        sortBy !== 'position' ? sortBy : undefined,
      ),
    initialPageParam: 1,
    getNextPageParam: (lastPage) =>
      lastPage.page < lastPage.pages ? lastPage.page + 1 : undefined,
    enabled: !!slug && !!wishList,
    retry: 1,
    // Keep previous results visible while new filter query loads — prevents flash
    placeholderData: keepPreviousData,
  });

  // True only when a filter/sort change is in flight (not initial load, not next-page load)
  const isFilterTransition =
    isFetchingGiftItems && !isLoadingGiftItems && !isFetchingNextPage;

  const isLoading = isLoadingWishList || isLoadingGiftItems;
  const isError = isErrorWishList || isErrorGiftItems;

  const giftItems = itemsData?.pages.flatMap((p) => p.items ?? []) ?? [];
  const totalItems = itemsData?.pages[0]?.total ?? 0;

  // Count reserved items from loaded pages
  const reservedCount = giftItems.filter((item) => item.is_reserved).length;

  const hasActiveFilters =
    searchQuery.trim() !== '' ||
    statusFilter !== 'all' ||
    sortBy !== 'position';

  const promoDelay = Math.min(giftItems.length, 9) + 3;

  const handleFetchNext = useCallback(() => {
    if (hasNextPage && !isFetchingNextPage) {
      fetchNextPage();
    }
  }, [hasNextPage, isFetchingNextPage, fetchNextPage]);

  const sentinelRef = useIntersectionObserver({
    onIntersect: handleFetchNext,
    enabled: !!hasNextPage && !isFetchingNextPage,
  });

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

  return (
    <div className="max-w-3xl mx-auto px-4 py-12">
      {/* Wishlist header with staggered entrance */}
      <div className="wl-fade-up wl-delay-0">
        <WishlistHeader wishlist={wishList} reservedCount={reservedCount} />
      </div>

      {/* List controls */}
      {totalItems > 0 && (
        <WishlistFilters
          searchQuery={searchQuery}
          onSearchChange={setSearchQuery}
          statusFilter={statusFilter}
          onStatusFilterChange={setStatusFilter}
          sortBy={sortBy}
          onSortByChange={setSortBy}
          hasActiveFilters={hasActiveFilters}
          onReset={() => {
            setSearchQuery('');
            setStatusFilter('all');
            setSortBy('position');
          }}
        />
      )}

      {/* Gift items */}
      <div
        className="space-y-3"
        style={{
          opacity: isFilterTransition ? 0.5 : 1,
          transition: 'opacity 150ms ease',
          pointerEvents: isFilterTransition ? 'none' : undefined,
        }}
      >
        {giftItems.length === 0 &&
        !isFetchingNextPage &&
        !isFilterTransition ? (
          hasActiveFilters ? (
            <div
              className="rounded-2xl px-5 py-8 text-center text-sm sm:text-base"
              style={{
                background: 'var(--wl-card)',
                border: '1px solid var(--wl-card-border)',
                color: 'var(--wl-muted)',
              }}
            >
              {t('publicWishlist.filters.noResults')}
            </div>
          ) : (
            <WishlistEmptyState />
          )
        ) : (
          giftItems.map((item, index) => (
            <div
              key={item.id}
              className={`wl-fade-up wl-delay-${Math.min(index + 2, 9)}`}
            >
              <GiftItemCard
                item={item}
                reserveAction={
                  <GuestReservationDialog
                    wishlistSlug={slug}
                    wishlistId={wishList.id}
                    itemId={item.id}
                    itemName={item.name}
                    isReserved={
                      item.is_reserved ??
                      !!(item.reserved_by_user_id || item.reserved_at)
                    }
                    isPurchased={!!item.purchased_by_user_id}
                  />
                }
              />
            </div>
          ))
        )}

        {/* Infinite scroll sentinel */}
        <div ref={sentinelRef} aria-hidden="true" className="h-1" />

        {/* Loading skeletons for next page */}
        {isFetchingNextPage &&
          Array.from({ length: 3 }).map((_, i) => (
            // biome-ignore lint/suspicious/noArrayIndexKey: skeleton list
            <GiftItemSkeleton key={i} />
          ))}
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
