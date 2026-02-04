'use client';

import { useTranslation } from 'react-i18next';
import { FeaturedListCard } from './FeaturedListCard';

export interface FeaturedListData {
  titleKey: string;
  themeKey: string;
  slug: string;
  itemCount: number;
  gradient: string;
  icon: string;
}

const FEATURED_LISTS: FeaturedListData[] = [
  {
    titleKey: 'featuredLists.lists.birthday.title',
    themeKey: 'featuredLists.lists.birthday.theme',
    slug: 'sarah-birthday-2026',
    itemCount: 12,
    gradient: 'from-rose-500 to-pink-500',
    icon: '🎂',
  },
  {
    titleKey: 'featuredLists.lists.babyShower.title',
    themeKey: 'featuredLists.lists.babyShower.theme',
    slug: 'baby-martinez-2026',
    itemCount: 24,
    gradient: 'from-sky-500 to-blue-500',
    icon: '👶',
  },
  {
    titleKey: 'featuredLists.lists.wedding.title',
    themeKey: 'featuredLists.lists.wedding.theme',
    slug: 'emma-james-wedding',
    itemCount: 18,
    gradient: 'from-amber-500 to-orange-500',
    icon: '💍',
  },
];

export function FeaturedListsSection() {
  const { t } = useTranslation();

  return (
    <section className="py-16 sm:py-24 px-4 sm:px-6 bg-muted/30 dark:bg-muted/10">
      <div className="max-w-6xl mx-auto">
        {/* Section header */}
        <div className="text-center mb-12 sm:mb-16">
          <h2 className="text-2xl sm:text-3xl md:text-4xl font-bold text-foreground mb-4">
            {t('featuredLists.title')}
          </h2>
          <p className="text-muted-foreground text-base sm:text-lg max-w-2xl mx-auto">
            {t('featuredLists.subtitle')}
          </p>
        </div>

        {/* Cards grid */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 sm:gap-6">
          {FEATURED_LISTS.map((list, index) => (
            <FeaturedListCard key={list.slug} list={list} index={index} />
          ))}
        </div>

        {/* View more hint */}
        <div className="mt-12 text-center">
          <p className="text-sm text-muted-foreground">
            {t('featuredLists.createHint')}{' '}
            <span className="font-medium text-foreground">
              {t('featuredLists.mobileApp')}
            </span>
          </p>
        </div>
      </div>
    </section>
  );
}
