import { MaterialCommunityIcons } from '@expo/vector-icons';
import { useQuery } from '@tanstack/react-query';
import { BlurView } from 'expo-blur';
import { Pressable, StyleSheet, View } from 'react-native';
import { ActivityIndicator, Text } from 'react-native-paper';
import { apiClient } from '@/lib/api';
import type { GiftItem, WishList } from '@/lib/api/types';
import { GiftItemCard } from './GiftItemCard';

interface RecentItemsSectionProps {
  onSeeAll: () => void;
  onItemPress: (listId: string) => void;
}

export function RecentItemsSection({
  onSeeAll,
  onItemPress,
}: RecentItemsSectionProps) {
  const { data: paginatedItems, isLoading } = useQuery({
    queryKey: ['recent-items'],
    queryFn: () =>
      apiClient.getUserGiftItems({
        limit: 8,
        sort: 'created_at',
        order: 'desc',
      }),
    retry: 2,
  });

  // Reuse cached wishlists to resolve list titles (no extra network request)
  const { data: wishlists = [] } = useQuery<WishList[]>({
    queryKey: ['wishlists'],
    queryFn: () => apiClient.getWishLists(),
    retry: 2,
  });

  const wishlistMap = new Map(wishlists.map((wl) => [wl.id, wl.title]));
  const items: GiftItem[] = paginatedItems?.items ?? [];

  const getListTitle = (item: GiftItem): string => {
    const firstId = item.wishlist_ids?.[0];
    if (firstId) return wishlistMap.get(firstId) ?? 'My wishlist';
    return 'Standalone';
  };

  const getListId = (item: GiftItem): string => {
    return item.wishlist_ids?.[0] ?? '';
  };

  return (
    <View style={styles.section}>
      <View style={styles.sectionHeader}>
        <Text style={styles.sectionTitle}>Recent Items</Text>
        {items.length > 0 && (
          <Pressable onPress={onSeeAll}>
            <Text style={styles.seeAllText}>See All</Text>
          </Pressable>
        )}
      </View>

      {isLoading ? (
        <View style={styles.loadingContainer}>
          <ActivityIndicator size="large" color="#FFD700" />
        </View>
      ) : items.length > 0 ? (
        <View style={styles.giftsGrid}>
          {items.map((item, index) => (
            <GiftItemCard
              key={item.id}
              item={item}
              listTitle={getListTitle(item)}
              onPress={() => onItemPress(getListId(item))}
              index={index}
            />
          ))}
        </View>
      ) : (
        <BlurView intensity={20} style={styles.emptyCard}>
          <MaterialCommunityIcons
            name="gift-outline"
            size={48}
            color="rgba(255, 255, 255, 0.3)"
          />
          <Text style={styles.emptyText}>No items yet</Text>
          <Text style={styles.emptySubtext}>
            Add items to your wishlists to see them here
          </Text>
        </BlurView>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  section: {
    marginBottom: 28,
  },
  sectionHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 16,
  },
  sectionTitle: {
    fontSize: 20,
    fontWeight: '700',
    color: '#ffffff',
  },
  seeAllText: {
    fontSize: 14,
    fontWeight: '600',
    color: '#FFD700',
  },
  giftsGrid: {
    gap: 10,
  },
  loadingContainer: {
    paddingVertical: 40,
    alignItems: 'center',
  },
  emptyCard: {
    borderRadius: 16,
    overflow: 'hidden',
    backgroundColor: 'rgba(255, 255, 255, 0.05)',
    borderWidth: 1,
    borderColor: 'rgba(255, 255, 255, 0.08)',
    padding: 32,
    alignItems: 'center',
  },
  emptyText: {
    fontSize: 16,
    fontWeight: '600',
    color: 'rgba(255, 255, 255, 0.6)',
    marginTop: 12,
  },
  emptySubtext: {
    fontSize: 13,
    color: 'rgba(255, 255, 255, 0.4)',
    marginTop: 4,
    textAlign: 'center',
  },
});
