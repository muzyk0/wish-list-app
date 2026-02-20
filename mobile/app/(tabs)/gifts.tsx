import { MaterialCommunityIcons } from '@expo/vector-icons';
import { useInfiniteQuery } from '@tanstack/react-query';
import { BlurView } from 'expo-blur';
import { LinearGradient } from 'expo-linear-gradient';
import { useRouter } from 'expo-router';
import { useCallback, useState } from 'react';
import {
  Dimensions,
  FlatList,
  Image,
  Pressable,
  RefreshControl,
  StyleSheet,
  View,
} from 'react-native';
import { ActivityIndicator, Text } from 'react-native-paper';
import GiftItemDetailModal from '@/components/gifts/GiftItemDetailModal';
import { apiClient } from '@/lib/api';
import type { GiftItem, PaginatedGiftItems } from '@/lib/api/types';

const { width: SCREEN_WIDTH } = Dimensions.get('window');
const CARD_PADDING = 12;
const CARD_GAP = 12;
const NUM_COLUMNS = 2;
const CARD_WIDTH =
  (SCREEN_WIDTH - CARD_PADDING * 2 - CARD_GAP * (NUM_COLUMNS - 1)) /
  NUM_COLUMNS;

// Gift Item Card Component
const GiftItemCard = ({
  item,
  onPress,
}: {
  item: GiftItem;
  onPress: (item: GiftItem) => void;
}) => {
  const isAttached = (item.wishlist_ids?.length ?? 0) > 0;

  return (
    <Pressable
      onPress={() => onPress(item)}
      style={[styles.giftItemCard, { width: CARD_WIDTH }]}
    >
      <BlurView intensity={20} style={styles.cardBlur}>
        {/* Image */}
        {item.image_url ? (
          <Image
            source={{ uri: item.image_url }}
            style={styles.itemImage}
            resizeMode="cover"
          />
        ) : (
          <View style={styles.placeholderImage}>
            <MaterialCommunityIcons
              name="gift"
              size={32}
              color="rgba(255, 215, 0, 0.3)"
            />
          </View>
        )}

        {/* Content */}
        <View style={styles.cardContent}>
          <Text style={styles.itemTitle} numberOfLines={2}>
            {item.title}
          </Text>

          <View style={styles.cardFooter}>
            {item.price !== undefined && item.price !== null ? (
              <LinearGradient
                colors={['#FFD700', '#FFA500']}
                style={styles.priceGradient}
              >
                <Text style={styles.itemPrice}>${item.price.toFixed(2)}</Text>
              </LinearGradient>
            ) : (
              <View />
            )}

            <MaterialCommunityIcons
              name={isAttached ? 'link' : 'link-off'}
              size={16}
              color={isAttached ? '#4CAF50' : '#FF9800'}
            />
          </View>
        </View>
      </BlurView>
    </Pressable>
  );
};

// Filter Tab Component
const FilterTab = ({
  label,
  isActive,
  onPress,
}: {
  label: string;
  isActive: boolean;
  onPress: () => void;
}) => (
  <Pressable
    onPress={onPress}
    style={[styles.filterTab, isActive && styles.filterTabActive]}
  >
    <Text
      style={[styles.filterTabText, isActive && styles.filterTabTextActive]}
    >
      {label}
    </Text>
  </Pressable>
);

// Stats Component
const StatsCard = ({ allItems }: { allItems: GiftItem[] }) => {
  const attachedCount = allItems.filter(
    (item) => (item.wishlist_ids?.length ?? 0) > 0,
  ).length;
  const standaloneCount = allItems.length - attachedCount;

  return (
    <BlurView intensity={20} style={styles.statsCard}>
      <View style={styles.statsContent}>
        <View style={styles.statItem}>
          <Text style={styles.statValue}>{allItems.length}</Text>
          <Text style={styles.statLabel}>Total</Text>
        </View>
        <View style={styles.statDivider} />
        <View style={styles.statItem}>
          <Text style={[styles.statValue, { color: '#4CAF50' }]}>
            {attachedCount}
          </Text>
          <Text style={styles.statLabel}>Attached</Text>
        </View>
        <View style={styles.statDivider} />
        <View style={styles.statItem}>
          <Text style={styles.statValue}>{standaloneCount}</Text>
          <Text style={styles.statLabel}>Standalone</Text>
        </View>
      </View>
    </BlurView>
  );
};

// Empty State Component
const EmptyState = ({
  filter,
}: {
  filter: 'all' | 'attached' | 'unattached';
}) => {
  const getMessage = () => {
    switch (filter) {
      case 'attached':
        return 'No gifts attached to wishlists yet';
      case 'unattached':
        return 'No standalone gifts yet';
      default:
        return 'Create your first gift item';
    }
  };

  return (
    <View style={styles.emptyState}>
      <MaterialCommunityIcons name="gift-off" size={64} color="#FFD700" />
      <Text style={styles.emptyStateTitle}>No gifts yet</Text>
      <Text style={styles.emptyStateText}>{getMessage()}</Text>
    </View>
  );
};

const PAGE_SIZE = 20;

export default function GiftsTab() {
  const router = useRouter();
  const [refreshing, setRefreshing] = useState(false);
  const [filter, setFilter] = useState<'all' | 'attached' | 'unattached'>(
    'all',
  );
  const [selectedItem, setSelectedItem] = useState<GiftItem | null>(null);
  const [modalVisible, setModalVisible] = useState(false);

  const {
    data,
    isLoading,
    error,
    refetch,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
  } = useInfiniteQuery<PaginatedGiftItems, Error>({
    queryKey: ['userGiftItems', filter],
    queryFn: async ({ pageParam = 1 }) =>
      apiClient.getUserGiftItems({
        page: pageParam as number,
        limit: PAGE_SIZE,
        unattached: filter === 'unattached' ? true : undefined,
        attached: filter === 'attached' ? true : undefined,
      }),
    getNextPageParam: (lastPage) => {
      const currentPage = lastPage?.page ?? 1;
      const totalPages = lastPage?.total_pages ?? 1;
      return currentPage < totalPages ? currentPage + 1 : undefined;
    },
    initialPageParam: 1,
  });

  const allItems = data?.pages?.flatMap((page) => page?.items ?? []) ?? [];

  const onRefresh = useCallback(async () => {
    setRefreshing(true);
    await refetch();
    setRefreshing(false);
  }, [refetch]);

  const handleEndReached = useCallback(() => {
    if (hasNextPage && !isFetchingNextPage) {
      fetchNextPage();
    }
  }, [hasNextPage, isFetchingNextPage, fetchNextPage]);

  const handleItemPress = (item: GiftItem) => {
    setSelectedItem(item);
    setModalVisible(true);
  };

  const handleCloseModal = () => {
    setModalVisible(false);
    setSelectedItem(null);
  };

  const handleCreateGift = () => {
    router.push('/gifts/create');
  };

  // Loading State
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

  // Error State
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
          <Text style={styles.headerTitle}>My Gifts</Text>
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
            <Pressable onPress={() => refetch()}>
              <LinearGradient
                colors={['#FFD700', '#FFA500']}
                style={styles.retryButton}
              >
                <Text style={styles.retryButtonText}>Retry</Text>
              </LinearGradient>
            </Pressable>
          </BlurView>
        </View>
      </View>
    );
  }

  // Main Content
  return (
    <View style={styles.container}>
      <LinearGradient
        colors={['#1a0a2e', '#2d1b4e', '#3d2a6e']}
        style={StyleSheet.absoluteFill}
      />

      <View style={styles.decorCircle1} />
      <View style={styles.decorCircle2} />

      <View style={styles.header}>
        <Text style={styles.headerTitle}>My Gifts</Text>
      </View>

      <View style={styles.contentContainer}>
        <BlurView intensity={20} style={styles.filterCard}>
          <View style={styles.filterContent}>
            <FilterTab
              label="All"
              isActive={filter === 'all'}
              onPress={() => setFilter('all')}
            />
            <FilterTab
              label="Attached"
              isActive={filter === 'attached'}
              onPress={() => setFilter('attached')}
            />
            <FilterTab
              label="Standalone"
              isActive={filter === 'unattached'}
              onPress={() => setFilter('unattached')}
            />
          </View>
        </BlurView>

        <StatsCard allItems={allItems} />

        <FlatList
          data={allItems}
          renderItem={({ item }) => (
            <GiftItemCard item={item} onPress={handleItemPress} />
          )}
          keyExtractor={(item) => item.id || ''}
          numColumns={NUM_COLUMNS}
          columnWrapperStyle={styles.columnWrapper}
          contentContainerStyle={styles.listContent}
          showsVerticalScrollIndicator={false}
          refreshControl={
            <RefreshControl
              refreshing={refreshing}
              onRefresh={onRefresh}
              tintColor="#FFD700"
            />
          }
          ListEmptyComponent={<EmptyState filter={filter} />}
          onEndReached={handleEndReached}
          onEndReachedThreshold={0.5}
          ListFooterComponent={
            isFetchingNextPage ? (
              <View style={styles.loadingMore}>
                <ActivityIndicator size="small" color="#FFD700" />
                <Text style={styles.loadingMoreText}>Loading more...</Text>
              </View>
            ) : null
          }
        />
      </View>

      <Pressable onPress={handleCreateGift} style={styles.fab}>
        <LinearGradient
          colors={['#FFD700', '#FFA500']}
          style={styles.fabGradient}
        >
          <MaterialCommunityIcons name="plus" size={28} color="#000000" />
        </LinearGradient>
      </Pressable>

      <GiftItemDetailModal
        item={selectedItem}
        visible={modalVisible}
        onClose={handleCloseModal}
      />
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
    justifyContent: 'center',
    alignItems: 'center',
    paddingHorizontal: 20,
    paddingTop: 60,
    paddingBottom: 20,
  },
  headerTitle: {
    fontSize: 24,
    fontWeight: '700',
    color: '#ffffff',
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
  retryButton: {
    paddingVertical: 12,
    paddingHorizontal: 32,
    borderRadius: 12,
  },
  retryButtonText: {
    fontSize: 16,
    fontWeight: '700',
    color: '#000000',
  },
  contentContainer: {
    flex: 1,
    paddingHorizontal: CARD_PADDING,
  },
  filterCard: {
    borderRadius: 20,
    overflow: 'hidden',
    backgroundColor: 'rgba(255, 255, 255, 0.08)',
    borderWidth: 1,
    borderColor: 'rgba(255, 255, 255, 0.1)',
    marginBottom: 12,
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
  statLabel: {
    fontSize: 12,
    color: 'rgba(255, 255, 255, 0.5)',
  },
  statDivider: {
    width: 1,
    height: 30,
    backgroundColor: 'rgba(255, 255, 255, 0.1)',
  },
  columnWrapper: {
    justifyContent: 'space-between',
    marginBottom: CARD_GAP,
  },
  listContent: {
    paddingBottom: 100,
  },
  giftItemCard: {
    borderRadius: 12,
    overflow: 'hidden',
    marginBottom: 0,
  },
  cardBlur: {
    flex: 1,
    backgroundColor: 'rgba(255, 255, 255, 0.08)',
    borderWidth: 1,
    borderColor: 'rgba(255, 255, 255, 0.1)',
    borderRadius: 12,
    overflow: 'hidden',
  },
  itemImage: {
    width: '100%',
    height: 100,
    backgroundColor: 'rgba(255, 255, 255, 0.05)',
  },
  placeholderImage: {
    width: '100%',
    height: 100,
    backgroundColor: 'rgba(255, 255, 255, 0.05)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  cardContent: {
    padding: 12,
    minHeight: 80,
    justifyContent: 'space-between',
  },
  itemTitle: {
    fontSize: 14,
    fontWeight: '600',
    color: '#ffffff',
    lineHeight: 18,
    marginBottom: 8,
  },
  cardFooter: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  priceGradient: {
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: 8,
  },
  itemPrice: {
    fontSize: 13,
    fontWeight: '700',
    color: '#000000',
  },
  emptyState: {
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 80,
    paddingHorizontal: 40,
  },
  emptyStateTitle: {
    fontSize: 20,
    fontWeight: '700',
    color: '#ffffff',
    marginTop: 16,
    marginBottom: 8,
  },
  emptyStateText: {
    fontSize: 14,
    color: 'rgba(255, 255, 255, 0.5)',
    textAlign: 'center',
    lineHeight: 20,
  },
  loadingMore: {
    flexDirection: 'row',
    justifyContent: 'center',
    alignItems: 'center',
    paddingVertical: 20,
    gap: 8,
  },
  loadingMoreText: {
    fontSize: 14,
    color: 'rgba(255, 255, 255, 0.6)',
  },
  fab: {
    position: 'absolute',
    bottom: 100,
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
