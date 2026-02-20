import { MaterialCommunityIcons } from '@expo/vector-icons';
import { BlurView } from 'expo-blur';
import { LinearGradient } from 'expo-linear-gradient';
import { Pressable, StyleSheet, View } from 'react-native';
import { Text } from 'react-native-paper';
import { dialog } from '@/stores/dialogStore';

interface Reservation {
  id: string;
  giftItem: {
    id: string;
    name: string;
    imageUrl?: string;
    price?: number;
  };
  wishlist: {
    id: string;
    title: string;
    ownerFirstName?: string;
    ownerLastName?: string;
  };
  status: 'active' | 'canceled' | 'fulfilled' | 'expired';
  reservedAt: string;
  expiresAt?: string;
}

interface ReservationItemProps {
  reservation: Reservation;
  onRefresh: () => void;
}

export function ReservationItem({
  reservation,
  onRefresh,
}: ReservationItemProps) {
  const handleCancelReservation = async () => {
    dialog.confirm({
      title: 'Cancel Reservation',
      message: 'Are you sure you want to cancel this reservation?',
      confirmLabel: 'Yes, Cancel',
      cancelLabel: 'No',
      destructive: true,
      onConfirm: async () => {
        try {
          const response = await fetch(
            `/api/reservations/${reservation.id}/cancel`,
            {
              method: 'POST',
              headers: {
                'Content-Type': 'application/json',
              },
            },
          );

          if (response.ok) {
            dialog.success('Reservation canceled successfully');
            onRefresh();
          } else {
            const data = await response.json();
            dialog.error(data.error || 'Failed to cancel reservation');
          }
        } catch {
          dialog.error('An error occurred while cancelling the reservation');
        }
      },
    });
  };

  const getStatusConfig = () => {
    switch (reservation.status) {
      case 'active':
        return {
          color: '#FFD700',
          icon: 'clock-outline',
          label: 'Active',
        };
      case 'canceled':
        return {
          color: '#9E9E9E',
          icon: 'close-circle-outline',
          label: 'Canceled',
        };
      case 'fulfilled':
        return {
          color: '#4CAF50',
          icon: 'check-circle',
          label: 'Fulfilled',
        };
      case 'expired':
        return {
          color: '#FF6B6B',
          icon: 'alert-circle-outline',
          label: 'Expired',
        };
      default:
        return {
          color: '#9E9E9E',
          icon: 'help-circle-outline',
          label: reservation.status,
        };
    }
  };

  const statusConfig = getStatusConfig();
  const ownerName =
    `${reservation.wishlist.ownerFirstName || ''} ${reservation.wishlist.ownerLastName || ''}`.trim();

  return (
    <BlurView intensity={20} style={styles.card}>
      <View style={styles.cardContent}>
        {/* Header */}
        <View style={styles.header}>
          <View style={styles.itemInfo}>
            <Text style={styles.itemName} numberOfLines={2}>
              {reservation.giftItem.name}
            </Text>
            {reservation.giftItem.price !== undefined && (
              <View style={styles.priceContainer}>
                <LinearGradient
                  colors={['#FFD700', '#FFA500']}
                  style={styles.priceGradient}
                >
                  <Text style={styles.priceText}>
                    ${reservation.giftItem.price.toFixed(2)}
                  </Text>
                </LinearGradient>
              </View>
            )}
          </View>

          <View
            style={[
              styles.statusBadge,
              { backgroundColor: `${statusConfig.color}20` },
            ]}
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

        {/* Wishlist Info */}
        <View style={styles.wishlistInfo}>
          <MaterialCommunityIcons
            name="gift-outline"
            size={16}
            color="rgba(255, 255, 255, 0.5)"
          />
          <Text style={styles.wishlistText} numberOfLines={1}>
            {reservation.wishlist.title}
          </Text>
        </View>

        {ownerName && (
          <View style={styles.ownerInfo}>
            <MaterialCommunityIcons
              name="account-outline"
              size={16}
              color="rgba(255, 255, 255, 0.5)"
            />
            <Text style={styles.ownerText} numberOfLines={1}>
              {ownerName}
            </Text>
          </View>
        )}

        {/* Dates */}
        <View style={styles.datesContainer}>
          <View style={styles.dateItem}>
            <MaterialCommunityIcons
              name="calendar-check"
              size={14}
              color="rgba(255, 255, 255, 0.4)"
            />
            <Text style={styles.dateText}>
              {new Date(reservation.reservedAt).toLocaleDateString()}
            </Text>
          </View>

          {reservation.expiresAt && (
            <View style={styles.dateItem}>
              <MaterialCommunityIcons
                name="calendar-alert"
                size={14}
                color="#FF6B6B"
              />
              <Text style={[styles.dateText, { color: '#FF6B6B' }]}>
                Expires {new Date(reservation.expiresAt).toLocaleDateString()}
              </Text>
            </View>
          )}
        </View>

        {/* Cancel Button */}
        {reservation.status === 'active' && (
          <Pressable
            onPress={handleCancelReservation}
            style={styles.cancelButtonWrapper}
          >
            <View style={styles.cancelButton}>
              <MaterialCommunityIcons
                name="close"
                size={16}
                color="rgba(255, 255, 255, 0.7)"
              />
              <Text style={styles.cancelButtonText}>Cancel Reservation</Text>
            </View>
          </Pressable>
        )}
      </View>
    </BlurView>
  );
}

const styles = StyleSheet.create({
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
  wishlistInfo: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    marginBottom: 6,
  },
  wishlistText: {
    fontSize: 14,
    color: 'rgba(255, 255, 255, 0.8)',
    flex: 1,
  },
  ownerInfo: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    marginBottom: 12,
  },
  ownerText: {
    fontSize: 13,
    color: 'rgba(255, 255, 255, 0.6)',
    flex: 1,
  },
  datesContainer: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 12,
    marginBottom: 12,
  },
  dateItem: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
  },
  dateText: {
    fontSize: 12,
    color: 'rgba(255, 255, 255, 0.5)',
  },
  cancelButtonWrapper: {
    marginTop: 4,
  },
  cancelButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 10,
    paddingHorizontal: 16,
    borderRadius: 10,
    backgroundColor: 'rgba(255, 255, 255, 0.05)',
    borderWidth: 1,
    borderColor: 'rgba(255, 107, 107, 0.3)',
    gap: 6,
  },
  cancelButtonText: {
    fontSize: 13,
    fontWeight: '600',
    color: 'rgba(255, 255, 255, 0.7)',
  },
});
