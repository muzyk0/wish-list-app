import { Skeleton } from '@/shared/ui/skeleton';

export function GiftItemSkeleton() {
  return (
    <div
      className="rounded-2xl overflow-hidden flex"
      style={{
        background: 'var(--wl-card)',
        border: '1px solid var(--wl-card-border)',
      }}
    >
      {/* Image placeholder */}
      <div className="flex-shrink-0 w-24 sm:w-32 min-h-[110px]">
        <Skeleton
          className="h-full w-full rounded-none"
          style={{ minHeight: '110px' }}
        />
      </div>

      {/* Content placeholder */}
      <div className="flex-1 p-5 flex flex-col justify-between gap-3">
        <div className="space-y-2">
          <div className="flex justify-between gap-4">
            <Skeleton className="h-5 w-1/2" />
            <Skeleton className="h-5 w-20 rounded-full flex-shrink-0" />
          </div>
          <Skeleton className="h-4 w-3/4" />
          <Skeleton className="h-4 w-2/5" />
        </div>
        <div className="flex justify-between items-center">
          <Skeleton className="h-6 w-16" />
          <Skeleton className="h-9 w-28 rounded-lg" />
        </div>
      </div>
    </div>
  );
}
