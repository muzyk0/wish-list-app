'use client';

import Link from 'next/link';
import { useTranslation } from 'react-i18next';
import { Badge } from '@/shared/ui/badge';
import { Card, CardContent } from '@/shared/ui/card';
import type { FeaturedListData } from './FeaturedListsSection';

interface FeaturedListCardProps {
  list: FeaturedListData;
  index: number;
}

export function FeaturedListCard({ list, index }: FeaturedListCardProps) {
  const { t } = useTranslation();

  return (
    <Link
      href={`/public/${list.slug}`}
      className="group block animate-fade-in-up"
      style={{ animationDelay: `${index * 100}ms` }}
    >
      <Card className="relative overflow-hidden h-full border-0 bg-gradient-to-br from-card to-muted/30 dark:from-card dark:to-muted/10 shadow-sm hover:shadow-xl transition-all duration-500 hover:-translate-y-1">
        {/* Gradient accent bar */}
        <div
          className={`absolute top-0 left-0 right-0 h-1 bg-gradient-to-r ${list.gradient} opacity-80 group-hover:opacity-100 transition-opacity`}
        />

        {/* Hover glow effect */}
        <div
          className={`absolute inset-0 bg-gradient-to-br ${list.gradient} opacity-0 group-hover:opacity-5 transition-opacity duration-500`}
        />

        <CardContent className="relative p-6 sm:p-8">
          {/* Icon */}
          <div className="mb-4 text-3xl sm:text-4xl">{list.icon}</div>

          {/* Theme badge */}
          <Badge
            variant="secondary"
            className="mb-3 text-xs font-medium tracking-wide bg-muted/50 dark:bg-muted/30"
          >
            {t(list.themeKey)}
          </Badge>

          {/* Title */}
          <h3 className="text-lg sm:text-xl font-semibold text-foreground mb-2 group-hover:text-primary transition-colors line-clamp-2">
            {t(list.titleKey)}
          </h3>

          {/* Meta info */}
          <div className="flex items-center gap-3 text-sm text-muted-foreground">
            <span className="flex items-center gap-1.5">
              <svg
                role="presentation"
                className="size-4"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth={1.5}
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M21 11.25v8.25a1.5 1.5 0 01-1.5 1.5H5.25a1.5 1.5 0 01-1.5-1.5v-8.25M12 4.875A2.625 2.625 0 109.375 7.5H12m0-2.625V7.5m0-2.625A2.625 2.625 0 1114.625 7.5H12m0 0V21m-8.625-9.75h18c.621 0 1.125-.504 1.125-1.125v-1.5c0-.621-.504-1.125-1.125-1.125h-18c-.621 0-1.125.504-1.125 1.125v1.5c0 .621.504 1.125 1.125 1.125z"
                />
              </svg>
              {list.itemCount} {t('featuredLists.items')}
            </span>
            <span className="text-muted-foreground/50">•</span>
            <span>{t('featuredLists.public')}</span>
          </div>

          {/* Arrow indicator */}
          <div className="absolute bottom-6 right-6 sm:bottom-8 sm:right-8 opacity-0 group-hover:opacity-100 transform translate-x-2 group-hover:translate-x-0 transition-all duration-300">
            <svg
              role="presentation"
              className="size-5 text-muted-foreground"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              strokeWidth={2}
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M17.25 8.25L21 12m0 0l-3.75 3.75M21 12H3"
              />
            </svg>
          </div>
        </CardContent>
      </Card>
    </Link>
  );
}
