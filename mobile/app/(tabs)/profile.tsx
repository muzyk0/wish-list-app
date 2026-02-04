import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "expo-router";
import { Controller, useForm } from "react-hook-form";
import { Alert, ScrollView, StyleSheet, View } from "react-native";
import {
  ActivityIndicator,
  Avatar,
  Button,
  Card,
  Divider,
  HelperText,
  Switch,
  Text,
  TextInput,
  useTheme,
} from "react-native-paper";
import { z } from "zod";
import { useThemeContext } from "@/contexts/ThemeContext";
import { apiClient } from "@/lib/api";

// Validation Schemas
const profileUpdateSchema = z.object({
  first_name: z.string().optional(),
  last_name: z.string().optional(),
  avatar_url: z.string().url("Invalid URL").optional().or(z.literal("")),
});

const emailChangeSchema = z.object({
  new_email: z.string().email("Invalid email address"),
  current_password: z.string().min(6, "Password must be at least 6 characters"),
});

const passwordChangeSchema = z
  .object({
    current_password: z
      .string()
      .min(6, "Password must be at least 6 characters"),
    new_password: z.string().min(6, "Password must be at least 6 characters"),
    confirm_password: z
      .string()
      .min(6, "Password must be at least 6 characters"),
  })
  .refine((data) => data.new_password === data.confirm_password, {
    message: "Passwords do not match",
    path: ["confirm_password"],
  });

type ProfileUpdateForm = z.infer<typeof profileUpdateSchema>;
type EmailChangeForm = z.infer<typeof emailChangeSchema>;
type PasswordChangeForm = z.infer<typeof passwordChangeSchema>;

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
    queryKey: ["profile"],
    queryFn: () => apiClient.getProfile(),
    retry: 1,
  });

  // Profile Update Form
  const profileForm = useForm<ProfileUpdateForm>({
    resolver: zodResolver(profileUpdateSchema),
    defaultValues: {
      first_name: user?.first_name || "",
      last_name: user?.last_name || "",
      avatar_url: user?.avatar_url || "",
    },
    values: {
      first_name: user?.first_name || "",
      last_name: user?.last_name || "",
      avatar_url: user?.avatar_url || "",
    },
  });

  // Email Change Form
  const emailForm = useForm<EmailChangeForm>({
    resolver: zodResolver(emailChangeSchema),
    defaultValues: {
      new_email: "",
      current_password: "",
    },
  });

  // Password Change Form
  const passwordForm = useForm<PasswordChangeForm>({
    resolver: zodResolver(passwordChangeSchema),
    defaultValues: {
      current_password: "",
      new_password: "",
      confirm_password: "",
    },
  });

  // Mutations
  const updateProfileMutation = useMutation({
    mutationFn: (data: ProfileUpdateForm) => apiClient.updateProfile(data),
    onSuccess: () => {
      Alert.alert("Success", "Profile updated successfully!");
      queryClient.invalidateQueries({ queryKey: ["profile"] });
    },
    // biome-ignore lint/suspicious/noExplicitAny: Error type
    onError: (error: any) => {
      Alert.alert(
        "Error",
        error.message || "Failed to update profile. Please try again.",
      );
    },
  });

  const changeEmailMutation = useMutation({
    mutationFn: (data: EmailChangeForm) =>
      apiClient.changeEmail(data.current_password, data.new_email),
    onSuccess: () => {
      Alert.alert("Success", "Email changed successfully!");
      emailForm.reset();
      queryClient.invalidateQueries({ queryKey: ["profile"] });
    },
    // biome-ignore lint/suspicious/noExplicitAny: Error type
    onError: (error: any) => {
      Alert.alert(
        "Error",
        error.message || "Failed to change email. Please try again.",
      );
    },
  });

  const changePasswordMutation = useMutation({
    mutationFn: (data: PasswordChangeForm) =>
      apiClient.changePassword(data.current_password, data.new_password),
    onSuccess: () => {
      Alert.alert("Success", "Password changed successfully!");
      passwordForm.reset();
    },
    // biome-ignore lint/suspicious/noExplicitAny: Error type
    onError: (error: any) => {
      Alert.alert(
        "Error",
        error.message || "Failed to change password. Please try again.",
      );
    },
  });

  const logoutMutation = useMutation({
    mutationFn: () => apiClient.logout(),
    onSuccess: () => {
      queryClient.clear();
      router.replace("/auth/login");
    },
    // biome-ignore lint/suspicious/noExplicitAny: Error type
    onError: (error: any) => {
      Alert.alert(
        "Error",
        error.message || "Failed to logout. Please try again.",
      );
    },
  });

  const deleteMutation = useMutation({
    mutationFn: () => apiClient.deleteAccount(),
    onSuccess: async () => {
      await apiClient.logout();
      queryClient.clear();

      Alert.alert("Success", "Account deleted successfully!", [
        {
          text: "OK",
          onPress: () => {
            router.replace("/auth/login");
          },
        },
      ]);
    },
    // biome-ignore lint/suspicious/noExplicitAny: Error type
    onError: (error: any) => {
      Alert.alert(
        "Error",
        error.message || "Failed to delete account. Please try again.",
      );
    },
  });

  const handleLogout = () => {
    Alert.alert("Confirm Logout", "Are you sure you want to logout?", [
      { text: "Cancel", style: "cancel" },
      {
        text: "Logout",
        onPress: () => logoutMutation.mutate(),
      },
    ]);
  };

  const handleDeleteAccount = () => {
    Alert.alert(
      "Confirm Delete",
      "Are you sure you want to delete your account? This action cannot be undone.",
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
      {/* Header Section */}
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
                      (user.last_name?.charAt(0) || "")
                    ).toUpperCase()
                  : user?.email.charAt(0).toUpperCase() || "?"
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
            : user?.first_name || user?.email.split("@")[0]}
        </Text>
        <Text
          variant="bodyLarge"
          style={[styles.email, { color: colors.outline }]}
        >
          {user?.email}
        </Text>
      </View>

      {/* Profile Update Section (Low Security) */}
      <Card style={styles.card}>
        <Card.Content>
          <Text
            variant="titleLarge"
            style={[styles.sectionTitle, { color: colors.onSurface }]}
          >
            Profile Information
          </Text>
          <Text
            variant="bodySmall"
            style={[styles.sectionDescription, { color: colors.outline }]}
          >
            Update your public profile details
          </Text>

          <Divider style={styles.divider} />

          <Controller
            control={profileForm.control}
            name="first_name"
            render={({ field: { onChange, onBlur, value } }) => (
              <>
                <TextInput
                  label="First Name"
                  value={value || ""}
                  onChangeText={onChange}
                  onBlur={onBlur}
                  mode="outlined"
                  disabled={updateProfileMutation.isPending}
                  style={styles.input}
                  left={<TextInput.Icon icon="account" />}
                />
                {profileForm.formState.errors.first_name && (
                  <HelperText type="error">
                    {profileForm.formState.errors.first_name.message}
                  </HelperText>
                )}
              </>
            )}
          />

          <Controller
            control={profileForm.control}
            name="last_name"
            render={({ field: { onChange, onBlur, value } }) => (
              <>
                <TextInput
                  label="Last Name"
                  value={value || ""}
                  onChangeText={onChange}
                  onBlur={onBlur}
                  mode="outlined"
                  disabled={updateProfileMutation.isPending}
                  style={styles.input}
                  left={<TextInput.Icon icon="account" />}
                />
                {profileForm.formState.errors.last_name && (
                  <HelperText type="error">
                    {profileForm.formState.errors.last_name.message}
                  </HelperText>
                )}
              </>
            )}
          />

          <Controller
            control={profileForm.control}
            name="avatar_url"
            render={({ field: { onChange, onBlur, value } }) => (
              <>
                <TextInput
                  label="Avatar URL"
                  value={value || ""}
                  onChangeText={onChange}
                  onBlur={onBlur}
                  keyboardType="url"
                  autoCapitalize="none"
                  mode="outlined"
                  disabled={updateProfileMutation.isPending}
                  style={styles.input}
                  left={<TextInput.Icon icon="image" />}
                />
                {profileForm.formState.errors.avatar_url && (
                  <HelperText type="error">
                    {profileForm.formState.errors.avatar_url.message}
                  </HelperText>
                )}
              </>
            )}
          />

          <Button
            mode="contained"
            onPress={profileForm.handleSubmit((data) =>
              updateProfileMutation.mutate(data),
            )}
            loading={updateProfileMutation.isPending}
            disabled={updateProfileMutation.isPending}
            style={styles.button}
            labelStyle={styles.buttonLabel}
          >
            Update Profile
          </Button>
        </Card.Content>
      </Card>

      {/* Email Change Section (High Security) */}
      <Card style={styles.card}>
        <Card.Content>
          <Text
            variant="titleLarge"
            style={[styles.sectionTitle, { color: colors.onSurface }]}
          >
            Change Email
          </Text>
          <Text
            variant="bodySmall"
            style={[styles.sectionDescription, { color: colors.outline }]}
          >
            Update your email address (requires password verification)
          </Text>

          <Divider style={styles.divider} />

          <Controller
            control={emailForm.control}
            name="new_email"
            render={({ field: { onChange, onBlur, value } }) => (
              <>
                <TextInput
                  label="New Email"
                  value={value}
                  onChangeText={onChange}
                  onBlur={onBlur}
                  keyboardType="email-address"
                  autoCapitalize="none"
                  mode="outlined"
                  disabled={changeEmailMutation.isPending}
                  style={styles.input}
                  left={<TextInput.Icon icon="email" />}
                />
                {emailForm.formState.errors.new_email && (
                  <HelperText type="error">
                    {emailForm.formState.errors.new_email.message}
                  </HelperText>
                )}
              </>
            )}
          />

          <Controller
            control={emailForm.control}
            name="current_password"
            render={({ field: { onChange, onBlur, value } }) => (
              <>
                <TextInput
                  label="Current Password"
                  value={value}
                  onChangeText={onChange}
                  onBlur={onBlur}
                  secureTextEntry
                  autoCapitalize="none"
                  mode="outlined"
                  disabled={changeEmailMutation.isPending}
                  style={styles.input}
                  left={<TextInput.Icon icon="lock" />}
                />
                {emailForm.formState.errors.current_password && (
                  <HelperText type="error">
                    {emailForm.formState.errors.current_password.message}
                  </HelperText>
                )}
              </>
            )}
          />

          <Button
            mode="contained"
            onPress={emailForm.handleSubmit((data) =>
              changeEmailMutation.mutate(data),
            )}
            loading={changeEmailMutation.isPending}
            disabled={changeEmailMutation.isPending}
            style={styles.button}
            labelStyle={styles.buttonLabel}
          >
            Change Email
          </Button>
        </Card.Content>
      </Card>

      {/* Password Change Section (High Security) */}
      <Card style={styles.card}>
        <Card.Content>
          <Text
            variant="titleLarge"
            style={[styles.sectionTitle, { color: colors.onSurface }]}
          >
            Change Password
          </Text>
          <Text
            variant="bodySmall"
            style={[styles.sectionDescription, { color: colors.outline }]}
          >
            Update your password (requires current password)
          </Text>

          <Divider style={styles.divider} />

          <Controller
            control={passwordForm.control}
            name="current_password"
            render={({ field: { onChange, onBlur, value } }) => (
              <>
                <TextInput
                  label="Current Password"
                  value={value}
                  onChangeText={onChange}
                  onBlur={onBlur}
                  secureTextEntry
                  autoCapitalize="none"
                  mode="outlined"
                  disabled={changePasswordMutation.isPending}
                  style={styles.input}
                  left={<TextInput.Icon icon="lock" />}
                />
                {passwordForm.formState.errors.current_password && (
                  <HelperText type="error">
                    {passwordForm.formState.errors.current_password.message}
                  </HelperText>
                )}
              </>
            )}
          />

          <Controller
            control={passwordForm.control}
            name="new_password"
            render={({ field: { onChange, onBlur, value } }) => (
              <>
                <TextInput
                  label="New Password"
                  value={value}
                  onChangeText={onChange}
                  onBlur={onBlur}
                  secureTextEntry
                  autoCapitalize="none"
                  mode="outlined"
                  disabled={changePasswordMutation.isPending}
                  style={styles.input}
                  left={<TextInput.Icon icon="lock-plus" />}
                />
                {passwordForm.formState.errors.new_password && (
                  <HelperText type="error">
                    {passwordForm.formState.errors.new_password.message}
                  </HelperText>
                )}
              </>
            )}
          />

          <Controller
            control={passwordForm.control}
            name="confirm_password"
            render={({ field: { onChange, onBlur, value } }) => (
              <>
                <TextInput
                  label="Confirm New Password"
                  value={value}
                  onChangeText={onChange}
                  onBlur={onBlur}
                  secureTextEntry
                  autoCapitalize="none"
                  mode="outlined"
                  disabled={changePasswordMutation.isPending}
                  style={styles.input}
                  left={<TextInput.Icon icon="lock-check" />}
                />
                {passwordForm.formState.errors.confirm_password && (
                  <HelperText type="error">
                    {passwordForm.formState.errors.confirm_password.message}
                  </HelperText>
                )}
              </>
            )}
          />

          <Button
            mode="contained"
            onPress={passwordForm.handleSubmit((data) =>
              changePasswordMutation.mutate(data),
            )}
            loading={changePasswordMutation.isPending}
            disabled={changePasswordMutation.isPending}
            style={styles.button}
            labelStyle={styles.buttonLabel}
          >
            Change Password
          </Button>
        </Card.Content>
      </Card>

      {/* Appearance Section */}
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

      {/* Account Actions Section */}
      <Card style={styles.card}>
        <Card.Content>
          <Text
            variant="titleLarge"
            style={[styles.sectionTitle, { color: colors.onSurface }]}
          >
            Account
          </Text>

          <Divider style={styles.divider} />

          <Button
            mode="outlined"
            onPress={handleLogout}
            loading={logoutMutation.isPending}
            disabled={logoutMutation.isPending}
            style={styles.button}
            labelStyle={styles.buttonLabel}
            icon="logout"
          >
            Logout
          </Button>
        </Card.Content>
      </Card>

      {/* Danger Zone Section */}
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
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    justifyContent: "center",
    alignItems: "center",
    padding: 16,
  },
  title: {
    fontSize: 20,
    fontWeight: "bold",
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
    alignItems: "center",
    paddingVertical: 32,
    paddingHorizontal: 16,
    backgroundColor: "transparent",
  },
  avatarContainer: {
    marginBottom: 16,
  },
  avatar: {
    backgroundColor: "#6200ee",
  },
  name: {
    fontSize: 24,
    fontWeight: "bold",
    marginTop: 12,
    marginBottom: 4,
  },
  email: {
    fontSize: 16,
    opacity: 0.7,
  },
  sectionTitle: {
    fontSize: 20,
    fontWeight: "600",
    marginBottom: 4,
  },
  sectionDescription: {
    fontSize: 12,
    marginBottom: 12,
    opacity: 0.7,
  },
  divider: {
    marginVertical: 12,
  },
  input: {
    marginBottom: 4,
  },
  button: {
    marginTop: 8,
    borderRadius: 8,
    paddingVertical: 6,
  },
  buttonLabel: {
    fontWeight: "600",
    fontSize: 16,
  },
  dangerDescription: {
    marginBottom: 16,
    lineHeight: 20,
  },
  dangerButton: {
    borderRadius: 8,
    paddingVertical: 6,
    backgroundColor: "transparent",
  },
  dangerButtonLabel: {
    fontWeight: "600",
    fontSize: 16,
  },
  themeToggleContainer: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    paddingVertical: 8,
  },
});
