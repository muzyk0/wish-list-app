import { MaterialCommunityIcons } from '@expo/vector-icons';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { BlurView } from 'expo-blur';
import { LinearGradient } from 'expo-linear-gradient';
import { router } from 'expo-router';
import { useState } from 'react';
import { Dimensions, Pressable, StyleSheet, View } from 'react-native';
import { ActivityIndicator, Chip, Text } from 'react-native-paper';
import { TabsLayout } from '@/components/TabsLayout';
import { apiClient } from '@/lib/api';
import type { WishList } from '@/lib/api/types';
import { dialog } from '@/stores/dialogStore';

const { width } = Dimensions.get('window');

interface WishListCardProps {
  item: WishList;
  onDelete: (id: string) => void;
}

const WishListCard: React.FC<WishListCardProps> = ({ item, onDelete }) => {
  return (
    <Pressable
      onPress={() =>
        router.push({
          pathname: '/lists/[id]',
          params: { id: item.id },
        })
      }
    >
      <BlurView intensity={20} style={styles.card}>
        <View style={styles.cardInner}>
          {/* Header */}
          <View style={styles.cardHeader}>
            <View style={styles.titleRow}>
              <LinearGradient
                colors={['#6B4EE6', '#9B6DFF']}
                style={styles.iconContainer}
              >
                <MaterialCommunityIcons name="gift" size={24} color="#ffffff" />
              </LinearGradient>
              <View style={styles.titleContainer}>
                <Text style={styles.title} numberOfLines={1}>
                  {item.title}
                </Text>
                {item.occasion && (
                  <Text style={styles.occasion} numberOfLines={1}>
                    {item.occasion}
                  </Text>
                )}
              </View>
            </View>
            {item.is_public && (
              <Chip
                mode="flat"
                style={styles.publicChip}
                textStyle={styles.publicChipText}
                icon={() => (
                  <MaterialCommunityIcons
                    name="earth"
                    size={14}
                    color="#FFD700"
                  />
                )}
              >
                Public
              </Chip>
            )}
          </View>

          {/* Description */}
          {item.description && (
            <Text style={styles.description} numberOfLines={2}>
              {item.description}
            </Text>
          )}

          {/* Stats */}
          <View style={styles.statsRow}>
            <View style={styles.stat}>
              <MaterialCommunityIcons
                name="eye-outline"
                size={16}
                color="rgba(255, 255, 255, 0.5)"
              />
              <Text style={styles.statText}>
                {item.view_count !== '0' ? item.view_count : '0'} views
              </Text>
            </View>
            {item.occasion_date && (
              <View style={styles.stat}>
                <MaterialCommunityIcons
                  name="calendar"
                  size={16}
                  color="rgba(255, 255, 255, 0.5)"
                />
                <Text style={styles.statText}>{item.occasion_date}</Text>
              </View>
            )}
          </View>

          {/* Actions */}
          <View style={styles.actions}>
            <Pressable
              onPress={() =>
                router.push({
                  pathname: '/lists/[id]/edit',
                  params: { id: item.id },
                })
              }
              style={{ flex: 1 }}
            >
              <View style={[styles.actionButton, styles.editButton]}>
                <MaterialCommunityIcons
                  name="pencil"
                  size={18}
                  color="#FFD700"
                />
                <Text style={styles.editButtonText}>Edit</Text>
              </View>
            </Pressable>
            <Pressable onPress={() => onDelete(item.id)} style={{ flex: 1 }}>
              <View style={[styles.actionButton, styles.deleteButton]}>
                <MaterialCommunityIcons
                  name="delete-outline"
                  size={18}
                  color="#FF6B6B"
                />
                <Text style={styles.deleteButtonText}>Delete</Text>
              </View>
            </Pressable>
          </View>
        </View>
      </BlurView>
    </Pressable>
  );
};

export default function ListsTab() {
  const [refreshing, setRefreshing] = useState(false);
  const queryClient = useQueryClient();

  const {
    data: wishLists,
    isLoading,
    isError,
    refetch,
  } = useQuery<WishList[]>({
    queryKey: ['wishlists'],
    queryFn: () => apiClient.getWishLists(),
    retry: 2,
  });

  const onRefresh = async () => {
    setRefreshing(true);
    await refetch();
    setRefreshing(false);
  };

  const handleDelete = (id: string) => {
    dialog.confirm({
      title: 'Confirm Delete',
      message:
        'Are you sure you want to delete this wishlist? This action cannot be undone.',
      confirmLabel: 'Delete',
      cancelLabel: 'Cancel',
      destructive: true,
      onConfirm: async () => {
        try {
          await apiClient.deleteWishList(id);
          dialog.success('Wishlist deleted successfully!');
          queryClient.invalidateQueries({ queryKey: ['wishlists'] });
        } catch (error: any) {
          dialog.error(
            error.message || 'Failed to delete wishlist. Please try again.',
          );
        }
      },
    });
  };

  if (isLoading) {
    return (
      <TabsLayout title="My Lists" subtitle="All your wishlists">
        <View style={styles.loadingContainer}>
          <ActivityIndicator size="large" color="#FFD700" />
          <Text style={styles.loadingText}>Loading wishlists...</Text>
        </View>
      </TabsLayout>
    );
  }

  if (isError) {
    return (
      <TabsLayout title="My Lists" subtitle="All your wishlists">
        <View style={styles.emptyContainer}>
          <MaterialCommunityIcons
            name="alert-circle-outline"
            size={64}
            color="rgba(255, 107, 107, 0.5)"
          />
          <Text style={styles.errorText}>Failed to load wishlists</Text>
          <Pressable onPress={() => refetch()}>
            <View style={styles.retryButton}>
              <Text style={styles.retryButtonText}>Retry</Text>
            </View>
          </Pressable>
        </View>
      </TabsLayout>
    );
  }

  return (
    <TabsLayout
      title="My Lists"
      subtitle="All your wishlists"
      refreshing={refreshing}
      onRefresh={onRefresh}
    >
      {/* Create New Button */}
      <Pressable
        onPress={() => router.push('/lists/create')}
        style={{ marginBottom: 24 }}
      >
        <LinearGradient
          colors={['#FFD700', '#FFA500']}
          start={{ x: 0, y: 0 }}
          end={{ x: 1, y: 0 }}
          style={styles.createButton}
        >
          <MaterialCommunityIcons name="plus" size={24} color="#000000" />
          <Text style={styles.createButtonText}>Create New List</Text>
        </LinearGradient>
      </Pressable>

      {/* Lists */}
      {wishLists && wishLists.length > 0 ? (
        <View style={styles.listsContainer}>
          {wishLists.map((list) => (
            <WishListCard key={list.id} item={list} onDelete={handleDelete} />
          ))}
        </View>
      ) : (
        <View style={styles.emptyContainer}>
          <MaterialCommunityIcons
            name="gift-outline"
            size={64}
            color="rgba(255, 255, 255, 0.3)"
          />
          <Text style={styles.emptyText}>No wishlists yet</Text>
          <Text style={styles.emptySubtext}>
            Create your first wishlist to get started
          </Text>
        </View>
      )}
    </TabsLayout>
  );
}

const styles = StyleSheet.create({
  loadingContainer: {
    paddingVertical: 60,
    alignItems: 'center',
  },
  loadingText: {
    fontSize: 16,
    color: 'rgba(255, 255, 255, 0.6)',
    marginTop: 16,
  },
  createButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 16,
    borderRadius: 16,
    gap: 8,
  },
  createButtonText: {
    fontSize: 16,
    fontWeight: '700',
    color: '#000000',
  },
  listsContainer: {
    gap: 16,
  },
  card: {
    borderRadius: 20,
    overflow: 'hidden',
    backgroundColor: 'rgba(255, 255, 255, 0.08)',
    borderWidth: 1,
    borderColor: 'rgba(255, 255, 255, 0.1)',
  },
  cardInner: {
    padding: 20,
  },
  cardHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    marginBottom: 12,
  },
  titleRow: {
    flexDirection: 'row',
    alignItems: 'center',
    flex: 1,
    marginRight: 12,
  },
  iconContainer: {
    width: 48,
    height: 48,
    borderRadius: 12,
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 12,
  },
  titleContainer: {
    flex: 1,
  },
  title: {
    fontSize: 18,
    fontWeight: '700',
    color: '#ffffff',
    marginBottom: 2,
  },
  occasion: {
    fontSize: 14,
    color: 'rgba(255, 255, 255, 0.6)',
  },
  publicChip: {
    backgroundColor: 'rgba(255, 215, 0, 0.15)',
    borderWidth: 1,
    borderColor: 'rgba(255, 215, 0, 0.3)',
  },
  publicChipText: {
    fontSize: 12,
    fontWeight: '600',
    color: '#FFD700',
  },
  description: {
    fontSize: 14,
    color: 'rgba(255, 255, 255, 0.6)',
    marginBottom: 12,
    lineHeight: 20,
  },
  statsRow: {
    flexDirection: 'row',
    gap: 16,
    marginBottom: 16,
  },
  stat: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
  },
  statText: {
    fontSize: 13,
    color: 'rgba(255, 255, 255, 0.5)',
  },
  actions: {
    flexDirection: 'row',
    gap: 12,
  },
  actionButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 12,
    borderRadius: 12,
    gap: 6,
  },
  editButton: {
    backgroundColor: 'rgba(255, 215, 0, 0.15)',
    borderWidth: 1,
    borderColor: 'rgba(255, 215, 0, 0.3)',
  },
  editButtonText: {
    fontSize: 14,
    fontWeight: '600',
    color: '#FFD700',
  },
  deleteButton: {
    backgroundColor: 'rgba(255, 107, 107, 0.15)',
    borderWidth: 1,
    borderColor: 'rgba(255, 107, 107, 0.3)',
  },
  deleteButtonText: {
    fontSize: 14,
    fontWeight: '600',
    color: '#FF6B6B',
  },
  emptyContainer: {
    paddingVertical: 80,
    alignItems: 'center',
  },
  emptyText: {
    fontSize: 18,
    fontWeight: '600',
    color: 'rgba(255, 255, 255, 0.7)',
    marginTop: 16,
  },
  emptySubtext: {
    fontSize: 14,
    color: 'rgba(255, 255, 255, 0.5)',
    marginTop: 8,
    textAlign: 'center',
  },
  errorText: {
    fontSize: 18,
    fontWeight: '600',
    color: 'rgba(255, 107, 107, 0.8)',
    marginTop: 16,
  },
  retryButton: {
    marginTop: 16,
    paddingVertical: 12,
    paddingHorizontal: 24,
    backgroundColor: 'rgba(255, 215, 0, 0.2)',
    borderRadius: 12,
    borderWidth: 1,
    borderColor: 'rgba(255, 215, 0, 0.4)',
  },
  retryButtonText: {
    fontSize: 14,
    fontWeight: '600',
    color: '#FFD700',
  },
});
