import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useRouter } from 'expo-router';
import { useEffect, useState } from 'react';
import { Alert, ScrollView, StyleSheet, View } from 'react-native';
import {
  ActivityIndicator,
  Avatar,
  Button,
  Card,
  Divider,
  Switch,
  Text,
  TextInput,
  useTheme,
} from 'react-native-paper';
import { useThemeContext } from '@/contexts/ThemeContext';
import { apiClient } from '@/lib/api';

export default function ProfileScreen() {
  const queryClient = useQueryClient();
  const router = useRouter();
  const { colors } = useTheme();
  const { isDark, toggleTheme } = useThemeContext();

  const {
    data: user,
    isLoading,
    error,
    refetch,
  } = useQuery({
    queryKey: ['profile'],
    queryFn: () => apiClient.getProfile(),
    retry: 1,
  });

  const [email, setEmail] = useState('');
  const [firstName, setFirstName] = useState('');
  const [lastName, setLastName] = useState('');
  const [avatarUrl, setAvatarUrl] = useState('');

  useEffect(() => {
    if (user) {
      setEmail(user.email);
      setFirstName(user.first_name || '');
      setLastName(user.last_name || '');
      setAvatarUrl(user.avatar_url || '');
    }
  }, [user]);

  const updateMutation = useMutation({
    mutationFn: (userData: {
      email: string;
      first_name?: string;
      last_name?: string;
      avatar_url?: string;
    }) =>
      apiClient.updateProfile({
        // email: userData.email,
        first_name: userData.first_name,
        last_name: userData.last_name,
        avatar_url: userData.avatar_url,
      }),
    onSuccess: () => {
      Alert.alert('Success', 'Profile updated successfully!');
      queryClient.invalidateQueries({ queryKey: ['profile'] });
    },
    // biome-ignore lint/suspicious/noExplicitAny: Error type
    onError: (error: any) => {
      Alert.alert(
        'Error',
        error.message || 'Failed to update profile. Please try again.',
      );
    },
  });

  const handleUpdateProfile = () => {
    if (!email) {
      Alert.alert('Error', 'Email is required.');
      return;
    }

    updateMutation.mutate({
      email,
      first_name: firstName.trim() || undefined,
      last_name: lastName.trim() || undefined,
      avatar_url: avatarUrl.trim() || undefined,
    });
  };

  const deleteMutation = useMutation({
    mutationFn: () => apiClient.deleteAccount(),
    onSuccess: async () => {
      // Clear auth session and cached data
      await apiClient.logout();
      queryClient.clear();

      Alert.alert('Success', 'Account deleted successfully!', [
        {
          text: 'OK',
          onPress: () => {
            // Navigate to auth flow
            router.replace('/auth/login');
          },
        },
      ]);
    },
    // biome-ignore lint/suspicious/noExplicitAny: Error type
    onError: (error: any) => {
      Alert.alert(
        'Error',
        error.message || 'Failed to delete account. Please try again.',
      );
    },
  });

  const handleDeleteAccount = () => {
    Alert.alert(
      'Confirm Delete',
      'Are you sure you want to delete your account? This action cannot be undone.',
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
      <View style={[styles.container, { backgroundColor: colors.background }]}>
        <ActivityIndicator size="large" animating={true} />
      </View>
    );
  }

  if (error) {
    return (
      <View style={[styles.container, { backgroundColor: colors.background }]}>
        <Card style={styles.card}>
          <Card.Content>
            <Text variant="headlineMedium" style={styles.title}>
              Error loading profile
            </Text>
            <Text variant="bodyLarge">{error.message}</Text>
            <Button
              mode="contained"
              onPress={() => refetch()}
              style={styles.button}
            >
              Retry
            </Button>
          </Card.Content>
        </Card>
      </View>
    );
  }

  return (
    <ScrollView style={{ flex: 1, backgroundColor: colors.background }}>
      <View style={styles.headerSection}>
        <View style={styles.avatarContainer}>
          {user?.avatar_url ? (
            <Avatar.Image
              size={100}
              source={{ uri: user.avatar_url }}
              style={styles.avatar}
            />
          ) : (
            <Avatar.Text
              size={100}
              label={
                user?.first_name
                  ? (
                      user.first_name.charAt(0) +
                      (user.last_name?.charAt(0) || '')
                    ).toUpperCase()
                  : user?.email.charAt(0).toUpperCase() || '?'
              }
              style={styles.avatar}
            />
          )}
        </View>

        <Text
          variant="headlineMedium"
          style={[styles.name, { color: colors.onSurface }]}
        >
          {user?.first_name && user?.last_name
            ? `${user.first_name} ${user.last_name}`
            : user?.first_name || user?.email.split('@')[0]}
        </Text>
        <Text
          variant="bodyLarge"
          style={[styles.email, { color: colors.outline }]}
        >
          {user?.email}
        </Text>
      </View>

      <Card style={styles.card}>
        <Card.Content>
          <Text
            variant="titleLarge"
            style={[styles.sectionTitle, { color: colors.onSurface }]}
          >
            Account Information
          </Text>

          <Divider style={styles.divider} />

          <TextInput
            label="Email"
            value={email}
            onChangeText={setEmail}
            keyboardType="email-address"
            autoCapitalize="none"
            mode="outlined"
            disabled={updateMutation.isPending}
            style={styles.input}
            left={<TextInput.Icon icon="email" />}
          />

          <TextInput
            label="First Name"
            value={firstName}
            onChangeText={setFirstName}
            mode="outlined"
            disabled={updateMutation.isPending}
            style={styles.input}
            left={<TextInput.Icon icon="account" />}
          />

          <TextInput
            label="Last Name"
            value={lastName}
            onChangeText={setLastName}
            mode="outlined"
            disabled={updateMutation.isPending}
            style={styles.input}
            left={<TextInput.Icon icon="account" />}
          />

          <TextInput
            label="Avatar URL"
            value={avatarUrl}
            onChangeText={setAvatarUrl}
            keyboardType="url"
            autoCapitalize="none"
            mode="outlined"
            disabled={updateMutation.isPending}
            style={styles.input}
            left={<TextInput.Icon icon="image" />}
          />

          <Button
            mode="contained"
            onPress={handleUpdateProfile}
            loading={updateMutation.isPending}
            disabled={updateMutation.isPending}
            style={styles.button}
            labelStyle={styles.buttonLabel}
          >
            Update Profile
          </Button>
        </Card.Content>
      </Card>

      <Card style={styles.card}>
        <Card.Content>
          <Text
            variant="titleLarge"
            style={[styles.sectionTitle, { color: colors.onSurface }]}
          >
            Danger Zone
          </Text>

          <Divider style={styles.divider} />

          <Text
            variant="bodyMedium"
            style={[styles.dangerDescription, { color: colors.onSurface }]}
          >
            Permanently delete your account and all associated data. This action
            cannot be undone.
          </Text>

          <Button
            mode="contained-tonal"
            onPress={handleDeleteAccount}
            loading={deleteMutation.isPending}
            disabled={deleteMutation.isPending}
            style={styles.dangerButton}
            labelStyle={styles.dangerButtonLabel}
            textColor={colors.error}
          >
            Delete Account
          </Button>
        </Card.Content>
      </Card>

      <Card style={styles.card}>
        <Card.Content>
          <Text
            variant="titleLarge"
            style={[styles.sectionTitle, { color: colors.onSurface }]}
          >
            Appearance
          </Text>

          <Divider style={styles.divider} />

          <View style={styles.themeToggleContainer}>
            <Text variant="bodyLarge" style={{ color: colors.onSurface }}>
              Dark Mode
            </Text>
            <Switch
              value={isDark}
              onValueChange={toggleTheme}
              color={colors.primary}
            />
          </View>
        </Card.Content>
      </Card>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 16,
  },
  title: {
    fontSize: 20,
    fontWeight: 'bold',
    marginBottom: 8,
  },
  card: {
    margin: 16,
    borderRadius: 12,
    elevation: 4,
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
  },
  headerSection: {
    alignItems: 'center',
    paddingVertical: 32,
    paddingHorizontal: 16,
    backgroundColor: 'transparent',
  },
  avatarContainer: {
    marginBottom: 16,
  },
  avatar: {
    backgroundColor: '#6200ee',
  },
  name: {
    fontSize: 24,
    fontWeight: 'bold',
    marginTop: 12,
    marginBottom: 4,
  },
  email: {
    fontSize: 16,
    opacity: 0.7,
  },
  sectionTitle: {
    fontSize: 20,
    fontWeight: '600',
    marginBottom: 16,
  },
  divider: {
    marginVertical: 12,
  },
  input: {
    marginBottom: 16,
  },
  button: {
    marginTop: 8,
    borderRadius: 8,
    paddingVertical: 6,
  },
  buttonLabel: {
    fontWeight: '600',
    fontSize: 16,
  },
  dangerDescription: {
    marginBottom: 16,
    lineHeight: 20,
  },
  dangerButton: {
    borderRadius: 8,
    paddingVertical: 6,
    backgroundColor: 'transparent',
  },
  dangerButtonLabel: {
    fontWeight: '600',
    fontSize: 16,
  },
  themeToggleContainer: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingVertical: 8,
  },
});
