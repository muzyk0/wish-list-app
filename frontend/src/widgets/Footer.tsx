'use client';

import Link from 'next/link';
import { useTranslation } from 'react-i18next';

export function Footer() {
  const { t } = useTranslation();

  return (
    <footer className="py-12 px-4 sm:px-6 border-t border-border/50">
      <div className="max-w-6xl mx-auto">
        <div className="flex flex-col sm:flex-row items-center justify-between gap-6">
          {/* Left side */}
          <div className="flex items-center gap-3">
            <div className="size-8 rounded-lg bg-gradient-to-br from-amber-500 to-orange-500 flex items-center justify-center">
              <svg
                className="size-4 text-white"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                strokeWidth={2}
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M21 11.25v8.25a1.5 1.5 0 01-1.5 1.5H5.25a1.5 1.5 0 01-1.5-1.5v-8.25M12 4.875A2.625 2.625 0 109.375 7.5H12m0-2.625V7.5m0-2.625A2.625 2.625 0 1114.625 7.5H12m0 0V21m-8.625-9.75h18c.621 0 1.125-.504 1.125-1.125v-1.5c0-.621-.504-1.125-1.125-1.125h-18c-.621 0-1.125.504-1.125 1.125v1.5c0 .621.504 1.125 1.125 1.125z"
                />
              </svg>
            </div>
            <p className="text-sm text-muted-foreground">
              {t('footer.description.line1')}
              <br className="hidden sm:block" />
              <span className="sm:hidden"> </span>
              {t('footer.description.line2')}
            </p>
          </div>

          {/* Right side */}
          <div className="flex items-center gap-6 text-sm">
            <Link
              href="/auth/login"
              className="text-muted-foreground hover:text-foreground transition-colors"
            >
              {t('footer.login')}
            </Link>
            <Link
              href="/my/reservations"
              className="text-muted-foreground hover:text-foreground transition-colors"
            >
              {t('footer.myReservations')}
            </Link>
          </div>
        </div>

        {/* Bottom */}
        <div className="mt-8 pt-6 border-t border-border/30 flex items-center justify-center text-sm text-muted-foreground/60">
          <span>{t('footer.crafted')}</span>
          <span className="mx-2 text-amber-500">✦</span>
          <span>© 2026</span>
        </div>
      </div>
    </footer>
  );
}
