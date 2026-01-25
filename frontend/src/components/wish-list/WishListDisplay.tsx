'use client';

import { ExternalLink, Eye, Share2 } from 'lucide-react';
import Image from 'next/image';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { ReservationButton } from './ReservationButton';

interface GiftItem {
  id: string;
  name: string;
  description: string;
  link: string;
  image_url: string;
  price: string;
  priority: number;
  reserved_by_user_id: string;
  reserved_at: string;
  purchased_by_user_id: string;
  purchased_at: string;
  notes: string;
  position: number;
  created_at: string;
  updated_at: string;
}

interface WishList {
  id: string;
  owner_id: string;
  title: string;
  description: string;
  occasion: string;
  occasion_date: string;
  template_id: string;
  is_public: boolean;
  public_slug: string;
  view_count: number;
  created_at: string;
  updated_at: string;
  gift_items: GiftItem[];
}

interface WishListDisplayProps {
  wishList: WishList;
  showActions?: boolean; // Whether to show action buttons (edit, etc.) - typically for owner
  onReserve?: (itemId: string) => void; // Handler for reservation actions
}

export default function WishListDisplay({
  wishList,
  showActions = false,
  onReserve,
}: WishListDisplayProps) {
  // Filter out purchased items if we want to hide them
  const availableItems = wishList.gift_items.filter(
    (item) => !item.purchased_by_user_id,
  );

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-2">
            <div>
              <CardTitle className="text-2xl">{wishList.title}</CardTitle>
              {wishList.occasion && (
                <p className="text-lg text-muted-foreground mt-1">
                  {wishList.occasion}
                </p>
              )}
            </div>

            {showActions && (
              <div className="flex space-x-2 sm:ml-auto">
                <Button variant="outline" size="sm">
                  Edit
                </Button>
                <Button variant="outline" size="sm">
                  Share
                </Button>
              </div>
            )}
          </div>

          {wishList.description && (
            <p className="text-muted-foreground mt-2">{wishList.description}</p>
          )}
        </CardHeader>

        <CardContent>
          <div className="flex flex-wrap gap-2 mb-4">
            <Badge variant={wishList.is_public ? 'default' : 'secondary'}>
              {wishList.is_public ? 'Public' : 'Private'}
            </Badge>
            {wishList.occasion_date && (
              <Badge variant="outline">
                {new Date(wishList.occasion_date).toLocaleDateString()}
              </Badge>
            )}
            <Badge variant="outline">{availableItems.length} items</Badge>
            {wishList.view_count > 0 && (
              <Badge variant="outline" className="flex items-center">
                <Eye className="mr-1 h-3 w-3" /> {wishList.view_count} views
              </Badge>
            )}
          </div>

          {wishList.public_slug && wishList.is_public && (
            <div className="mb-4 p-3 bg-muted rounded-md flex items-center justify-between">
              <div>
                <p className="text-sm font-medium">Share this list</p>
                <p className="text-xs text-muted-foreground break-all">
                  {`${window.location.origin}/public/${wishList.public_slug}`}
                </p>
              </div>
              <Button
                variant="outline"
                size="sm"
                onClick={() =>
                  navigator.clipboard.writeText(
                    `${window.location.origin}/public/${wishList.public_slug}`,
                  )
                }
              >
                <Share2 className="h-4 w-4" />
              </Button>
            </div>
          )}
        </CardContent>
      </Card>

      <div className="space-y-4">
        {availableItems.length === 0 ? (
          <Card>
            <CardContent className="p-8 text-center">
              <p className="text-muted-foreground">
                No gift items in this list yet.
              </p>
            </CardContent>
          </Card>
        ) : (
          availableItems
            .sort((a, b) => a.position - b.position)
            .map((item) => {
              const isReserved = !!item.reserved_by_user_id;

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
                          <div className="h-6 w-6 text-muted-foreground">
                            🎁
                          </div>
                        </div>
                      )}

                      <div className="flex-1 min-w-0">
                        <div className="flex items-start justify-between">
                          <h3 className="text-lg font-semibold truncate">
                            {item.name}
                          </h3>
                          <div className="flex space-x-2 ml-2">
                            {isReserved && (
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
                            View Product{' '}
                            <ExternalLink className="ml-1 h-4 w-4" />
                          </a>
                        )}

                        <div className="mt-4 flex items-center justify-between">
                          <div className="flex items-center space-x-2">
                            {item.priority > 0 && (
                              <span className="text-xs bg-muted px-2 py-1 rounded">
                                Priority: {item.priority}/10
                              </span>
                            )}
                          </div>

                          {onReserve && (
                            <ReservationButton
                              giftItemId={item.id}
                              wishlistId={wishList.id}
                              isReserved={isReserved}
                              reservedByName={
                                item.reserved_by_user_id ? 'Someone' : undefined
                              }
                              onReservationSuccess={() => onReserve(item.id)}
                            />
                          )}
                        </div>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              );
            })
        )}
      </div>
    </div>
  );
}
