import MaterialCommunityIcons from '@expo/vector-icons/MaterialCommunityIcons';
import type React from 'react';
import { ActivityIndicator, Pressable, StyleSheet, View } from 'react-native';

type OAuthProvider = 'google' | 'apple' | 'facebook';

interface OAuthButtonProps {
  provider: OAuthProvider;
  onPress: () => void;
  loading?: boolean;
  disabled?: boolean;
}

const PROVIDER_ICONS: Record<
  OAuthProvider,
  keyof typeof MaterialCommunityIcons.glyphMap
> = {
  google: 'google',
  apple: 'apple',
  facebook: 'facebook',
};

export const OAuthButton: React.FC<OAuthButtonProps> = ({
  provider,
  onPress,
  loading = false,
  disabled = false,
}) => {
  const iconName = PROVIDER_ICONS[provider];

  return (
    <Pressable
      style={({ pressed }) => [
        styles.button,
        pressed && styles.buttonPressed,
        disabled && styles.buttonDisabled,
      ]}
      onPress={onPress}
      disabled={disabled || loading}
    >
      {loading ? (
        <ActivityIndicator size="small" color="#ffffff" />
      ) : (
        <MaterialCommunityIcons name={iconName} size={24} color="#ffffff" />
      )}
    </Pressable>
  );
};

interface OAuthButtonGroupProps {
  onGooglePress: () => void;
  onApplePress: () => void;
  onFacebookPress: () => void;
  loadingProvider: OAuthProvider | null;
  disabled?: boolean;
}

export const OAuthButtonGroup: React.FC<OAuthButtonGroupProps> = ({
  onGooglePress,
  onApplePress,
  onFacebookPress,
  loadingProvider,
  disabled = false,
}) => {
  const isDisabled = disabled || loadingProvider !== null;

  return (
    <View style={styles.container}>
      <OAuthButton
        provider="google"
        onPress={onGooglePress}
        loading={loadingProvider === 'google'}
        disabled={isDisabled}
      />
      <OAuthButton
        provider="apple"
        onPress={onApplePress}
        loading={loadingProvider === 'apple'}
        disabled={isDisabled}
      />
      <OAuthButton
        provider="facebook"
        onPress={onFacebookPress}
        loading={loadingProvider === 'facebook'}
        disabled={isDisabled}
      />
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    justifyContent: 'center',
    gap: 20,
  },
  button: {
    width: 60,
    height: 60,
    borderRadius: 30,
    backgroundColor: 'rgba(255, 255, 255, 0.1)',
    justifyContent: 'center',
    alignItems: 'center',
    borderWidth: 1,
    borderColor: 'rgba(255, 255, 255, 0.15)',
  },
  buttonPressed: {
    backgroundColor: 'rgba(255, 255, 255, 0.2)',
  },
  buttonDisabled: {
    opacity: 0.5,
  },
});

export default OAuthButton;
