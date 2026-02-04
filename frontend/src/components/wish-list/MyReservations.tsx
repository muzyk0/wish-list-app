"use client";

import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

interface Reservation {
  id: string;
  giftItem: {
    id: string;
    name: string;
    imageUrl?: string;
    price?: number;
  };
  wishlist: {
    id: string;
    title: string;
    ownerFirstName?: string;
    ownerLastName?: string;
  };
  status: "active" | "cancelled" | "fulfilled" | "expired";
  reservedAt: string;
  expiresAt?: string;
}

export function MyReservations() {
  const [reservations, setReservations] = useState<Reservation[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchReservations = useCallback(async () => {
    try {
      setLoading(true);

      // Check if user is authenticated
      const userResponse = await fetch("/api/auth/me");
      if (userResponse.ok) {
        // Authenticated user - fetch user reservations
        const response = await fetch("/api/users/me/reservations");
        if (response.ok) {
          const data = await response.json();
          setReservations(data.data || []);
        }
      } else {
        // Guest user - check for reservation token in localStorage
        const token = localStorage.getItem("reservationToken");
        if (token) {
          const response = await fetch(
            `/api/guest/reservations?token=${token}`,
          );
          if (response.ok) {
            const data = await response.json();
            setReservations(data || []);
          }
        }
      }
    } catch (_error) {
      toast.error("Failed to load reservations");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchReservations();
  }, [fetchReservations]);

  if (loading) {
    return (
      <div className="flex justify-center items-center h-32">
        <p>Loading reservations...</p>
      </div>
    );
  }

  if (reservations.length === 0) {
    return (
      <Card>
        <CardContent className="pt-6">
          <p className="text-center text-gray-500">
            You have no reservations yet.
          </p>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-4">
      <h2 className="text-xl font-bold">My Reservations</h2>

      {reservations.map((reservation) => (
        <Card key={reservation.id}>
          <CardHeader className="pb-2">
            <div className="flex justify-between items-start">
              <CardTitle className="text-lg">
                {reservation.giftItem.name}
              </CardTitle>
              <Badge
                variant={
                  reservation.status === "active"
                    ? "default"
                    : reservation.status === "cancelled"
                      ? "secondary"
                      : reservation.status === "fulfilled"
                        ? "outline"
                        : "destructive"
                }
              >
                {reservation.status.charAt(0).toUpperCase() +
                  reservation.status.slice(1)}
              </Badge>
            </div>
            <p className="text-sm text-gray-500">
              Reserved for: {reservation.wishlist.title} by{" "}
              {reservation.wishlist.ownerFirstName}{" "}
              {reservation.wishlist.ownerLastName}
            </p>
          </CardHeader>

          <CardContent>
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm">
                  Reserved on:{" "}
                  {new Date(reservation.reservedAt).toLocaleDateString()}
                </p>
                {reservation.expiresAt && (
                  <p className="text-sm text-gray-500">
                    Expires on:{" "}
                    {new Date(reservation.expiresAt).toLocaleDateString()}
                  </p>
                )}
              </div>

              {reservation.status === "active" && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={async () => {
                    try {
                      const response = await fetch(
                        `/api/reservations/${reservation.id}/cancel`,
                        {
                          method: "POST",
                          headers: {
                            "Content-Type": "application/json",
                          },
                        },
                      );

                      if (response.ok) {
                        toast.success("Reservation cancelled successfully");
                        fetchReservations(); // Refresh the list
                      } else {
                        const data = await response.json();
                        toast.error(
                          data.error || "Failed to cancel reservation",
                        );
                      }
                    } catch (_error) {
                      toast.error(
                        "An error occurred while cancelling the reservation",
                      );
                    }
                  }}
                >
                  Cancel Reservation
                </Button>
              )}
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  );
}
