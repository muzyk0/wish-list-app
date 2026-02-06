import type React from 'react';
import { Pressable, StyleSheet, View } from 'react-native';
import { Text } from 'react-native-paper';

interface AuthFooterProps {
  text: string;
  linkText: string;
  onLinkPress: () => void;
}

export const AuthFooter: React.FC<AuthFooterProps> = ({
  text,
  linkText,
  onLinkPress,
}) => {
  return (
    <View style={styles.container}>
      <Text style={styles.text}>{text}</Text>
      <Pressable onPress={onLinkPress}>
        <Text style={styles.link}>{linkText}</Text>
      </Pressable>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    justifyContent: 'center',
    alignItems: 'center',
    marginTop: 32,
  },
  text: {
    color: 'rgba(255, 255, 255, 0.6)',
    fontSize: 15,
  },
  link: {
    color: '#FFD700',
    fontSize: 15,
    fontWeight: '600',
  },
});
