'use client';

import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Heart } from 'lucide-react';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { z } from 'zod';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { apiClient } from '@/lib/api/client';
import { addReservation } from '@/lib/guest-reservations';

type GuestReservationFormData = {
  guestName: string;
  guestEmail?: string;
};

interface GuestReservationDialogProps {
  wishlistSlug: string;
  wishlistId: string;
  itemId: string;
  itemName: string;
  isReserved: boolean;
  isPurchased: boolean;
}

interface ReservationResponse {
  id: string;
  gift_item_id: string;
  reservation_token: string;
  status: string;
  reserved_at: string;
}

export function GuestReservationDialog({
  wishlistSlug,
  wishlistId,
  itemId,
  itemName,
  isReserved,
  isPurchased,
}: GuestReservationDialogProps) {
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);
  const queryClient = useQueryClient();

  const schema = z.object({
    guestName: z
      .string()
      .trim()
      .min(1, t('reservation.dialog.validation.nameRequired'))
      .max(255, t('reservation.dialog.validation.nameTooLong')),
    guestEmail: z
      .string()
      .optional()
      .refine((value) => {
        if (!value) return true;
        const trimmed = value.trim();
        if (!trimmed) return true;
        return z.string().email().safeParse(trimmed).success;
      }, t('reservation.dialog.validation.emailInvalid')),
  });

  const {
    register,
    handleSubmit,
    reset,
    getValues,
    formState: { errors },
  } = useForm<GuestReservationFormData>({
    resolver: zodResolver(schema),
    defaultValues: { guestName: '', guestEmail: '' },
  });

  const reservationMutation = useMutation({
    mutationFn: async (data: GuestReservationFormData) => {
      return apiClient.createReservation(wishlistId, itemId, {
        guest_name: data.guestName.trim(),
        guest_email: data.guestEmail?.trim() || undefined,
      }) as Promise<ReservationResponse>;
    },
    onSuccess: (data) => {
      const { guestName, guestEmail } = getValues();

      toast.success(t('reservation.success.title'), {
        description: t('reservation.success.description', {
          itemName,
        }),
      });

      // Persist reservation token using localStorage utility (R-004, R-005)
      addReservation({
        itemId,
        itemName,
        reservationToken: data.reservation_token,
        reservedAt: data.reserved_at,
        guestName,
        guestEmail: guestEmail?.trim() || undefined,
        wishlistId,
      });

      // Invalidate gift items list so status badges update immediately (T017)
      queryClient.invalidateQueries({
        queryKey: ['public-gift-items', wishlistSlug],
      });
      queryClient.invalidateQueries({
        queryKey: ['guest-reservations'],
      });

      setOpen(false);
      reset();
    },
    onError: (error: Error) => {
      // Race condition: item was reserved by someone else between load and submit (T018)
      const isAlreadyReserved =
        error.message.toLowerCase().includes('already reserved') ||
        error.message.toLowerCase().includes('already been reserved') ||
        error.message.includes('409') ||
        error.message.includes('conflict');

      if (isAlreadyReserved) {
        toast.error(t('reservation.errors.failed'), {
          description: t('reservation.errors.alreadyReserved'),
        });
        // Refresh item list to show current reservation status
        queryClient.invalidateQueries({
          queryKey: ['public-gift-items', wishlistSlug],
        });
        setOpen(false);
        reset();
      } else {
        toast.error(t('reservation.errors.failed'), {
          description: error.message || t('reservation.errors.generic'),
        });
      }
    },
  });

  const onSubmit = (data: GuestReservationFormData) => {
    reservationMutation.mutate(data);
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button
          variant={isReserved ? 'secondary' : 'default'}
          size="sm"
          disabled={isReserved || isPurchased}
        >
          {isPurchased ? (
            t('publicWishlist.item.alreadyPurchased')
          ) : isReserved ? (
            t('publicWishlist.item.alreadyReserved')
          ) : (
            <>
              <Heart className="mr-2 h-4 w-4" />
              {t('publicWishlist.item.reserve')}
            </>
          )}
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <form onSubmit={handleSubmit(onSubmit)}>
          <DialogHeader>
            <DialogTitle>
              {t('reservation.dialog.title', { itemName })}
            </DialogTitle>
            <DialogDescription>
              {t('reservation.dialog.description')}
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="guest-name">
                {t('reservation.dialog.nameLabel')}
              </Label>
              <Input
                id="guest-name"
                placeholder={t('reservation.dialog.namePlaceholder')}
                {...register('guestName')}
                maxLength={255}
                aria-invalid={!!errors.guestName}
              />
              {errors.guestName && (
                <p className="text-sm text-red-500">
                  {errors.guestName.message}
                </p>
              )}
            </div>
            <div className="grid gap-2">
              <Label htmlFor="guest-email">
                {t('reservation.dialog.emailLabel')}
              </Label>
              <Input
                id="guest-email"
                type="email"
                placeholder={t('reservation.dialog.emailPlaceholder')}
                {...register('guestEmail')}
                aria-invalid={!!errors.guestEmail}
              />
              {errors.guestEmail && (
                <p className="text-sm text-red-500">
                  {errors.guestEmail.message}
                </p>
              )}
              <p className="text-sm text-muted-foreground">
                {t('reservation.dialog.emailHint')}
              </p>
            </div>
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => setOpen(false)}
            >
              {t('reservation.dialog.cancel')}
            </Button>
            <Button type="submit" disabled={reservationMutation.isPending}>
              {reservationMutation.isPending
                ? t('reservation.dialog.submitting')
                : t('reservation.dialog.submit')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
