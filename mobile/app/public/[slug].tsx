import { useQuery } from '@tanstack/react-query';
import { useLocalSearchParams } from 'expo-router';
import {
  ActivityIndicator,
  FlatList,
  RefreshControl,
  StyleSheet,
  Text,
  View,
} from 'react-native';
import { ReservationButton } from '@/components/wish-list/ReservationButton';
import { apiClient } from '@/lib/api';

import type {
  GiftItem as GiftItemType,
  WishList as WishListType,
} from '../../lib/types';

type GiftItem = GiftItemType;
type WishList = WishListType & { giftItems?: GiftItem[] }; // Extend to include giftItems if needed

export default function PublicWishListScreen() {
  const { slug } = useLocalSearchParams<{ slug: string }>();

  const {
    data: wishList,
    isLoading,
    isError,
    refetch,
    isRefetching,
  } = useQuery<WishList>({
    queryKey: ['public-wishlist', slug],
    queryFn: async () => {
      const resp = await apiClient.getPublicWishList(slug as string);
      // Normalize response: API returns 'items', component expects 'giftItems'
      return { ...resp, giftItems: (resp as any).items ?? [] } as WishList;
    },
    enabled: !!slug,
    retry: 1,
  });

  if (isLoading) {
    return (
      <View style={styles.centerContainer}>
        <ActivityIndicator size="large" color="#007AFF" />
      </View>
    );
  }

  if (isError || !wishList) {
    return (
      <View style={styles.centerContainer}>
        <Text>Failed to load wishlist</Text>
        <Text>Error occurred while fetching the wishlist</Text>
      </View>
    );
  }

  const renderGiftItem = ({ item }: { item: GiftItem }) => {
    const isReserved = !!item.reserved_by_user_id;
    const isPurchased = !!item.purchased_by_user_id;

    return (
      <View style={styles.itemContainer}>
        <View style={styles.itemHeader}>
          <Text style={styles.itemName}>{item.name}</Text>
          {isPurchased && <Text style={styles.purchasedBadge}>Purchased</Text>}
          {isReserved && !isPurchased && (
            <Text style={styles.reservedBadge}>Reserved</Text>
          )}
        </View>

        {item.description ? (
          <Text style={styles.itemDescription}>{item.description}</Text>
        ) : null}

        {item.price !== 0 && item.price !== undefined && (
          <Text style={styles.itemPrice}>${item.price}</Text>
        )}

        {item.link ? (
          <Text
            style={styles.itemLink}
            numberOfLines={1}
            ellipsizeMode="middle"
          >
            {item.link}
          </Text>
        ) : null}

        <View style={styles.itemActions}>
          <ReservationButton
            giftItemId={item.id}
            wishlistId={wishList.id}
            isReserved={isReserved}
            onReservationSuccess={() => refetch()}
          />
        </View>
      </View>
    );
  };

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        <Text style={styles.title}>{wishList.title}</Text>
        {wishList.occasion ? (
          <Text style={styles.occasion}>{wishList.occasion}</Text>
        ) : null}
        {wishList.description ? (
          <Text style={styles.description}>{wishList.description}</Text>
        ) : null}
      </View>

      <FlatList
        data={wishList.giftItems || []} // For now, just use an empty array if property doesn't exist
        renderItem={renderGiftItem}
        keyExtractor={(item) => item.id}
        contentContainerStyle={styles.listContent}
        refreshControl={
          <RefreshControl refreshing={isRefetching} onRefresh={refetch} />
        }
        ListEmptyComponent={
          <View style={styles.emptyContainer}>
            <Text>No gift items found</Text>
          </View>
        }
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#fff',
  },
  centerContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  header: {
    padding: 20,
    borderBottomWidth: 1,
    borderBottomColor: '#eee',
  },
  title: {
    fontSize: 24,
    fontWeight: 'bold',
    marginBottom: 5,
  },
  occasion: {
    fontSize: 16,
    color: '#666',
    marginBottom: 5,
  },
  description: {
    fontSize: 14,
    color: '#888',
  },
  listContent: {
    padding: 10,
  },
  itemContainer: {
    backgroundColor: '#f9f9f9',
    padding: 15,
    marginVertical: 5,
    borderRadius: 8,
    borderWidth: 1,
    borderColor: '#eee',
  },
  itemHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 5,
  },
  itemName: {
    fontSize: 16,
    fontWeight: '600',
    flex: 1,
  },
  purchasedBadge: {
    backgroundColor: '#4CAF50',
    color: 'white',
    paddingHorizontal: 6,
    paddingVertical: 2,
    borderRadius: 4,
    fontSize: 12,
    marginLeft: 5,
  },
  reservedBadge: {
    backgroundColor: '#FF9800',
    color: 'white',
    paddingHorizontal: 6,
    paddingVertical: 2,
    borderRadius: 4,
    fontSize: 12,
    marginLeft: 5,
  },
  itemDescription: {
    fontSize: 14,
    color: '#666',
    marginBottom: 5,
  },
  itemPrice: {
    fontSize: 14,
    fontWeight: '600',
    color: '#007AFF',
    marginBottom: 5,
  },
  itemLink: {
    fontSize: 12,
    color: '#007AFF',
    marginBottom: 10,
  },
  itemActions: {
    flexDirection: 'row',
    justifyContent: 'flex-end',
  },
  reserveButton: {
    backgroundColor: '#007AFF',
    color: 'white',
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 4,
    fontSize: 12,
  },
  reservedByText: {
    backgroundColor: '#FF9800',
    color: 'white',
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 4,
    fontSize: 12,
  },
  emptyContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingTop: 50,
  },
});
