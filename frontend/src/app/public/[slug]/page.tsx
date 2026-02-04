"use client";

import { useQuery } from "@tanstack/react-query";
import { ExternalLink, Image as ImageIcon } from "lucide-react";
import Image from "next/image";
import { useParams } from "next/navigation";
import { GuestReservationDialog } from "@/components/guest/GuestReservationDialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import {
  DOMAIN_CONSTANTS,
  MOBILE_APP_REDIRECT_PATHS,
} from "@/constants/domains";
import { apiClient } from "@/lib/api/client";

export default function PublicWishListPage() {
  const { slug } = useParams<{ slug: string }>();

  const {
    data: wishList,
    isLoading: isLoadingWishList,
    isError: isErrorWishList,
  } = useQuery({
    queryKey: ["public-wishlist", slug],
    queryFn: async () => {
      return apiClient.getPublicWishList(slug);
    },
    enabled: !!slug,
    retry: 1,
  });

  const {
    data: giftItemsData,
    isLoading: isLoadingGiftItems,
    isError: isErrorGiftItems,
  } = useQuery({
    queryKey: ["public-gift-items", slug],
    queryFn: async () => {
      return apiClient.getPublicGiftItems(slug, 1, 100);
    },
    enabled: !!slug && !!wishList,
    retry: 1,
  });

  const isLoading = isLoadingWishList || isLoadingGiftItems;
  const isError = isErrorWishList || isErrorGiftItems;
  const giftItems = giftItemsData?.items || [];

  if (isLoading) {
    return (
      <div className="container mx-auto py-8 px-4">
        <Skeleton className="h-10 w-1/2 mb-6" />
        <Skeleton className="h-6 w-1/4 mb-4" />
        <Skeleton className="h-4 w-2/3 mb-8" />

        {[...Array(5)].map((_, i) => (
          // biome-ignore lint/suspicious/noArrayIndexKey: Mock
          <Card key={i} className="mb-4">
            <CardContent className="p-4">
              <div className="flex space-x-4">
                <Skeleton className="h-16 w-16 rounded" />
                <div className="flex-1 space-y-2">
                  <Skeleton className="h-4 w-3/4" />
                  <Skeleton className="h-4 w-1/2" />
                  <Skeleton className="h-4 w-1/4" />
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    );
  }

  if (isError || !wishList) {
    return (
      <div className="container mx-auto py-8 px-4">
        <Card>
          <CardHeader>
            <CardTitle>Error Loading Wishlist</CardTitle>
          </CardHeader>
          <CardContent>
            <p>
              Could not load the requested wishlist. Please try again later.
            </p>
            <Button onClick={() => window.location.reload()} className="mt-4">
              Retry
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="container mx-auto py-8 px-4">
      <Card className="mb-6">
        <CardHeader>
          <CardTitle className="text-2xl">{wishList.title}</CardTitle>
          {wishList.occasion && (
            <p className="text-lg text-muted-foreground">{wishList.occasion}</p>
          )}
        </CardHeader>
        <CardContent>
          {wishList.description && (
            <p className="text-muted-foreground mb-4">{wishList.description}</p>
          )}
          <div className="flex items-center space-x-2">
            <Badge variant="secondary">Public List</Badge>
            {wishList.view_count !== "0" && (
              <Badge variant="outline">{wishList.view_count} views</Badge>
            )}
          </div>
        </CardContent>
      </Card>

      <div className="space-y-4">
        {giftItems.length === 0 ? (
          <Card>
            <CardContent className="p-8 text-center">
              <p className="text-muted-foreground">
                No gift items in this list yet.
              </p>
            </CardContent>
          </Card>
        ) : (
          giftItems
            .sort((a, b) => (a.position || 0) - (b.position || 0))
            .map((item) => {
              const isReserved = !!item.reserved_by_user_id;
              const isPurchased = !!item.purchased_by_user_id;

              return (
                <Card key={item.id}>
                  <CardContent className="p-4">
                    <div className="flex flex-col md:flex-row gap-4">
                      {item.image_url ? (
                        <div className="flex-shrink-0">
                          <Image
                            src={item.image_url}
                            alt={item.name}
                            width={16}
                            height={16}
                            className="w-16 h-16 object-cover rounded-md"
                          />
                        </div>
                      ) : (
                        <div className="flex-shrink-0 flex items-center justify-center w-16 h-16 bg-muted rounded-md">
                          <ImageIcon className="h-6 w-6 text-muted-foreground" />
                        </div>
                      )}

                      <div className="flex-1 min-w-0">
                        <div className="flex items-start justify-between">
                          <h3 className="text-lg font-semibold truncate">
                            {item.name}
                          </h3>
                          <div className="flex space-x-2 ml-2">
                            {isPurchased && (
                              <Badge variant="destructive">Purchased</Badge>
                            )}
                            {isReserved && !isPurchased && (
                              <Badge variant="secondary">Reserved</Badge>
                            )}
                          </div>
                        </div>

                        {item.description && (
                          <p className="text-sm text-muted-foreground mt-1 line-clamp-2">
                            {item.description}
                          </p>
                        )}

                        {item.price && (
                          <p className="text-lg font-bold mt-2">
                            ${item.price}
                          </p>
                        )}

                        {item.link && (
                          <a
                            href={item.link}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="inline-flex items-center text-sm text-blue-600 hover:underline mt-2"
                          >
                            View Product{" "}
                            <ExternalLink className="ml-1 h-4 w-4" />
                          </a>
                        )}

                        <div className="mt-4 flex items-center justify-between">
                          <div className="flex items-center space-x-2">
                            {item.priority && item.priority > 0 && (
                              <span className="text-xs bg-muted px-2 py-1 rounded">
                                Priority: {item.priority}/10
                              </span>
                            )}
                          </div>

                          <GuestReservationDialog
                            wishlistSlug={slug}
                            wishlistId={wishList.id}
                            itemId={item.id}
                            itemName={item.name}
                            isReserved={isReserved}
                            isPurchased={isPurchased}
                          />
                        </div>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              );
            })
        )}
      </div>

      {/* Redirect to mobile app suggestion */}
      <div className="mt-8 p-4 bg-blue-50 rounded-lg border border-blue-200">
        <h3 className="font-semibold text-blue-900 mb-2">
          Manage your own wishlists?
        </h3>
        <p className="text-blue-700 mb-3">
          Download our mobile app to create and manage your own wishlists!
        </p>
        <Button
          variant="outline"
          onClick={() => {
            // Redirect to mobile app or mobile web version
            const appScheme = `wishlistapp://${MOBILE_APP_REDIRECT_PATHS.HOME}`;
            const webFallback = DOMAIN_CONSTANTS.MOBILE_APP_BASE_URL;

            window.location.href = appScheme;

            setTimeout(() => {
              // Only redirect if page is still visible (app didn't open)
              if (!document.hidden && document.visibilityState !== "hidden") {
                window.location.href = webFallback;
              }
            }, 1500);
          }}
        >
          Open Mobile App
        </Button>
      </div>
    </div>
  );
}
