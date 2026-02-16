import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation } from '@tanstack/react-query';
import { useRouter } from 'expo-router';
import { useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { Alert, StyleSheet, View } from 'react-native';
import { z } from 'zod';
import {
  AuthDivider,
  AuthFooter,
  AuthGradientButton,
  AuthInput,
  AuthLayout,
} from '@/components/auth';
import { OAuthButtonGroup } from '@/components/OAuthButton';
import { useOAuthHandler } from '@/hooks/useOAuthHandler';
import { registerUser } from '@/lib/api';

// Zod schema for registration form validation
const registerSchema = z.object({
  firstName: z
    .string()
    .min(1, 'First name is required')
    .max(100, 'First name is too long'),
  lastName: z
    .string()
    .min(1, 'Last name is required')
    .max(100, 'Last name is too long'),
  email: z.string().min(1, 'Email is required').email('Invalid email address'),
  password: z
    .string()
    .min(6, 'Password must be at least 6 characters')
    .max(100, 'Password is too long'),
});

type RegisterFormData = z.infer<typeof registerSchema>;

export default function RegisterScreen() {
  const [showPassword, setShowPassword] = useState(false);
  const router = useRouter();
  const { oauthLoading, handleOAuth } = useOAuthHandler();

  const {
    control,
    handleSubmit,
    formState: { errors },
  } = useForm<RegisterFormData>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      firstName: '',
      lastName: '',
      email: '',
      password: '',
    },
  });

  const mutation = useMutation({
    mutationFn: ({
      email,
      password,
      firstName,
      lastName,
    }: {
      email: string;
      password: string;
      firstName: string;
      lastName: string;
    }) => registerUser({ email, password, firstName, lastName }),
    onSuccess: () => {
      router.replace('/(tabs)');
    },
    onError: (error: Error) => {
      Alert.alert(
        'Error',
        error.message || 'Registration failed. Please try again.',
      );
    },
  });

  const onSubmit = (data: RegisterFormData) => {
    mutation.mutate(data);
  };

  return (
    <AuthLayout title="Wish List" subtitle="Create Your Account">
      <View style={styles.nameRow}>
        <Controller
          control={control}
          name="firstName"
          render={({ field: { onChange, onBlur, value } }) => (
            <AuthInput
              testID="register-firstname-input"
              placeholder="First Name"
              value={value}
              onChangeText={onChange}
              onBlur={onBlur}
              icon="account-outline"
              style={styles.nameInput}
              error={errors.firstName?.message}
            />
          )}
        />

        <Controller
          control={control}
          name="lastName"
          render={({ field: { onChange, onBlur, value } }) => (
            <AuthInput
              testID="register-lastname-input"
              placeholder="Last Name"
              value={value}
              onChangeText={onChange}
              onBlur={onBlur}
              style={styles.nameInput}
              error={errors.lastName?.message}
            />
          )}
        />
      </View>

      <Controller
        control={control}
        name="email"
        render={({ field: { onChange, onBlur, value } }) => (
          <AuthInput
            testID="register-email-input"
            placeholder="Email"
            value={value}
            onChangeText={onChange}
            onBlur={onBlur}
            icon="email-outline"
            keyboardType="email-address"
            error={errors.email?.message}
          />
        )}
      />

      <Controller
        control={control}
        name="password"
        render={({ field: { onChange, onBlur, value } }) => (
          <AuthInput
            testID="register-password-input"
            placeholder="Password"
            value={value}
            onChangeText={onChange}
            onBlur={onBlur}
            icon="lock-outline"
            secureTextEntry
            showPasswordToggle
            showPassword={showPassword}
            onTogglePassword={() => setShowPassword(!showPassword)}
            error={errors.password?.message}
          />
        )}
      />

      <AuthGradientButton
        testID="register-submit-button"
        label="Create Account"
        loadingLabel="Creating Account..."
        loading={mutation.isPending}
        onPress={handleSubmit(onSubmit)}
      />

      <AuthDivider />

      <OAuthButtonGroup
        onGooglePress={() => handleOAuth('google')}
        onApplePress={() => handleOAuth('apple')}
        onFacebookPress={() => handleOAuth('facebook')}
        loadingProvider={oauthLoading}
      />

      <AuthFooter
        text="Already have an account? "
        linkText="Sign in"
        onLinkPress={() => router.push('/auth/login')}
      />
    </AuthLayout>
  );
}

const styles = StyleSheet.create({
  nameRow: {
    flexDirection: 'row',
    gap: 12,
  },
  nameInput: {
    flex: 1,
  },
});
