import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation } from '@tanstack/react-query';
import { useRouter } from 'expo-router';
import { Controller, useForm } from 'react-hook-form';
import { Alert, ScrollView, StyleSheet, Text, View } from 'react-native';
import {
  Button as PaperButton,
  HelperText,
  TextInput as PaperTextInput,
} from 'react-native-paper';
import { z } from 'zod';
import { apiClient } from '@/lib/api';
import type { GiftItem } from '@/lib/api/types';
import ImageUpload from './ImageUpload';

// Zod schema for form validation
const giftItemSchema = z.object({
  name: z
    .string()
    .min(1, 'Name is required')
    .max(255, 'Name must be less than 255 characters'),
  description: z.string().optional(),
  link: z.string().url('Invalid URL').or(z.literal('')).optional(),
  imageUrl: z.string().optional(),
  price: z
    .string()
    .optional()
    .refine(
      (val) => !val || val === '' || !Number.isNaN(parseFloat(val)),
      'Invalid price',
    ),
  priority: z
    .string()
    .optional()
    .refine((val) => {
      if (!val || val === '') return true;
      const num = parseInt(val, 10);
      return !Number.isNaN(num) && num >= 0 && num <= 10;
    }, 'Priority must be between 0 and 10'),
  notes: z.string().optional(),
  position: z.string().optional(),
});

type GiftItemFormData = z.infer<typeof giftItemSchema>;

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
  const router = useRouter();

  const {
    control,
    handleSubmit,
    setValue,
    watch,
    formState: { errors },
  } = useForm<GiftItemFormData>({
    resolver: zodResolver(giftItemSchema),
    defaultValues: {
      name: existingItem?.name || '',
      description: existingItem?.description || '',
      link: existingItem?.link || '',
      imageUrl: existingItem?.image_url || '',
      price: existingItem?.price?.toString() || '',
      priority: existingItem?.priority?.toString() || '0',
      notes: existingItem?.notes || '',
      position: existingItem?.position?.toString() || '0',
    },
  });

  const imageUrl = watch('imageUrl');

  const mutation = useMutation({
    mutationFn: (data: GiftItemFormData) => {
      // Parse numeric fields, set to undefined if empty or invalid
      const parsedPrice = data.price ? parseFloat(data.price) : undefined;
      const parsedPriority = data.priority
        ? parseInt(data.priority, 10)
        : undefined;

      if (existingItem) {
        // Update existing item
        return apiClient.updateGiftItem(wishlistId, existingItem.id, {
          title: data.name,
          description: data.description || '',
          link: data.link || '',
          image_url: data.imageUrl || '',
          price: parsedPrice,
          priority: parsedPriority,
          notes: data.notes || '',
        });
      }
      // Create new item
      return apiClient.createGiftItem(wishlistId, {
        title: data.name,
        description: data.description || '',
        link: data.link || '',
        image_url: data.imageUrl || '',
        price: parsedPrice,
        priority: parsedPriority,
        notes: data.notes || '',
      });
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

  const onSubmit = (data: GiftItemFormData) => {
    mutation.mutate(data);
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

      <Controller
        control={control}
        name="name"
        render={({ field: { onChange, onBlur, value } }) => (
          <View>
            <PaperTextInput
              label="Name *"
              value={value}
              onChangeText={onChange}
              onBlur={onBlur}
              mode="outlined"
              maxLength={255}
              style={styles.input}
              disabled={mutation.isPending}
              error={!!errors.name}
            />
            {errors.name && (
              <HelperText type="error" visible={!!errors.name}>
                {errors.name.message}
              </HelperText>
            )}
          </View>
        )}
      />

      <Controller
        control={control}
        name="description"
        render={({ field: { onChange, onBlur, value } }) => (
          <PaperTextInput
            label="Description"
            value={value}
            onChangeText={onChange}
            onBlur={onBlur}
            mode="outlined"
            multiline
            numberOfLines={3}
            style={styles.multilineInput}
            disabled={mutation.isPending}
          />
        )}
      />

      <Controller
        control={control}
        name="link"
        render={({ field: { onChange, onBlur, value } }) => (
          <View>
            <PaperTextInput
              label="Link (URL)"
              value={value}
              onChangeText={onChange}
              onBlur={onBlur}
              mode="outlined"
              keyboardType="url"
              style={styles.input}
              disabled={mutation.isPending}
              error={!!errors.link}
            />
            {errors.link && (
              <HelperText type="error" visible={!!errors.link}>
                {errors.link.message}
              </HelperText>
            )}
          </View>
        )}
      />

      <ImageUpload
        onImageUpload={(url) => setValue('imageUrl', url)}
        currentImageUrl={imageUrl}
        disabled={mutation.isPending}
      />

      <Controller
        control={control}
        name="price"
        render={({ field: { onChange, onBlur, value } }) => (
          <View>
            <PaperTextInput
              label="Price"
              value={value}
              onChangeText={onChange}
              onBlur={onBlur}
              mode="outlined"
              keyboardType="decimal-pad"
              style={styles.input}
              disabled={mutation.isPending}
              error={!!errors.price}
            />
            {errors.price && (
              <HelperText type="error" visible={!!errors.price}>
                {errors.price.message}
              </HelperText>
            )}
          </View>
        )}
      />

      <Controller
        control={control}
        name="priority"
        render={({ field: { onChange, onBlur, value } }) => (
          <View>
            <PaperTextInput
              label="Priority (0-10)"
              value={value}
              onChangeText={onChange}
              onBlur={onBlur}
              mode="outlined"
              keyboardType="numeric"
              style={styles.input}
              disabled={mutation.isPending}
              error={!!errors.priority}
            />
            {errors.priority && (
              <HelperText type="error" visible={!!errors.priority}>
                {errors.priority.message}
              </HelperText>
            )}
          </View>
        )}
      />

      <Controller
        control={control}
        name="notes"
        render={({ field: { onChange, onBlur, value } }) => (
          <PaperTextInput
            label="Notes"
            value={value}
            onChangeText={onChange}
            onBlur={onBlur}
            mode="outlined"
            multiline
            numberOfLines={3}
            style={styles.multilineInput}
            disabled={mutation.isPending}
          />
        )}
      />

      <Controller
        control={control}
        name="position"
        render={({ field: { onChange, onBlur, value } }) => (
          <PaperTextInput
            label="Position"
            value={value}
            onChangeText={onChange}
            onBlur={onBlur}
            mode="outlined"
            keyboardType="numeric"
            style={styles.input}
            disabled={mutation.isPending}
          />
        )}
      />

      <PaperButton
        mode="contained"
        onPress={handleSubmit(onSubmit)}
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
