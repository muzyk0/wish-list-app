import { MaterialCommunityIcons } from '@expo/vector-icons';
import { useQuery } from '@tanstack/react-query';
import { BlurView } from 'expo-blur';
import { LinearGradient } from 'expo-linear-gradient';
import { useRouter } from 'expo-router';
import { useState } from 'react';
import {
  Linking,
  Pressable,
  RefreshControl,
  ScrollView,
  StyleSheet,
  View,
} from 'react-native';
import { ActivityIndicator, Text } from 'react-native-paper';
import { apiClient } from '@/lib/api';
import type { GiftItem } from '@/lib/api/types';

// Gift Item Card Component
const GiftItemCard = ({
  item,
  onEdit,
}: {
  item: GiftItem;
  onEdit: (item: GiftItem) => void;
}) => {
  const isAttached = !!item.wishlist_id;

  return (
    <BlurView intensity={20} style={styles.giftItemCard}>
      <View style={styles.cardContent}>
        {/* Header */}
        <View style={styles.itemHeader}>
          <View style={styles.itemTitleContainer}>
            <Text style={styles.itemTitle}>{item.name}</Text>
            {item.price !== undefined && item.price !== null && (
              <View style={styles.priceContainer}>
                <LinearGradient
                  colors={['#FFD700', '#FFA500']}
                  style={styles.priceGradient}
                >
                  <Text style={styles.itemPrice}>${item.price.toFixed(2)}</Text>
                </LinearGradient>
              </View>
            )}
          </View>
        </View>

        {/* Description */}
        {item.description && (
          <Text style={styles.itemDescription}>{item.description}</Text>
        )}

        {/* Link */}
        {item.link && (
          <Pressable onPress={() => Linking.openURL(item.link || '')}>
            <View style={styles.linkContainer}>
              <MaterialCommunityIcons
                name="open-in-new"
                size={16}
                color="#FFD700"
              />
              <Text style={styles.linkText}>Visit Website</Text>
            </View>
          </Pressable>
        )}

        {/* Footer with status and priority */}
        <View style={styles.itemFooter}>
          {isAttached ? (
            <View style={styles.statusBadge}>
              <MaterialCommunityIcons name="link" size={16} color="#4CAF50" />
              <Text style={styles.statusText}>Attached to List</Text>
            </View>
          ) : (
            <View style={styles.statusBadge}>
              <MaterialCommunityIcons
                name="link-off"
                size={16}
                color="#FF9800"
              />
              <Text style={styles.statusText}>Not Attached</Text>
            </View>
          )}

          {item.priority && item.priority > 0 && (
            <View style={styles.priorityBadge}>
              <MaterialCommunityIcons name="star" size={14} color="#FFD700" />
              <Text style={styles.priorityText}>{item.priority}/10</Text>
            </View>
          )}
        </View>

        {/* Edit Button */}
        <Pressable
          onPress={() => onEdit(item)}
          style={styles.editButton}
        >
          <MaterialCommunityIcons
            name="pencil"
            size={18}
            color="rgba(255, 255, 255, 0.7)"
          />
          <Text style={styles.editButtonText}>Edit</Text>
        </Pressable>
      </View>
    </BlurView>
  );
};

export default function MyGiftsScreen() {
  const router = useRouter();
  const [refreshing, setRefreshing] = useState(false);
  const [filter, setFilter] = useState<'all' | 'attached' | 'unattached'>('all');

  const {
    data: itemsResponse,
    isLoading,
    error,
    refetch,
  } = useQuery({
    queryKey: ['userGiftItems', filter],
    queryFn: () =>
      apiClient.getUserGiftItems({
        unattached: filter === 'unattached' ? true : undefined,
        limit: 100,
      }),
  });

  const onRefresh = async () => {
    setRefreshing(true);
    await refetch();
    setRefreshing(false);
  };

  const handleEditGift = (item: GiftItem) => {
    router.push(`/gift-items/${item.id}/edit?wishlistId=${item.wishlist_id || ''}`);
  };

  const handleCreateGift = () => {
    router.push('/gifts/create');
  };

  const allItems = itemsResponse?.items || [];
  const filteredItems =
    filter === 'attached'
      ? allItems.filter((item) => item.wishlist_id)
      : filter === 'unattached'
        ? allItems.filter((item) => !item.wishlist_id)
        : allItems;

  if (isLoading) {
    return (
      <View style={styles.container}>
        <LinearGradient
          colors={['#1a0a2e', '#2d1b4e', '#3d2a6e']}
          style={StyleSheet.absoluteFill}
        />
        <View style={styles.centerContainer}>
          <ActivityIndicator size="large" color="#FFD700" />
          <Text style={styles.loadingText}>Loading your gifts...</Text>
        </View>
      </View>
    );
  }

  if (error) {
    return (
      <View style={styles.container}>
        <LinearGradient
          colors={['#1a0a2e', '#2d1b4e', '#3d2a6e']}
          style={StyleSheet.absoluteFill}
        />
        <View style={styles.decorCircle1} />
        <View style={styles.decorCircle2} />

        <View style={styles.header}>
          <Pressable onPress={() => router.back()} style={styles.backButton}>
            <MaterialCommunityIcons
              name="arrow-left"
              size={24}
              color="#ffffff"
            />
          </Pressable>
          <Text style={styles.headerTitle}>Error</Text>
          <View style={{ width: 40 }} />
        </View>

        <View style={styles.centerContainer}>
          <BlurView intensity={20} style={styles.errorCard}>
            <MaterialCommunityIcons
              name="alert-circle"
              size={64}
              color="#FF6B6B"
            />
            <Text style={styles.errorTitle}>Error loading gifts</Text>
            <Text style={styles.errorMessage}>{error.message}</Text>
            <Pressable onPress={() => router.back()}>
              <LinearGradient
                colors={['#FFD700', '#FFA500']}
                style={styles.backHomeButton}
              >
                <Text style={styles.backHomeButtonText}>Go Back</Text>
              </LinearGradient>
            </Pressable>
          </BlurView>
        </View>
      </View>
    );
  }

  return (
    <View style={styles.container}>
      <LinearGradient
        colors={['#1a0a2e', '#2d1b4e', '#3d2a6e']}
        style={StyleSheet.absoluteFill}
      />

      {/* Decorative elements */}
      <View style={styles.decorCircle1} />
      <View style={styles.decorCircle2} />

      {/* Header */}
      <View style={styles.header}>
        <Pressable onPress={() => router.back()} style={styles.backButton}>
          <MaterialCommunityIcons name="arrow-left" size={24} color="#ffffff" />
        </Pressable>
        <Text style={styles.headerTitle}>My Gifts</Text>
        <View style={{ width: 40 }} />
      </View>

      <View style={styles.contentContainer}>
        {/* Filter Tabs */}
        <BlurView intensity={20} style={styles.filterCard}>
          <View style={styles.filterContent}>
            <Pressable
              onPress={() => setFilter('all')}
              style={[
                styles.filterTab,
                filter === 'all' && styles.filterTabActive,
              ]}
            >
              <Text
                style={[
                  styles.filterTabText,
                  filter === 'all' && styles.filterTabTextActive,
                ]}
              >
                All
              </Text>
            </Pressable>
            <Pressable
              onPress={() => setFilter('attached')}
              style={[
                styles.filterTab,
                filter === 'attached' && styles.filterTabActive,
              ]}
            >
              <Text
                style={[
                  styles.filterTabText,
                  filter === 'attached' && styles.filterTabTextActive,
                ]}
              >
                Attached
              </Text>
            </Pressable>
            <Pressable
              onPress={() => setFilter('unattached')}
              style={[
                styles.filterTab,
                filter === 'unattached' && styles.filterTabActive,
              ]}
            >
              <Text
                style={[
                  styles.filterTabText,
                  filter === 'unattached' && styles.filterTabTextActive,
                ]}
              >
                Standalone
              </Text>
            </Pressable>
          </View>
        </BlurView>

        {/* Stats Card */}
        <BlurView intensity={20} style={styles.statsCard}>
          <View style={styles.statsContent}>
            <View style={styles.statItem}>
              <Text style={styles.statValue}>{allItems.length}</Text>
              <Text style={styles.statLabel}>Total</Text>
            </View>
            <View style={styles.statDivider} />
            <View style={styles.statItem}>
              <Text style={styles.statValueSecondary}>
                {allItems.filter((item) => item.wishlist_id).length}
              </Text>
              <Text style={styles.statLabel}>Attached</Text>
            </View>
            <View style={styles.statDivider} />
            <View style={styles.statItem}>
              <Text style={styles.statValue}>
                {allItems.filter((item) => !item.wishlist_id).length}
              </Text>
              <Text style={styles.statLabel}>Standalone</Text>
            </View>
          </View>
        </BlurView>

        {/* Gift Items List */}
        <ScrollView
          style={styles.scrollContainer}
          showsVerticalScrollIndicator={false}
          refreshControl={
            <RefreshControl
              refreshing={refreshing}
              onRefresh={onRefresh}
              tintColor="#FFD700"
            />
          }
        >
          {filteredItems.length > 0 ? (
            <View style={styles.listContainer}>
              {filteredItems.map((item) => (
                <GiftItemCard key={item.id} item={item} onEdit={handleEditGift} />
              ))}
            </View>
          ) : (
            <View style={styles.emptyState}>
              <MaterialCommunityIcons name="gift-off" size={64} color="#FFD700" />
              <Text style={styles.emptyStateTitle}>No gifts yet</Text>
              <Text style={styles.emptyStateText}>
                {filter === 'all'
                  ? 'Create your first gift item'
                  : filter === 'attached'
                    ? 'No gifts attached to wishlists yet'
                    : 'No standalone gifts yet'}
              </Text>
            </View>
          )}
        </ScrollView>
      </View>

      {/* Floating Action Button */}
      <Pressable onPress={handleCreateGift} style={styles.fab}>
        <LinearGradient
          colors={['#FFD700', '#FFA500']}
          style={styles.fabGradient}
        >
          <MaterialCommunityIcons name="plus" size={28} color="#000000" />
        </LinearGradient>
      </Pressable>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  decorCircle1: {
    position: 'absolute',
    width: 250,
    height: 250,
    borderRadius: 125,
    backgroundColor: 'rgba(255, 215, 0, 0.06)',
    top: -80,
    right: -60,
  },
  decorCircle2: {
    position: 'absolute',
    width: 180,
    height: 180,
    borderRadius: 90,
    backgroundColor: 'rgba(107, 78, 230, 0.12)',
    bottom: 200,
    left: -40,
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: 20,
    paddingTop: 60,
    paddingBottom: 20,
  },
  backButton: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: 'rgba(255, 255, 255, 0.1)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  headerTitle: {
    fontSize: 18,
    fontWeight: '700',
    color: '#ffffff',
    flex: 1,
    textAlign: 'center',
    marginHorizontal: 12,
  },
  centerContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 20,
  },
  loadingText: {
    fontSize: 16,
    color: 'rgba(255, 255, 255, 0.7)',
    marginTop: 16,
  },
  errorCard: {
    borderRadius: 24,
    overflow: 'hidden',
    backgroundColor: 'rgba(255, 255, 255, 0.08)',
    borderWidth: 1,
    borderColor: 'rgba(255, 255, 255, 0.1)',
    padding: 32,
    alignItems: 'center',
    maxWidth: 400,
  },
  errorTitle: {
    fontSize: 20,
    fontWeight: '700',
    color: '#FF6B6B',
    marginTop: 16,
    marginBottom: 8,
    textAlign: 'center',
  },
  errorMessage: {
    fontSize: 14,
    color: 'rgba(255, 255, 255, 0.6)',
    textAlign: 'center',
    marginBottom: 24,
  },
  backHomeButton: {
    paddingVertical: 12,
    paddingHorizontal: 32,
    borderRadius: 12,
  },
  backHomeButtonText: {
    fontSize: 16,
    fontWeight: '700',
    color: '#000000',
  },
  contentContainer: {
    flex: 1,
    paddingHorizontal: 20,
  },
  filterCard: {
    borderRadius: 20,
    overflow: 'hidden',
    backgroundColor: 'rgba(255, 255, 255, 0.08)',
    borderWidth: 1,
    borderColor: 'rgba(255, 255, 255, 0.1)',
    marginBottom: 16,
  },
  filterContent: {
    flexDirection: 'row',
    padding: 8,
    gap: 8,
  },
  filterTab: {
    flex: 1,
    paddingVertical: 10,
    paddingHorizontal: 16,
    borderRadius: 12,
    alignItems: 'center',
    backgroundColor: 'transparent',
  },
  filterTabActive: {
    backgroundColor: 'rgba(255, 215, 0, 0.15)',
  },
  filterTabText: {
    fontSize: 14,
    fontWeight: '600',
    color: 'rgba(255, 255, 255, 0.5)',
  },
  filterTabTextActive: {
    color: '#FFD700',
  },
  statsCard: {
    borderRadius: 20,
    overflow: 'hidden',
    backgroundColor: 'rgba(255, 255, 255, 0.08)',
    borderWidth: 1,
    borderColor: 'rgba(255, 255, 255, 0.1)',
    marginBottom: 16,
  },
  statsContent: {
    flexDirection: 'row',
    justifyContent: 'space-around',
    alignItems: 'center',
    padding: 16,
  },
  statItem: {
    alignItems: 'center',
  },
  statValue: {
    fontSize: 24,
    fontWeight: '700',
    color: '#FFD700',
    marginBottom: 4,
  },
  statValueSecondary: {
    fontSize: 24,
    fontWeight: '700',
    color: '#4CAF50',
    marginBottom: 4,
  },
  statLabel: {
    fontSize: 12,
    color: 'rgba(255, 255, 255, 0.5)',
  },
  statDivider: {
    width: 1,
    height: 30,
    backgroundColor: 'rgba(255, 255, 255, 0.1)',
  },
  scrollContainer: {
    flex: 1,
  },
  listContainer: {
    gap: 12,
    paddingBottom: 100,
  },
  giftItemCard: {
    borderRadius: 16,
    overflow: 'hidden',
    backgroundColor: 'rgba(255, 255, 255, 0.08)',
    borderWidth: 1,
    borderColor: 'rgba(255, 255, 255, 0.1)',
    marginBottom: 12,
  },
  cardContent: {
    padding: 16,
  },
  itemHeader: {
    marginBottom: 8,
  },
  itemTitleContainer: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  itemTitle: {
    fontSize: 16,
    fontWeight: '600',
    color: '#ffffff',
    flex: 1,
  },
  priceContainer: {
    marginLeft: 12,
  },
  priceGradient: {
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 12,
  },
  itemPrice: {
    fontSize: 14,
    fontWeight: '700',
    color: '#000000',
  },
  itemDescription: {
    fontSize: 14,
    color: 'rgba(255, 255, 255, 0.6)',
    lineHeight: 20,
    marginBottom: 8,
  },
  linkContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
    marginBottom: 12,
  },
  linkText: {
    fontSize: 13,
    color: '#FFD700',
    fontWeight: '500',
  },
  itemFooter: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    marginBottom: 16,
  },
  statusBadge: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
    paddingHorizontal: 10,
    paddingVertical: 6,
    borderRadius: 12,
    backgroundColor: 'rgba(255, 255, 255, 0.08)',
  },
  statusText: {
    fontSize: 12,
    color: 'rgba(255, 255, 255, 0.7)',
    fontWeight: '600',
  },
  priorityBadge: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
    paddingHorizontal: 8,
    paddingVertical: 4,
    borderRadius: 10,
    backgroundColor: 'rgba(255, 215, 0, 0.15)',
  },
  priorityText: {
    fontSize: 11,
    color: '#FFD700',
    fontWeight: '600',
  },
  editButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 12,
    paddingHorizontal: 16,
    borderRadius: 12,
    backgroundColor: 'rgba(255, 255, 255, 0.08)',
    borderWidth: 1,
    borderColor: 'rgba(255, 255, 255, 0.15)',
    gap: 6,
  },
  editButtonText: {
    fontSize: 14,
    fontWeight: '600',
    color: 'rgba(255, 255, 255, 0.7)',
  },
  emptyState: {
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 60,
  },
  emptyStateTitle: {
    fontSize: 18,
    fontWeight: '700',
    color: '#ffffff',
    marginTop: 16,
    marginBottom: 8,
  },
  emptyStateText: {
    fontSize: 14,
    color: 'rgba(255, 255, 255, 0.5)',
    textAlign: 'center',
    paddingHorizontal: 40,
  },
  fab: {
    position: 'absolute',
    bottom: 24,
    right: 24,
    borderRadius: 28,
    shadowColor: '#FFD700',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.4,
    shadowRadius: 12,
    elevation: 8,
  },
  fabGradient: {
    width: 56,
    height: 56,
    borderRadius: 28,
    justifyContent: 'center',
    alignItems: 'center',
  },
});
