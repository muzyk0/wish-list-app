import { MaterialCommunityIcons } from '@expo/vector-icons';
import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { BlurView } from 'expo-blur';
import { LinearGradient } from 'expo-linear-gradient';
import { useRouter } from 'expo-router';
import { Controller, useForm } from 'react-hook-form';
import { Alert, Pressable, ScrollView, StyleSheet, View } from 'react-native';
import {
  ActivityIndicator,
  Avatar,
  HelperText,
  Text,
  TextInput,
} from 'react-native-paper';
import { z } from 'zod';
import { apiClient } from '@/lib/api';
import { dialog } from '@/stores/dialogStore';

const profileUpdateSchema = z.object({
  first_name: z.string().optional(),
  last_name: z.string().optional(),
  avatar_url: z.string().url('Invalid URL').optional().or(z.literal('')),
});

type ProfileUpdateForm = z.infer<typeof profileUpdateSchema>;

export default function EditProfileScreen() {
  const router = useRouter();
  const queryClient = useQueryClient();

  const { data: user, isLoading } = useQuery({
    queryKey: ['profile'],
    queryFn: () => apiClient.getProfile(),
    retry: 1,
  });

  const profileForm = useForm<ProfileUpdateForm>({
    resolver: zodResolver(profileUpdateSchema),
    defaultValues: {
      first_name: user?.first_name || '',
      last_name: user?.last_name || '',
      avatar_url: user?.avatar_url || '',
    },
    values: {
      first_name: user?.first_name || '',
      last_name: user?.last_name || '',
      avatar_url: user?.avatar_url || '',
    },
  });

  const updateProfileMutation = useMutation({
    mutationFn: (data: ProfileUpdateForm) => apiClient.updateProfile(data),
    onSuccess: () => {
      dialog.success('Profile updated successfully!');
      queryClient.invalidateQueries({ queryKey: ['profile'] });
      router.back();
    },
    onError: (error: Error) => {
      dialog.error(error.message || 'Failed to update profile');
    },
  });

  if (isLoading) {
    return (
      <View style={styles.container}>
        <LinearGradient
          colors={['#1a0a2e', '#2d1b4e', '#3d2a6e']}
          style={StyleSheet.absoluteFill}
        />
        <View style={styles.loadingContainer}>
          <ActivityIndicator size="large" color="#FFD700" />
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
        <Text style={styles.headerTitle}>Edit Profile</Text>
        <View style={{ width: 40 }} />
      </View>

      <ScrollView
        style={styles.scrollView}
        contentContainerStyle={styles.scrollContent}
      >
        {/* Avatar */}
        <View style={styles.avatarSection}>
          <LinearGradient
            colors={['#FFD700', '#FFA500']}
            style={styles.avatarGradient}
          >
            <Avatar.Text
              size={100}
              label={`${user?.first_name?.[0] || ''}${user?.last_name?.[0] || ''}`.toUpperCase()}
              color="#000000"
              style={styles.avatar}
            />
          </LinearGradient>
          <Pressable>
            <Text style={styles.changePhotoText}>Change Photo</Text>
          </Pressable>
        </View>

        {/* Form */}
        <BlurView intensity={20} style={styles.formCard}>
          <View style={styles.formContent}>
            <Controller
              control={profileForm.control}
              name="first_name"
              render={({
                field: { onChange, value },
                fieldState: { error },
              }) => (
                <>
                  <TextInput
                    label="First Name"
                    value={value}
                    onChangeText={onChange}
                    style={styles.input}
                    textColor="#ffffff"
                    underlineColor="transparent"
                    activeUnderlineColor="#FFD700"
                    placeholderTextColor="rgba(255, 255, 255, 0.4)"
                    theme={{
                      colors: {
                        primary: '#FFD700',
                        onSurfaceVariant: 'rgba(255, 255, 255, 0.5)',
                      },
                    }}
                  />
                  {error && (
                    <HelperText type="error" style={styles.errorText}>
                      {error.message}
                    </HelperText>
                  )}
                </>
              )}
            />

            <Controller
              control={profileForm.control}
              name="last_name"
              render={({
                field: { onChange, value },
                fieldState: { error },
              }) => (
                <>
                  <TextInput
                    label="Last Name"
                    value={value}
                    onChangeText={onChange}
                    style={styles.input}
                    textColor="#ffffff"
                    underlineColor="transparent"
                    activeUnderlineColor="#FFD700"
                    placeholderTextColor="rgba(255, 255, 255, 0.4)"
                    theme={{
                      colors: {
                        primary: '#FFD700',
                        onSurfaceVariant: 'rgba(255, 255, 255, 0.5)',
                      },
                    }}
                  />
                  {error && (
                    <HelperText type="error" style={styles.errorText}>
                      {error.message}
                    </HelperText>
                  )}
                </>
              )}
            />

            <Pressable
              onPress={profileForm.handleSubmit((data) =>
                updateProfileMutation.mutate(data),
              )}
              disabled={updateProfileMutation.isPending}
            >
              <LinearGradient
                colors={['#FFD700', '#FFA500']}
                style={styles.saveButton}
              >
                <Text style={styles.saveButtonText}>
                  {updateProfileMutation.isPending
                    ? 'Saving...'
                    : 'Save Changes'}
                </Text>
              </LinearGradient>
            </Pressable>
          </View>
        </BlurView>
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
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
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
  avatarSection: {
    alignItems: 'center',
    marginBottom: 32,
  },
  avatarGradient: {
    width: 120,
    height: 120,
    borderRadius: 60,
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: 16,
    shadowColor: '#FFD700',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.3,
    shadowRadius: 8,
    elevation: 8,
  },
  avatar: {
    backgroundColor: 'transparent',
  },
  changePhotoText: {
    fontSize: 15,
    fontWeight: '600',
    color: '#FFD700',
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
  input: {
    backgroundColor: 'rgba(255, 255, 255, 0.08)',
    marginBottom: 16,
    borderRadius: 12,
  },
  errorText: {
    color: '#FF6B6B',
    marginTop: -12,
    marginBottom: 8,
  },
  saveButton: {
    paddingVertical: 16,
    borderRadius: 12,
    alignItems: 'center',
    marginTop: 8,
  },
  saveButtonText: {
    fontSize: 16,
    fontWeight: '700',
    color: '#000000',
  },
});
