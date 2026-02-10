import { MaterialCommunityIcons } from '@expo/vector-icons';
import { useQuery } from '@tanstack/react-query';
import { BlurView } from 'expo-blur';
import { LinearGradient } from 'expo-linear-gradient';
import { StyleSheet, View } from 'react-native';
import { Text } from 'react-native-paper';

interface HomeStats {
  totalItems: number;
  reserved: number;
  purchased: number;
}

export function StatsRow() {
  // TODO: Add real endpoint to fetch home stats
  // Backend should calculate counts, not send all items
  const { data: stats = { totalItems: 0, reserved: 0, purchased: 0 } } =
    useQuery<HomeStats>({
      queryKey: ['home-stats'],
      queryFn: async () => {
        // Mock: return empty stats
        // TODO: Replace with: apiClient.getHomeStats()
        return {
          totalItems: 0,
          reserved: 0,
          purchased: 0,
        };
      },
      retry: 2,
    });

  const { totalItems, reserved, purchased } = stats;
  return (
    <View style={styles.statsRow}>
      <BlurView intensity={20} style={styles.statCard}>
        <LinearGradient
          colors={['#FFD700', '#FFA500']}
          style={styles.statGradient}
        >
          <MaterialCommunityIcons name="gift" size={32} color="#000000" />
          <Text style={styles.statValue}>{totalItems}</Text>
          <Text style={styles.statLabelDark}>Total Items</Text>
        </LinearGradient>
      </BlurView>

      <BlurView intensity={20} style={styles.statCard}>
        <View style={styles.statContent}>
          <MaterialCommunityIcons name="lock" size={28} color="#FF9800" />
          <Text style={[styles.statValue, { color: '#FF9800' }]}>
            {reserved}
          </Text>
          <Text style={styles.statLabel}>Reserved</Text>
        </View>
      </BlurView>

      <BlurView intensity={20} style={styles.statCard}>
        <View style={styles.statContent}>
          <MaterialCommunityIcons
            name="check-circle"
            size={28}
            color="#4CAF50"
          />
          <Text style={[styles.statValue, { color: '#4CAF50' }]}>
            {purchased}
          </Text>
          <Text style={styles.statLabel}>Purchased</Text>
        </View>
      </BlurView>
    </View>
  );
}

const styles = StyleSheet.create({
  statsRow: {
    flexDirection: 'row',
    gap: 10,
    marginBottom: 24,
  },
  statCard: {
    flex: 1,
    borderRadius: 16,
    overflow: 'hidden',
    backgroundColor: 'rgba(255, 255, 255, 0.08)',
    borderWidth: 1,
    borderColor: 'rgba(255, 255, 255, 0.1)',
  },
  statGradient: {
    padding: 16,
    alignItems: 'center',
    justifyContent: 'center',
  },
  statContent: {
    padding: 16,
    alignItems: 'center',
    justifyContent: 'center',
  },
  statValue: {
    fontSize: 24,
    fontWeight: '800',
    color: '#000000',
    marginTop: 8,
    marginBottom: 2,
  },
  statLabel: {
    fontSize: 12,
    color: 'rgba(255, 255, 255, 0.5)',
    fontWeight: '600',
  },
  statLabelDark: {
    fontSize: 11,
    color: 'rgba(0, 0, 0, 0.7)',
    fontWeight: '600',
  },
});
