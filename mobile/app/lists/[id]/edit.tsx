import { MaterialCommunityIcons } from '@expo/vector-icons';
import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation, useQuery } from '@tanstack/react-query';
import { BlurView } from 'expo-blur';
import { LinearGradient } from 'expo-linear-gradient';
import { useLocalSearchParams, useRouter } from 'expo-router';
import { useEffect } from 'react';
import { Controller, useForm } from 'react-hook-form';
import {
  Alert,
  Pressable,
  ScrollView,
  StyleSheet,
  Switch,
  View,
} from 'react-native';
import {
  ActivityIndicator,
  HelperText,
  Text,
  TextInput,
} from 'react-native-paper';
import { z } from 'zod';

import { apiClient } from '@/lib/api';

const updateWishListSchema = z.object({
  title: z.string().min(1, 'Please enter a title for your wishlist.').max(200),
  description: z.string().optional(),
  occasion: z.string().max(100).optional(),
  is_public: z.boolean(),
});

type UpdateWishListFormData = z.infer<typeof updateWishListSchema>;

export default function EditWishListScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const router = useRouter();

  const {
    control,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<UpdateWishListFormData>({
    resolver: zodResolver(updateWishListSchema),
    defaultValues: {
      title: '',
      description: '',
      occasion: '',
      is_public: false,
    },
  });

  const {
    data: wishList,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['wishlist', id],
    queryFn: () => apiClient.getWishListById(id),
    enabled: !!id,
  });

  useEffect(() => {
    if (wishList) {
      reset({
        title: wishList.title,
        description: wishList.description || '',
        occasion: wishList.occasion || '',
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
      Alert.alert('Success', 'Wishlist updated successfully!', [
        {
          text: 'OK',
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
        'Error',
        error.message || 'Failed to update wishlist. Please try again.',
      );
    },
  });

  const deleteMutation = useMutation({
    mutationFn: () => apiClient.deleteWishList(id),
    onSuccess: () => {
      Alert.alert('Success', 'Wishlist deleted successfully!', [
        { text: 'OK', onPress: () => router.push('/lists') },
      ]);
    },
    onError: (error) => {
      Alert.alert(
        'Error',
        error.message || 'Failed to delete wishlist. Please try again.',
      );
    },
  });

  const onSubmit = (data: UpdateWishListFormData) => {
    updateMutation.mutate(data);
  };

  const handleDelete = () => {
    Alert.alert(
      'Confirm Delete',
      'Are you sure you want to delete this wishlist? This action cannot be undone and will also delete all associated gift items.',
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

  if (isLoading) {
    return (
      <View style={styles.container}>
        <LinearGradient
          colors={['#1a0a2e', '#2d1b4e', '#3d2a6e']}
          style={StyleSheet.absoluteFill}
        />
        <View style={styles.centerContainer}>
          <ActivityIndicator size="large" color="#FFD700" />
          <Text style={styles.loadingText}>Loading wishlist...</Text>
        </View>
      </View>
    );
  }

  if (error) {
    return (
      <View style={styles.container}>
        <LinearGradient
          colors={['#1a0a2e', '#2d1b4e', '#3d2a6e']}
          style={StyleSheet.absoluteFill}
        />
        <View style={styles.decorCircle1} />
        <View style={styles.decorCircle2} />

        <View style={styles.header}>
          <Pressable onPress={() => router.back()} style={styles.backButton}>
            <MaterialCommunityIcons
              name="arrow-left"
              size={24}
              color="#ffffff"
            />
          </Pressable>
          <Text style={styles.headerTitle}>Error</Text>
          <View style={{ width: 40 }} />
        </View>

        <View style={styles.centerContainer}>
          <BlurView intensity={20} style={styles.errorCard}>
            <MaterialCommunityIcons
              name="alert-circle"
              size={64}
              color="#FF6B6B"
            />
            <Text style={styles.errorTitle}>Error loading wishlist</Text>
            <Text style={styles.errorMessage}>{error.message}</Text>
            <Pressable onPress={() => router.back()}>
              <LinearGradient
                colors={['#FFD700', '#FFA500']}
                style={styles.errorButton}
              >
                <Text style={styles.errorButtonText}>Go Back</Text>
              </LinearGradient>
            </Pressable>
          </BlurView>
        </View>
      </View>
    );
  }

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
        <Text style={styles.headerTitle}>Edit Wishlist</Text>
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
                <MaterialCommunityIcons
                  name="pencil"
                  size={32}
                  color="#000000"
                />
              </LinearGradient>
            </View>

            {/* Title */}
            <Controller
              control={control}
              name="title"
              render={({ field: { onChange, onBlur, value } }) => (
                <>
                  <TextInput
                    label="Title *"
                    value={value}
                    onChangeText={onChange}
                    onBlur={onBlur}
                    maxLength={200}
                    style={styles.input}
                    textColor="#ffffff"
                    underlineColor="transparent"
                    activeUnderlineColor="#FFD700"
                    placeholderTextColor="rgba(255, 255, 255, 0.4)"
                    disabled={updateMutation.isPending}
                    theme={{
                      colors: {
                        primary: '#FFD700',
                        onSurfaceVariant: 'rgba(255, 255, 255, 0.5)',
                      },
                    }}
                  />
                  {errors.title && (
                    <HelperText type="error" style={styles.errorText}>
                      {errors.title.message}
                    </HelperText>
                  )}
                </>
              )}
            />

            {/* Description */}
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
                    disabled={updateMutation.isPending}
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

            {/* Occasion */}
            <Controller
              control={control}
              name="occasion"
              render={({ field: { onChange, onBlur, value } }) => (
                <TextInput
                  label="Occasion (e.g., Birthday, Wedding)"
                  value={value}
                  onChangeText={onChange}
                  onBlur={onBlur}
                  maxLength={100}
                  style={styles.input}
                  textColor="#ffffff"
                  underlineColor="transparent"
                  activeUnderlineColor="#FFD700"
                  placeholderTextColor="rgba(255, 255, 255, 0.4)"
                  disabled={updateMutation.isPending}
                  theme={{
                    colors: {
                      primary: '#FFD700',
                      onSurfaceVariant: 'rgba(255, 255, 255, 0.5)',
                    },
                  }}
                />
              )}
            />

            {/* Public Toggle */}
            <View style={styles.toggleContainer}>
              <View style={styles.toggleLeft}>
                <MaterialCommunityIcons
                  name="earth"
                  size={20}
                  color="#FFD700"
                />
                <Text style={styles.toggleLabel}>Make Public</Text>
              </View>
              <Controller
                control={control}
                name="is_public"
                render={({ field: { onChange, value } }) => (
                  <Switch
                    value={value}
                    onValueChange={onChange}
                    disabled={updateMutation.isPending}
                    trackColor={{ false: '#767577', true: '#FFD700' }}
                    thumbColor={value ? '#FFA500' : '#f4f3f4'}
                  />
                )}
              />
            </View>

            {/* Update Button */}
            <Pressable
              onPress={handleSubmit(onSubmit)}
              disabled={updateMutation.isPending}
            >
              <LinearGradient
                colors={['#FFD700', '#FFA500']}
                style={styles.updateButton}
              >
                {updateMutation.isPending ? (
                  <Text style={styles.updateButtonText}>Updating...</Text>
                ) : (
                  <>
                    <MaterialCommunityIcons
                      name="check"
                      size={20}
                      color="#000000"
                    />
                    <Text style={styles.updateButtonText}>Update Wishlist</Text>
                  </>
                )}
              </LinearGradient>
            </Pressable>

            {/* Cancel Button */}
            <Pressable
              onPress={() => router.back()}
              disabled={updateMutation.isPending}
              style={styles.cancelButton}
            >
              <Text style={styles.cancelButtonText}>Cancel</Text>
            </Pressable>
          </View>
        </BlurView>

        {/* Danger Zone */}
        <BlurView intensity={20} style={[styles.formCard, styles.dangerCard]}>
          <View style={styles.formContent}>
            <View style={styles.dangerHeader}>
              <MaterialCommunityIcons
                name="alert-circle"
                size={24}
                color="#FF6B6B"
              />
              <Text style={styles.dangerTitle}>Danger Zone</Text>
            </View>
            <Text style={styles.dangerText}>
              Deleting this wishlist will permanently remove it and all
              associated gift items. This action cannot be undone.
            </Text>
            <Pressable
              onPress={handleDelete}
              disabled={updateMutation.isPending || deleteMutation.isPending}
            >
              <LinearGradient
                colors={['#FF6B6B', '#FF4444']}
                style={styles.deleteButton}
              >
                {deleteMutation.isPending ? (
                  <Text style={styles.deleteButtonText}>Deleting...</Text>
                ) : (
                  <>
                    <MaterialCommunityIcons
                      name="delete-forever"
                      size={20}
                      color="#ffffff"
                    />
                    <Text style={styles.deleteButtonText}>Delete Wishlist</Text>
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
            Changes will be reflected immediately after saving
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
  centerContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 20,
  },
  loadingText: {
    fontSize: 16,
    color: 'rgba(255, 255, 255, 0.7)',
    marginTop: 16,
  },
  errorCard: {
    borderRadius: 24,
    overflow: 'hidden',
    backgroundColor: 'rgba(255, 255, 255, 0.08)',
    borderWidth: 1,
    borderColor: 'rgba(255, 255, 255, 0.1)',
    padding: 32,
    alignItems: 'center',
    maxWidth: 400,
  },
  errorTitle: {
    fontSize: 20,
    fontWeight: '700',
    color: '#FF6B6B',
    marginTop: 16,
    marginBottom: 8,
    textAlign: 'center',
  },
  errorMessage: {
    fontSize: 14,
    color: 'rgba(255, 255, 255, 0.6)',
    textAlign: 'center',
    marginBottom: 24,
  },
  errorButton: {
    paddingVertical: 12,
    paddingHorizontal: 32,
    borderRadius: 12,
  },
  errorButtonText: {
    fontSize: 16,
    fontWeight: '700',
    color: '#000000',
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
    marginBottom: 16,
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
  errorText: {
    color: '#FF6B6B',
    marginTop: -12,
    marginBottom: 8,
  },
  toggleContainer: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingVertical: 16,
    paddingHorizontal: 16,
    backgroundColor: 'rgba(255, 255, 255, 0.05)',
    borderRadius: 12,
    marginBottom: 24,
  },
  toggleLeft: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 12,
  },
  toggleLabel: {
    fontSize: 16,
    color: '#ffffff',
    fontWeight: '500',
  },
  updateButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 16,
    borderRadius: 12,
    gap: 8,
    marginBottom: 12,
  },
  updateButtonText: {
    fontSize: 16,
    fontWeight: '700',
    color: '#000000',
  },
  cancelButton: {
    paddingVertical: 16,
    borderRadius: 12,
    alignItems: 'center',
    backgroundColor: 'rgba(255, 255, 255, 0.05)',
    borderWidth: 1,
    borderColor: 'rgba(255, 255, 255, 0.1)',
  },
  cancelButtonText: {
    fontSize: 16,
    fontWeight: '600',
    color: 'rgba(255, 255, 255, 0.7)',
  },
  dangerCard: {
    backgroundColor: 'rgba(255, 107, 107, 0.08)',
    borderColor: 'rgba(255, 107, 107, 0.2)',
  },
  dangerHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 12,
    marginBottom: 12,
  },
  dangerTitle: {
    fontSize: 18,
    fontWeight: '700',
    color: '#FF6B6B',
  },
  dangerText: {
    fontSize: 14,
    color: 'rgba(255, 255, 255, 0.6)',
    lineHeight: 20,
    marginBottom: 16,
  },
  deleteButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 16,
    borderRadius: 12,
    gap: 8,
  },
  deleteButtonText: {
    fontSize: 16,
    fontWeight: '700',
    color: '#ffffff',
  },
  helpContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    paddingHorizontal: 4,
  },
  helpText: {
    fontSize: 13,
    color: 'rgba(255, 255, 255, 0.5)',
    flex: 1,
  },
});
