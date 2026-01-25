'use client';

import { MyReservations } from '@/components/wish-list/MyReservations';
import MobileRedirect from '@/components/common/MobileRedirect';
import { useAuthRedirect } from '@/hooks/useAuthRedirect';

export default function MyReservationsPage() {
  const { isAuthenticated, isLoading } = useAuthRedirect(true);

  // If still checking authentication, show loading state
  if (isLoading) {
    return (
      <div className="container mx-auto py-10">
        <div className="flex justify-center items-center h-32">
          <p>Loading...</p>
        </div>
      </div>
    );
  }

  // If authenticated user, redirect to mobile app for account management
  if (isAuthenticated) {
    return (
      <MobileRedirect
        redirectPath="my/reservations"
        fallbackUrl="https://lk.domain.com/my/reservations"
      >
        <div className="container mx-auto py-10">
          <div className="flex justify-center items-center h-32">
            <p>Redirecting to mobile app for account access...</p>
          </div>
        </div>
      </MobileRedirect>
    );
  }

  // Guest users can view their reservations in the frontend
  return (
    <div className="container mx-auto py-10">
      <h1 className="text-3xl font-bold mb-8">My Reservations</h1>
      <MyReservations />
    </div>
  );
}
