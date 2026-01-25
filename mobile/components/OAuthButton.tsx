import MaterialIcons from '@expo/vector-icons/MaterialIcons';
import type React from 'react';
import { StyleSheet, View } from 'react-native';
import { Button, Text, useTheme } from 'react-native-paper';
import { useThemeContext } from '@/contexts/ThemeContext';

interface OAuthButtonProps {
  provider: 'google' | 'facebook' | 'apple';
  onPress: () => void;
  loading?: boolean;
  disabled?: boolean;
}

const OAuthButton: React.FC<OAuthButtonProps> = ({
  provider,
  onPress,
  loading = false,
  disabled = false,
}) => {
  const { colors } = useTheme();
  const { isDark } = useThemeContext();

  // Define provider-specific styling and text
  const getProviderConfig = () => {
    switch (provider) {
      case 'google':
        return {
          text: 'Continue with Google',
          backgroundColor: isDark ? colors.surface : '#FFFFFF',
          textColor: isDark ? colors.onSurface : '#000000',
          borderColor: isDark ? colors.outline : '#CCCCCC',
          icon: 'mail' as const, // Using mail icon as Google icon
        };
      case 'facebook':
        return {
          text: 'Continue with Facebook',
          backgroundColor: isDark ? colors.surface : '#1877F2',
          textColor: isDark ? colors.onSurface : '#FFFFFF',
          borderColor: isDark ? colors.outline : '#1877F2',
          icon: 'thumb-up' as const, // Using thumb-up icon for Facebook
        };
      case 'apple':
        return {
          text: 'Continue with Apple',
          backgroundColor: isDark ? colors.surface : '#000000',
          textColor: isDark ? colors.onSurface : '#FFFFFF',
          borderColor: isDark ? colors.outline : '#000000',
          icon: 'phone-iphone' as const, // Using phone-iphone icon for Apple
        };
      default:
        return {
          text: 'Continue',
          backgroundColor: isDark ? colors.surface : '#FFFFFF',
          textColor: isDark ? colors.onSurface : '#000000',
          borderColor: isDark ? colors.outline : '#CCCCCC',
          icon: 'person' as const, // Using person icon as default
        };
    }
  };

  const config = getProviderConfig();

  return (
    <Button
      mode="outlined"
      onPress={onPress}
      loading={loading}
      disabled={disabled}
      style={[
        styles.button,
        {
          backgroundColor: config.backgroundColor,
          borderColor: config.borderColor,
        },
      ]}
      labelStyle={[styles.buttonLabel, { color: config.textColor }]}
      contentStyle={styles.buttonContent}
      uppercase={false}
    >
      <View style={styles.buttonContent}>
        <MaterialIcons name={config.icon} size={20} color={config.textColor} />
        <Text style={[styles.buttonText, { color: config.textColor }]}>
          {config.text}
        </Text>
      </View>
    </Button>
  );
};

const styles = StyleSheet.create({
  button: {
    borderRadius: 8,
    borderWidth: 1,
    paddingVertical: 6,
    marginVertical: 8,
  },
  buttonContent: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    gap: 10,
  },
  buttonLabel: {
    fontWeight: '600',
    fontSize: 16,
    textTransform: 'none',
  },
  buttonText: {
    fontWeight: '600',
    fontSize: 16,
  },
});

export default OAuthButton;
