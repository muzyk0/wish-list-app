import type React from "react";
import { StyleSheet, Text, View } from "react-native";
import { useTheme } from "react-native-paper";

interface BadgeProps {
  children: React.ReactNode;
  variant?: "default" | "secondary" | "destructive" | "outline";
  className?: string;
}

export function Badge({ children, variant = "default" }: BadgeProps) {
  const theme = useTheme();

  const getVariantStyles = () => {
    switch (variant) {
      case "secondary":
        return {
          backgroundColor: theme.colors.surfaceVariant,
          borderColor: theme.colors.outline,
        };
      case "destructive":
        return {
          backgroundColor: "#FEE2E2", // red-100 equivalent
          borderColor: "#EF4444", // red-500 equivalent
        };
      case "outline":
        return {
          backgroundColor: "transparent",
          borderColor: theme.colors.outline,
        };
      default:
        return {
          backgroundColor: theme.colors.primaryContainer,
          borderColor: theme.colors.primary,
        };
    }
  };

  return (
    <View style={[styles.badge, getVariantStyles()]}>
      <Text style={styles.badgeText}>{children}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  badge: {
    paddingHorizontal: 8,
    paddingVertical: 4,
    borderRadius: 12,
    borderWidth: 1,
    alignSelf: "flex-start",
  },
  badgeText: {
    fontSize: 12,
    fontWeight: "600",
    textAlign: "center",
  },
});
