import { MaterialCommunityIcons } from '@expo/vector-icons';
import { BlurView } from 'expo-blur';
import { LinearGradient } from 'expo-linear-gradient';
import { useEffect, useState } from 'react';
import { FlatList, RefreshControl, StyleSheet, View } from 'react-native';
import { ActivityIndicator, Text } from 'react-native-paper';

interface WishReservation {
  id: string;
  giftItem: {
    id: string;
    name: string;
    price?: number;
    imageUrl?: string;
  };
  wishlist: {
    id: string;
    title: string;
  };
  reservedBy: {
    id: string;
    firstName?: string;
    lastName?: string;
    email: string;
  };
  status: 'active' | 'canceled' | 'fulfilled';
  reservedAt: string;
  isPurchased: boolean;
}

export function MyWishesReserved() {
  const [reservations, setReservations] = useState<WishReservation[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);

  const fetchReservations = async () => {
    try {
      setLoading(true);
      // Fetch reservations of items from user's wishlists
      const response = await fetch('/api/users/me/wishlists/reservations');
      if (response.ok) {
        const data = await response.json();
        setReservations(data.data || []);
      }
    } catch (error) {
      console.error('Failed to load wish reservations:', error);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  };

  // biome-ignore lint/correctness/useExhaustiveDependencies: Execute once
  useEffect(() => {
    void fetchReservations();
  }, []);

  const onRefresh = () => {
    setRefreshing(true);
    fetchReservations();
  };

  const renderItem = ({ item }: { item: WishReservation }) => {
    const reservedByName =
      `${item.reservedBy.firstName || ''} ${item.reservedBy.lastName || ''}`.trim() ||
      item.reservedBy.email;

    const getStatusConfig = () => {
      if (item.isPurchased) {
        return {
          color: '#4CAF50',
          icon: 'check-circle',
          label: 'Purchased',
          bg: 'rgba(76, 175, 80, 0.15)',
        };
      }
      switch (item.status) {
        case 'active':
          return {
            color: '#FFD700',
            icon: 'clock-outline',
            label: 'Reserved',
            bg: 'rgba(255, 215, 0, 0.15)',
          };
        case 'canceled':
          return {
            color: '#9E9E9E',
            icon: 'close-circle-outline',
            label: 'Canceled',
            bg: 'rgba(158, 158, 158, 0.15)',
          };
        case 'fulfilled':
          return {
            color: '#4CAF50',
            icon: 'check-circle',
            label: 'Fulfilled',
            bg: 'rgba(76, 175, 80, 0.15)',
          };
        default:
          return {
            color: '#FFD700',
            icon: 'help-circle-outline',
            label: item.status,
            bg: 'rgba(255, 215, 0, 0.15)',
          };
      }
    };

    const statusConfig = getStatusConfig();

    return (
      <BlurView intensity={20} style={styles.card}>
        <View style={styles.cardContent}>
          {/* Header with Status */}
          <View style={styles.header}>
            <View style={styles.itemInfo}>
              <Text style={styles.itemName} numberOfLines={2}>
                {item.giftItem.name}
              </Text>
              {item.giftItem.price !== undefined && (
                <View style={styles.priceContainer}>
                  <LinearGradient
                    colors={['#FFD700', '#FFA500']}
                    style={styles.priceGradient}
                  >
                    <Text style={styles.priceText}>
                      ${item.giftItem.price.toFixed(2)}
                    </Text>
                  </LinearGradient>
                </View>
              )}
            </View>

            <View
              style={[styles.statusBadge, { backgroundColor: statusConfig.bg }]}
            >
              <MaterialCommunityIcons
                name={statusConfig.icon as any}
                size={14}
                color={statusConfig.color}
              />
              <Text style={[styles.statusText, { color: statusConfig.color }]}>
                {statusConfig.label}
              </Text>
            </View>
          </View>

          {/* Wishlist */}
          <View style={styles.infoRow}>
            <MaterialCommunityIcons
              name="gift-outline"
              size={16}
              color="rgba(255, 255, 255, 0.5)"
            />
            <Text style={styles.infoText} numberOfLines={1}>
              From: {item.wishlist.title}
            </Text>
          </View>

          {/* Reserved By */}
          <View style={styles.reservedByContainer}>
            <View style={styles.avatarCircle}>
              <MaterialCommunityIcons
                name="account"
                size={16}
                color="#FFD700"
              />
            </View>
            <View style={styles.reservedByInfo}>
              <Text style={styles.reservedByLabel}>Reserved by</Text>
              <Text style={styles.reservedByName} numberOfLines={1}>
                {reservedByName}
              </Text>
            </View>
          </View>

          {/* Date */}
          <View style={styles.dateRow}>
            <MaterialCommunityIcons
              name="calendar-check"
              size={14}
              color="rgba(255, 255, 255, 0.4)"
            />
            <Text style={styles.dateText}>
              {new Date(item.reservedAt).toLocaleDateString('en-US', {
                month: 'short',
                day: 'numeric',
                year: 'numeric',
              })}
            </Text>
          </View>
        </View>
      </BlurView>
    );
  };

  if (loading) {
    return (
      <View style={styles.centerContainer}>
        <ActivityIndicator size="large" color="#FFD700" />
        <Text style={styles.loadingText}>Loading reservations...</Text>
      </View>
    );
  }

  if (reservations.length === 0) {
    return (
      <View style={styles.centerContainer}>
        <BlurView intensity={20} style={styles.emptyCard}>
          <MaterialCommunityIcons
            name="gift-outline"
            size={64}
            color="#FFD700"
          />
          <Text style={styles.emptyTitle}>No reservations yet</Text>
          <Text style={styles.emptyText}>
            When someone reserves your wishes, they'll appear here
          </Text>
        </BlurView>
      </View>
    );
  }

  return (
    <View style={styles.container}>
      <View style={styles.statsContainer}>
        <BlurView intensity={15} style={styles.statCard}>
          <Text style={styles.statValue}>{reservations.length}</Text>
          <Text style={styles.statLabel}>Total Reserved</Text>
        </BlurView>
        <BlurView intensity={15} style={styles.statCard}>
          <Text style={styles.statValueGreen}>
            {
              reservations.filter(
                (r) => r.isPurchased || r.status === 'fulfilled',
              ).length
            }
          </Text>
          <Text style={styles.statLabel}>Purchased</Text>
        </BlurView>
      </View>

      <FlatList
        data={reservations}
        keyExtractor={(item) => item.id}
        renderItem={renderItem}
        contentContainerStyle={styles.listContent}
        showsVerticalScrollIndicator={false}
        refreshControl={
          <RefreshControl
            refreshing={refreshing}
            onRefresh={onRefresh}
            tintColor="#FFD700"
          />
        }
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  centerContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 20,
  },
  statsContainer: {
    flexDirection: 'row',
    gap: 12,
    marginBottom: 16,
  },
  statCard: {
    flex: 1,
    borderRadius: 16,
    overflow: 'hidden',
    backgroundColor: 'rgba(255, 255, 255, 0.08)',
    borderWidth: 1,
    borderColor: 'rgba(255, 255, 255, 0.1)',
    padding: 16,
    alignItems: 'center',
  },
  statValue: {
    fontSize: 28,
    fontWeight: '700',
    color: '#FFD700',
    marginBottom: 4,
  },
  statValueGreen: {
    fontSize: 28,
    fontWeight: '700',
    color: '#4CAF50',
    marginBottom: 4,
  },
  statLabel: {
    fontSize: 12,
    color: 'rgba(255, 255, 255, 0.5)',
  },
  listContent: {
    paddingBottom: 100,
  },
  card: {
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
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    marginBottom: 12,
    gap: 12,
  },
  itemInfo: {
    flex: 1,
  },
  itemName: {
    fontSize: 16,
    fontWeight: '600',
    color: '#ffffff',
    marginBottom: 6,
  },
  priceContainer: {
    alignSelf: 'flex-start',
  },
  priceGradient: {
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: 8,
  },
  priceText: {
    fontSize: 13,
    fontWeight: '700',
    color: '#000000',
  },
  statusBadge: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
    paddingHorizontal: 10,
    paddingVertical: 6,
    borderRadius: 12,
  },
  statusText: {
    fontSize: 11,
    fontWeight: '600',
  },
  infoRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    marginBottom: 12,
  },
  infoText: {
    fontSize: 13,
    color: 'rgba(255, 255, 255, 0.7)',
    flex: 1,
  },
  reservedByContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 12,
    paddingVertical: 12,
    paddingHorizontal: 12,
    backgroundColor: 'rgba(255, 215, 0, 0.08)',
    borderRadius: 12,
    marginBottom: 12,
  },
  avatarCircle: {
    width: 36,
    height: 36,
    borderRadius: 18,
    backgroundColor: 'rgba(255, 215, 0, 0.15)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  reservedByInfo: {
    flex: 1,
  },
  reservedByLabel: {
    fontSize: 11,
    color: 'rgba(255, 255, 255, 0.5)',
    marginBottom: 2,
  },
  reservedByName: {
    fontSize: 14,
    fontWeight: '600',
    color: '#FFD700',
  },
  dateRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
  },
  dateText: {
    fontSize: 12,
    color: 'rgba(255, 255, 255, 0.5)',
  },
  loadingText: {
    fontSize: 16,
    color: 'rgba(255, 255, 255, 0.7)',
    marginTop: 16,
  },
  emptyCard: {
    borderRadius: 24,
    overflow: 'hidden',
    backgroundColor: 'rgba(255, 255, 255, 0.08)',
    borderWidth: 1,
    borderColor: 'rgba(255, 255, 255, 0.1)',
    padding: 40,
    alignItems: 'center',
    maxWidth: 400,
  },
  emptyTitle: {
    fontSize: 20,
    fontWeight: '700',
    color: '#ffffff',
    marginTop: 16,
    marginBottom: 8,
  },
  emptyText: {
    fontSize: 14,
    color: 'rgba(255, 255, 255, 0.6)',
    textAlign: 'center',
    lineHeight: 20,
  },
});
