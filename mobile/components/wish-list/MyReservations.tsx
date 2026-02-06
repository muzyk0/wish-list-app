import { MaterialCommunityIcons } from '@expo/vector-icons';
import { useEffect, useState } from 'react';
import { FlatList, RefreshControl, StyleSheet, View } from 'react-native';
import { ActivityIndicator, Text } from 'react-native-paper';
import { BlurView } from 'expo-blur';
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
          <MaterialCommunityIcons name="bookmark-outline" size={64} color="#FFD700" />
          <Text style={styles.emptyTitle}>No reservations yet</Text>
          <Text style={styles.emptyText}>
            When you reserve items, they'll appear here
          </Text>
        </BlurView>
      </View>
    );
  }

  return (
    <View style={styles.container}>
      <FlatList
        data={reservations}
        keyExtractor={(item) => item.id}
        renderItem={({ item }) => (
          <ReservationItem reservation={item} onRefresh={fetchReservations} />
        )}
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
  listContent: {
    paddingBottom: 100,
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
