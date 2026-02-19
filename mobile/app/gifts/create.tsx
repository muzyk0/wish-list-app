import { MaterialCommunityIcons } from '@expo/vector-icons';
import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { BlurView } from 'expo-blur';
import { LinearGradient } from 'expo-linear-gradient';
import { useRouter } from 'expo-router';
import { Controller, useForm } from 'react-hook-form';
import { Alert, Pressable, ScrollView, StyleSheet, View } from 'react-native';
import { HelperText, Text, TextInput } from 'react-native-paper';
import { z } from 'zod';
import { apiClient } from '@/lib/api';
import { dialog } from '@/stores/dialogStore';

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
});

type GiftItemFormData = z.infer<typeof giftItemSchema>;

export default function CreateStandaloneGiftScreen() {
  const router = useRouter();
  const queryClient = useQueryClient();

  const {
    control,
    handleSubmit,
    formState: { errors },
  } = useForm<GiftItemFormData>({
    resolver: zodResolver(giftItemSchema),
    defaultValues: {
      name: '',
      description: '',
      link: '',
      imageUrl: '',
      price: '',
      priority: '0',
      notes: '',
    },
  });

  const mutation = useMutation({
    mutationFn: (data: GiftItemFormData) => {
      const parsedPrice = data.price ? parseFloat(data.price) : undefined;
      const parsedPriority = data.priority
        ? parseInt(data.priority, 10)
        : undefined;

      return apiClient.createStandaloneGiftItem({
        title: data.name,
        description: data.description || undefined,
        link: data.link || undefined,
        image_url: data.imageUrl || undefined,
        price: parsedPrice,
        priority: parsedPriority,
        notes: data.notes || undefined,
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['userGiftItems'] });
      queryClient.invalidateQueries({ queryKey: ['standaloneGiftItems'] });
      dialog.message({
        title: 'Success',
        message: 'Gift item created successfully!',
        onPress: () => router.push('/gifts'),
      });
    },
    onError: (error: Error) => {
      dialog.error(
        error.message || 'Failed to create gift item. Please try again.',
      );
    },
  });

  const onSubmit = (data: GiftItemFormData) => {
    mutation.mutate(data);
  };

  return (
    <View style={styles.container}>
      <LinearGradient
        colors={['#1a0a2e', '#2d1b4e', '#3d2a6e']}
        style={StyleSheet.absoluteFill}
      />

      {/* Decorative elements */}
      <View style={styles.decorCircle1} />
      <View style={styles.decorCircle2} />

      {/* Header */}
      <View style={styles.header}>
        <Pressable onPress={() => router.back()} style={styles.backButton}>
          <MaterialCommunityIcons name="arrow-left" size={24} color="#ffffff" />
        </Pressable>
        <Text style={styles.headerTitle}>Create Gift</Text>
        <View style={{ width: 40 }} />
      </View>

      <ScrollView
        style={styles.scrollView}
        contentContainerStyle={styles.scrollContent}
        showsVerticalScrollIndicator={false}
      >
        {/* Form Card */}
        <BlurView intensity={20} style={styles.formCard}>
          <View style={styles.formContent}>
            <View style={styles.iconContainer}>
              <LinearGradient
                colors={['#FFD700', '#FFA500']}
                style={styles.iconGradient}
              >
                <MaterialCommunityIcons name="gift" size={32} color="#000000" />
              </LinearGradient>
            </View>

            {/* Name Field */}
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

            {/* Description Field */}
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

            {/* Link Field */}
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

            {/* Price Field */}
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

            {/* Priority Field */}
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

            {/* Notes Field */}
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

            {/* Create Button */}
            <Pressable
              onPress={handleSubmit(onSubmit)}
              disabled={mutation.isPending}
            >
              <LinearGradient
                colors={['#FFD700', '#FFA500']}
                style={styles.createButton}
              >
                {mutation.isPending ? (
                  <Text style={styles.createButtonText}>Creating...</Text>
                ) : (
                  <>
                    <MaterialCommunityIcons
                      name="plus"
                      size={20}
                      color="#000000"
                    />
                    <Text style={styles.createButtonText}>Create Gift</Text>
                  </>
                )}
              </LinearGradient>
            </Pressable>
          </View>
        </BlurView>

        {/* Help Text */}
        <View style={styles.helpContainer}>
          <MaterialCommunityIcons
            name="information"
            size={16}
            color="rgba(255, 255, 255, 0.5)"
          />
          <Text style={styles.helpText}>
            This gift won't be attached to any wishlist. You can attach it
            later.
          </Text>
        </View>
      </ScrollView>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  decorCircle1: {
    position: 'absolute',
    width: 250,
    height: 250,
    borderRadius: 125,
    backgroundColor: 'rgba(255, 215, 0, 0.06)',
    top: -80,
    right: -60,
  },
  decorCircle2: {
    position: 'absolute',
    width: 180,
    height: 180,
    borderRadius: 90,
    backgroundColor: 'rgba(107, 78, 230, 0.12)',
    bottom: 200,
    left: -40,
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: 20,
    paddingTop: 60,
    paddingBottom: 20,
  },
  backButton: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: 'rgba(255, 255, 255, 0.1)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  headerTitle: {
    fontSize: 18,
    fontWeight: '700',
    color: '#ffffff',
  },
  scrollView: {
    flex: 1,
  },
  scrollContent: {
    padding: 24,
    paddingBottom: 100,
  },
  formCard: {
    borderRadius: 24,
    overflow: 'hidden',
    backgroundColor: 'rgba(255, 255, 255, 0.08)',
    borderWidth: 1,
    borderColor: 'rgba(255, 255, 255, 0.1)',
  },
  formContent: {
    padding: 24,
  },
  iconContainer: {
    alignItems: 'center',
    marginBottom: 24,
  },
  iconGradient: {
    width: 80,
    height: 80,
    borderRadius: 20,
    justifyContent: 'center',
    alignItems: 'center',
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
  createButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 16,
    borderRadius: 12,
    gap: 8,
    marginTop: 24,
  },
  createButtonText: {
    fontSize: 16,
    fontWeight: '700',
    color: '#000000',
  },
  helpContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    marginTop: 16,
    paddingHorizontal: 4,
  },
  helpText: {
    fontSize: 13,
    color: 'rgba(255, 255, 255, 0.5)',
    flex: 1,
  },
});
