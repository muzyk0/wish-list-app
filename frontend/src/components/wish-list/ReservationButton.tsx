"use client";

import { useState } from "react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

interface ReservationButtonProps {
  giftItemId: string;
  wishlistId: string;
  isReserved?: boolean;
  reservedByName?: string;
  onReservationSuccess?: () => void;
}

export function ReservationButton({
  giftItemId,
  wishlistId,
  isReserved = false,
  reservedByName,
  onReservationSuccess,
}: ReservationButtonProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [guestName, setGuestName] = useState("");
  const [guestEmail, setGuestEmail] = useState("");
  const [isLoading, setIsLoading] = useState(false);

  const handleReservation = async () => {
    if (!guestName.trim() || !guestEmail.trim()) {
      toast.error("Please enter your name and email");
      return;
    }

    setIsLoading(true);

    try {
      const response = await fetch(
        `/api/wishlists/${wishlistId}/items/${giftItemId}/reserve`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            guestName: guestName.trim(),
            guestEmail: guestEmail.trim(),
          }),
        },
      );

      if (response.ok) {
        toast.success("Gift item reserved successfully!");
        setIsOpen(false);
        setGuestName("");
        setGuestEmail("");
        onReservationSuccess?.();
      } else {
        const data = await response.json();
        toast.error(data.error || "Failed to reserve gift item");
      }
    } catch (_error) {
      toast.error("An error occurred while reserving the gift item");
    } finally {
      setIsLoading(false);
    }
  };

  if (isReserved) {
    return (
      <Button variant="outline" disabled className="cursor-not-allowed">
        Reserved by {reservedByName || "someone"}
      </Button>
    );
  }

  return (
    <>
      <Button onClick={() => setIsOpen(true)}>Reserve this gift</Button>

      <Dialog open={isOpen} onOpenChange={setIsOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Reserve this gift</DialogTitle>
            <DialogDescription>
              Enter your details to reserve this gift item. This will prevent
              others from reserving the same gift.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="guestName">Your Name</Label>
              <Input
                id="guestName"
                value={guestName}
                onChange={(e) => setGuestName(e.target.value)}
                placeholder="Enter your name"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="guestEmail">Your Email</Label>
              <Input
                id="guestEmail"
                type="email"
                value={guestEmail}
                onChange={(e) => setGuestEmail(e.target.value)}
                placeholder="Enter your email"
              />
            </div>
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setIsOpen(false)}
              disabled={isLoading}
            >
              Cancel
            </Button>
            <Button onClick={handleReservation} disabled={isLoading}>
              {isLoading ? "Reserving..." : "Reserve Gift"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
