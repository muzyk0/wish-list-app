'use client';

import { useTranslation } from 'react-i18next';
import { MobileRedirect } from '@/features/mobile-redirect';
import { MyReservationsList } from '@/features/my-reservations';
import {
  MOBILE_APP_REDIRECT_PATHS,
  MOBILE_APP_URLS,
} from '@/shared/config/domains';
import { useAuthRedirect } from '@/shared/lib/auth';

export function MyReservationsPage() {
  const { t } = useTranslation();
  const { isAuthenticated, isLoading } = useAuthRedirect(true);

  if (isLoading) {
    return (
      <div className="container mx-auto py-10">
        <div className="flex justify-center items-center h-32">
          <p className="text-muted-foreground">{t('myReservations.loading')}</p>
        </div>
      </div>
    );
  }

  // Authenticated users are redirected to the mobile app for account management
  if (isAuthenticated) {
    return (
      <MobileRedirect
        redirectPath={MOBILE_APP_REDIRECT_PATHS.MY_RESERVATIONS}
        fallbackUrl={MOBILE_APP_URLS.MY_RESERVATIONS}
      >
        <div className="container mx-auto py-10">
          <div className="flex justify-center items-center h-32">
            <p className="text-muted-foreground">
              {t('myReservations.loading')}
            </p>
          </div>
        </div>
      </MobileRedirect>
    );
  }

  return (
    <div className="container mx-auto py-10 px-4">
      <h1 className="text-3xl font-bold mb-8">{t('myReservations.title')}</h1>
      <MyReservationsList />
    </div>
  );
}
