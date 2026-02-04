import { StyleSheet, View } from "react-native";
import { Text, useTheme } from "react-native-paper";
import { MyReservations } from "../../components/wish-list/MyReservations";

export default function ReservationsScreen() {
  const theme = useTheme();

  return (
    <View
      style={[styles.container, { backgroundColor: theme.colors.background }]}
    >
      <Text variant="headlineLarge" style={styles.title}>
        My Reservations
      </Text>

      <MyReservations />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  title: {
    padding: 16,
    textAlign: "center",
  },
});
