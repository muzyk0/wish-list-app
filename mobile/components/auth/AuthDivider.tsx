import type React from 'react';
import { StyleSheet, View } from 'react-native';
import { Text } from 'react-native-paper';

interface AuthDividerProps {
  text?: string;
}

export const AuthDivider: React.FC<AuthDividerProps> = ({
  text = 'or continue with',
}) => {
  return (
    <View style={styles.container}>
      <View style={styles.line} />
      <Text style={styles.text}>{text}</Text>
      <View style={styles.line} />
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    alignItems: 'center',
    marginVertical: 28,
  },
  line: {
    flex: 1,
    height: 1,
    backgroundColor: 'rgba(255, 255, 255, 0.15)',
  },
  text: {
    color: 'rgba(255, 255, 255, 0.5)',
    fontSize: 13,
    marginHorizontal: 16,
  },
});
