'use client';

import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { zodResolver } from '@hookform/resolvers/zod';
import { useForm } from 'react-hook-form';
import { Heart } from 'lucide-react';
import { toast } from 'sonner';
import { z } from 'zod';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { apiClient } from '@/lib/api/client';

// Zod schema for guest reservation form validation
const guestReservationSchema = z.object({
  guestName: z
    .string()
    .min(1, 'Name is required')
    .max(255, 'Name must be less than 255 characters'),
  guestEmail: z
    .string()
    .min(1, 'Email is required')
    .email('Invalid email address'),
});

type GuestReservationFormData = z.infer<typeof guestReservationSchema>;

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
  const [open, setOpen] = useState(false);
  const queryClient = useQueryClient();

  const {
    register,
    handleSubmit,
    reset,
    getValues,
    formState: { errors },
  } = useForm<GuestReservationFormData>({
    resolver: zodResolver(guestReservationSchema),
    defaultValues: {
      guestName: '',
      guestEmail: '',
    },
  });

  const reservationMutation = useMutation({
    mutationFn: async (data: GuestReservationFormData) => {
      return apiClient.createReservation(wishlistId, itemId, {
        guest_name: data.guestName.trim(),
        guest_email: data.guestEmail.trim(),
      }) as Promise<ReservationResponse>;
    },
    onSuccess: (data) => {
      const { guestName, guestEmail } = getValues();
      toast.success('Reservation Successful!', {
        description: `You have reserved "${itemName}". A confirmation has been sent to ${guestEmail}`,
      });

      // Save reservation token to localStorage for future reference
      const reservations = JSON.parse(
        localStorage.getItem('guest_reservations') || '[]',
      );
      reservations.push({
        itemId,
        itemName,
        reservationToken: data.reservation_token,
        reservedAt: data.reserved_at,
        guestName,
        guestEmail,
      });
      localStorage.setItem('guest_reservations', JSON.stringify(reservations));

      // Invalidate and refetch the wishlist query to update UI
      queryClient.invalidateQueries({
        queryKey: ['public-wishlist', wishlistSlug],
      });

      // Close dialog and reset form
      setOpen(false);
      reset();
    },
    onError: (error: Error) => {
      toast.error('Reservation Failed', {
        description: error.message,
      });
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
            'Already Purchased'
          ) : isReserved ? (
            'Reserved'
          ) : (
            <>
              <Heart className="mr-2 h-4 w-4" /> Reserve Gift
            </>
          )}
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <form onSubmit={handleSubmit(onSubmit)}>
          <DialogHeader>
            <DialogTitle>Reserve "{itemName}"</DialogTitle>
            <DialogDescription>
              Enter your name and email to reserve this gift. You'll receive a
              confirmation with a reservation token.
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="guest-name">Your Name</Label>
              <Input
                id="guest-name"
                placeholder="John Doe"
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
              <Label htmlFor="guest-email">Your Email</Label>
              <Input
                id="guest-email"
                type="email"
                placeholder="john@example.com"
                {...register('guestEmail')}
                aria-invalid={!!errors.guestEmail}
              />
              {errors.guestEmail && (
                <p className="text-sm text-red-500">
                  {errors.guestEmail.message}
                </p>
              )}
              <p className="text-sm text-muted-foreground">
                We'll send you a confirmation email with your reservation token
              </p>
            </div>
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => setOpen(false)}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={reservationMutation.isPending}>
              {reservationMutation.isPending ? 'Reserving...' : 'Reserve Gift'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
