import { MaterialCommunityIcons } from '@expo/vector-icons';
import { useQuery } from '@tanstack/react-query';
import { BlurView } from 'expo-blur';
import { Pressable, StyleSheet, View } from 'react-native';
import { ActivityIndicator, Text } from 'react-native-paper';
import { GiftItemCard } from './GiftItemCard';

interface RecentItemsSectionProps {
  onSeeAll: () => void;
  onItemPress: (listId: string) => void;
}

export function RecentItemsSection({
  onSeeAll,
  onItemPress,
}: RecentItemsSectionProps) {
  // TODO: Add real endpoint to fetch recent items (last 8)
  // Backend should return only recent items, not all
  const { data: items = [], isLoading } = useQuery<any[]>({
    queryKey: ['recent-items'],
    queryFn: async () => {
      // Mock: return empty array
      // TODO: Replace with: apiClient.getRecentItems()
      return [];
    },
    retry: 2,
  });
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
              listTitle={item.listTitle}
              onPress={() => onItemPress(item.listId)}
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
