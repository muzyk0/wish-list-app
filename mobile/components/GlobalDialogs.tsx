import { StyleSheet } from 'react-native';
import { Button, Dialog, Portal, Text } from 'react-native-paper';
import { useDialogStore } from '@/stores/dialogStore';

export function GlobalDialogs() {
  const {
    confirmDialog,
    confirmDialogVisible,
    confirmDialogLoading,
    hideConfirm,
    setConfirmLoading,
    messageDialog,
    messageDialogVisible,
    hideMessage,
  } = useDialogStore();

  const handleConfirm = async () => {
    if (confirmDialog?.onConfirm) {
      setConfirmLoading(true);
      try {
        await confirmDialog.onConfirm();
      } finally {
        setConfirmLoading(false);
        hideConfirm();
      }
    } else {
      hideConfirm();
    }
  };

  const handleMessagePress = async () => {
    if (messageDialog?.onPress) {
      await messageDialog.onPress();
    }
    hideMessage();
  };

  return (
    <>
      <Portal>
        <Dialog
          visible={confirmDialogVisible}
          onDismiss={hideConfirm}
          style={styles.dialog}
        >
          <Dialog.Title style={styles.title}>
            {confirmDialog?.title}
          </Dialog.Title>
          <Dialog.Content>
            <Text style={styles.content}>{confirmDialog?.message}</Text>
          </Dialog.Content>
          <Dialog.Actions>
            <Button
              onPress={hideConfirm}
              textColor="rgba(255,255,255,0.6)"
              disabled={confirmDialogLoading}
            >
              {confirmDialog?.cancelLabel || 'Cancel'}
            </Button>
            <Button
              onPress={handleConfirm}
              textColor={confirmDialog?.destructive ? '#FF6B6B' : '#FFD700'}
              loading={confirmDialogLoading}
              disabled={confirmDialogLoading}
            >
              {confirmDialog?.confirmLabel || 'Confirm'}
            </Button>
          </Dialog.Actions>
        </Dialog>
      </Portal>

      <Portal>
        <Dialog
          visible={messageDialogVisible}
          onDismiss={hideMessage}
          style={styles.dialog}
        >
          <Dialog.Title style={styles.title}>
            {messageDialog?.title}
          </Dialog.Title>
          <Dialog.Content>
            <Text style={styles.content}>{messageDialog?.message}</Text>
          </Dialog.Content>
          <Dialog.Actions>
            <Button onPress={handleMessagePress} textColor="#FFD700">
              {messageDialog?.buttonLabel || 'OK'}
            </Button>
          </Dialog.Actions>
        </Dialog>
      </Portal>
    </>
  );
}

const styles = StyleSheet.create({
  dialog: {
    backgroundColor: '#2d1b4e',
    borderRadius: 20,
  },
  title: {
    color: '#ffffff',
  },
  content: {
    color: 'rgba(255,255,255,0.7)',
  },
});
