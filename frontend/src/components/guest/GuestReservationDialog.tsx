"use client";

import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Heart } from "lucide-react";
import { toast } from "sonner";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { apiClient } from "@/lib/api/client";

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
  const [guestName, setGuestName] = useState("");
  const [guestEmail, setGuestEmail] = useState("");
  const queryClient = useQueryClient();

  const reservationMutation = useMutation({
    mutationFn: async () => {
      return apiClient.createReservation(wishlistId, itemId, {
        guest_name: guestName.trim(),
        guest_email: guestEmail.trim(),
      }) as Promise<ReservationResponse>;
    },
    onSuccess: (data) => {
      toast.success("Reservation Successful!", {
        description: `You have reserved "${itemName}". A confirmation has been sent to ${guestEmail}`,
      });

      // Save reservation token to localStorage for future reference
      const reservations = JSON.parse(
        localStorage.getItem("guest_reservations") || "[]",
      );
      reservations.push({
        itemId,
        itemName,
        reservationToken: data.reservation_token,
        reservedAt: data.reserved_at,
        guestName,
        guestEmail,
      });
      localStorage.setItem("guest_reservations", JSON.stringify(reservations));

      // Invalidate and refetch the wishlist query to update UI
      queryClient.invalidateQueries({
        queryKey: ["public-wishlist", wishlistSlug],
      });

      // Close dialog and reset form
      setOpen(false);
      setGuestName("");
      setGuestEmail("");
    },
    onError: (error: Error) => {
      toast.error("Reservation Failed", {
        description: error.message,
      });
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    // Basic validation
    if (!guestName.trim()) {
      toast.error("Validation Error", {
        description: "Please enter your name",
      });
      return;
    }

    if (!guestEmail.trim() || !guestEmail.includes("@")) {
      toast.error("Validation Error", {
        description: "Please enter a valid email address",
      });
      return;
    }

    reservationMutation.mutate();
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button
          variant={isReserved ? "secondary" : "default"}
          size="sm"
          disabled={isReserved || isPurchased}
        >
          {isPurchased ? (
            "Already Purchased"
          ) : isReserved ? (
            "Reserved"
          ) : (
            <>
              <Heart className="mr-2 h-4 w-4" /> Reserve Gift
            </>
          )}
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <form onSubmit={handleSubmit}>
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
                value={guestName}
                onChange={(e) => setGuestName(e.target.value)}
                required
                maxLength={200}
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="guest-email">Your Email</Label>
              <Input
                id="guest-email"
                type="email"
                placeholder="john@example.com"
                value={guestEmail}
                onChange={(e) => setGuestEmail(e.target.value)}
                required
              />
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
              {reservationMutation.isPending ? "Reserving..." : "Reserve Gift"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
