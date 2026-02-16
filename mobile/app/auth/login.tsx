import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation } from '@tanstack/react-query';
import { useRouter } from 'expo-router';
import { useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { Alert } from 'react-native';
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
import { loginUser } from '@/lib/api';

// Zod schema for login form validation
const loginSchema = z.object({
  email: z.string().min(1, 'Email is required').email('Invalid email address'),
  password: z
    .string()
    .min(6, 'Password must be at least 6 characters')
    .max(100, 'Password is too long'),
});

type LoginFormData = z.infer<typeof loginSchema>;

export default function LoginScreen() {
  const [showPassword, setShowPassword] = useState(false);
  const router = useRouter();
  const { oauthLoading, handleOAuth } = useOAuthHandler();

  const {
    control,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      email: '',
      password: '',
    },
  });

  const mutation = useMutation({
    mutationFn: ({ email, password }: { email: string; password: string }) =>
      loginUser({ email, password }),
    onSuccess: async () => {
      router.replace('/(tabs)');
    },
    onError: (error: Error) => {
      Alert.alert('Error', error.message || 'Login failed. Please try again.');
    },
  });

  const onSubmit = (data: LoginFormData) => {
    mutation.mutate(data);
  };

  return (
    <AuthLayout title="Wish List" subtitle="Welcome back!">
      <Controller
        control={control}
        name="email"
        render={({ field: { onChange, onBlur, value } }) => (
          <AuthInput
            testID="login-email-input"
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
            testID="login-password-input"
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
        testID="login-submit-button"
        label="Sign In"
        loadingLabel="Signing in..."
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
        text="Don't have an account? "
        linkText="Create one"
        onLinkPress={() => router.push('/auth/register')}
      />
    </AuthLayout>
  );
}
