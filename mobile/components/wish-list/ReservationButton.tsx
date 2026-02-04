import { useMutation } from '@tanstack/react-query';
import { useState } from 'react';
import { Alert, StyleSheet } from 'react-native';
import {
  Button,
  Dialog,
  Paragraph,
  Portal,
  Text,
  TextInput,
  useTheme,
} from 'react-native-paper';
import { apiClient } from '@/lib/api';

interface ReservationButtonProps {
  giftItemId: string;
  wishlistId: string;
  isReserved?: boolean;
  reservedByName?: string;
  onReservationSuccess?: () => void;
}

export function ReservationButton({
  giftItemId,
  wishlistId,
  isReserved = false,
  reservedByName,
  onReservationSuccess,
}: ReservationButtonProps) {
  const theme = useTheme();
  const [modalVisible, setModalVisible] = useState(false);
  const [guestName, setGuestName] = useState('');
  const [guestEmail, setGuestEmail] = useState('');

  const reservationMutation = useMutation({
    mutationFn: (data: { guest_name: string; guest_email: string }) =>
      apiClient.createReservation(wishlistId, giftItemId, {
        guest_name: data.guest_name,
        guest_email: data.guest_email,
      }),
    onSuccess: () => {
      Alert.alert('Success', 'Gift item reserved successfully!');
      setModalVisible(false);
      setGuestName('');
      setGuestEmail('');
      onReservationSuccess?.();
    },
    onError: (error: Error) => {
      Alert.alert('Error', error?.message || 'Failed to reserve gift item');
    },
  });

  const handleReservation = () => {
    if (!guestName.trim() || !guestEmail.trim()) {
      Alert.alert('Error', 'Please enter your name and email');
      return;
    }

    reservationMutation.mutate({
      guest_name: guestName.trim(),
      guest_email: guestEmail.trim(),
    });
  };

  if (isReserved) {
    return (
      <Button
        mode="contained-tonal"
        disabled
        style={styles.disabledButton}
        labelStyle={styles.disabledButtonText}
      >
        Reserved by {reservedByName || 'someone'}
      </Button>
    );
  }

  return (
    <>
      <Button
        mode="contained"
        onPress={() => setModalVisible(true)}
        style={styles.button}
        labelStyle={styles.buttonText}
      >
        Reserve this gift
      </Button>

      <Portal>
        <Dialog
          visible={modalVisible}
          onDismiss={() => setModalVisible(false)}
          style={{ backgroundColor: theme.colors.surface }}
        >
          <Dialog.Title>Reserve this gift</Dialog.Title>
          <Dialog.Content>
            <Paragraph style={styles.modalText}>
              Enter your details to reserve this gift item. This will prevent
              others from reserving the same gift.
            </Paragraph>

            <TextInput
              label="Your Name"
              value={guestName}
              onChangeText={setGuestName}
              style={styles.input}
              mode="outlined"
            />

            <TextInput
              label="Your Email"
              value={guestEmail}
              onChangeText={setGuestEmail}
              style={[styles.input, { marginTop: 16 }]}
              mode="outlined"
              keyboardType="email-address"
            />
          </Dialog.Content>
          <Dialog.Actions>
            <Button onPress={() => setModalVisible(false)}>
              <Text>Cancel</Text>
            </Button>
            <Button
              onPress={handleReservation}
              loading={reservationMutation.isPending}
              disabled={reservationMutation.isPending}
            >
              <Text>
                {reservationMutation.isPending
                  ? 'Reserving...'
                  : 'Reserve Gift'}
              </Text>
            </Button>
          </Dialog.Actions>
        </Dialog>
      </Portal>
    </>
  );
}

const styles = StyleSheet.create({
  button: {
    marginVertical: 8,
    marginHorizontal: 16,
  },
  disabledButton: {
    marginVertical: 8,
    marginHorizontal: 16,
    opacity: 0.6,
  },
  buttonText: {
    color: '#fff',
    fontWeight: '600',
  },
  disabledButtonText: {
    color: '#666',
    fontWeight: 'normal',
  },
  modalText: {
    marginBottom: 16,
    textAlign: 'center',
  },
  input: {
    width: '100%',
  },
});
