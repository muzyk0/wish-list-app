import { useQuery } from "@tanstack/react-query";
import { useLocalSearchParams, useRouter } from "expo-router";

import {
  ActivityIndicator,
  ScrollView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from "react-native";
import { apiClient } from "@/lib/api";
import GiftItemForm from "../../../components/wish-list/GiftItemForm";

export default function EditGiftItemScreen() {
  // biome-ignore lint/correctness/noUnusedVariables: Temp
  const { id, wishlistId } = useLocalSearchParams<{
    id: string;
    wishlistId: string;
  }>();
  const router = useRouter();

  const {
    data: giftItem,
    isLoading,
    error,
  } = useQuery({
    queryKey: ["giftItem", id],
    queryFn: () =>
      apiClient.getGiftItemById(wishlistId as string, id as string),
    enabled: !!id && !!wishlistId,
  });

  if (isLoading) {
    return (
      <View style={styles.centerContainer}>
        <ActivityIndicator size="large" color="#007AFF" />
      </View>
    );
  }

  if (error) {
    return (
      <View style={styles.centerContainer}>
        <Text>Error loading gift item</Text>
        <Text>{error.message}</Text>
        <TouchableOpacity
          style={styles.retryButton}
          onPress={() => router.back()}
        >
          <Text style={styles.buttonText}>Back</Text>
        </TouchableOpacity>
      </View>
    );
  }

  return (
    <ScrollView style={styles.container}>
      <Text style={styles.title}>Edit Gift Item</Text>

      {giftItem && (
        <GiftItemForm
          wishlistId={giftItem.wishlist_id}
          existingItem={giftItem}
          onComplete={() =>
            router.push({
              pathname: `/lists/[id]`,
              params: { id: giftItem.wishlist_id },
            })
          }
        />
      )}
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    padding: 20,
    backgroundColor: "#fff",
  },
  centerContainer: {
    flex: 1,
    justifyContent: "center",
    alignItems: "center",
  },
  title: {
    fontSize: 24,
    fontWeight: "bold",
    textAlign: "center",
    marginBottom: 20,
  },
  retryButton: {
    backgroundColor: "#007AFF",
    padding: 10,
    borderRadius: 4,
    marginTop: 10,
  },
  buttonText: {
    color: "#fff",
    fontSize: 16,
    fontWeight: "bold",
  },
});
