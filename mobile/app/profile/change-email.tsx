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

const emailChangeSchema = z.object({
  new_email: z.email('Invalid email address'),
  current_password: z.string().min(6, 'Password must be at least 6 characters'),
});

type EmailChangeForm = z.infer<typeof emailChangeSchema>;

export default function ChangeEmailScreen() {
  const router = useRouter();
  const queryClient = useQueryClient();

  const emailForm = useForm<EmailChangeForm>({
    resolver: zodResolver(emailChangeSchema),
    defaultValues: {
      new_email: '',
      current_password: '',
    },
  });

  const changeEmailMutation = useMutation({
    mutationFn: (data: EmailChangeForm) =>
      apiClient.changeEmail({
        newEmail: data.new_email,
        currentPassword: data.current_password,
      }),
    onSuccess: () => {
      dialog.success('Email updated successfully!');
      emailForm.reset();
      queryClient.invalidateQueries({ queryKey: ['profile'] });
      router.back();
    },
    onError: (error: Error) => {
      dialog.error(error.message || 'Failed to change email');
    },
  });

  return (
    <View style={styles.container}>
      <LinearGradient
        colors={['#1a0a2e', '#2d1b4e', '#3d2a6e']}
        style={StyleSheet.absoluteFill}
      />

      {/* Header */}
      <View style={styles.header}>
        <Pressable onPress={() => router.back()} style={styles.backButton}>
          <MaterialCommunityIcons name="arrow-left" size={24} color="#ffffff" />
        </Pressable>
        <Text style={styles.headerTitle}>Change Email</Text>
        <View style={{ width: 40 }} />
      </View>

      <ScrollView
        style={styles.scrollView}
        contentContainerStyle={styles.scrollContent}
      >
        <BlurView intensity={20} style={styles.formCard}>
          <View style={styles.formContent}>
            <Text style={styles.description}>
              Enter your new email address and current password to confirm the
              change.
            </Text>

            <Controller
              control={emailForm.control}
              name="new_email"
              render={({
                field: { onChange, value },
                fieldState: { error },
              }) => (
                <>
                  <TextInput
                    label="New Email"
                    value={value}
                    onChangeText={onChange}
                    keyboardType="email-address"
                    autoCapitalize="none"
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
              control={emailForm.control}
              name="current_password"
              render={({
                field: { onChange, value },
                fieldState: { error },
              }) => (
                <>
                  <TextInput
                    label="Current Password"
                    value={value}
                    onChangeText={onChange}
                    secureTextEntry
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
              onPress={emailForm.handleSubmit((data) =>
                changeEmailMutation.mutate(data),
              )}
              disabled={changeEmailMutation.isPending}
            >
              <LinearGradient
                colors={['#6B4EE6', '#9B6DFF']}
                style={styles.saveButton}
              >
                <Text style={styles.saveButtonText}>
                  {changeEmailMutation.isPending
                    ? 'Changing Email...'
                    : 'Change Email'}
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
    borderRadius: 20,
    overflow: 'hidden',
    backgroundColor: 'rgba(255, 255, 255, 0.08)',
    borderWidth: 1,
    borderColor: 'rgba(255, 255, 255, 0.1)',
  },
  formContent: {
    padding: 20,
  },
  description: {
    fontSize: 14,
    color: 'rgba(255, 255, 255, 0.6)',
    marginBottom: 24,
    lineHeight: 20,
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
    color: '#ffffff',
  },
});
