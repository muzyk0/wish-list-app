import { MaterialCommunityIcons } from '@expo/vector-icons';
import type React from 'react';
import { Pressable, StyleSheet, View, type ViewStyle } from 'react-native';
import { HelperText, TextInput } from 'react-native-paper';

type IconName = keyof typeof MaterialCommunityIcons.glyphMap;

interface AuthInputProps {
  testID?: string;
  placeholder: string;
  value: string;
  onChangeText: (text: string) => void;
  onBlur?: () => void;
  icon?: IconName;
  secureTextEntry?: boolean;
  showPasswordToggle?: boolean;
  showPassword?: boolean;
  onTogglePassword?: () => void;
  keyboardType?: 'default' | 'email-address' | 'numeric' | 'phone-pad';
  autoCapitalize?: 'none' | 'sentences' | 'words' | 'characters';
  style?: ViewStyle;
  error?: string;
}

export const AuthInput: React.FC<AuthInputProps> = ({
  testID,
  placeholder,
  value,
  onChangeText,
  onBlur,
  icon,
  secureTextEntry = false,
  showPasswordToggle = false,
  showPassword = false,
  onTogglePassword,
  keyboardType = 'default',
  autoCapitalize = 'none',
  style,
  error,
}) => {
  return (
    <View style={style}>
      <View style={[styles.container, error && styles.containerError]}>
        {icon && (
          <MaterialCommunityIcons
            name={icon}
            size={22}
            color="rgba(255, 255, 255, 0.5)"
            style={styles.icon}
          />
        )}
        <TextInput
          testID={testID}
          placeholder={placeholder}
          placeholderTextColor="rgba(255, 255, 255, 0.4)"
          value={value}
          onChangeText={onChangeText}
          onBlur={onBlur}
          secureTextEntry={secureTextEntry && !showPassword}
          keyboardType={keyboardType}
          autoCapitalize={autoCapitalize}
          style={styles.input}
          textColor="#ffffff"
          underlineColor="transparent"
          activeUnderlineColor="#FFD700"
          contentStyle={styles.inputContent}
          error={!!error}
        />
        {showPasswordToggle && onTogglePassword && (
          <Pressable onPress={onTogglePassword} style={styles.eyeIcon}>
            <MaterialCommunityIcons
              name={showPassword ? 'eye-off' : 'eye'}
              size={22}
              color="rgba(255, 255, 255, 0.5)"
            />
          </Pressable>
        )}
      </View>
      {error && (
        <HelperText type="error" visible={!!error} style={styles.errorText}>
          {error}
        </HelperText>
      )}
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: 'rgba(255, 255, 255, 0.08)',
    borderRadius: 16,
    marginBottom: 4,
    paddingHorizontal: 16,
    borderWidth: 1,
    borderColor: 'rgba(255, 255, 255, 0.1)',
  },
  containerError: {
    borderColor: 'rgba(255, 82, 82, 0.5)',
  },
  icon: {
    marginRight: 12,
  },
  input: {
    flex: 1,
    backgroundColor: 'transparent',
    fontSize: 16,
    height: 56,
  },
  inputContent: {
    paddingLeft: 0,
  },
  eyeIcon: {
    padding: 8,
  },
  errorText: {
    marginTop: -8,
    marginBottom: 8,
    color: '#ff5252',
  },
});
