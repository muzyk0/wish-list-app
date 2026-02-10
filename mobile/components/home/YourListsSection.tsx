import { MaterialCommunityIcons } from '@expo/vector-icons';
import { useQuery } from '@tanstack/react-query';
import { Pressable, StyleSheet, View } from 'react-native';
import { Text } from 'react-native-paper';
import { apiClient } from '@/lib/api';
import type { WishList } from '@/lib/api/types';
import { CompactListCard } from './CompactListCard';

interface YourListsSectionProps {
  onAddPress: () => void;
  onListPress: (listId: string) => void;
}

export function YourListsSection({
  onAddPress,
  onListPress,
}: YourListsSectionProps) {
  // Fetch wishlists
  const { data: lists = [] } = useQuery<WishList[]>({
    queryKey: ['wishlists'],
    queryFn: () => apiClient.getWishLists(),
    retry: 2,
  });

  if (lists.length === 0) {
    return null;
  }

  return (
    <View style={styles.section}>
      <View style={styles.sectionHeader}>
        <Text style={styles.sectionTitle}>Your Lists</Text>
        <Pressable onPress={onAddPress}>
          <MaterialCommunityIcons
            name="plus-circle"
            size={24}
            color="#FFD700"
          />
        </Pressable>
      </View>

      <View style={styles.compactListsContainer}>
        {lists.slice(0, 4).map((list, index) => (
          <CompactListCard
            key={list.id}
            title={list.title}
            itemCount={list.item_count || 0}
            onPress={() => onListPress(list.id)}
            index={index}
          />
        ))}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  section: {
    marginBottom: 28,
  },
  sectionHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 16,
  },
  sectionTitle: {
    fontSize: 20,
    fontWeight: '700',
    color: '#ffffff',
  },
  compactListsContainer: {
    gap: 8,
  },
});
