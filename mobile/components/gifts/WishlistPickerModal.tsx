import { MaterialCommunityIcons } from '@expo/vector-icons';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import * as Haptics from 'expo-haptics';
import { LinearGradient } from 'expo-linear-gradient';
import {
  ActivityIndicator,
  Modal,
  Pressable,
  ScrollView,
  StyleSheet,
  View,
} from 'react-native';
import { Text } from 'react-native-paper';
import { apiClient } from '@/lib/api';
import type { WishList } from '@/lib/api/types';
import { dialog } from '@/stores/dialogStore';

// ─── Color tokens ────────────────────────────────────────────────────
const C = {
  bg0: '#060411',
  bg1: '#0d0920',
  bg2: '#16112e',
  gold: '#E2B96C',
  goldBright: '#F5D38A',
  goldDim: 'rgba(226, 185, 108, 0.14)',
  goldBorder: 'rgba(226, 185, 108, 0.22)',
  surface: 'rgba(255, 255, 255, 0.04)',
  surfaceMid: 'rgba(255, 255, 255, 0.07)',
  border: 'rgba(255, 255, 255, 0.07)',
  white: '#EEE8FF',
  muted: 'rgba(238, 232, 255, 0.45)',
  faint: 'rgba(238, 232, 255, 0.18)',
  green: '#5EEAD4',
  greenBg: 'rgba(94, 234, 212, 0.13)',
} as const;

interface WishlistPickerModalProps {
  visible: boolean;
  itemId: string;
  itemName: string;
  /** IDs of wishlists this item is already attached to */
  attachedWishlistIds: string[];
  onClose: () => void;
  onSuccess?: (wishlistId: string) => void;
}

export default function WishlistPickerModal({
  visible,
  itemId,
  itemName,
  attachedWishlistIds,
  onClose,
  onSuccess,
}: WishlistPickerModalProps) {
  const queryClient = useQueryClient();

  const {
    data: wishlists = [],
    isLoading,
    error,
  } = useQuery({
    queryKey: ['wishlists'],
    queryFn: () => apiClient.getWishLists(),
    enabled: visible,
  });

  const attachMutation = useMutation({
    mutationFn: ({ wishlistId }: { wishlistId: string }) =>
      apiClient.attachGiftItemToWishlist(wishlistId, itemId),
    onSuccess: (_, { wishlistId }) => {
      Haptics.notificationAsync(Haptics.NotificationFeedbackType.Success);
      dialog.success(`"${itemName}" attached to wishlist!`);
      queryClient.invalidateQueries({ queryKey: ['giftItems'] });
      queryClient.invalidateQueries({ queryKey: ['standaloneGiftItems'] });
      queryClient.invalidateQueries({
        queryKey: ['userGiftItems'],
      });
      onSuccess?.(wishlistId);
      onClose();
    },
    onError: (err: Error) => {
      dialog.error(err.message || 'Failed to attach gift. Please try again.');
    },
  });

  const handleAttach = (wishlist: WishList) => {
    if (attachMutation.isPending) return;
    Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Light);
    attachMutation.mutate({ wishlistId: wishlist.id });
  };

  return (
    <Modal
      visible={visible}
      transparent
      animationType="slide"
      onRequestClose={onClose}
    >
      <View style={s.overlay}>
        <Pressable style={s.backdrop} onPress={onClose} />

        <View style={s.sheet}>
          <LinearGradient
            colors={[C.bg2, C.bg1, C.bg0]}
            style={StyleSheet.absoluteFill}
          />

          {/* Handle */}
          <View style={s.handle} />

          {/* Header */}
          <View style={s.header}>
            <View>
              <Text style={s.headerTitle}>Attach to Wishlist</Text>
              <Text style={s.headerSubtitle} numberOfLines={1}>
                "{itemName}"
              </Text>
            </View>
            <Pressable onPress={onClose} style={s.closeBtn}>
              <MaterialCommunityIcons name="close" size={20} color={C.white} />
            </Pressable>
          </View>

          {/* Content */}
          {isLoading ? (
            <View style={s.center}>
              <ActivityIndicator color={C.gold} />
              <Text style={s.loadingText}>Loading wishlists…</Text>
            </View>
          ) : error ? (
            <View style={s.center}>
              <MaterialCommunityIcons
                name="alert-circle-outline"
                size={40}
                color="#F87171"
              />
              <Text style={s.errorText}>Failed to load wishlists</Text>
            </View>
          ) : wishlists.length === 0 ? (
            <View style={s.center}>
              <MaterialCommunityIcons
                name="playlist-remove"
                size={48}
                color={C.muted}
              />
              <Text style={s.emptyTitle}>No wishlists yet</Text>
              <Text style={s.emptyText}>Create a wishlist first</Text>
            </View>
          ) : (
            <ScrollView
              style={s.scroll}
              contentContainerStyle={s.scrollContent}
              showsVerticalScrollIndicator={false}
            >
              {wishlists.map((wl) => {
                const isAttached = attachedWishlistIds.includes(wl.id);
                const isAttaching =
                  attachMutation.isPending &&
                  attachMutation.variables?.wishlistId === wl.id;

                return (
                  <Pressable
                    key={wl.id}
                    onPress={() => !isAttached && handleAttach(wl)}
                    disabled={isAttached || attachMutation.isPending}
                    style={[s.row, isAttached && s.rowAttached]}
                  >
                    {/* Icon */}
                    <View style={[s.rowIcon, isAttached && s.rowIconAttached]}>
                      {isAttaching ? (
                        <ActivityIndicator size={16} color={C.gold} />
                      ) : (
                        <MaterialCommunityIcons
                          name={isAttached ? 'check-circle' : 'gift-outline'}
                          size={20}
                          color={isAttached ? C.green : C.gold}
                        />
                      )}
                    </View>

                    {/* Info */}
                    <View style={s.rowInfo}>
                      <Text
                        style={[s.rowTitle, isAttached && s.rowTitleMuted]}
                        numberOfLines={1}
                      >
                        {wl.title}
                      </Text>
                      <View style={s.rowMeta}>
                        {wl.occasion ? (
                          <Text style={s.rowOccasion}>{wl.occasion}</Text>
                        ) : null}
                        {wl.item_count != null && (
                          <Text style={s.rowCount}>
                            {wl.item_count} gift{wl.item_count !== 1 ? 's' : ''}
                          </Text>
                        )}
                      </View>
                    </View>

                    {/* State */}
                    {isAttached ? (
                      <View style={s.alreadyBadge}>
                        <Text style={s.alreadyText}>Added</Text>
                      </View>
                    ) : (
                      <MaterialCommunityIcons
                        name="chevron-right"
                        size={20}
                        color={C.faint}
                      />
                    )}
                  </Pressable>
                );
              })}
            </ScrollView>
          )}
        </View>
      </View>
    </Modal>
  );
}

const s = StyleSheet.create({
  overlay: {
    flex: 1,
    justifyContent: 'flex-end',
  },
  backdrop: {
    ...StyleSheet.absoluteFillObject,
    backgroundColor: 'rgba(0,0,0,0.6)',
  },
  sheet: {
    maxHeight: '75%',
    borderTopLeftRadius: 28,
    borderTopRightRadius: 28,
    overflow: 'hidden',
    paddingBottom: 32,
  },
  handle: {
    width: 40,
    height: 4,
    borderRadius: 2,
    backgroundColor: 'rgba(255,255,255,0.2)',
    alignSelf: 'center',
    marginTop: 12,
    marginBottom: 4,
  },
  header: {
    flexDirection: 'row',
    alignItems: 'flex-start',
    justifyContent: 'space-between',
    paddingHorizontal: 20,
    paddingVertical: 16,
    borderBottomWidth: 1,
    borderBottomColor: 'rgba(255,255,255,0.07)',
  },
  headerTitle: {
    fontSize: 18,
    fontWeight: '700',
    color: '#EEE8FF',
    lineHeight: 24,
  },
  headerSubtitle: {
    fontSize: 13,
    color: 'rgba(238, 232, 255, 0.45)',
    marginTop: 2,
    maxWidth: 260,
  },
  closeBtn: {
    width: 36,
    height: 36,
    borderRadius: 18,
    backgroundColor: 'rgba(255,255,255,0.07)',
    alignItems: 'center',
    justifyContent: 'center',
  },
  scroll: {
    flex: 1,
  },
  scrollContent: {
    paddingHorizontal: 16,
    paddingTop: 12,
    paddingBottom: 8,
    gap: 8,
  },
  center: {
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 48,
    gap: 12,
  },
  loadingText: {
    fontSize: 14,
    color: 'rgba(238,232,255,0.45)',
  },
  errorText: {
    fontSize: 15,
    color: '#F87171',
    fontWeight: '600',
  },
  emptyTitle: {
    fontSize: 17,
    fontWeight: '700',
    color: '#EEE8FF',
  },
  emptyText: {
    fontSize: 13,
    color: 'rgba(238,232,255,0.45)',
  },
  row: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 14,
    paddingVertical: 14,
    paddingHorizontal: 14,
    borderRadius: 16,
    backgroundColor: 'rgba(255,255,255,0.04)',
    borderWidth: 1,
    borderColor: 'rgba(255,255,255,0.07)',
  },
  rowAttached: {
    borderColor: 'rgba(94, 234, 212, 0.2)',
    backgroundColor: 'rgba(94, 234, 212, 0.06)',
  },
  rowIcon: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: 'rgba(226, 185, 108, 0.12)',
    alignItems: 'center',
    justifyContent: 'center',
  },
  rowIconAttached: {
    backgroundColor: 'rgba(94, 234, 212, 0.13)',
  },
  rowInfo: {
    flex: 1,
    gap: 3,
  },
  rowTitle: {
    fontSize: 15,
    fontWeight: '600',
    color: '#EEE8FF',
  },
  rowTitleMuted: {
    color: 'rgba(238, 232, 255, 0.5)',
  },
  rowMeta: {
    flexDirection: 'row',
    gap: 10,
    alignItems: 'center',
  },
  rowOccasion: {
    fontSize: 12,
    color: 'rgba(226, 185, 108, 0.8)',
  },
  rowCount: {
    fontSize: 12,
    color: 'rgba(238, 232, 255, 0.4)',
  },
  alreadyBadge: {
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: 10,
    backgroundColor: 'rgba(94, 234, 212, 0.15)',
  },
  alreadyText: {
    fontSize: 12,
    fontWeight: '600',
    color: '#5EEAD4',
  },
});
