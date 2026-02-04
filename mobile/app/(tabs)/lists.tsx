import { useQuery, useQueryClient } from "@tanstack/react-query";
import { router } from "expo-router";
import { useState } from "react";
import {
  Alert,
  RefreshControl,
  ScrollView,
  StyleSheet,
  View,
} from "react-native";
import {
  ActivityIndicator,
  Appbar,
  Button,
  Card,
  Chip,
  List,
  Text,
  useTheme,
} from "react-native-paper";
import { apiClient } from "@/lib/api";
import type { WishList } from "@/lib/api/types";

export default function ListsTab() {
  const [refreshing, setRefreshing] = useState(false);
  const queryClient = useQueryClient();
  const { colors } = useTheme();

  const {
    data: wishLists,
    isLoading,
    isError,
    refetch,
  } = useQuery<WishList[]>({
    queryKey: ["wishlists"],
    queryFn: () => apiClient.getWishLists(),
    retry: 2,
  });

  const onRefresh = async () => {
    setRefreshing(true);
    await refetch();
    setRefreshing(false);
  };

  const handleDelete = (id: string) => {
    Alert.alert(
      "Confirm Delete",
      "Are you sure you want to delete this wishlist? This action cannot be undone and will also delete all associated gift items.",
      [
        { text: "Cancel", style: "cancel" },
        {
          text: "Delete",
          style: "destructive",
          onPress: async () => {
            try {
              await apiClient.deleteWishList(id);
              Alert.alert("Success", "Wishlist deleted successfully!");
              queryClient.invalidateQueries({ queryKey: ["wishlists"] });
              // biome-ignore lint/suspicious/noExplicitAny: Error type
            } catch (error: any) {
              Alert.alert(
                "Error",
                error.message || "Failed to delete wishlist. Please try again.",
              );
            }
          },
        },
      ],
    );
  };

  const renderWishList = ({ item }: { item: WishList }) => (
    <Card style={[styles.listItem, { backgroundColor: colors.surface }]}>
      <Card.Content style={styles.cardContent}>
        <View style={styles.listHeader}>
          <View style={styles.titleContainer}>
            <Text
              variant="titleMedium"
              style={[styles.listTitle, { color: colors.onSurface }]}
              numberOfLines={1}
            >
              {item.title}
            </Text>
            {item.is_public && (
              <Chip
                mode="outlined"
                style={styles.publicBadge}
                textStyle={{ fontSize: 12 }}
              >
                Public
              </Chip>
            )}
          </View>

          {item.occasion && (
            <Text
              variant="bodyMedium"
              style={[styles.listOccasion, { color: colors.outline }]}
              numberOfLines={1}
            >
              {item.occasion}
            </Text>
          )}

          {item.description && (
            <Text
              variant="bodySmall"
              style={[
                styles.listDescription,
                { color: colors.onSurfaceVariant },
              ]}
              numberOfLines={2}
            >
              {item.description}
            </Text>
          )}
        </View>

        <View style={styles.listStats}>
          <Text
            variant="bodySmall"
            style={[styles.listStat, { color: colors.outline }]}
          >
            {item.view_count !== "0"
              ? `${item.view_count} views`
              : "Not viewed"}
          </Text>
          <Text
            variant="bodySmall"
            style={[styles.listStat, { color: colors.outline }]}
          >
            {item.occasion_date || "No date set"}
          </Text>
        </View>

        <View style={styles.listActions}>
          <Button
            mode="contained-tonal"
            onPress={() =>
              router.push({
                pathname: "/lists/[id]/edit",
                params: { id: item.id },
              })
            }
            style={styles.actionButton}
            labelStyle={styles.actionButtonText}
          >
            Edit
          </Button>
          <Button
            mode="contained-tonal"
            onPress={() => handleDelete(item.id)}
            style={[
              styles.actionButton,
              { backgroundColor: colors.errorContainer },
            ]}
            labelStyle={[
              styles.actionButtonText,
              { color: colors.onErrorContainer },
            ]}
          >
            Delete
          </Button>
        </View>
      </Card.Content>

      <Card.Actions style={styles.cardActions}>
        <Button
          onPress={() =>
            router.push({
              pathname: "/lists/[id]",
              params: { id: item.id },
            })
          }
          mode="contained"
          style={styles.viewButton}
          labelStyle={styles.viewButtonText}
        >
          View List
        </Button>
      </Card.Actions>
    </Card>
  );

  if (isLoading) {
    return (
      <View
        style={[styles.centerContainer, { backgroundColor: colors.background }]}
      >
        <ActivityIndicator size="large" animating={true} />
        <Text
          variant="bodyLarge"
          style={{ marginTop: 10, color: colors.onSurface }}
        >
          Loading wishlists...
        </Text>
      </View>
    );
  }

  if (isError) {
    return (
      <View
        style={[styles.centerContainer, { backgroundColor: colors.background }]}
      >
        <Text
          variant="headlineSmall"
          style={{ color: colors.error, marginBottom: 10 }}
        >
          Error loading wishlists
        </Text>
        <Button
          mode="contained"
          onPress={() => refetch()}
          style={styles.retryButton}
        >
          Retry
        </Button>
      </View>
    );
  }

  return (
    <View style={{ flex: 1, backgroundColor: colors.background }}>
      <Appbar.Header style={{ backgroundColor: colors.primary }}>
        <Appbar.Content
          title="My Wish Lists"
          titleStyle={{ color: colors.onPrimary }}
        />
        <Appbar.Action
          icon="plus"
          onPress={() => router.push("/lists/create")}
          color={colors.onPrimary}
        />
      </Appbar.Header>

      <ScrollView
        contentContainerStyle={styles.listContainer}
        refreshControl={
          <RefreshControl
            refreshing={refreshing}
            onRefresh={onRefresh}
            colors={[colors.primary]}
            progressBackgroundColor={colors.surface}
          />
        }
      >
        {wishLists && wishLists.length > 0 ? (
          wishLists.map((item) => (
            <View key={item.id} style={styles.listItemContainer}>
              {renderWishList({ item })}
            </View>
          ))
        ) : (
          <View style={styles.emptyContainer}>
            <List.Icon icon="playlist-star" />
            <Text
              variant="headlineSmall"
              style={{ color: colors.onSurface, textAlign: "center" }}
            >
              No wish lists yet
            </Text>
            <Text
              variant="bodyMedium"
              style={{
                color: colors.onSurfaceVariant,
                textAlign: "center",
                marginTop: 8,
              }}
            >
              Create your first wish list to get started
            </Text>
            <Button
              mode="contained"
              onPress={() => router.push("/lists/create")}
              style={styles.createButton}
              labelStyle={styles.createButtonText}
            >
              Create List
            </Button>
          </View>
        )}
      </ScrollView>
    </View>
  );
}

const styles = StyleSheet.create({
  centerContainer: {
    flex: 1,
    justifyContent: "center",
    alignItems: "center",
  },
  listContainer: {
    padding: 16,
  },
  listItemContainer: {
    marginBottom: 12,
  },
  listItem: {
    borderRadius: 12,
    elevation: 4,
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
  },
  cardContent: {
    padding: 16,
  },
  listHeader: {
    marginBottom: 12,
  },
  titleContainer: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "flex-start",
    marginBottom: 8,
  },
  listTitle: {
    flex: 1,
    fontSize: 18,
    fontWeight: "600",
  },
  publicBadge: {
    alignSelf: "flex-start",
    marginLeft: 8,
    marginTop: 2,
  },
  listOccasion: {
    fontSize: 14,
    marginBottom: 4,
  },
  listDescription: {
    fontSize: 12,
    lineHeight: 16,
  },
  listStats: {
    flexDirection: "row",
    justifyContent: "space-between",
    marginBottom: 12,
  },
  listStat: {
    fontSize: 12,
  },
  listActions: {
    flexDirection: "row",
    justifyContent: "flex-end",
    gap: 8,
  },
  actionButton: {
    borderRadius: 8,
    minWidth: 70,
  },
  actionButtonText: {
    fontSize: 14,
    fontWeight: "500",
  },
  cardActions: {
    padding: 16,
    paddingTop: 0,
  },
  viewButton: {
    borderRadius: 8,
    flex: 1,
  },
  viewButtonText: {
    fontWeight: "600",
  },
  emptyContainer: {
    flex: 1,
    justifyContent: "center",
    alignItems: "center",
    paddingTop: 50,
  },
  retryButton: {
    marginTop: 10,
  },
  createButton: {
    marginTop: 20,
    paddingHorizontal: 24,
  },
  createButtonText: {
    fontWeight: "600",
  },
});
