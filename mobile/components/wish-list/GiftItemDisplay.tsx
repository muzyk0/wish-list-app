import { useMutation } from '@tanstack/react-query';
import { Alert, StyleSheet, Text, TouchableOpacity, View } from 'react-native';
import { Badge } from '@/components/ui/Badge';
import { apiClient } from '@/lib/api';
import type { GiftItem } from '@/lib/types';

interface GiftItemDisplayProps {
  item: GiftItem;
  onRefresh?: () => void; // Callback to refresh the list after purchase
  showPurchaseOption?: boolean; // Whether to show purchase option (typically for owner)
}

export default function GiftItemDisplay({
  item,
  onRefresh,
  showPurchaseOption = false,
}: GiftItemDisplayProps) {
  const isReserved = !!item.reserved_by_user_id;
  const isPurchased = !!item.purchased_by_user_id;

  const purchaseMutation = useMutation({
    mutationFn: (giftItemId: string) =>
      apiClient.markGiftItemAsPurchased(giftItemId, item.price || 0),
    onSuccess: () => {
      Alert.alert('Success', 'Gift item marked as purchased successfully!');
      if (onRefresh) {
        onRefresh();
      }
    },
    // biome-ignore lint/suspicious/noExplicitAny: Error type
    onError: (error: any) => {
      Alert.alert(
        'Error',
        error.message ||
          'Failed to mark gift item as purchased. Please try again.',
      );
    },
  });

  const handlePurchase = () => {
    if (!item.id) return;

    Alert.alert(
      'Confirm Purchase',
      `Are you sure you want to mark "${item.name}" as purchased?`,
      [
        { text: 'Cancel', style: 'cancel' },
        {
          text: 'Mark as Purchased',
          style: 'default',
          onPress: () => purchaseMutation.mutate(item.id),
        },
      ],
    );
  };

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        <Text style={styles.name} numberOfLines={2}>
          {item.name}
        </Text>
        <View style={styles.statusContainer}>
          {isPurchased && <Badge>Purchased</Badge>}
          {isReserved && !isPurchased && <Badge>Reserved</Badge>}
        </View>
      </View>

      {item.description ? (
        <Text style={styles.description} numberOfLines={3}>
          {item.description}
        </Text>
      ) : null}

      <View style={styles.details}>
        {item.price ? <Text style={styles.price}>${item.price}</Text> : null}

        {item.priority > 0 ? (
          <Text style={styles.priority}>Priority: {item.priority}/10</Text>
        ) : null}
      </View>

      {item.link ? (
        <Text style={styles.link} numberOfLines={1}>
          {item.link}
        </Text>
      ) : null}

      {showPurchaseOption && !isPurchased && (
        <View style={styles.purchaseContainer}>
          <TouchableOpacity
            style={styles.purchaseButton}
            onPress={handlePurchase}
            disabled={purchaseMutation.isPending}
          >
            {purchaseMutation.isPending ? (
              <Text style={styles.buttonText}>Processing...</Text>
            ) : (
              <Text style={styles.buttonText}>Mark as Purchased</Text>
            )}
          </TouchableOpacity>
        </View>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    backgroundColor: '#f9f9f9',
    padding: 15,
    marginVertical: 5,
    borderRadius: 8,
    borderWidth: 1,
    borderColor: '#eee',
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    marginBottom: 8,
  },
  name: {
    fontSize: 16,
    fontWeight: '600',
    flex: 1,
    marginRight: 10,
  },
  statusContainer: {
    flexDirection: 'row',
    gap: 5,
  },
  description: {
    fontSize: 14,
    color: '#666',
    marginBottom: 8,
  },
  details: {
    flexDirection: 'row',
    gap: 15,
    marginBottom: 8,
  },
  price: {
    fontSize: 14,
    fontWeight: 'bold',
    color: '#007AFF',
  },
  priority: {
    fontSize: 12,
    backgroundColor: '#e5e7eb',
    paddingHorizontal: 6,
    paddingVertical: 2,
    borderRadius: 4,
  },
  link: {
    fontSize: 12,
    color: '#007AFF',
    textDecorationLine: 'underline',
    marginBottom: 10,
  },
  purchaseContainer: {
    marginTop: 10,
    alignItems: 'flex-end',
  },
  purchaseButton: {
    backgroundColor: '#34C759',
    paddingHorizontal: 15,
    paddingVertical: 8,
    borderRadius: 6,
  },
  buttonText: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '600',
  },
});
