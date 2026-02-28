'use client';

import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { apiClient } from '@/lib/api/client';
import {
  getStoredReservations,
  removeReservation,
} from '@/lib/guest-reservations';
import type { ReservationDetailsResponse } from '@/lib/api/types';

type ReservationWithToken = ReservationDetailsResponse & {
  reservationToken: string;
};

function getStatusVariant(
  status: string,
): 'default' | 'secondary' | 'outline' | 'destructive' {
  switch (status) {
    case 'active':
      return 'default';
    case 'canceled':
      return 'secondary';
    case 'fulfilled':
      return 'outline';
    default:
      return 'destructive';
  }
}

export function MyReservations() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [confirmingId, setConfirmingId] = useState<string | null>(null);

  const {
    data: reservations = [],
    isLoading,
    isError,
    refetch,
  } = useQuery({
    queryKey: ['guest-reservations'],
    queryFn: async (): Promise<ReservationWithToken[]> => {
      const stored = getStoredReservations();
      if (stored.length === 0) return [];

      const results = await Promise.allSettled(
        stored.map(async (s) => {
          const details = await apiClient.getGuestReservations(
            s.reservationToken,
          );
          return details.map((d) => ({
            ...d,
            reservationToken: s.reservationToken,
          }));
        }),
      );

      const successful = results.flatMap((result) =>
        result.status === 'fulfilled' ? result.value : [],
      );
      const hasFailures = results.some(
        (result) => result.status === 'rejected',
      );

      if (successful.length === 0 && hasFailures) {
        throw new Error('Failed to load guest reservations');
      }

      return successful;
    },
    staleTime: 30_000,
  });

  const cancelMutation = useMutation({
    mutationFn: async ({
      reservation,
      token,
    }: {
      reservation: ReservationDetailsResponse;
      token: string;
    }) => {
      return apiClient.cancelReservation(
        reservation.wishlist.id,
        reservation.gift_item.id,
        { reservation_token: token },
      );
    },
    onSuccess: (_, { token }) => {
      removeReservation(token);
      toast.success(t('myReservations.cancel.success'));
      queryClient.invalidateQueries({ queryKey: ['guest-reservations'] });
      setConfirmingId(null);
    },
    onError: (error, { token }) => {
      const message = (error as Error).message.toLowerCase();
      if (message.includes('invalid') || message.includes('expired')) {
        // T025: Remove invalid/expired token from localStorage
        removeReservation(token);
        toast.error(t('myReservations.cancel.invalidToken'));
        queryClient.invalidateQueries({ queryKey: ['guest-reservations'] });
      } else {
        toast.error(t('myReservations.cancel.error'));
      }
      setConfirmingId(null);
    },
  });

  if (isLoading) {
    return (
      <div className="space-y-4">
        {[1, 2, 3].map((i) => (
          <Card key={i}>
            <CardHeader className="pb-2">
              <Skeleton className="h-5 w-48" />
              <Skeleton className="h-4 w-64 mt-1" />
            </CardHeader>
            <CardContent>
              <Skeleton className="h-4 w-32" />
            </CardContent>
          </Card>
        ))}
      </div>
    );
  }

  if (isError) {
    return (
      <Card>
        <CardContent className="pt-6 text-center space-y-4">
          <p className="text-destructive">
            {t('myReservations.errors.loadFailed')}
          </p>
          <Button variant="outline" onClick={() => refetch()}>
            {t('myReservations.errors.retry')}
          </Button>
        </CardContent>
      </Card>
    );
  }

  if (reservations.length === 0) {
    return (
      <Card>
        <CardContent className="pt-6 text-center space-y-2">
          <p className="font-medium text-lg">
            {t('myReservations.empty.title')}
          </p>
          <p className="text-muted-foreground">
            {t('myReservations.empty.description')}
          </p>
          <p className="text-sm text-muted-foreground/70">
            {t('myReservations.empty.hint')}
          </p>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-4">
      {reservations.map((reservation) => {
        const isConfirming = confirmingId === reservation.id;
        const owner = [
          reservation.wishlist.owner_first_name,
          reservation.wishlist.owner_last_name,
        ]
          .filter(Boolean)
          .join(' ');

        return (
          <Card key={reservation.id}>
            <CardHeader className="pb-2">
              <div className="flex justify-between items-start gap-2">
                <CardTitle className="text-lg">
                  {reservation.gift_item.name}
                </CardTitle>
                <Badge variant={getStatusVariant(reservation.status)}>
                  {t(
                    `myReservations.item.status.${reservation.status}` as Parameters<
                      typeof t
                    >[0],
                    { defaultValue: reservation.status },
                  )}
                </Badge>
              </div>
              <p className="text-sm text-muted-foreground">
                {t('myReservations.item.reservedFor', {
                  wishlist: reservation.wishlist.title,
                  owner: owner || '—',
                })}
              </p>
            </CardHeader>

            <CardContent>
              <div className="flex items-center justify-between flex-wrap gap-4">
                <div className="text-sm text-muted-foreground space-y-0.5">
                  <p>
                    {t('myReservations.item.reservedOn', {
                      date: new Date(
                        reservation.reserved_at,
                      ).toLocaleDateString(),
                    })}
                  </p>
                  {reservation.expires_at && (
                    <p>
                      {t('myReservations.item.expiresOn', {
                        date: new Date(
                          reservation.expires_at,
                        ).toLocaleDateString(),
                      })}
                    </p>
                  )}
                </div>

                {reservation.status === 'active' && (
                  <div className="flex items-center gap-2 flex-wrap">
                    {isConfirming ? (
                      <>
                        <span className="text-sm text-muted-foreground">
                          {t('myReservations.item.cancelConfirm')}
                        </span>
                        <Button
                          variant="destructive"
                          size="sm"
                          disabled={cancelMutation.isPending}
                          onClick={() =>
                            cancelMutation.mutate({
                              reservation,
                              token: reservation.reservationToken,
                            })
                          }
                        >
                          {t('myReservations.item.cancelConfirmYes')}
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          disabled={cancelMutation.isPending}
                          onClick={() => setConfirmingId(null)}
                        >
                          {t('myReservations.item.cancelConfirmNo')}
                        </Button>
                      </>
                    ) : (
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => setConfirmingId(reservation.id)}
                      >
                        {t('myReservations.item.cancelButton')}
                      </Button>
                    )}
                  </div>
                )}
              </div>
            </CardContent>
          </Card>
        );
      })}
    </div>
  );
}
