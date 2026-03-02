'use client';

import { useTranslation } from 'react-i18next';
import { Button } from '@/shared/ui/button';
import { Input } from '@/shared/ui/input';

type StatusFilter = 'all' | 'available' | 'reserved' | 'purchased';
type SortOption =
  | 'position'
  | 'name_asc'
  | 'name_desc'
  | 'price_asc'
  | 'price_desc'
  | 'priority_desc';

interface WishlistFiltersProps {
  searchQuery: string;
  onSearchChange: (query: string) => void;
  statusFilter: StatusFilter;
  onStatusFilterChange: (filter: StatusFilter) => void;
  sortBy: SortOption;
  onSortByChange: (sort: SortOption) => void;
  hasActiveFilters: boolean;
  onReset: () => void;
}

export function WishlistFilters({
  searchQuery,
  onSearchChange,
  statusFilter,
  onStatusFilterChange,
  sortBy,
  onSortByChange,
  hasActiveFilters,
  onReset,
}: WishlistFiltersProps) {
  const { t } = useTranslation();

  return (
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
            onChange={(e) => onSearchChange(e.target.value)}
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
              onStatusFilterChange(e.target.value as StatusFilter)
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
            onChange={(e) => onSortByChange(e.target.value as SortOption)}
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
              onClick={onReset}
            >
              {t('publicWishlist.filters.reset')}
            </Button>
          </div>
        ) : null}
      </div>
    </div>
  );
}
