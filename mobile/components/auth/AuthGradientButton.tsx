import { MaterialCommunityIcons } from '@expo/vector-icons';
import { LinearGradient } from 'expo-linear-gradient';
import type React from 'react';
import { Pressable, StyleSheet } from 'react-native';
import { Text } from 'react-native-paper';

interface AuthGradientButtonProps {
  testID?: string;
  label: string;
  loadingLabel?: string;
  loading?: boolean;
  disabled?: boolean;
  onPress: () => void;
  showArrow?: boolean;
}

export const AuthGradientButton: React.FC<AuthGradientButtonProps> = ({
  testID,
  label,
  loadingLabel,
  loading = false,
  disabled = false,
  onPress,
  showArrow = true,
}) => {
  const displayLabel = loading && loadingLabel ? loadingLabel : label;

  return (
    <Pressable testID={testID} onPress={onPress} disabled={disabled || loading}>
      <LinearGradient
        colors={['#FFD700', '#FFA500']}
        start={{ x: 0, y: 0 }}
        end={{ x: 1, y: 0 }}
        style={styles.button}
      >
        <Text style={styles.buttonText}>{displayLabel}</Text>
        {showArrow && !loading && (
          <MaterialCommunityIcons
            name="arrow-right"
            size={24}
            color="#000000"
          />
        )}
      </LinearGradient>
    </Pressable>
  );
};

const styles = StyleSheet.create({
  button: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 18,
    borderRadius: 16,
    gap: 8,
    marginTop: 8,
  },
  buttonText: {
    fontSize: 18,
    fontWeight: '700',
    color: '#000000',
  },
});
