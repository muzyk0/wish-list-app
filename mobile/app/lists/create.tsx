import { useMutation } from "@tanstack/react-query";
import { useRouter } from "expo-router";
import { useState } from "react";
import {
  ActivityIndicator,
  Alert,
  StyleSheet,
  Switch,
  Text,
  TextInput,
  TouchableOpacity,
  View,
} from "react-native";
import { apiClient } from "@/lib/api";

export default function CreateWishListScreen() {
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [occasion, setOccasion] = useState("");
  const [isPublic, setIsPublic] = useState(false);
  const router = useRouter();

  const mutation = useMutation({
    mutationFn: (data: {
      title: string;
      description?: string;
      occasion?: string;
      is_public: boolean;
    }) =>
      apiClient.createWishList({
        title: data.title,
        description: data.description || "",
        occasion: data.occasion || "",
        is_public: data.is_public,
        template_id: "default",
      }),
    onSuccess: (_data) => {
      Alert.alert("Success", "Wishlist created successfully!", [
        { text: "OK", onPress: () => router.push("/") },
      ]);
    },
    onError: (error: Error) => {
      Alert.alert(
        "Error",
        error.message || "Failed to create wishlist. Please try again.",
      );
    },
  });

  const handleCreate = () => {
    if (!title.trim()) {
      Alert.alert("Error", "Please enter a title for your wishlist.");
      return;
    }

    mutation.mutate({
      title: title.trim(),
      description: description.trim(),
      occasion: occasion.trim(),
      is_public: isPublic,
    });
  };

  return (
    <View style={styles.container}>
      <Text style={styles.title}>Create New Wishlist</Text>

      <TextInput
        style={styles.input}
        placeholder="Title *"
        value={title}
        onChangeText={setTitle}
        maxLength={200}
      />

      <TextInput
        style={styles.input}
        placeholder="Description"
        value={description}
        onChangeText={setDescription}
        multiline
        numberOfLines={3}
      />

      <TextInput
        style={styles.input}
        placeholder="Occasion (e.g., Birthday, Wedding)"
        value={occasion}
        onChangeText={setOccasion}
        maxLength={100}
      />

      <View style={styles.toggleContainer}>
        <Text style={styles.label}>Make Public</Text>
        <Switch value={isPublic} onValueChange={setIsPublic} />
      </View>

      <TouchableOpacity
        style={styles.button}
        onPress={handleCreate}
        disabled={mutation.isPending}
      >
        {mutation.isPending ? (
          <ActivityIndicator color="#fff" />
        ) : (
          <Text style={styles.buttonText}>Create Wishlist</Text>
        )}
      </TouchableOpacity>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    padding: 20,
    backgroundColor: "#fff",
  },
  title: {
    fontSize: 24,
    fontWeight: "bold",
    textAlign: "center",
    marginBottom: 30,
  },
  input: {
    borderWidth: 1,
    borderColor: "#ddd",
    padding: 12,
    borderRadius: 8,
    marginBottom: 15,
    fontSize: 16,
    minHeight: 40,
  },
  toggleContainer: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    marginBottom: 20,
    paddingVertical: 10,
  },
  label: {
    fontSize: 16,
  },
  button: {
    backgroundColor: "#007AFF",
    padding: 15,
    borderRadius: 8,
    alignItems: "center",
    marginTop: 10,
  },
  buttonText: {
    color: "#fff",
    fontSize: 16,
    fontWeight: "bold",
  },
});
