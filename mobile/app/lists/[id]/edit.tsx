import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery } from "@tanstack/react-query";
import { useLocalSearchParams, useRouter } from "expo-router";
import { useEffect } from "react";
import { Alert, ScrollView, StyleSheet, View } from "react-native";
import {
  ActivityIndicator,
  Button,
  HelperText,
  Switch,
  Text,
  TextInput,
  useTheme,
} from "react-native-paper";
import { Controller, useForm } from "react-hook-form";
import { z } from "zod";

import { apiClient } from "@/lib/api";

const updateWishListSchema = z.object({
  title: z.string().min(1, "Please enter a title for your wishlist.").max(200),
  description: z.string().optional(),
  occasion: z.string().max(100).optional(),
  is_public: z.boolean(),
});

type UpdateWishListFormData = z.infer<typeof updateWishListSchema>;

export default function EditWishListScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const router = useRouter();
  const { colors } = useTheme();

  const {
    control,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<UpdateWishListFormData>({
    resolver: zodResolver(updateWishListSchema),
    defaultValues: {
      title: "",
      description: "",
      occasion: "",
      is_public: false,
    },
  });

  const {
    data: wishList,
    isLoading,
    error,
  } = useQuery({
    queryKey: ["wishlist", id],
    queryFn: () => apiClient.getWishListById(id),
    enabled: !!id,
  });

  useEffect(() => {
    if (wishList) {
      reset({
        title: wishList.title,
        description: wishList.description || "",
        occasion: wishList.occasion || "",
        is_public: wishList.is_public,
      });
    }
  }, [wishList, reset]);

  const updateMutation = useMutation({
    mutationFn: (data: UpdateWishListFormData) =>
      apiClient.updateWishList(id, {
        title: data.title.trim(),
        description: data.description?.trim() || undefined,
        occasion: data.occasion?.trim() || undefined,
        is_public: data.is_public,
      }),
    onSuccess: () => {
      Alert.alert("Success", "Wishlist updated successfully!", [
        {
          text: "OK",
          onPress: () =>
            router.push({
              pathname: `/lists/[id]`,
              params: { id },
            }),
        },
      ]);
    },
    onError: (error) => {
      Alert.alert(
        "Error",
        error.message || "Failed to update wishlist. Please try again.",
      );
    },
  });

  const deleteMutation = useMutation({
    mutationFn: () => apiClient.deleteWishList(id),
    onSuccess: () => {
      Alert.alert("Success", "Wishlist deleted successfully!", [
        { text: "OK", onPress: () => router.push("/lists") },
      ]);
    },
    onError: (error) => {
      Alert.alert(
        "Error",
        error.message || "Failed to delete wishlist. Please try again.",
      );
    },
  });

  const onSubmit = (data: UpdateWishListFormData) => {
    updateMutation.mutate(data);
  };

  const handleDelete = () => {
    Alert.alert(
      "Confirm Delete",
      "Are you sure you want to delete this wishlist? This action cannot be undone and will also delete all associated gift items.",
      [
        { text: "Cancel", style: "cancel" },
        {
          text: "Delete",
          style: "destructive",
          onPress: () => deleteMutation.mutate(),
        },
      ],
    );
  };

  if (isLoading) {
    return (
      <View
        style={[styles.centerContainer, { backgroundColor: colors.background }]}
      >
        <ActivityIndicator
          size="large"
          animating={true}
          color={colors.primary}
        />
        <Text
          variant="bodyLarge"
          style={{ marginTop: 10, color: colors.onSurface }}
        >
          Loading wishlist...
        </Text>
      </View>
    );
  }

  if (error) {
    return (
      <View
        style={[styles.centerContainer, { backgroundColor: colors.background }]}
      >
        <Text
          variant="headlineMedium"
          style={{ color: colors.error, marginBottom: 10 }}
        >
          Error loading wishlist
        </Text>
        <HelperText type="error">{error.message}</HelperText>
        <Button
          mode="contained"
          onPress={() => router.back()}
          style={{ marginTop: 16 }}
          buttonColor={colors.primary}
          textColor={colors.onPrimary}
        >
          Back
        </Button>
      </View>
    );
  }

  return (
    <ScrollView style={{ flex: 1, backgroundColor: colors.background }}>
      <View style={styles.container}>
        <Text
          variant="headlineLarge"
          style={[styles.title, { color: colors.onSurface }]}
        >
          Edit Wish List
        </Text>

        <View style={styles.form}>
          <View style={styles.inputGroup}>
            <Text
              variant="labelLarge"
              style={[styles.label, { color: colors.onSurface }]}
            >
              Title *
            </Text>
            <Controller
              control={control}
              name="title"
              render={({ field: { onChange, onBlur, value } }) => (
                <TextInput
                  mode="outlined"
                  value={value}
                  onChangeText={onChange}
                  onBlur={onBlur}
                  placeholder="Wishlist title"
                  maxLength={200}
                  disabled={updateMutation.isPending}
                  error={!!errors.title}
                  style={styles.input}
                />
              )}
            />
            {errors.title && (
              <HelperText type="error" visible={!!errors.title}>
                {errors.title.message}
              </HelperText>
            )}
          </View>

          <View style={styles.inputGroup}>
            <Text
              variant="labelLarge"
              style={[styles.label, { color: colors.onSurface }]}
            >
              Description
            </Text>
            <Controller
              control={control}
              name="description"
              render={({ field: { onChange, onBlur, value } }) => (
                <TextInput
                  mode="outlined"
                  value={value}
                  onChangeText={onChange}
                  onBlur={onBlur}
                  placeholder="Wishlist description"
                  multiline
                  numberOfLines={3}
                  disabled={updateMutation.isPending}
                  style={[styles.input, styles.multiline]}
                />
              )}
            />
          </View>

          <View style={styles.inputGroup}>
            <Text
              variant="labelLarge"
              style={[styles.label, { color: colors.onSurface }]}
            >
              Occasion
            </Text>
            <Controller
              control={control}
              name="occasion"
              render={({ field: { onChange, onBlur, value } }) => (
                <TextInput
                  mode="outlined"
                  value={value}
                  onChangeText={onChange}
                  onBlur={onBlur}
                  placeholder="Occasion (e.g., Birthday, Wedding)"
                  maxLength={100}
                  disabled={updateMutation.isPending}
                  style={styles.input}
                />
              )}
            />
          </View>

          <View style={styles.toggleContainer}>
            <Text variant="labelLarge" style={{ color: colors.onSurface }}>
              Make Public
            </Text>
            <Controller
              control={control}
              name="is_public"
              render={({ field: { onChange, value } }) => (
                <Switch
                  value={value}
                  onValueChange={onChange}
                  disabled={updateMutation.isPending}
                  color={colors.primary}
                />
              )}
            />
          </View>

          <Button
            mode="contained"
            onPress={handleSubmit(onSubmit)}
            loading={updateMutation.isPending}
            disabled={updateMutation.isPending}
            style={styles.button}
            buttonColor={colors.primary}
            textColor={colors.onPrimary}
          >
            Update Wish List
          </Button>

          <Button
            mode="outlined"
            onPress={() => router.back()}
            disabled={updateMutation.isPending}
            style={styles.cancelButton}
            textColor={colors.onSurfaceVariant}
          >
            Cancel
          </Button>

          <Button
            mode="contained-tonal"
            onPress={handleDelete}
            loading={deleteMutation.isPending}
            disabled={updateMutation.isPending || deleteMutation.isPending}
            style={styles.deleteButton}
            buttonColor={colors.errorContainer}
            textColor={colors.onErrorContainer}
          >
            Delete Wish List
          </Button>
        </View>
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    padding: 20,
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
  form: {
    flex: 1,
  },
  inputGroup: {
    marginBottom: 15,
  },
  label: {
    fontSize: 16,
    fontWeight: "600",
    marginBottom: 5,
  },
  input: {
    marginTop: 4,
  },
  multiline: {
    height: 100,
  },
  toggleContainer: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    paddingVertical: 10,
    marginBottom: 20,
  },
  button: {
    marginTop: 10,
    paddingVertical: 6,
    borderRadius: 8,
  },
  cancelButton: {
    marginTop: 10,
    paddingVertical: 6,
    borderRadius: 8,
  },
  deleteButton: {
    marginTop: 10,
    paddingVertical: 6,
    borderRadius: 8,
  },
});
