import { useEffect, useState } from 'react';
import { FlatList, RefreshControl, StyleSheet, View } from 'react-native';
import { ActivityIndicator, Card, Text, useTheme } from 'react-native-paper';
import { ReservationItem } from './ReservationItem';

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

export function MyReservations() {
  const theme = useTheme();
  const [reservations, setReservations] = useState<Reservation[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);

  const fetchReservations = async () => {
    try {
      setLoading(true);

      // Check if user is authenticated
      const userResponse = await fetch('/api/auth/me');
      if (userResponse.ok) {
        // Authenticated user - fetch user reservations
        const response = await fetch('/api/users/me/reservations');
        if (response.ok) {
          const data = await response.json();
          setReservations(data.data || []);
        }
      } else {
        // Guest user - check for reservation token in AsyncStorage
        const AsyncStorage = await import(
          '@react-native-async-storage/async-storage'
        );
        const token = await AsyncStorage.default.getItem('reservationToken');
        if (token) {
          const response = await fetch(
            `/api/guest/reservations?token=${token}`,
          );
          if (response.ok) {
            const data = await response.json();
            setReservations(data || []);
          }
        }
      }
    } catch (error) {
      console.error('Failed to load reservations:', error);
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

  if (loading) {
    return (
      <View
        style={[styles.container, { backgroundColor: theme.colors.background }]}
      >
        <ActivityIndicator
          animating={true}
          color={theme.colors.primary}
          size="large"
        />
        <Text style={styles.loadingText}>Loading reservations...</Text>
      </View>
    );
  }

  if (reservations.length === 0) {
    return (
      <View
        style={[styles.container, { backgroundColor: theme.colors.background }]}
      >
        <Card
          style={[styles.emptyCard, { backgroundColor: theme.colors.surface }]}
        >
          <Card.Content>
            <Text style={styles.emptyText}>You have no reservations yet.</Text>
          </Card.Content>
        </Card>
      </View>
    );
  }

  return (
    <View
      style={[styles.container, { backgroundColor: theme.colors.background }]}
    >
      <Text variant="headlineMedium" style={styles.title}>
        My Reservations
      </Text>

      <FlatList
        data={reservations}
        keyExtractor={(item) => item.id}
        renderItem={({ item }) => (
          <ReservationItem reservation={item} onRefresh={fetchReservations} />
        )}
        contentContainerStyle={styles.listContent}
        refreshControl={
          <RefreshControl
            refreshing={refreshing}
            onRefresh={onRefresh}
            colors={[theme.colors.primary]}
            progressBackgroundColor={theme.colors.surface}
          />
        }
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    padding: 16,
  },
  title: {
    marginBottom: 16,
  },
  listContent: {
    paddingBottom: 16,
  },
  loadingText: {
    marginTop: 16,
    textAlign: 'center',
  },
  emptyCard: {
    margin: 16,
    padding: 24,
    borderRadius: 12,
  },
  emptyText: {
    textAlign: 'center',
    fontSize: 16,
  },
});
