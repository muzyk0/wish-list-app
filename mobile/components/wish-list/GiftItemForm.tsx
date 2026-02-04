import { useMutation } from '@tanstack/react-query';
import { useRouter } from 'expo-router';
import { useState } from 'react';
import { Alert, ScrollView, StyleSheet, Text } from 'react-native';
import {
  Button as PaperButton,
  TextInput as PaperTextInput,
} from 'react-native-paper';
import { apiClient } from '@/lib/api';
import type { GiftItem } from '@/lib/api/types';
import ImageUpload from './ImageUpload';

interface GiftItemFormProps {
  wishlistId: string;
  existingItem?: GiftItem; // Optional existing item for editing
  onComplete?: () => void; // Callback when form is completed
}

export default function GiftItemForm({
  wishlistId,
  existingItem,
  onComplete,
}: GiftItemFormProps) {
  const [name, setName] = useState(existingItem?.name || '');
  const [description, setDescription] = useState(
    existingItem?.description || '',
  );
  const [link, setLink] = useState(existingItem?.link || '');
  const [imageUrl, setImageUrl] = useState(existingItem?.image_url || '');
  const [price, setPrice] = useState(existingItem?.price?.toString() || '');
  const [priority, setPriority] = useState(
    existingItem?.priority?.toString() || '0',
  );
  const [notes, setNotes] = useState(existingItem?.notes || '');
  const [position, setPosition] = useState(
    existingItem?.position?.toString() || '0',
  );

  const router = useRouter();

  const mutation = useMutation({
    mutationFn: (data: {
      name: string;
      description: string;
      link: string;
      image_url: string;
      price: string;
      priority: string;
      notes: string;
      position: string;
    }) => {
      // Parse numeric fields, set to undefined if empty or invalid
      const parsedPrice =
        data.price.trim() !== '' ? parseFloat(data.price) : NaN;
      const parsedPriority =
        data.priority.trim() !== '' ? parseInt(data.priority, 10) : NaN;
      const parsedPosition =
        data.position.trim() !== '' ? parseInt(data.position, 10) : NaN;

      if (existingItem) {
        // Update existing item
        return apiClient.updateGiftItem(wishlistId, existingItem.id, {
          name: data.name,
          description: data.description,
          link: data.link,
          image_url: data.image_url,
          price: !isNaN(parsedPrice) ? parsedPrice : undefined,
          priority: !isNaN(parsedPriority) ? parsedPriority : undefined,
          notes: data.notes,
          position: !isNaN(parsedPosition) ? parsedPosition : undefined,
        });
      } else {
        // Create new item
        return apiClient.createGiftItem(wishlistId, {
          name: data.name,
          description: data.description,
          link: data.link,
          image_url: data.image_url,
          price: !isNaN(parsedPrice) ? parsedPrice : undefined,
          priority: !isNaN(parsedPriority) ? parsedPriority : undefined,
          notes: data.notes,
          position: !isNaN(parsedPosition) ? parsedPosition : undefined,
        });
      }
    },
    onSuccess: (_data) => {
      Alert.alert(
        'Success',
        `Gift item ${existingItem ? 'updated' : 'created'} successfully!`,
        [
          {
            text: 'OK',
            onPress: () => {
              if (onComplete) {
                onComplete();
              } else {
                router.back();
              }
            },
          },
        ],
      );
    },
    onError: (error: Error) => {
      Alert.alert(
        'Error',
        error.message ||
          `Failed to ${existingItem ? 'update' : 'create'} gift item. Please try again.`,
      );
    },
  });

  const deleteMutation = useMutation({
    mutationFn: () => {
      if (!existingItem?.id) {
        throw new Error('No item to delete');
      }
      return apiClient.deleteGiftItem(wishlistId, existingItem.id);
    },
    onSuccess: () => {
      Alert.alert('Success', 'Gift item deleted successfully!', [
        {
          text: 'OK',
          onPress: () => {
            if (onComplete) {
              onComplete();
            } else {
              router.back();
            }
          },
        },
      ]);
    },
    onError: (error: Error) => {
      Alert.alert(
        'Error',
        error.message || 'Failed to delete gift item. Please try again.',
      );
    },
  });

  const handleSubmit = () => {
    if (!name.trim()) {
      Alert.alert('Error', 'Please enter a name for the gift item.');
      return;
    }

    // Validate price if provided
    if (price && Number.isNaN(parseFloat(price))) {
      Alert.alert('Error', 'Please enter a valid price.');
      return;
    }

    // Validate priority (0-10)
    const priorityNum = parseInt(priority, 10);
    if (
      priority &&
      (Number.isNaN(priorityNum) || priorityNum < 0 || priorityNum > 10)
    ) {
      Alert.alert('Error', 'Please enter a priority between 0 and 10.');
      return;
    }

    mutation.mutate({
      name: name.trim(),
      description: description.trim(),
      link: link.trim(),
      image_url: imageUrl.trim(),
      price,
      priority,
      notes: notes.trim(),
      position: position.trim(),
    });
  };

  const handleDelete = () => {
    if (!existingItem) return;

    Alert.alert(
      'Confirm Delete',
      'Are you sure you want to delete this gift item? This action cannot be undone.',
      [
        { text: 'Cancel', style: 'cancel' },
        {
          text: 'Delete',
          style: 'destructive',
          onPress: () => deleteMutation.mutate(),
        },
      ],
    );
  };

  return (
    <ScrollView style={styles.container}>
      <Text style={styles.title}>
        {existingItem ? 'Edit Gift Item' : 'Add New Gift Item'}
      </Text>

      <PaperTextInput
        label="Name *"
        value={name}
        onChangeText={setName}
        mode="outlined"
        maxLength={255}
        style={styles.input}
        disabled={mutation.isPending}
      />

      <PaperTextInput
        label="Description"
        value={description}
        onChangeText={setDescription}
        mode="outlined"
        multiline
        numberOfLines={3}
        style={styles.multilineInput}
        disabled={mutation.isPending}
      />

      <PaperTextInput
        label="Link (URL)"
        value={link}
        onChangeText={setLink}
        mode="outlined"
        keyboardType="url"
        style={styles.input}
        disabled={mutation.isPending}
      />

      <ImageUpload
        onImageUpload={setImageUrl}
        currentImageUrl={imageUrl}
        disabled={mutation.isPending}
      />

      <PaperTextInput
        label="Price"
        value={price}
        onChangeText={setPrice}
        mode="outlined"
        keyboardType="decimal-pad"
        style={styles.input}
        disabled={mutation.isPending}
      />

      <PaperTextInput
        label="Priority (0-10)"
        value={priority}
        onChangeText={setPriority}
        mode="outlined"
        keyboardType="numeric"
        style={styles.input}
        disabled={mutation.isPending}
      />

      <PaperTextInput
        label="Notes"
        value={notes}
        onChangeText={setNotes}
        mode="outlined"
        multiline
        numberOfLines={3}
        style={styles.multilineInput}
        disabled={mutation.isPending}
      />

      <PaperTextInput
        label="Position"
        value={position}
        onChangeText={setPosition}
        mode="outlined"
        keyboardType="numeric"
        style={styles.input}
        disabled={mutation.isPending}
      />

      <PaperButton
        mode="contained"
        onPress={handleSubmit}
        loading={mutation.isPending}
        disabled={mutation.isPending}
        style={styles.button}
      >
        {mutation.isPending ? (
          <Text style={styles.buttonText}>Processing...</Text>
        ) : (
          <Text style={styles.buttonText}>
            {existingItem ? 'Update Item' : 'Add Item'}
          </Text>
        )}
      </PaperButton>

      {existingItem?.id && (
        <PaperButton
          mode="contained-tonal"
          onPress={handleDelete}
          loading={deleteMutation.isPending}
          disabled={mutation.isPending || deleteMutation.isPending}
          style={styles.deleteButton}
        >
          {deleteMutation.isPending ? (
            <Text style={styles.buttonText}>Processing...</Text>
          ) : (
            <Text style={styles.buttonText}>Delete Item</Text>
          )}
        </PaperButton>
      )}
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    padding: 20,
    backgroundColor: '#fff',
  },
  title: {
    fontSize: 24,
    fontWeight: 'bold',
    textAlign: 'center',
    marginBottom: 20,
  },
  input: {
    marginBottom: 15,
  },
  multilineInput: {
    marginBottom: 15,
  },
  button: {
    marginTop: 10,
    paddingVertical: 5,
  },
  deleteButton: {
    marginTop: 10,
    paddingVertical: 5,
  },
  buttonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: 'bold',
  },
});
