import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useLocalSearchParams, useRouter } from 'expo-router';
import { useState } from 'react';
import {
  Alert,
  Linking,
  RefreshControl,
  ScrollView,
  StyleSheet,
  View,
} from 'react-native';
import {
  ActivityIndicator,
  Avatar,
  Button,
  Card,
  Chip,
  Divider,
  FAB,
  HelperText,
  IconButton,
  Surface,
  Text,
  useTheme,
} from 'react-native-paper';
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
  const { colors } = useTheme();
  const isReserved = !!item.reserved_by_user_id || item.purchased_by_user_id;
  const isPurchased = !!item.purchased_by_user_id;

  return (
    <Card style={[styles.giftItemCard, { backgroundColor: colors.surface }]}>
      {item.image_url && <Card.Cover source={{ uri: item.image_url }} />}

      <Card.Content>
        <View style={styles.itemHeader}>
          <Text
            variant="titleMedium"
            style={[styles.itemTitle, { color: colors.onSurface }]}
          >
            {item.name}
          </Text>
          {item.price !== undefined && item.price !== null && (
            <Text
              variant="titleMedium"
              style={[styles.itemPrice, { color: colors.primary }]}
            >
              ${item.price.toFixed(2)}
            </Text>
          )}
        </View>

        {item.description && (
          <Text
            variant="bodyMedium"
            style={{ color: colors.onSurfaceVariant, marginBottom: 8 }}
          >
            {item.description}
          </Text>
        )}

        {item.link && (
          <Button
            mode="text"
            // biome-ignore lint/style/noNonNullAssertion: Not nullable
            onPress={() => Linking.openURL(item.link!)}
            compact
            style={{ alignSelf: 'flex-start' }}
          >
            Visit Website
          </Button>
        )}

        <View style={styles.itemFooter}>
          {isPurchased ? (
            <Chip
              mode="outlined"
              icon="check-circle"
              style={{
                borderColor: colors.primary,
                backgroundColor: colors.primaryContainer,
              }}
            >
              <Text style={{ color: colors.onPrimary, fontWeight: '600' }}>
                Purchased
              </Text>
            </Chip>
          ) : isReserved ? (
            <Chip
              mode="outlined"
              icon="lock"
              style={{
                borderColor: colors.secondary,
                backgroundColor: colors.secondaryContainer,
              }}
            >
              <Text style={{ color: colors.onSecondary, fontWeight: '600' }}>
                Reserved
              </Text>
            </Chip>
          ) : (
            <Chip
              mode="outlined"
              icon="gift-open"
              style={{ borderColor: colors.primary }}
            >
              <Text style={{ color: colors.primary, fontWeight: '600' }}>
                Available
              </Text>
            </Chip>
          )}

          {item.priority > 0 && (
            <Chip mode="outlined">
              <Text>Priority: {item.priority}/10</Text>
            </Chip>
          )}
        </View>
      </Card.Content>

      <Card.Actions style={styles.cardActions}>
        {!isReserved && !isPurchased && (
          <Button
            mode="contained"
            onPress={() => onReserve(item)}
            icon="gift"
            buttonColor={colors.primary}
            textColor={colors.onPrimary}
          >
            Reserve
          </Button>
        )}
        <Button
          mode="outlined"
          onPress={() => onEdit(item)}
          icon="pencil"
          textColor={colors.onSurfaceVariant}
        >
          Edit
        </Button>
      </Card.Actions>
    </Card>
  );
};

export default function WishListScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const router = useRouter();
  const queryClient = useQueryClient();
  const { colors } = useTheme();

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
    // In a real app, this would call the reservation API
    // For now, we'll just show an alert
    Alert.alert(
      'Reserve Gift',
      `You are reserving "${item.name}". In a real app, this would create a reservation.`,
      [
        { text: 'Cancel', style: 'cancel' },
        {
          text: 'Reserve',
          onPress: () => {
            Alert.alert('Success', 'Gift reserved successfully!');
            // Invalidate the query to refetch data
            queryClient
              .invalidateQueries({ queryKey: ['giftItems', id] })
              .catch(console.error);
          },
        },
      ],
    );
  };

  const handleEditGift = (item: GiftItem) => {
    // Navigate to edit gift item screen
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
      <View style={[styles.container, { backgroundColor: colors.background }]}>
        <View style={styles.centerContainer}>
          <ActivityIndicator
            animating={true}
            size="large"
            color={colors.primary}
          />
          <Text
            variant="bodyLarge"
            style={{ marginTop: 10, color: colors.onSurface }}
          >
            Loading wishlist...
          </Text>
        </View>
      </View>
    );
  }

  if (error || itemsError) {
    return (
      <View style={[styles.container, { backgroundColor: colors.background }]}>
        <Surface style={styles.errorSurface}>
          <Text
            variant="headlineMedium"
            style={{ color: colors.error, marginBottom: 10 }}
          >
            Error loading wishlist
          </Text>
          <HelperText type="error">
            {error?.message || itemsError?.message}
          </HelperText>
          <Button
            mode="contained"
            onPress={() => router.back()}
            style={{ marginTop: 16 }}
            buttonColor={colors.primary}
            textColor={colors.onPrimary}
          >
            Back
          </Button>
        </Surface>
      </View>
    );
  }

  if (!wishList) {
    return (
      <View style={[styles.container, { backgroundColor: colors.background }]}>
        <Surface style={styles.errorSurface}>
          <Text
            variant="headlineMedium"
            style={{ color: colors.error, marginBottom: 10 }}
          >
            Wishlist not found
          </Text>
          <Button
            mode="contained"
            onPress={() => router.back()}
            style={{ marginTop: 16 }}
            buttonColor={colors.primary}
            textColor={colors.onPrimary}
          >
            Back
          </Button>
        </Surface>
      </View>
    );
  }

  return (
    <View style={[styles.container, { backgroundColor: colors.background }]}>
      {/* Header */}
      <Surface
        style={[styles.headerSurface, { backgroundColor: colors.surface }]}
      >
        <View style={styles.headerRow}>
          <View style={{ flex: 1 }}>
            <Text
              variant="headlineSmall"
              style={[styles.headerTitle, { color: colors.onSurface }]}
            >
              {wishList.title}
            </Text>
            <Text
              variant="bodyMedium"
              style={{ color: colors.onSurfaceVariant, marginVertical: 4 }}
            >
              {wishList.occasion ? `${wishList.occasion}` : 'Personal Wishlist'}
            </Text>
            <Text
              variant="bodyMedium"
              style={{ color: colors.onSurfaceVariant }}
            >
              {wishList.description || 'No description'}
            </Text>
          </View>

          <IconButton
            icon="pencil"
            size={24}
            onPress={handleEditWishList}
            mode="contained-tonal"
            containerColor={colors.primaryContainer}
            iconColor={colors.onPrimaryContainer}
          />
        </View>

        {/* Stats */}
        <Divider style={{ marginVertical: 16 }} />
        <View style={styles.statsContainer}>
          <View style={styles.statItem}>
            <Text variant="headlineMedium" style={{ color: colors.primary }}>
              {giftItems?.length || 0}
            </Text>
            <Text
              variant="labelLarge"
              style={{ color: colors.onSurfaceVariant }}
            >
              Items
            </Text>
          </View>
          <View style={styles.statItem}>
            <Text variant="headlineMedium" style={{ color: colors.secondary }}>
              {giftItems?.filter((item) => item.reserved_by_user_id).length ||
                0}
            </Text>
            <Text
              variant="labelLarge"
              style={{ color: colors.onSurfaceVariant }}
            >
              Reserved
            </Text>
          </View>
          <View style={styles.statItem}>
            <Text variant="headlineMedium" style={{ color: colors.primary }}>
              {giftItems?.filter((item) => item.purchased_by_user_id).length ||
                0}
            </Text>
            <Text
              variant="labelLarge"
              style={{ color: colors.onSurfaceVariant }}
            >
              Purchased
            </Text>
          </View>
        </View>
      </Surface>

      {/* Gift Items List */}
      <ScrollView
        style={styles.scrollContainer}
        refreshControl={
          <RefreshControl
            refreshing={refreshing}
            onRefresh={onRefresh}
            colors={[colors.primary]}
            progressBackgroundColor={colors.surface}
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
          <Surface style={styles.emptyState}>
            <Avatar.Icon
              size={64}
              icon="gift"
              style={{
                backgroundColor: colors.primaryContainer,
                marginBottom: 16,
              }}
              color={colors.onPrimaryContainer}
            />
            <Text
              variant="headlineMedium"
              style={{ color: colors.onSurface, marginBottom: 8 }}
            >
              No gift items yet
            </Text>
            <Text
              variant="bodyMedium"
              style={{ color: colors.onSurfaceVariant, textAlign: 'center' }}
            >
              Add your first gift item to get started
            </Text>
          </Surface>
        )}
      </ScrollView>

      {/* Floating Action Button for adding gift items */}
      <FAB
        icon="plus"
        style={styles.fab}
        onPress={handleAddGiftItem}
        label="Add Item"
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
  errorSurface: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 20,
    margin: 16,
    borderRadius: 8,
  },
  headerSurface: {
    padding: 16,
    borderBottomLeftRadius: 24,
    borderBottomRightRadius: 24,
    elevation: 4,
  },
  headerRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    marginBottom: 8,
  },
  headerTitle: {
    fontSize: 24,
    fontWeight: 'bold',
  },
  statsContainer: {
    flexDirection: 'row',
    justifyContent: 'space-around',
  },
  statItem: {
    alignItems: 'center',
  },
  scrollContainer: {
    flex: 1,
    padding: 16,
  },
  listContainer: {
    gap: 12,
  },
  giftItemCard: {
    marginBottom: 12,
    borderRadius: 12,
    elevation: 4,
  },
  itemHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 8,
  },
  itemTitle: {
    fontSize: 18,
    fontWeight: '600',
    flex: 1,
  },
  itemPrice: {
    fontWeight: 'bold',
    fontSize: 16,
  },
  itemFooter: {
    flexDirection: 'row',
    justifyContent: 'flex-start',
    alignItems: 'center',
    gap: 8,
    marginTop: 8,
  },
  cardActions: {
    justifyContent: 'flex-end',
    padding: 8,
  },
  emptyState: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 20,
    marginVertical: 16,
    borderRadius: 12,
  },
  fab: {
    position: 'absolute',
    margin: 16,
    right: 0,
    bottom: 0,
  },
});
