import { Alert, StyleSheet, View } from 'react-native';
import {
  Button,
  Card,
  Paragraph,
  Text,
  Title,
  useTheme,
} from 'react-native-paper';

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
  status: 'active' | 'cancelled' | 'fulfilled' | 'expired';
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
  const theme = useTheme();

  const handleCancelReservation = async () => {
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
        Alert.alert('Success', 'Reservation cancelled successfully');
        onRefresh(); // Refresh the list
      } else {
        const data = await response.json();
        Alert.alert('Error', data.error || 'Failed to cancel reservation');
      }
    } catch {
      Alert.alert(
        'Error',
        'An error occurred while cancelling the reservation',
      );
    }
  };

  const getStatusColor = () => {
    switch (reservation.status) {
      case 'active':
        return theme.colors.primary;
      case 'cancelled':
        return theme.colors.outline;
      case 'fulfilled':
        return theme.colors.secondary;
      case 'expired':
        return theme.colors.error;
      default:
        return theme.colors.outline;
    }
  };

  const getStatusBackgroundColor = () => {
    switch (reservation.status) {
      case 'active':
        return `${theme.colors.primary}20`;
      case 'cancelled':
        return `${theme.colors.outline}20`;
      case 'fulfilled':
        return `${theme.colors.secondary}20`;
      case 'expired':
        return `${theme.colors.error}20`;
      default:
        return `${theme.colors.outline}20`;
    }
  };

  return (
    <Card style={[styles.card, { backgroundColor: theme.colors.surface }]}>
      <Card.Content>
        <View style={styles.header}>
          <Title style={styles.itemName}>{reservation.giftItem.name}</Title>
          <View
            style={[
              styles.statusBadge,
              {
                backgroundColor: getStatusBackgroundColor(),
                borderColor: getStatusColor(),
              },
            ]}
          >
            <Text style={[styles.statusText, { color: getStatusColor() }]}>
              {reservation.status.charAt(0).toUpperCase() +
                reservation.status.slice(1)}
            </Text>
          </View>
        </View>

        <Paragraph style={styles.details}>
          Reserved for: {reservation.wishlist.title} by{' '}
          {reservation.wishlist.ownerFirstName}{' '}
          {reservation.wishlist.ownerLastName}
        </Paragraph>

        <View style={styles.footer}>
          <View style={styles.dateInfo}>
            <Text style={styles.smallText}>
              Reserved on:{' '}
              {new Date(reservation.reservedAt).toLocaleDateString()}
            </Text>
            {reservation.expiresAt && (
              <Text style={[styles.smallText, { color: theme.colors.error }]}>
                Expires on:{' '}
                {new Date(reservation.expiresAt).toLocaleDateString()}
              </Text>
            )}
          </View>

          {reservation.status === 'active' && (
            <Button
              mode="contained"
              onPress={handleCancelReservation}
              style={styles.cancelButton}
              buttonColor={theme.colors.error}
              labelStyle={styles.cancelButtonText}
            >
              Cancel
            </Button>
          )}
        </View>
      </Card.Content>
    </Card>
  );
}

const styles = StyleSheet.create({
  card: {
    margin: 8,
    borderRadius: 16,
    elevation: 4,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 12,
  },
  itemName: {
    flex: 1,
    marginRight: 12,
  },
  statusBadge: {
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 20,
    borderWidth: 1,
    minWidth: 80,
    alignItems: 'center',
  },
  statusText: {
    fontSize: 12,
    fontWeight: '600',
    textAlign: 'center',
  },
  details: {
    marginBottom: 12,
    lineHeight: 20,
  },
  footer: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  dateInfo: {
    flex: 1,
  },
  smallText: {
    fontSize: 12,
    opacity: 0.8,
  },
  cancelButton: {
    borderRadius: 20,
  },
  cancelButtonText: {
    color: '#fff',
    fontSize: 12,
    fontWeight: '600',
  },
});
