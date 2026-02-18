import { create } from 'zustand';

export interface DialogOptions {
  title: string;
  message: string;
  confirmLabel?: string;
  cancelLabel?: string;
  onConfirm?: () => void | Promise<void>;
  destructive?: boolean;
}

export interface MessageDialogOptions {
  title: string;
  message: string;
  buttonLabel?: string;
  onPress?: () => void | Promise<void>;
}

interface DialogState {
  confirmDialog: DialogOptions | null;
  confirmDialogVisible: boolean;
  confirmDialogLoading: boolean;
  messageDialog: MessageDialogOptions | null;
  messageDialogVisible: boolean;
  showConfirm: (options: DialogOptions) => void;
  hideConfirm: () => void;
  setConfirmLoading: (loading: boolean) => void;
  showMessage: (options: MessageDialogOptions) => void;
  hideMessage: () => void;
}

export const useDialogStore = create<DialogState>((set) => ({
  confirmDialog: null,
  confirmDialogVisible: false,
  confirmDialogLoading: false,
  messageDialog: null,
  messageDialogVisible: false,

  showConfirm: (options) => {
    set({
      confirmDialog: options,
      confirmDialogVisible: true,
      confirmDialogLoading: false,
    });
  },

  hideConfirm: () => {
    set({ confirmDialogVisible: false, confirmDialogLoading: false });
    setTimeout(() => set({ confirmDialog: null }), 300);
  },

  setConfirmLoading: (loading) => set({ confirmDialogLoading: loading }),

  showMessage: (options) => {
    set({ messageDialog: options, messageDialogVisible: true });
  },

  hideMessage: () => {
    set({ messageDialogVisible: false });
    setTimeout(() => set({ messageDialog: null }), 300);
  },
}));

export const dialog = {
  confirm: (options: DialogOptions) => {
    useDialogStore.getState().showConfirm(options);
  },

  message: (options: MessageDialogOptions) => {
    useDialogStore.getState().showMessage(options);
  },

  success: (message: string, title = 'Success') => {
    useDialogStore.getState().showMessage({ title, message });
  },

  error: (message: string, title = 'Error') => {
    useDialogStore.getState().showMessage({ title, message });
  },

  confirmDelete: (itemName: string, onConfirm: () => void | Promise<void>) => {
    useDialogStore.getState().showConfirm({
      title: 'Confirm Delete',
      message: `Are you sure you want to delete ${itemName}? This action cannot be undone.`,
      confirmLabel: 'Delete',
      cancelLabel: 'Cancel',
      destructive: true,
      onConfirm,
    });
  },

  comingSoon: () => {
    useDialogStore.getState().showMessage({
      title: 'Coming Soon',
      message: 'This feature is coming soon!',
    });
  },

  hide: () => {
    useDialogStore.getState().hideConfirm();
    useDialogStore.getState().hideMessage();
  },
};
