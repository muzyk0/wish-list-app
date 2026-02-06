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
import { registerUser } from '@/lib/api';

export default function RegisterScreen() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [firstName, setFirstName] = useState('');
  const [lastName, setLastName] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const router = useRouter();
  const { oauthLoading, handleOAuth } = useOAuthHandler();

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

  const handleRegister = () => {
    if (!email || !password) {
      Alert.alert('Error', 'Please fill in all required fields.');
      return;
    }
    mutation.mutate({ email, password, firstName, lastName });
  };

  return (
    <AuthLayout title="Wish List" subtitle="Create Your Account">
      <View style={styles.nameRow}>
        <AuthInput
          testID="register-firstname-input"
          placeholder="First Name"
          value={firstName}
          onChangeText={setFirstName}
          icon="account-outline"
          style={styles.nameInput}
        />

        <AuthInput
          testID="register-lastname-input"
          placeholder="Last Name"
          value={lastName}
          onChangeText={setLastName}
          style={styles.nameInput}
        />
      </View>

      <AuthInput
        testID="register-email-input"
        placeholder="Email"
        value={email}
        onChangeText={setEmail}
        icon="email-outline"
        keyboardType="email-address"
      />

      <AuthInput
        testID="register-password-input"
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
        testID="register-submit-button"
        label="Create Account"
        loadingLabel="Creating Account..."
        loading={mutation.isPending}
        onPress={handleRegister}
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
