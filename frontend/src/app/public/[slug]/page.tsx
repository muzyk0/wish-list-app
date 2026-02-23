'use client';

import { useQuery } from '@tanstack/react-query';
import { useParams } from 'next/navigation';
import { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { GuestReservationDialog } from '@/components/guest/GuestReservationDialog';
import { GiftItemCard } from '@/components/public-wishlist/GiftItemCard';
import { GiftItemSkeleton } from '@/components/public-wishlist/GiftItemSkeleton';
import { WishlistEmptyState } from '@/components/public-wishlist/WishlistEmptyState';
import { WishlistHeader } from '@/components/public-wishlist/WishlistHeader';
import { WishlistNotFound } from '@/components/public-wishlist/WishlistNotFound';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  DOMAIN_CONSTANTS,
  MOBILE_APP_REDIRECT_PATHS,
} from '@/constants/domains';
import { apiClient } from '@/lib/api/client';
import type { GiftItem } from '@/lib/api/types';

type StatusFilter = 'all' | 'available' | 'reserved' | 'purchased';
type SortOption =
  | 'position'
  | 'name_asc'
  | 'name_desc'
  | 'price_asc'
  | 'price_desc'
  | 'priority_desc';

const isItemReserved = (item: GiftItem) =>
  item.is_reserved ?? (!!item.reserved_by_user_id || !!item.reserved_at);

const isItemPurchased = (item: GiftItem) => !!item.purchased_by_user_id;

export default function PublicWishListPage() {
  const { slug } = useParams<{ slug: string }>();
  const { t } = useTranslation();
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all');
  const [sortBy, setSortBy] = useState<SortOption>('position');

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

  const filteredAndSortedItems = useMemo(() => {
    const query = searchQuery.trim().toLowerCase();
    let items = [...giftItems];

    if (query) {
      items = items.filter((item) => {
        const haystack = [item.name, item.description]
          .filter(Boolean)
          .join(' ')
          .toLowerCase();
        return haystack.includes(query);
      });
    }

    if (statusFilter !== 'all') {
      items = items.filter((item) => {
        const isPurchased = isItemPurchased(item);
        const isReserved = isItemReserved(item);

        if (statusFilter === 'purchased') return isPurchased;
        if (statusFilter === 'reserved') return !isPurchased && isReserved;
        return !isPurchased && !isReserved;
      });
    }

    items.sort((a, b) => {
      switch (sortBy) {
        case 'name_asc':
          return a.name.localeCompare(b.name);
        case 'name_desc':
          return b.name.localeCompare(a.name);
        case 'price_asc':
          return (
            (a.price ?? Number.POSITIVE_INFINITY) -
            (b.price ?? Number.POSITIVE_INFINITY)
          );
        case 'price_desc':
          return (
            (b.price ?? Number.NEGATIVE_INFINITY) -
            (a.price ?? Number.NEGATIVE_INFINITY)
          );
        case 'priority_desc':
          return (b.priority ?? 0) - (a.priority ?? 0);
        default:
          return (a.position ?? 0) - (b.position ?? 0);
      }
    });

    return items;
  }, [giftItems, searchQuery, statusFilter, sortBy]);

  const reservedCount = giftItems.filter(
    (item) => isItemReserved(item) || isItemPurchased(item),
  ).length;

  // Promo block appears after all items are done animating
  const promoDelay = Math.min(filteredAndSortedItems.length, 9) + 3;
  const hasActiveFilters =
    searchQuery.trim() !== '' ||
    statusFilter !== 'all' ||
    sortBy !== 'position';

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
      {giftItems.length > 0 && (
        <div
          className="wl-fade-up wl-delay-1 mb-6 rounded-2xl p-4 sm:p-5"
          style={{
            background: 'var(--wl-card)',
            border: '1px solid var(--wl-card-border)',
          }}
        >
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-[1fr_180px_220px_auto]">
            <div className="sm:col-span-2 lg:col-span-1">
              <label
                htmlFor="wishlist-search"
                className="mb-1.5 block text-xs font-medium uppercase tracking-[0.08em]"
                style={{ color: 'var(--wl-muted)' }}
              >
                {t('publicWishlist.filters.searchLabel')}
              </label>
              <Input
                id="wishlist-search"
                type="search"
                data-testid="wishlist-search-input"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder={t('publicWishlist.filters.searchPlaceholder')}
              />
            </div>

            <div>
              <label
                htmlFor="wishlist-status-filter"
                className="mb-1.5 block text-xs font-medium uppercase tracking-[0.08em]"
                style={{ color: 'var(--wl-muted)' }}
              >
                {t('publicWishlist.filters.statusLabel')}
              </label>
              <select
                id="wishlist-status-filter"
                data-testid="wishlist-status-filter"
                value={statusFilter}
                onChange={(e) =>
                  setStatusFilter(e.target.value as StatusFilter)
                }
                className="border-input focus-visible:border-ring focus-visible:ring-ring/50 h-9 w-full rounded-md border bg-transparent px-3 text-sm shadow-xs outline-none focus-visible:ring-[3px]"
              >
                <option value="all">
                  {t('publicWishlist.filters.status.all')}
                </option>
                <option value="available">
                  {t('publicWishlist.filters.status.available')}
                </option>
                <option value="reserved">
                  {t('publicWishlist.filters.status.reserved')}
                </option>
                <option value="purchased">
                  {t('publicWishlist.filters.status.purchased')}
                </option>
              </select>
            </div>

            <div>
              <label
                htmlFor="wishlist-sort"
                className="mb-1.5 block text-xs font-medium uppercase tracking-[0.08em]"
                style={{ color: 'var(--wl-muted)' }}
              >
                {t('publicWishlist.filters.sortLabel')}
              </label>
              <select
                id="wishlist-sort"
                data-testid="wishlist-sort"
                value={sortBy}
                onChange={(e) => setSortBy(e.target.value as SortOption)}
                className="border-input focus-visible:border-ring focus-visible:ring-ring/50 h-9 w-full rounded-md border bg-transparent px-3 text-sm shadow-xs outline-none focus-visible:ring-[3px]"
              >
                <option value="position">
                  {t('publicWishlist.filters.sort.position')}
                </option>
                <option value="name_asc">
                  {t('publicWishlist.filters.sort.nameAsc')}
                </option>
                <option value="name_desc">
                  {t('publicWishlist.filters.sort.nameDesc')}
                </option>
                <option value="price_asc">
                  {t('publicWishlist.filters.sort.priceAsc')}
                </option>
                <option value="price_desc">
                  {t('publicWishlist.filters.sort.priceDesc')}
                </option>
                <option value="priority_desc">
                  {t('publicWishlist.filters.sort.priorityDesc')}
                </option>
              </select>
            </div>

            {hasActiveFilters ? (
              <div className="flex items-end">
                <Button
                  variant="outline"
                  size="sm"
                  className="w-full lg:w-auto"
                  onClick={() => {
                    setSearchQuery('');
                    setStatusFilter('all');
                    setSortBy('position');
                  }}
                >
                  {t('publicWishlist.filters.reset')}
                </Button>
              </div>
            ) : null}
          </div>
        </div>
      )}

      {/* Gift items */}
      <div className="space-y-3">
        {giftItems.length === 0 ? (
          <WishlistEmptyState />
        ) : filteredAndSortedItems.length === 0 ? (
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
          filteredAndSortedItems.map((item, index) => (
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
                    isReserved={isItemReserved(item)}
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
