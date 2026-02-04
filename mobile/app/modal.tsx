import { Link, useRouter } from "expo-router";
import { StyleSheet, View } from "react-native";
import { Appbar, Text, useTheme } from "react-native-paper";

export default function ModalScreen() {
  const router = useRouter();
  const { colors } = useTheme();

  return (
    <View style={{ flex: 1, backgroundColor: colors.background }}>
      <Appbar.Header style={{ backgroundColor: colors.primary }}>
        <Appbar.BackAction
          onPress={() => router.back()}
          color={colors.onPrimary}
        />
        <Appbar.Content
          title="Modal"
          titleStyle={{ color: colors.onPrimary }}
        />
      </Appbar.Header>

      <View style={styles.container}>
        <Text
          variant="headlineMedium"
          style={{ color: colors.onSurface, textAlign: "center" }}
        >
          This is a modal
        </Text>
        <Link href="/" dismissTo style={styles.link}>
          <Text variant="bodyLarge" style={{ color: colors.primary }}>
            Go to home screen
          </Text>
        </Link>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    alignItems: "center",
    justifyContent: "center",
    padding: 20,
  },
  link: {
    marginTop: 15,
    paddingVertical: 15,
  },
});
