import { View } from "react-native";
import { Text } from "react-native-paper";
import { MyReservations } from "@/components/wish-list/MyReservations";

export default function ReservationsScreen() {
  return (
    <View style={{ flex: 1 }}>
      <Text style={{ padding: 16 }}>My Reservations</Text>

      <MyReservations />
    </View>
  );
}
