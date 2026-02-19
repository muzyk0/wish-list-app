import { MaterialCommunityIcons } from '@expo/vector-icons';
import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation } from '@tanstack/react-query';
import { LinearGradient } from 'expo-linear-gradient';
import { useRouter } from 'expo-router';
import { Controller, useForm } from 'react-hook-form';
import { Pressable, StyleSheet, View } from 'react-native';
import { HelperText, Text, TextInput } from 'react-native-paper';
import { z } from 'zod';
import { apiClient } from '@/lib/api';
import type { GiftItem } from '@/lib/api/types';
import { dialog } from '@/stores/dialogStore';
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
      name: existingItem?.title || '',
      description: existingItem?.description || '',
      link: existingItem?.link || '',
      imageUrl: existingItem?.image_url || '',
      price: existingItem?.price?.toString() || '',
      priority: existingItem?.priority?.toString() || '0',
      notes: existingItem?.notes || '',
      position: '0',
    },
  });

  const imageUrl = watch('imageUrl');

  const mutation = useMutation({
    mutationFn: (data: GiftItemFormData) => {
      const parsedPrice = data.price ? parseFloat(data.price) : undefined;
      const parsedPriority = data.priority
        ? parseInt(data.priority, 10)
        : undefined;

      if (existingItem?.id) {
        return apiClient.updateGiftItem(wishlistId, existingItem.id, {
          title: data.name,
          description: data.description || undefined,
          link: data.link || undefined,
          image_url: data.imageUrl || undefined,
          price: parsedPrice,
          priority: parsedPriority,
          notes: data.notes || undefined,
        });
      }
      return apiClient.createGiftItem(wishlistId, {
        title: data.name,
        description: data.description || undefined,
        link: data.link || undefined,
        image_url: data.imageUrl || undefined,
        price: parsedPrice,
        priority: parsedPriority,
        notes: data.notes || undefined,
      });
    },
    onSuccess: (_data) => {
      dialog.message({
        title: 'Success',
        message: `Gift item ${existingItem ? 'updated' : 'created'} successfully!`,
        onPress: () => {
          if (onComplete) {
            onComplete();
          } else {
            router.back();
          }
        },
      });
    },
    onError: (error: Error) => {
      dialog.error(
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
      dialog.message({
        title: 'Success',
        message: 'Gift item deleted successfully!',
        onPress: () => {
          if (onComplete) {
            onComplete();
          } else {
            router.back();
          }
        },
      });
    },
    onError: (error: Error) => {
      dialog.error(
        error.message || 'Failed to delete gift item. Please try again.',
      );
    },
  });

  const onSubmit = (data: GiftItemFormData) => {
    mutation.mutate(data);
  };

  const handleDelete = () => {
    if (!existingItem) return;

    dialog.confirmDelete('this gift item', () => deleteMutation.mutate());
  };

  return (
    <View style={styles.container}>
      <Controller
        control={control}
        name="name"
        render={({ field: { onChange, onBlur, value } }) => (
          <View>
            <TextInput
              label="Name *"
              value={value}
              onChangeText={onChange}
              onBlur={onBlur}
              maxLength={255}
              style={styles.input}
              textColor="#ffffff"
              underlineColor="transparent"
              activeUnderlineColor="#FFD700"
              placeholderTextColor="rgba(255, 255, 255, 0.4)"
              disabled={mutation.isPending}
              error={!!errors.name}
              theme={{
                colors: {
                  primary: '#FFD700',
                  onSurfaceVariant: 'rgba(255, 255, 255, 0.5)',
                },
              }}
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
          <View style={styles.textAreaWrapper}>
            <TextInput
              label="Description"
              value={value}
              onChangeText={onChange}
              onBlur={onBlur}
              multiline
              numberOfLines={3}
              mode="flat"
              style={styles.textArea}
              textColor="#ffffff"
              underlineColor="transparent"
              activeUnderlineColor="#FFD700"
              placeholderTextColor="rgba(255, 255, 255, 0.4)"
              disabled={mutation.isPending}
              contentStyle={{
                backgroundColor: 'transparent',
              }}
              theme={{
                colors: {
                  primary: '#FFD700',
                  onSurfaceVariant: 'rgba(255, 255, 255, 0.5)',
                  background: 'transparent',
                  surface: 'transparent',
                  surfaceVariant: 'transparent',
                  elevation: {
                    level0: 'transparent',
                    level1: 'transparent',
                    level2: 'transparent',
                    level3: 'transparent',
                    level4: 'transparent',
                    level5: 'transparent',
                  },
                },
              }}
            />
          </View>
        )}
      />

      <Controller
        control={control}
        name="link"
        render={({ field: { onChange, onBlur, value } }) => (
          <View>
            <TextInput
              label="Link (URL)"
              value={value}
              onChangeText={onChange}
              onBlur={onBlur}
              keyboardType="url"
              style={styles.input}
              textColor="#ffffff"
              underlineColor="transparent"
              activeUnderlineColor="#FFD700"
              placeholderTextColor="rgba(255, 255, 255, 0.4)"
              disabled={mutation.isPending}
              error={!!errors.link}
              theme={{
                colors: {
                  primary: '#FFD700',
                  onSurfaceVariant: 'rgba(255, 255, 255, 0.5)',
                },
              }}
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
            <TextInput
              label="Price"
              value={value}
              onChangeText={onChange}
              onBlur={onBlur}
              keyboardType="decimal-pad"
              style={styles.input}
              textColor="#ffffff"
              underlineColor="transparent"
              activeUnderlineColor="#FFD700"
              placeholderTextColor="rgba(255, 255, 255, 0.4)"
              disabled={mutation.isPending}
              error={!!errors.price}
              theme={{
                colors: {
                  primary: '#FFD700',
                  onSurfaceVariant: 'rgba(255, 255, 255, 0.5)',
                },
              }}
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
            <TextInput
              label="Priority (0-10)"
              value={value}
              onChangeText={onChange}
              onBlur={onBlur}
              keyboardType="numeric"
              style={styles.input}
              textColor="#ffffff"
              underlineColor="transparent"
              activeUnderlineColor="#FFD700"
              placeholderTextColor="rgba(255, 255, 255, 0.4)"
              disabled={mutation.isPending}
              error={!!errors.priority}
              theme={{
                colors: {
                  primary: '#FFD700',
                  onSurfaceVariant: 'rgba(255, 255, 255, 0.5)',
                },
              }}
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
          <View style={styles.textAreaWrapper}>
            <TextInput
              label="Notes"
              value={value}
              onChangeText={onChange}
              onBlur={onBlur}
              multiline
              numberOfLines={3}
              mode="flat"
              style={styles.textArea}
              textColor="#ffffff"
              underlineColor="transparent"
              activeUnderlineColor="#FFD700"
              placeholderTextColor="rgba(255, 255, 255, 0.4)"
              disabled={mutation.isPending}
              contentStyle={{
                backgroundColor: 'transparent',
              }}
              theme={{
                colors: {
                  primary: '#FFD700',
                  onSurfaceVariant: 'rgba(255, 255, 255, 0.5)',
                  background: 'transparent',
                  surface: 'transparent',
                  surfaceVariant: 'transparent',
                  elevation: {
                    level0: 'transparent',
                    level1: 'transparent',
                    level2: 'transparent',
                    level3: 'transparent',
                    level4: 'transparent',
                    level5: 'transparent',
                  },
                },
              }}
            />
          </View>
        )}
      />

      {/* Create/Update Button */}
      <Pressable onPress={handleSubmit(onSubmit)} disabled={mutation.isPending}>
        <LinearGradient colors={['#FFD700', '#FFA500']} style={styles.button}>
          {mutation.isPending ? (
            <Text style={styles.buttonText}>Processing...</Text>
          ) : (
            <>
              <MaterialCommunityIcons
                name={existingItem ? 'pencil' : 'plus'}
                size={20}
                color="#000000"
              />
              <Text style={styles.buttonText}>
                {existingItem ? 'Update Item' : 'Add Item'}
              </Text>
            </>
          )}
        </LinearGradient>
      </Pressable>

      {existingItem?.id && (
        <Pressable
          onPress={handleDelete}
          disabled={mutation.isPending || deleteMutation.isPending}
          style={{ marginTop: 12 }}
        >
          <View style={styles.deleteButton}>
            {deleteMutation.isPending ? (
              <Text style={styles.deleteButtonText}>Processing...</Text>
            ) : (
              <>
                <MaterialCommunityIcons
                  name="delete"
                  size={20}
                  color="#FF6B6B"
                />
                <Text style={styles.deleteButtonText}>Delete Item</Text>
              </>
            )}
          </View>
        </Pressable>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  input: {
    backgroundColor: 'rgba(255, 255, 255, 0.08)',
    marginBottom: 16,
    borderRadius: 12,
  },
  textAreaWrapper: {
    backgroundColor: 'rgba(255, 255, 255, 0.08)',
    borderTopLeftRadius: 4,
    borderTopRightRadius: 4,
    borderBottomRightRadius: 12,
    borderBottomLeftRadius: 12,
    marginBottom: 16,
  },
  textArea: {
    backgroundColor: 'transparent',
    minHeight: 100,
  },
  button: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 16,
    borderRadius: 12,
    gap: 8,
    marginTop: 24,
  },
  buttonText: {
    fontSize: 16,
    fontWeight: '700',
    color: '#000000',
  },
  deleteButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 16,
    borderRadius: 12,
    backgroundColor: 'rgba(255, 107, 107, 0.15)',
    borderWidth: 1,
    borderColor: 'rgba(255, 107, 107, 0.3)',
    gap: 8,
  },
  deleteButtonText: {
    fontSize: 16,
    fontWeight: '700',
    color: '#FF6B6B',
  },
});
