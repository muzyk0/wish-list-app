'use client';

import { ExternalLink } from 'lucide-react';
import Image from 'next/image';
import { useTranslation } from 'react-i18next';
import type { GiftItem } from '@/shared/api/types';
import {
  Sheet,
  SheetContent,
  SheetTitle,
  SheetTrigger,
} from '@/shared/ui/sheet';

interface GiftItemCardProps {
  item: GiftItem;
  /** Slot for the reserve action button/dialog */
  reserveAction?: React.ReactNode;
}

/** 5-dot priority indicator */
function PriorityDots({ priority }: { priority: number }) {
  const filled = Math.ceil(priority / 2);
  return (
    <div
      className="flex items-center gap-1"
      role="img"
      aria-label={`Priority: ${priority}/10`}
    >
      {Array.from({ length: 5 }).map((_, i) => (
        <div
          // biome-ignore lint/suspicious/noArrayIndexKey: static indicator
          key={i}
          className="w-1.5 h-1.5 rounded-full"
          style={{
            background:
              i < filled ? 'var(--wl-accent)' : 'var(--wl-card-border)',
          }}
        />
      ))}
    </div>
  );
}

export function GiftItemCard({ item, reserveAction }: GiftItemCardProps) {
  const { t } = useTranslation();

  const isReserved =
    item.is_reserved ?? (!!item.reserved_by_user_id || !!item.reserved_at);
  const isPurchased = !!item.purchased_by_user_id;
  const priority = item.priority ?? 0;

  const statusConfig = isPurchased
    ? {
        label: t('publicWishlist.item.purchased'),
        color: 'var(--wl-purchased)',
        bg: 'oklch(0.97 0.02 28)',
      }
    : isReserved
      ? {
          label: t('publicWishlist.item.reserved'),
          color: 'var(--wl-reserved)',
          bg: 'oklch(0.965 0.025 68)',
        }
      : {
          label: t('publicWishlist.item.available'),
          color: 'var(--wl-available)',
          bg: 'oklch(0.965 0.03 150)',
        };

  const monogram = item.name.charAt(0).toUpperCase();

  return (
    <Sheet>
      {/* ── Card trigger ─────────────────────────────── */}
      <SheetTrigger asChild>
        <article
          className="wl-item-card rounded-2xl overflow-hidden flex cursor-pointer"
          style={{
            background: 'var(--wl-card)',
            border: '1px solid var(--wl-card-border)',
            opacity: isReserved || isPurchased ? 0.8 : 1,
          }}
        >
          {/* Image column */}
          <div className="relative flex-shrink-0 self-stretch w-24 sm:w-32 min-h-[110px]">
            {item.image_url ? (
              <Image
                src={item.image_url}
                alt={item.name}
                fill
                className="object-cover"
                sizes="(min-width: 640px) 128px, 96px"
              />
            ) : (
              <div
                className="absolute inset-0 flex items-center justify-center"
                style={{ background: 'var(--wl-accent-light)' }}
              >
                <span
                  className="wl-display text-3xl font-bold select-none"
                  style={{ color: 'var(--wl-accent)', opacity: 0.45 }}
                  aria-hidden
                >
                  {monogram}
                </span>
              </div>
            )}
            {(isReserved || isPurchased) && (
              <div
                className="absolute inset-0"
                style={{ background: 'oklch(0.1 0 0 / 0.22)' }}
                aria-hidden
              />
            )}
          </div>

          {/* Content */}
          <div className="flex-1 min-w-0 p-4 sm:p-5 flex flex-col justify-between gap-3">
            <div>
              <div className="mb-1.5 flex flex-col gap-1 sm:flex-row sm:items-start sm:gap-2">
                <h3
                  className="wl-display text-base sm:text-lg font-semibold leading-snug flex-1 min-w-0 break-words"
                  style={{ color: 'var(--wl-text)' }}
                >
                  {item.name}
                </h3>
                <span
                  className="self-start text-xs font-medium px-2.5 py-0.5 rounded-full flex-shrink-0 mt-0.5"
                  style={{
                    color: statusConfig.color,
                    background: statusConfig.bg,
                  }}
                >
                  {statusConfig.label}
                </span>
              </div>

              {item.description && (
                <p
                  className="text-sm line-clamp-2 leading-relaxed"
                  style={{ color: 'var(--wl-muted)' }}
                >
                  {item.description}
                </p>
              )}
            </div>

            <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
              <div className="flex flex-col gap-1.5">
                {item.price !== undefined && item.price !== null && (
                  <span
                    className="wl-display text-xl font-bold"
                    style={{ color: 'var(--wl-accent)' }}
                  >
                    ${item.price}
                  </span>
                )}
                <div className="flex items-center gap-3">
                  {priority > 0 && <PriorityDots priority={priority} />}
                  {item.link && (
                    <span
                      className="text-xs"
                      style={{ color: 'var(--wl-accent)', opacity: 0.7 }}
                    >
                      {t('publicWishlist.item.viewProduct')} →
                    </span>
                  )}
                </div>
              </div>

              {/* Reserve button — stopPropagation so it doesn't open the sheet */}
              <div
                role="none"
                onClick={(e) => e.stopPropagation()}
                className="w-full sm:w-auto [&>button]:w-full sm:[&>button]:w-auto"
              >
                {reserveAction}
              </div>
            </div>
          </div>
        </article>
      </SheetTrigger>

      {/* ── Detail sheet (bottom, mobile-first) ────── */}
      <SheetContent
        side="bottom"
        className="wl-theme rounded-t-3xl p-0 max-h-[88vh] flex flex-col"
        style={{ borderTop: '1px solid oklch(0.91 0.012 72)' }}
      >
        {/* Hidden accessible title */}
        <SheetTitle className="sr-only">{item.name}</SheetTitle>

        {/* Drag handle */}
        <div className="flex-shrink-0 flex justify-center pt-3 pb-1">
          <div
            className="w-10 h-1 rounded-full"
            style={{ background: 'var(--wl-card-border)' }}
          />
        </div>

        {/* Scrollable body */}
        <div className="flex-1 overflow-y-auto">
          {/* Hero image */}
          {item.image_url && (
            <div className="relative w-full aspect-video flex-shrink-0">
              <Image
                src={item.image_url}
                alt={item.name}
                fill
                className="object-cover"
                sizes="100vw"
              />
            </div>
          )}

          {/* Content */}
          <div className="p-5 pb-2 space-y-5">
            {/* Name + status */}
            <div className="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
              <h2
                className="wl-display text-2xl sm:text-3xl font-bold leading-tight flex-1"
                style={{ color: 'var(--wl-text)' }}
              >
                {item.name}
              </h2>
              <span
                className="self-start text-xs font-semibold px-3 py-1 rounded-full flex-shrink-0 mt-1"
                style={{
                  color: statusConfig.color,
                  background: statusConfig.bg,
                }}
              >
                {statusConfig.label}
              </span>
            </div>

            {/* Description — full, no clamp */}
            {item.description && (
              <p
                className="text-base leading-relaxed"
                style={{ color: 'var(--wl-muted)' }}
              >
                {item.description}
              </p>
            )}

            {/* Separator */}
            <div
              className="h-px"
              style={{
                background:
                  'linear-gradient(to right, var(--wl-accent), var(--wl-card-border) 50%, transparent)',
              }}
            />

            {/* Price + priority */}
            <div className="flex items-center justify-between flex-wrap gap-4">
              {item.price !== undefined && item.price !== null && (
                <div>
                  <p
                    className="text-xs uppercase tracking-widest mb-0.5"
                    style={{ color: 'var(--wl-muted)' }}
                  >
                    {t('publicWishlist.item.price') ?? 'Price'}
                  </p>
                  <span
                    className="wl-display text-3xl font-bold"
                    style={{ color: 'var(--wl-accent)' }}
                  >
                    ${item.price}
                  </span>
                </div>
              )}

              {priority > 0 && (
                <div className="text-right">
                  <p
                    className="text-xs uppercase tracking-widest mb-1.5"
                    style={{ color: 'var(--wl-muted)' }}
                  >
                    {t('publicWishlist.item.priority')}
                  </p>
                  <PriorityDots priority={priority} />
                </div>
              )}
            </div>

            {/* External link — only if it exists */}
            {item.link && (
              <a
                href={item.link}
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center justify-between w-full px-4 py-3 rounded-xl text-sm font-medium transition-opacity hover:opacity-80"
                style={{
                  background: 'var(--wl-accent-light)',
                  color: 'var(--wl-accent)',
                  border: '1px solid oklch(0.85 0.04 68)',
                }}
              >
                <span>{t('publicWishlist.item.viewProduct')}</span>
                <ExternalLink className="h-4 w-4 flex-shrink-0" aria-hidden />
              </a>
            )}
          </div>
        </div>

        {/* Sticky footer — reserve action */}
        {reserveAction && (
          <div
            className="flex-shrink-0 p-4 pt-3"
            style={{
              borderTop: '1px solid oklch(0.91 0.012 72)',
            }}
          >
            {/* Render the reserve action in full-width mode via a wrapper */}
            <div className="[&>button]:w-full [&>div>button]:w-full">
              {reserveAction}
            </div>
          </div>
        )}
      </SheetContent>
    </Sheet>
  );
}
