import { MaterialCommunityIcons } from '@expo/vector-icons';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { BlurView } from 'expo-blur';
import { LinearGradient } from 'expo-linear-gradient';
import { useLocalSearchParams, useRouter } from 'expo-router';
import { useState } from 'react';
import {
  Alert,
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
  onReserve,
  onEdit,
}: {
  item: GiftItem;
  onReserve: (item: GiftItem) => void;
  onEdit: (item: GiftItem) => void;
}) => {
  const isReserved = !!item.reserved_by_user_id || item.purchased_by_user_id;
  const isPurchased = !!item.purchased_by_user_id;

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
          {isPurchased ? (
            <View style={styles.statusBadge}>
              <MaterialCommunityIcons
                name="check-circle"
                size={16}
                color="#4CAF50"
              />
              <Text style={styles.statusText}>Purchased</Text>
            </View>
          ) : isReserved ? (
            <View style={styles.statusBadge}>
              <MaterialCommunityIcons name="lock" size={16} color="#FF9800" />
              <Text style={styles.statusText}>Reserved</Text>
            </View>
          ) : (
            <View style={styles.statusBadge}>
              <MaterialCommunityIcons
                name="gift-open"
                size={16}
                color="#FFD700"
              />
              <Text style={styles.statusText}>Available</Text>
            </View>
          )}

          {item.priority && item.priority > 0 && (
            <View style={styles.priorityBadge}>
              <MaterialCommunityIcons
                name="star"
                size={14}
                color="#FFD700"
              />
              <Text style={styles.priorityText}>{item.priority}/10</Text>
            </View>
          )}
        </View>

        {/* Actions */}
        <View style={styles.cardActions}>
          {!isReserved && !isPurchased && (
            <Pressable onPress={() => onReserve(item)} style={{ flex: 1 }}>
              <LinearGradient
                colors={['#FFD700', '#FFA500']}
                style={styles.reserveButton}
              >
                <MaterialCommunityIcons name="gift" size={18} color="#000000" />
                <Text style={styles.reserveButtonText}>Reserve</Text>
              </LinearGradient>
            </Pressable>
          )}
          <Pressable
            onPress={() => onEdit(item)}
            style={[
              styles.editButton,
              !isReserved && !isPurchased && { flex: 1 },
            ]}
          >
            <MaterialCommunityIcons
              name="pencil"
              size={18}
              color="rgba(255, 255, 255, 0.7)"
            />
            <Text style={styles.editButtonText}>Edit</Text>
          </Pressable>
        </View>
      </View>
    </BlurView>
  );
};

export default function WishListScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const router = useRouter();
  const queryClient = useQueryClient();

  const [refreshing, setRefreshing] = useState(false);

  const {
    data: wishList,
    isLoading,
    error,
    refetch,
  } = useQuery({
    queryKey: ['wishlist', id],
    queryFn: () => apiClient.getWishListById(id),
    enabled: !!id,
  });

  const {
    data: giftItems,
    isLoading: itemsLoading,
    error: itemsError,
    refetch: refetchItems,
  } = useQuery({
    queryKey: ['giftItems', id],
    queryFn: () => apiClient.getGiftItems(id),
    enabled: !!id,
  });

  const onRefresh = async () => {
    setRefreshing(true);
    await Promise.all([refetch(), refetchItems()]);
    setRefreshing(false);
  };

  const handleReserveGift = (item: GiftItem) => {
    Alert.alert(
      'Reserve Gift',
      `You are reserving "${item.name}". In a real app, this would create a reservation.`,
      [
        { text: 'Cancel', style: 'cancel' },
        {
          text: 'Reserve',
          onPress: () => {
            Alert.alert('Success', 'Gift reserved successfully!');
            queryClient
              .invalidateQueries({ queryKey: ['giftItems', id] })
              .catch(console.error);
          },
        },
      ],
    );
  };

  const handleEditGift = (item: GiftItem) => {
    router.push(`/lists/${id}/gifts/${item.id}/edit`);
  };

  const handleAddGiftItem = () => {
    router.push(`/lists/${id}/gifts/create`);
  };

  const handleEditWishList = () => {
    router.push(`/lists/${id}/edit`);
  };

  if (isLoading || itemsLoading) {
    return (
      <View style={styles.container}>
        <LinearGradient
          colors={['#1a0a2e', '#2d1b4e', '#3d2a6e']}
          style={StyleSheet.absoluteFill}
        />
        <View style={styles.centerContainer}>
          <ActivityIndicator size="large" color="#FFD700" />
          <Text style={styles.loadingText}>Loading wishlist...</Text>
        </View>
      </View>
    );
  }

  if (error || itemsError || !wishList) {
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
            <Text style={styles.errorTitle}>
              {!wishList ? 'Wishlist not found' : 'Error loading wishlist'}
            </Text>
            {(error || itemsError) && (
              <Text style={styles.errorMessage}>
                {error?.message || itemsError?.message}
              </Text>
            )}
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
          <MaterialCommunityIcons
            name="arrow-left"
            size={24}
            color="#ffffff"
          />
        </Pressable>
        <Text style={styles.headerTitle} numberOfLines={1}>
          {wishList.title}
        </Text>
        <Pressable onPress={handleEditWishList} style={styles.editIconButton}>
          <MaterialCommunityIcons name="pencil" size={20} color="#FFD700" />
        </Pressable>
      </View>

      {/* Wishlist Info Card */}
      <View style={styles.contentContainer}>
        <BlurView intensity={20} style={styles.infoCard}>
          <View style={styles.infoContent}>
            {wishList.occasion && (
              <View style={styles.occasionContainer}>
                <MaterialCommunityIcons
                  name="calendar"
                  size={16}
                  color="#FFD700"
                />
                <Text style={styles.occasionText}>{wishList.occasion}</Text>
              </View>
            )}

            {wishList.description && (
              <Text style={styles.descriptionText}>{wishList.description}</Text>
            )}

            {/* Stats */}
            <View style={styles.statsContainer}>
              <View style={styles.statItem}>
                <Text style={styles.statValue}>
                  {giftItems?.length || 0}
                </Text>
                <Text style={styles.statLabel}>Items</Text>
              </View>
              <View style={styles.statDivider} />
              <View style={styles.statItem}>
                <Text style={styles.statValueSecondary}>
                  {giftItems?.filter((item) => item.reserved_by_user_id)
                    .length || 0}
                </Text>
                <Text style={styles.statLabel}>Reserved</Text>
              </View>
              <View style={styles.statDivider} />
              <View style={styles.statItem}>
                <Text style={styles.statValue}>
                  {giftItems?.filter((item) => item.purchased_by_user_id)
                    .length || 0}
                </Text>
                <Text style={styles.statLabel}>Purchased</Text>
              </View>
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
          {giftItems && giftItems.length > 0 ? (
            <View style={styles.listContainer}>
              {giftItems.map((item) => (
                <GiftItemCard
                  key={item.id}
                  item={item}
                  onReserve={handleReserveGift}
                  onEdit={handleEditGift}
                />
              ))}
            </View>
          ) : (
            <View style={styles.emptyState}>
              <MaterialCommunityIcons name="gift" size={64} color="#FFD700" />
              <Text style={styles.emptyStateTitle}>No gift items yet</Text>
              <Text style={styles.emptyStateText}>
                Add your first gift item to get started
              </Text>
            </View>
          )}
        </ScrollView>
      </View>

      {/* Floating Action Button */}
      <Pressable onPress={handleAddGiftItem} style={styles.fab}>
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
  editIconButton: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: 'rgba(255, 215, 0, 0.15)',
    justifyContent: 'center',
    alignItems: 'center',
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
  occasionContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    marginBottom: 12,
  },
  occasionText: {
    fontSize: 14,
    color: '#FFD700',
    fontWeight: '600',
  },
  descriptionText: {
    fontSize: 14,
    color: 'rgba(255, 255, 255, 0.7)',
    lineHeight: 20,
    marginBottom: 16,
  },
  statsContainer: {
    flexDirection: 'row',
    justifyContent: 'space-around',
    alignItems: 'center',
    paddingTop: 16,
    borderTopWidth: 1,
    borderTopColor: 'rgba(255, 255, 255, 0.1)',
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
    color: '#FF9800',
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
  cardActions: {
    flexDirection: 'row',
    gap: 8,
  },
  reserveButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 12,
    paddingHorizontal: 16,
    borderRadius: 12,
    gap: 6,
  },
  reserveButtonText: {
    fontSize: 14,
    fontWeight: '700',
    color: '#000000',
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
