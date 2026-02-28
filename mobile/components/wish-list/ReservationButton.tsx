import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation } from '@tanstack/react-query';
import { useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { StyleSheet, View } from 'react-native';
import {
  Button,
  Dialog,
  HelperText,
  Paragraph,
  Portal,
  Text,
  TextInput,
  useTheme,
} from 'react-native-paper';
import { z } from 'zod';
import { apiClient } from '@/lib/api';
import { dialog } from '@/stores/dialogStore';

// Zod schema for guest reservation form validation
const guestReservationSchema = z.object({
  guestName: z
    .string()
    .trim()
    .min(1, 'Name is required')
    .max(255, 'Name must be less than 255 characters'),
  guestEmail: z
    .string()
    .optional()
    .refine((value) => {
      if (!value) return true;
      const trimmed = value.trim();
      if (!trimmed) return true;
      return z.string().email().safeParse(trimmed).success;
    }, 'Invalid email address'),
});

type GuestReservationFormData = z.infer<typeof guestReservationSchema>;

interface ReservationButtonProps {
  giftItemId: string;
  wishlistId: string;
  isReserved?: boolean;
  onReservationSuccess?: () => void;
}

export function ReservationButton({
  giftItemId,
  wishlistId,
  isReserved = false,
  onReservationSuccess,
}: ReservationButtonProps) {
  const theme = useTheme();
  const [modalVisible, setModalVisible] = useState(false);

  const {
    control,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<GuestReservationFormData>({
    resolver: zodResolver(guestReservationSchema),
    defaultValues: {
      guestName: '',
      guestEmail: '',
    },
  });

  const reservationMutation = useMutation({
    mutationFn: (data: GuestReservationFormData) =>
      apiClient.createReservation(wishlistId, giftItemId, {
        guest_name: data.guestName.trim(),
        guest_email: data.guestEmail?.trim() || undefined,
      }),
    onSuccess: () => {
      dialog.success('Gift item reserved successfully!');
      setModalVisible(false);
      reset();
      onReservationSuccess?.();
    },
    onError: (error: Error) => {
      dialog.error(error?.message || 'Failed to reserve gift item');
    },
  });

  const onSubmit = (data: GuestReservationFormData) => {
    reservationMutation.mutate(data);
  };

  if (isReserved) {
    return (
      <Button
        mode="contained-tonal"
        disabled
        style={styles.disabledButton}
        labelStyle={styles.disabledButtonText}
      >
        Reserved
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
              Enter your name to reserve this gift item. Email is optional.
            </Paragraph>

            <Controller
              control={control}
              name="guestName"
              render={({ field: { onChange, onBlur, value } }) => (
                <View>
                  <TextInput
                    label="Your Name"
                    value={value}
                    onChangeText={onChange}
                    onBlur={onBlur}
                    style={styles.input}
                    mode="outlined"
                    error={!!errors.guestName}
                  />
                  {errors.guestName && (
                    <HelperText type="error" visible={!!errors.guestName}>
                      {errors.guestName.message}
                    </HelperText>
                  )}
                </View>
              )}
            />

            <Controller
              control={control}
              name="guestEmail"
              render={({ field: { onChange, onBlur, value } }) => (
                <View>
                  <TextInput
                    label="Your Email (optional)"
                    value={value}
                    onChangeText={onChange}
                    onBlur={onBlur}
                    style={[styles.input, { marginTop: 16 }]}
                    mode="outlined"
                    keyboardType="email-address"
                    error={!!errors.guestEmail}
                  />
                  {errors.guestEmail && (
                    <HelperText type="error" visible={!!errors.guestEmail}>
                      {errors.guestEmail.message}
                    </HelperText>
                  )}
                  <HelperText type="info" visible>
                    Optional: if you sign up later, we can link these
                    reservations to your account.
                  </HelperText>
                </View>
              )}
            />
          </Dialog.Content>
          <Dialog.Actions>
            <Button onPress={() => setModalVisible(false)}>
              <Text>Cancel</Text>
            </Button>
            <Button
              onPress={handleSubmit(onSubmit)}
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
