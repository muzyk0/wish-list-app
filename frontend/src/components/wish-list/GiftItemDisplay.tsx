/** biome-ignore-all lint/correctness/noUnusedFunctionParameters: Temp */
import { ExternalLink, Heart, ShoppingCart } from "lucide-react";
import Image from "next/image";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";

interface GiftItemProps {
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
  onReserve?: () => void;
  onPurchase?: () => void;
  showActions?: boolean;
}

export default function GiftItemDisplay({
  id,
  name,
  description,
  link,
  image_url,
  price,
  priority,
  reserved_by_user_id,
  reserved_at,
  purchased_by_user_id,
  purchased_at,
  notes,
  position,
  created_at,
  updated_at,
  onReserve,
  onPurchase,
  showActions = true,
}: GiftItemProps) {
  const isReserved = !!reserved_by_user_id;
  const isPurchased = !!purchased_by_user_id;

  return (
    <Card className="overflow-hidden">
      <CardContent className="p-4">
        <div className="flex flex-col md:flex-row gap-4">
          {image_url ? (
            <div className="flex-shrink-0">
              <div className="relative w-16 h-16">
                <Image
                  src={image_url}
                  alt={name}
                  fill
                  className="object-cover rounded-md"
                  sizes="64px"
                />
              </div>
            </div>
          ) : (
            <div className="flex-shrink-0 flex items-center justify-center w-16 h-16 bg-muted rounded-md">
              <div className="h-6 w-6 text-muted-foreground">🎁</div>
            </div>
          )}

          <div className="flex-1 min-w-0">
            <div className="flex items-start justify-between">
              <h3 className="text-lg font-semibold truncate">{name}</h3>
              <div className="flex space-x-2 ml-2">
                {isPurchased && <Badge variant="destructive">Purchased</Badge>}
                {isReserved && !isPurchased && (
                  <Badge variant="secondary">Reserved</Badge>
                )}
              </div>
            </div>

            {description && (
              <p className="text-sm text-muted-foreground mt-1 line-clamp-2">
                {description}
              </p>
            )}

            {price && <p className="text-lg font-bold mt-2">${price}</p>}

            {link && (
              <a
                href={link}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center text-sm text-blue-600 hover:underline mt-2"
              >
                View Product <ExternalLink className="ml-1 h-4 w-4" />
              </a>
            )}

            {showActions && (
              <div className="mt-4 flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  {priority > 0 && (
                    <span className="text-xs bg-muted px-2 py-1 rounded">
                      Priority: {priority}/10
                    </span>
                  )}
                </div>

                <div className="flex space-x-2">
                  {onReserve && !isPurchased && (
                    <Button
                      variant={isReserved ? "secondary" : "default"}
                      size="sm"
                      onClick={onReserve}
                      disabled={isReserved}
                    >
                      {isReserved ? (
                        <>
                          <Heart className="mr-2 h-4 w-4 fill-current" />{" "}
                          Reserved
                        </>
                      ) : (
                        <>
                          <Heart className="mr-2 h-4 w-4" /> Reserve
                        </>
                      )}
                    </Button>
                  )}

                  {onPurchase && !isPurchased && (
                    <Button variant="outline" size="sm" onClick={onPurchase}>
                      <ShoppingCart className="mr-2 h-4 w-4" /> Purchase
                    </Button>
                  )}
                </div>
              </div>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
