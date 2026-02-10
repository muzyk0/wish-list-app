import { useMutation } from '@tanstack/react-query';
import { useRouter } from 'expo-router';
import { useState } from 'react';
import { Alert, StyleSheet, View } from 'react-native';
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

export default function LoginScreen() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const router = useRouter();
  const { oauthLoading, handleOAuth } = useOAuthHandler();

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

  const handleLogin = () => {
    if (!email || !password) {
      Alert.alert('Error', 'Please fill in all required fields.');
      return;
    }
    mutation.mutate({ email, password });
  };

  return (
    <AuthLayout title="Wish List" subtitle="Welcome back!">
      <AuthInput
        testID="login-email-input"
        placeholder="Email"
        value={email}
        onChangeText={setEmail}
        icon="email-outline"
        keyboardType="email-address"
      />

      <AuthInput
        testID="login-password-input"
        placeholder="Password"
        value={password}
        onChangeText={setPassword}
        icon="lock-outline"
        secureTextEntry
        showPasswordToggle
        showPassword={showPassword}
        onTogglePassword={() => setShowPassword(!showPassword)}
      />

      <AuthGradientButton
        testID="login-submit-button"
        label="Sign In"
        loadingLabel="Signing in..."
        loading={mutation.isPending}
        onPress={handleLogin}
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
