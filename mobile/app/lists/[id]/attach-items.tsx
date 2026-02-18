import { MaterialCommunityIcons } from '@expo/vector-icons';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { BlurView } from 'expo-blur';
import { LinearGradient } from 'expo-linear-gradient';
import { useLocalSearchParams, useRouter } from 'expo-router';
import { useState } from 'react';
import {
  Alert,
  Pressable,
  RefreshControl,
  ScrollView,
  StyleSheet,
  View,
} from 'react-native';
import { ActivityIndicator, Text } from 'react-native-paper';
import { apiClient } from '@/lib/api';
import type { GiftItem } from '@/lib/api/types';

// Gift Item Card Component for Selection
const SelectableGiftItemCard = ({
  item,
  onAttach,
  isAttaching,
}: {
  item: GiftItem;
  onAttach: (item: GiftItem) => void;
  isAttaching: boolean;
}) => {
  return (
    <BlurView intensity={20} style={styles.giftItemCard}>
      <View style={styles.cardContent}>
        {/* Header */}
        <View style={styles.itemHeader}>
          <View style={styles.itemTitleContainer}>
            <Text style={styles.itemTitle}>{item.title}</Text>
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

        {/* Priority Badge */}
        {item.priority && item.priority > 0 && (
          <View style={styles.priorityBadge}>
            <MaterialCommunityIcons name="star" size={14} color="#FFD700" />
            <Text style={styles.priorityText}>
              Priority: {item.priority}/10
            </Text>
          </View>
        )}

        {/* Attach Button */}
        <Pressable
          onPress={() => onAttach(item)}
          disabled={isAttaching}
          style={{ marginTop: 12 }}
        >
          <LinearGradient
            colors={['#FFD700', '#FFA500']}
            style={styles.attachButton}
          >
            <MaterialCommunityIcons
              name="link-plus"
              size={18}
              color="#000000"
            />
            <Text style={styles.attachButtonText}>
              {isAttaching ? 'Attaching...' : 'Attach to List'}
            </Text>
          </LinearGradient>
        </Pressable>
      </View>
    </BlurView>
  );
};

export default function AttachItemsScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const router = useRouter();
  const queryClient = useQueryClient();

  const [refreshing, setRefreshing] = useState(false);
  const [attachingItemId, setAttachingItemId] = useState<string | null>(null);

  // Fetch wishlist details
  const {
    data: wishList,
    isLoading: wishlistLoading,
    error: wishlistError,
  } = useQuery({
    queryKey: ['wishlist', id],
    queryFn: () => apiClient.getWishListById(id),
    enabled: !!id,
  });

  // Fetch all standalone gift items (items not attached to any wishlist)
  const {
    data: itemsResponse,
    isLoading: itemsLoading,
    error: itemsError,
    refetch,
  } = useQuery({
    queryKey: ['standaloneGiftItems'],
    queryFn: () => apiClient.getUserGiftItems({ unattached: true }),
    enabled: !!id,
  });

  const allItems = itemsResponse?.items || [];

  const onRefresh = async () => {
    setRefreshing(true);
    await refetch();
    setRefreshing(false);
  };

  const attachMutation = useMutation({
    mutationFn: ({ itemId }: { itemId: string }) =>
      apiClient.attachGiftItemToWishlist(id, itemId),
    onSuccess: () => {
      Alert.alert('Success', 'Gift item attached to wishlist successfully!', [
        {
          text: 'OK',
          onPress: () => {
            queryClient
              .invalidateQueries({ queryKey: ['giftItems', id] })
              .catch(console.error);
            queryClient
              .invalidateQueries({ queryKey: ['standaloneGiftItems'] })
              .catch(console.error);
            setAttachingItemId(null);
          },
        },
      ]);
    },
    onError: (error: Error) => {
      Alert.alert(
        'Error',
        error.message || 'Failed to attach gift item. Please try again.',
      );
      setAttachingItemId(null);
    },
  });

  const handleAttachItem = (item: GiftItem) => {
    setAttachingItemId(item.id ?? null);
    attachMutation.mutate({ itemId: item.id ?? '' });
  };

  if (wishlistLoading || itemsLoading) {
    return (
      <View style={styles.container}>
        <LinearGradient
          colors={['#1a0a2e', '#2d1b4e', '#3d2a6e']}
          style={StyleSheet.absoluteFill}
        />
        <View style={styles.centerContainer}>
          <ActivityIndicator size="large" color="#FFD700" />
          <Text style={styles.loadingText}>Loading items...</Text>
        </View>
      </View>
    );
  }

  if (wishlistError || itemsError || !wishList) {
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
            <Text style={styles.errorTitle}>Error loading data</Text>
            <Text style={styles.errorMessage}>
              {wishlistError?.message || itemsError?.message}
            </Text>
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
        <Text style={styles.headerTitle} numberOfLines={1}>
          Attach Items
        </Text>
        <View style={{ width: 40 }} />
      </View>

      {/* Info Card */}
      <View style={styles.contentContainer}>
        <BlurView intensity={20} style={styles.infoCard}>
          <View style={styles.infoContent}>
            <Text style={styles.infoTitle}>Attaching to: {wishList.title}</Text>
            <Text style={styles.infoText}>
              Select existing gift items to attach to this wishlist
            </Text>
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
          {allItems && allItems.length > 0 ? (
            <View style={styles.listContainer}>
              {allItems.map((item) => (
                <SelectableGiftItemCard
                  key={item.id}
                  item={item}
                  onAttach={handleAttachItem}
                  isAttaching={attachingItemId === item.id}
                />
              ))}
            </View>
          ) : (
            <View style={styles.emptyState}>
              <MaterialCommunityIcons
                name="gift-off"
                size={64}
                color="#FFD700"
              />
              <Text style={styles.emptyStateTitle}>No items available</Text>
              <Text style={styles.emptyStateText}>
                Create standalone gift items first to attach them to wishlists
              </Text>
            </View>
          )}
        </ScrollView>
      </View>
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
  infoCard: {
    borderRadius: 20,
    overflow: 'hidden',
    backgroundColor: 'rgba(255, 255, 255, 0.08)',
    borderWidth: 1,
    borderColor: 'rgba(255, 255, 255, 0.1)',
    marginBottom: 16,
  },
  infoContent: {
    padding: 20,
  },
  infoTitle: {
    fontSize: 16,
    fontWeight: '600',
    color: '#FFD700',
    marginBottom: 8,
  },
  infoText: {
    fontSize: 14,
    color: 'rgba(255, 255, 255, 0.7)',
    lineHeight: 20,
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
  priorityBadge: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
    paddingHorizontal: 8,
    paddingVertical: 4,
    borderRadius: 10,
    backgroundColor: 'rgba(255, 215, 0, 0.15)',
    alignSelf: 'flex-start',
  },
  priorityText: {
    fontSize: 11,
    color: '#FFD700',
    fontWeight: '600',
  },
  attachButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 12,
    paddingHorizontal: 16,
    borderRadius: 12,
    gap: 6,
  },
  attachButtonText: {
    fontSize: 14,
    fontWeight: '700',
    color: '#000000',
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
});
