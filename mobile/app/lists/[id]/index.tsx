import { MaterialCommunityIcons } from '@expo/vector-icons';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import * as Haptics from 'expo-haptics';
import { Image } from 'expo-image';
import { LinearGradient } from 'expo-linear-gradient';
import { useLocalSearchParams, useRouter } from 'expo-router';
import { useEffect, useState } from 'react';
import {
  Keyboard,
  Linking,
  Modal,
  Pressable,
  RefreshControl,
  ScrollView,
  Share,
  StyleSheet,
  TextInput,
  View,
} from 'react-native';
import { ActivityIndicator, Text } from 'react-native-paper';
import { apiClient } from '@/lib/api';
import type { WishlistItem } from '@/lib/api/types';
import { dialog } from '@/stores/dialogStore';

const WEB_DOMAIN = process.env.EXPO_PUBLIC_WEB_DOMAIN ?? 'wishlist.com';

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
  amber: '#FCD34D',
  amberBg: 'rgba(252, 211, 77, 0.13)',
  violet: '#C4B5FD',
  violetBg: 'rgba(196, 181, 253, 0.13)',
  slate: '#94A3B8',
  slateBg: 'rgba(148, 163, 184, 0.13)',
} as const;

// ─── Placeholder gradient palette ───────────────────────────────────
const GRADIENTS: [string, string][] = [
  ['#FF7BAC', '#C13584'],
  ['#4FD1C5', '#2B6CB0'],
  ['#81E6D9', '#2F855A'],
  ['#B794F4', '#553C9A'],
  ['#FBD38D', '#DD6B20'],
  ['#FC8181', '#C53030'],
];

function gradientForItem(
  id: string | undefined,
  idx: number,
): [string, string] {
  if (!id) return GRADIENTS[idx % GRADIENTS.length];
  let h = 0;
  for (let i = 0; i < id.length; i++) h = (h * 31 + id.charCodeAt(i)) | 0;
  return GRADIENTS[Math.abs(h) % GRADIENTS.length];
}

// ─── Gift Item Card ──────────────────────────────────────────────────
const GiftItemCard = ({
  item,
  index,
  onMarkReserved,
  onEdit,
}: {
  item: WishlistItem;
  index: number;
  onMarkReserved: (item: WishlistItem) => void;
  onEdit: (item: WishlistItem) => void;
}) => {
  const isPurchased = !!item.is_purchased;
  const isManuallyReserved = !!item.is_manually_reserved;
  const isReserved = (!!item.is_reserved && !isPurchased) || isManuallyReserved;
  const isArchived = !!item.is_archived;
  const isAvailable = !isReserved && !isPurchased && !isArchived;

  const status = isPurchased
    ? {
        color: C.violet,
        bg: C.violetBg,
        label: 'Purchased',
        icon: 'check-circle' as const,
      }
    : isReserved
      ? {
          color: C.amber,
          bg: C.amberBg,
          label: 'Reserved',
          icon: 'lock' as const,
        }
      : isArchived
        ? {
            color: C.slate,
            bg: C.slateBg,
            label: 'Archived',
            icon: 'archive' as const,
          }
        : {
            color: C.green,
            bg: C.greenBg,
            label: 'Available',
            icon: 'gift-open-outline' as const,
          };

  const grad = gradientForItem(item.id, index);
  const manualReservedByName = item.manual_reserved_by_name;

  return (
    <View style={card.wrapper}>
      {/* ── Image strip ── */}
      <View style={card.imageWrap}>
        {item.image_url ? (
          <Image
            source={{ uri: item.image_url }}
            style={card.image}
            contentFit="cover"
            transition={300}
          />
        ) : (
          <LinearGradient
            colors={grad}
            style={card.imagePlaceholder}
            start={{ x: 0, y: 0 }}
            end={{ x: 1, y: 1 }}
          >
            <MaterialCommunityIcons
              name="gift-outline"
              size={44}
              color="rgba(255,255,255,0.4)"
            />
          </LinearGradient>
        )}

        {/* Darken overlay when not available */}
        {!isAvailable && <View style={card.imageScrim} />}

        {/* Priority chip — top-left */}
        {item.priority !== undefined &&
          item.priority !== null &&
          item.priority > 0 && (
            <View style={card.priorityChip}>
              <MaterialCommunityIcons name="star" size={10} color="#FFD166" />
              <Text style={card.priorityChipText}>{item.priority}</Text>
            </View>
          )}

        {/* Status badge — top-right */}
        <View style={[card.statusChip, { backgroundColor: status.bg }]}>
          <MaterialCommunityIcons
            name={status.icon}
            size={12}
            color={status.color}
          />
          <Text style={[card.statusChipText, { color: status.color }]}>
            {status.label}
          </Text>
        </View>
      </View>

      {/* ── Content ── */}
      <View style={card.body}>
        {/* Title + price */}
        <View style={card.titleRow}>
          <Text
            style={[card.title, !isAvailable && card.titleFaded]}
            numberOfLines={2}
          >
            {item.title ?? 'Unnamed Gift'}
          </Text>
          {item.price != null && (
            <Text style={card.price}>${item.price.toFixed(0)}</Text>
          )}
        </View>

        {/* Description / notes */}
        {item.description || item.notes ? (
          <Text style={card.subtitle} numberOfLines={2}>
            {item.description || item.notes}
          </Text>
        ) : null}

        {/* Manual reservation info */}
        {isManuallyReserved && manualReservedByName ? (
          <View style={card.manualReservationBadge}>
            <MaterialCommunityIcons
              name="account-check"
              size={13}
              color={C.amber}
            />
            <Text style={card.manualReservationText} numberOfLines={1}>
              {manualReservedByName}
            </Text>
          </View>
        ) : null}

        {/* Action bar */}
        <View style={card.actions}>
          {/* Link pill */}
          {item.link ? (
            <Pressable
              onPress={() => Linking.openURL(item.link ?? '')}
              style={card.linkPill}
            >
              <MaterialCommunityIcons
                name="open-in-new"
                size={12}
                color={C.gold}
              />
              <Text style={card.linkPillText}>Open</Text>
            </Pressable>
          ) : null}

          <View style={{ flex: 1 }} />

          {/* Edit */}
          <Pressable
            onPress={() => onEdit(item)}
            style={card.iconBtn}
            hitSlop={8}
          >
            <MaterialCommunityIcons
              name="pencil-outline"
              size={15}
              color={C.faint}
            />
          </Pressable>

          {/* Mark as Reserved (owner action — available items only) */}
          {isAvailable && (
            <Pressable
              onPress={() => {
                Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Light);
                onMarkReserved(item);
              }}
              style={card.reserveBtn}
            >
              <LinearGradient
                colors={[C.goldBright, '#C48A3A']}
                style={card.reserveBtnGrad}
                start={{ x: 0, y: 0 }}
                end={{ x: 1, y: 0 }}
              >
                <MaterialCommunityIcons
                  name="account-check"
                  size={13}
                  color="#1a0f05"
                />
                <Text style={card.reserveBtnText}>Mark Reserved</Text>
              </LinearGradient>
            </Pressable>
          )}
        </View>
      </View>
    </View>
  );
};

// ─── Mark Reserved Modal ─────────────────────────────────────────────
const MarkReservedModal = ({
  visible,
  itemTitle,
  loading,
  onConfirm,
  onCancel,
}: {
  visible: boolean;
  itemTitle: string;
  loading: boolean;
  onConfirm: (name: string, note: string) => void;
  onCancel: () => void;
}) => {
  const [name, setName] = useState('');
  const [note, setNote] = useState('');

  useEffect(() => {
    if (visible) {
      setName('');
      setNote('');
    }
  }, [visible]);

  const handleConfirm = () => {
    if (!name.trim()) return;
    onConfirm(name.trim(), note.trim());
  };

  const handleCancel = () => {
    setName('');
    setNote('');
    onCancel();
  };

  return (
    <Modal
      visible={visible}
      transparent
      animationType="slide"
      onRequestClose={handleCancel}
    >
      <Pressable
        style={modal.overlay}
        onPress={() => {
          Keyboard.dismiss();
          handleCancel();
        }}
      >
        <View style={modal.sheet}>
          {/* Drag handle */}
          <View style={modal.handle} />

          <Text style={modal.title}>Mark as Reserved</Text>
          <Text style={modal.subtitle} numberOfLines={2}>
            {itemTitle}
          </Text>

          <Text style={modal.label}>Who will buy this gift?</Text>
          <TextInput
            style={modal.input}
            placeholder="e.g. Grandma & Grandpa"
            placeholderTextColor={C.muted}
            value={name}
            onChangeText={setName}
            autoFocus
            returnKeyType="next"
            maxLength={255}
          />

          <Text style={modal.label}>Note (optional)</Text>
          <TextInput
            style={[modal.input, modal.inputMultiline]}
            placeholder="e.g. Said they'll buy the bicycle"
            placeholderTextColor={C.muted}
            value={note}
            onChangeText={setNote}
            multiline
            numberOfLines={3}
            maxLength={1000}
            returnKeyType="done"
          />

          <View style={modal.actions}>
            <Pressable onPress={handleCancel} style={modal.cancelBtn}>
              <Text style={modal.cancelText}>Cancel</Text>
            </Pressable>
            <Pressable
              onPress={handleConfirm}
              style={[
                modal.confirmBtn,
                !name.trim() && modal.confirmBtnDisabled,
              ]}
              disabled={!name.trim() || loading}
            >
              {loading ? (
                <ActivityIndicator size={16} color="#1a0f05" />
              ) : (
                <>
                  <MaterialCommunityIcons
                    name="account-check"
                    size={16}
                    color="#1a0f05"
                  />
                  <Text style={modal.confirmText}>Confirm</Text>
                </>
              )}
            </Pressable>
          </View>
        </View>
      </Pressable>
    </Modal>
  );
};

// ─── Share strip ─────────────────────────────────────────────────────
const ShareStrip = ({ slug }: { slug: string }) => {
  const publicUrl = `https://${WEB_DOMAIN}/public/${slug}`;

  const handleShare = async () => {
    try {
      await Share.share({ message: publicUrl, url: publicUrl });
    } catch {
      // user dismissed
    }
  };

  const handleOpen = () => Linking.openURL(publicUrl);

  return (
    <View style={share.strip}>
      <MaterialCommunityIcons name="earth" size={14} color={C.green} />
      <Text style={share.url} numberOfLines={1}>
        {WEB_DOMAIN}/public/{slug}
      </Text>
      <Pressable onPress={handleOpen} style={share.iconBtn} hitSlop={8}>
        <MaterialCommunityIcons name="open-in-new" size={16} color={C.gold} />
      </Pressable>
      <Pressable onPress={handleShare} style={share.iconBtn} hitSlop={8}>
        <MaterialCommunityIcons name="share-variant" size={16} color={C.gold} />
      </Pressable>
    </View>
  );
};

// ─── Empty state ─────────────────────────────────────────────────────
const EmptyGifts = ({
  onAdd,
  onAttach,
}: {
  onAdd: () => void;
  onAttach: () => void;
}) => (
  <View style={empty.wrap}>
    <LinearGradient
      colors={['rgba(226, 185, 108, 0.18)', 'rgba(226, 185, 108, 0.04)']}
      style={empty.iconCircle}
    >
      <MaterialCommunityIcons name="gift-outline" size={40} color={C.gold} />
    </LinearGradient>
    <Text style={empty.title}>No gifts yet</Text>
    <Text style={empty.subtitle}>
      Create a new gift or attach an existing one to this wishlist
    </Text>
    <View style={empty.actions}>
      <Pressable onPress={onAdd} style={empty.cta}>
        <LinearGradient
          colors={[C.goldBright, '#C48A3A']}
          style={empty.ctaGrad}
          start={{ x: 0, y: 0 }}
          end={{ x: 1, y: 0 }}
        >
          <MaterialCommunityIcons name="plus" size={16} color="#1a0f05" />
          <Text style={empty.ctaText}>New Gift</Text>
        </LinearGradient>
      </Pressable>
      <Pressable onPress={onAttach} style={empty.ctaSecondary}>
        <MaterialCommunityIcons name="link-plus" size={16} color={C.gold} />
        <Text style={empty.ctaSecondaryText}>Attach Existing</Text>
      </Pressable>
    </View>
  </View>
);

// ─── Main screen ─────────────────────────────────────────────────────
export default function WishListScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const router = useRouter();
  const queryClient = useQueryClient();
  const [refreshing, setRefreshing] = useState(false);
  const [markReservedItem, setMarkReservedItem] = useState<WishlistItem | null>(
    null,
  );

  const {
    data: wishList,
    isLoading,
    error,
    refetch,
  } = useQuery({
    queryKey: ['wishlist', id],
    queryFn: () => apiClient.getWishListById(id),
    enabled: !!id,
  });

  const {
    data: giftItems,
    isLoading: itemsLoading,
    error: itemsError,
    refetch: refetchItems,
  } = useQuery({
    queryKey: ['giftItems', id],
    queryFn: () => apiClient.getGiftItems(id),
    enabled: !!id,
  });

  const markReservedMutation = useMutation({
    mutationFn: ({
      itemId,
      reservedByName,
      note,
    }: {
      itemId: string;
      reservedByName: string;
      note: string;
    }) =>
      apiClient.markItemAsManuallyReserved(id, itemId, {
        reserved_by_name: reservedByName,
        note: note || undefined,
      }),
    onSuccess: () => {
      setMarkReservedItem(null);
      queryClient
        .invalidateQueries({ queryKey: ['giftItems', id] })
        .catch(console.error);
      dialog.success('Gift marked as reserved!');
    },
    onError: () => {
      dialog.error('Failed to mark gift as reserved. Please try again.');
    },
  });

  const onRefresh = async () => {
    setRefreshing(true);
    await Promise.all([refetch(), refetchItems()]);
    setRefreshing(false);
  };

  const handleMarkReserved = (item: WishlistItem) => {
    setMarkReservedItem(item);
  };

  const handleMarkReservedConfirm = (name: string, note: string) => {
    if (!markReservedItem?.id) return;
    markReservedMutation.mutate({
      itemId: markReservedItem.id,
      reservedByName: name,
      note,
    });
  };

  const handleEditGift = (item: WishlistItem) => {
    if (!item.id) return;
    router.push({
      pathname: '/gift-items/[id]/edit',
      params: { id: item.id, wishlistId: id ?? '' },
    });
  };

  const handleAddGiftItem = () => router.push(`/lists/${id}/gifts/create`);
  const handleAttachExistingItem = () =>
    router.push(`/lists/${id}/attach-items`);
  const handleEditWishList = () => router.push(`/lists/${id}/edit`);

  // ── Loading ──
  if (isLoading || itemsLoading) {
    return (
      <View style={s.root}>
        <LinearGradient
          colors={[C.bg0, C.bg1, C.bg2]}
          style={StyleSheet.absoluteFill}
        />
        <View style={s.center}>
          <ActivityIndicator size="large" color={C.gold} />
          <Text style={s.loadingText}>Loading wishlist…</Text>
        </View>
      </View>
    );
  }

  // ── Error ──
  if (error || itemsError || !wishList) {
    return (
      <View style={s.root}>
        <LinearGradient
          colors={[C.bg0, C.bg1, C.bg2]}
          style={StyleSheet.absoluteFill}
        />
        <View style={s.headerRow}>
          <Pressable onPress={() => router.back()} style={s.circleBtn}>
            <MaterialCommunityIcons
              name="arrow-left"
              size={22}
              color={C.white}
            />
          </Pressable>
        </View>
        <View style={s.center}>
          <MaterialCommunityIcons
            name="alert-circle-outline"
            size={56}
            color="#F87171"
          />
          <Text style={s.errorTitle}>
            {!wishList ? 'Wishlist not found' : 'Something went wrong'}
          </Text>
          <Pressable onPress={() => router.back()} style={s.backBtn}>
            <Text style={s.backBtnText}>Go Back</Text>
          </Pressable>
        </View>
      </View>
    );
  }

  const items = giftItems?.items ?? [];
  const totalCount = items.length;
  const reservedCount = items.filter((i) => i.is_reserved).length;
  const purchasedCount = items.filter((i) => i.is_purchased).length;

  return (
    <View style={s.root}>
      {/* Background */}
      <LinearGradient
        colors={[C.bg0, C.bg1, C.bg2]}
        style={StyleSheet.absoluteFill}
      />

      {/* Ambient glow circles */}
      <View style={s.glow1} />
      <View style={s.glow2} />

      {/* ── Header ── */}
      <View style={s.headerRow}>
        <Pressable onPress={() => router.back()} style={s.circleBtn}>
          <MaterialCommunityIcons name="arrow-left" size={22} color={C.white} />
        </Pressable>

        <View style={s.headerCenter}>
          {wishList.occasion && (
            <View style={s.occasionPill}>
              <MaterialCommunityIcons
                name="calendar-star"
                size={11}
                color={C.gold}
              />
              <Text style={s.occasionPillText}>{wishList.occasion}</Text>
            </View>
          )}
          <Text style={s.headerTitle} numberOfLines={1}>
            {wishList.title}
          </Text>
        </View>

        <Pressable onPress={handleEditWishList} style={s.circleBtn}>
          <MaterialCommunityIcons name="pencil" size={18} color={C.gold} />
        </Pressable>
      </View>

      {/* ── Stats strip ── */}
      <View style={s.statsStrip}>
        <View style={s.statCell}>
          <Text style={s.statNum}>{totalCount}</Text>
          <Text style={s.statLabel}>Gifts</Text>
        </View>
        <View style={s.statSep} />
        <View style={s.statCell}>
          <Text style={[s.statNum, { color: C.amber }]}>{reservedCount}</Text>
          <Text style={s.statLabel}>Reserved</Text>
        </View>
        <View style={s.statSep} />
        <View style={s.statCell}>
          <Text style={[s.statNum, { color: C.violet }]}>{purchasedCount}</Text>
          <Text style={s.statLabel}>Purchased</Text>
        </View>
      </View>

      {/* ── Share strip (public wishlist with slug) ── */}
      {wishList.is_public && wishList.public_slug && (
        <ShareStrip slug={wishList.public_slug} />
      )}

      {/* ── Description (if exists) ── */}
      {wishList.description && (
        <View style={s.descStrip}>
          <Text style={s.descText} numberOfLines={2}>
            {wishList.description}
          </Text>
        </View>
      )}

      {/* ── Gift items scroll ── */}
      <ScrollView
        style={s.scroll}
        contentContainerStyle={s.scrollContent}
        showsVerticalScrollIndicator={false}
        refreshControl={
          <RefreshControl
            refreshing={refreshing}
            onRefresh={onRefresh}
            tintColor={C.gold}
          />
        }
      >
        {items.length > 0 ? (
          items.map((item, idx) => (
            <GiftItemCard
              key={item.id}
              item={item}
              index={idx}
              onMarkReserved={handleMarkReserved}
              onEdit={handleEditGift}
            />
          ))
        ) : (
          <EmptyGifts
            onAdd={handleAddGiftItem}
            onAttach={handleAttachExistingItem}
          />
        )}
      </ScrollView>

      {/* ── FABs ── */}
      {items.length > 0 && (
        <View style={s.fabGroup}>
          {/* Attach existing gift */}
          <Pressable
            onPress={() => {
              Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Light);
              handleAttachExistingItem();
            }}
            style={s.fabSecondary}
          >
            <View style={s.fabSecondaryInner}>
              <MaterialCommunityIcons
                name="link-plus"
                size={22}
                color={C.gold}
              />
            </View>
          </Pressable>

          {/* Create new gift */}
          <Pressable
            onPress={() => {
              Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Medium);
              handleAddGiftItem();
            }}
            style={s.fab}
          >
            <LinearGradient
              colors={[C.goldBright, '#C48A3A']}
              style={s.fabGrad}
              start={{ x: 0, y: 0 }}
              end={{ x: 1, y: 1 }}
            >
              <MaterialCommunityIcons name="plus" size={26} color="#1a0f05" />
            </LinearGradient>
          </Pressable>
        </View>
      )}

      {/* ── Mark Reserved Modal ── */}
      <MarkReservedModal
        visible={!!markReservedItem}
        itemTitle={markReservedItem?.title ?? ''}
        loading={markReservedMutation.isPending}
        onConfirm={handleMarkReservedConfirm}
        onCancel={() => setMarkReservedItem(null)}
      />
    </View>
  );
}

// ─── Card styles ─────────────────────────────────────────────────────
const card = StyleSheet.create({
  wrapper: {
    borderRadius: 20,
    overflow: 'hidden',
    backgroundColor: C.surface,
    borderWidth: 1,
    borderColor: C.border,
    marginBottom: 16,
  },
  imageWrap: {
    height: 168,
    position: 'relative',
  },
  image: {
    width: '100%',
    height: '100%',
  },
  imagePlaceholder: {
    width: '100%',
    height: '100%',
    alignItems: 'center',
    justifyContent: 'center',
  },
  imageScrim: {
    ...StyleSheet.absoluteFillObject,
    backgroundColor: 'rgba(6, 4, 17, 0.50)',
  },
  priorityChip: {
    position: 'absolute',
    top: 10,
    left: 10,
    flexDirection: 'row',
    alignItems: 'center',
    gap: 3,
    backgroundColor: 'rgba(0,0,0,0.55)',
    paddingHorizontal: 8,
    paddingVertical: 4,
    borderRadius: 20,
    backdropFilter: 'blur(4px)',
  },
  priorityChipText: {
    fontSize: 11,
    fontWeight: '700',
    color: '#FFD166',
  },
  statusChip: {
    position: 'absolute',
    top: 10,
    right: 10,
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
    paddingHorizontal: 9,
    paddingVertical: 5,
    borderRadius: 20,
  },
  statusChipText: {
    fontSize: 11,
    fontWeight: '700',
    letterSpacing: 0.2,
  },
  body: {
    padding: 16,
  },
  titleRow: {
    flexDirection: 'row',
    alignItems: 'flex-start',
    justifyContent: 'space-between',
    gap: 12,
    marginBottom: 6,
  },
  title: {
    flex: 1,
    fontSize: 17,
    fontWeight: '700',
    color: C.white,
    lineHeight: 22,
  },
  titleFaded: {
    color: C.muted,
  },
  price: {
    fontSize: 18,
    fontWeight: '800',
    color: C.goldBright,
    letterSpacing: -0.3,
  },
  subtitle: {
    fontSize: 13,
    color: C.muted,
    lineHeight: 19,
    marginBottom: 14,
  },
  manualReservationBadge: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 5,
    marginBottom: 10,
    paddingHorizontal: 10,
    paddingVertical: 5,
    borderRadius: 10,
    backgroundColor: 'rgba(252, 211, 77, 0.10)',
    borderWidth: 1,
    borderColor: 'rgba(252, 211, 77, 0.20)',
    alignSelf: 'flex-start',
  },
  manualReservationText: {
    fontSize: 12,
    fontWeight: '600',
    color: C.amber,
  },
  actions: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    marginTop: 2,
  },
  linkPill: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
    paddingHorizontal: 10,
    paddingVertical: 6,
    borderRadius: 20,
    borderWidth: 1,
    borderColor: C.goldBorder,
    backgroundColor: C.goldDim,
  },
  linkPillText: {
    fontSize: 12,
    fontWeight: '600',
    color: C.gold,
  },
  iconBtn: {
    width: 32,
    height: 32,
    borderRadius: 16,
    backgroundColor: C.surface,
    borderWidth: 1,
    borderColor: C.border,
    alignItems: 'center',
    justifyContent: 'center',
  },
  reserveBtn: {
    borderRadius: 20,
    overflow: 'hidden',
  },
  reserveBtnGrad: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 5,
    paddingHorizontal: 12,
    paddingVertical: 7,
  },
  reserveBtnText: {
    fontSize: 12,
    fontWeight: '700',
    color: '#1a0f05',
  },
});

// ─── Modal styles ─────────────────────────────────────────────────────
const modal = StyleSheet.create({
  overlay: {
    flex: 1,
    backgroundColor: 'rgba(0,0,0,0.6)',
    justifyContent: 'flex-end',
  },
  sheet: {
    backgroundColor: '#16112e',
    borderTopLeftRadius: 28,
    borderTopRightRadius: 28,
    paddingHorizontal: 24,
    paddingTop: 12,
    paddingBottom: 40,
    borderTopWidth: 1,
    borderColor: 'rgba(226, 185, 108, 0.15)',
  },
  handle: {
    width: 40,
    height: 4,
    borderRadius: 2,
    backgroundColor: 'rgba(255,255,255,0.15)',
    alignSelf: 'center',
    marginBottom: 20,
  },
  title: {
    fontSize: 20,
    fontWeight: '800',
    color: '#EEE8FF',
    marginBottom: 4,
  },
  subtitle: {
    fontSize: 13,
    color: 'rgba(238, 232, 255, 0.5)',
    marginBottom: 24,
  },
  label: {
    fontSize: 12,
    fontWeight: '600',
    color: 'rgba(238, 232, 255, 0.55)',
    textTransform: 'uppercase',
    letterSpacing: 0.5,
    marginBottom: 8,
  },
  input: {
    backgroundColor: 'rgba(255,255,255,0.06)',
    borderWidth: 1,
    borderColor: 'rgba(226, 185, 108, 0.2)',
    borderRadius: 14,
    paddingHorizontal: 16,
    paddingVertical: 13,
    fontSize: 15,
    color: '#EEE8FF',
    marginBottom: 20,
  },
  inputMultiline: {
    minHeight: 80,
    textAlignVertical: 'top',
  },
  actions: {
    flexDirection: 'row',
    gap: 12,
    marginTop: 4,
  },
  cancelBtn: {
    flex: 1,
    paddingVertical: 14,
    borderRadius: 16,
    backgroundColor: 'rgba(255,255,255,0.06)',
    borderWidth: 1,
    borderColor: 'rgba(255,255,255,0.1)',
    alignItems: 'center',
  },
  cancelText: {
    fontSize: 15,
    fontWeight: '600',
    color: 'rgba(238, 232, 255, 0.7)',
  },
  confirmBtn: {
    flex: 2,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    gap: 8,
    paddingVertical: 14,
    borderRadius: 16,
    backgroundColor: '#E2B96C',
  },
  confirmBtnDisabled: {
    opacity: 0.4,
  },
  confirmText: {
    fontSize: 15,
    fontWeight: '700',
    color: '#1a0f05',
  },
});

// ─── Share strip styles ──────────────────────────────────────────────
const share = StyleSheet.create({
  strip: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    marginHorizontal: 20,
    marginBottom: 8,
    paddingVertical: 10,
    paddingHorizontal: 14,
    borderRadius: 14,
    backgroundColor: 'rgba(94, 234, 212, 0.08)',
    borderWidth: 1,
    borderColor: 'rgba(94, 234, 212, 0.2)',
  },
  url: {
    flex: 1,
    fontSize: 12,
    color: 'rgba(94, 234, 212, 0.85)',
  },
  iconBtn: {
    width: 30,
    height: 30,
    borderRadius: 15,
    backgroundColor: C.goldDim,
    alignItems: 'center',
    justifyContent: 'center',
  },
});

// ─── Empty state styles ──────────────────────────────────────────────
const empty = StyleSheet.create({
  wrap: {
    alignItems: 'center',
    paddingVertical: 64,
    paddingHorizontal: 32,
  },
  iconCircle: {
    width: 96,
    height: 96,
    borderRadius: 48,
    alignItems: 'center',
    justifyContent: 'center',
    marginBottom: 24,
  },
  title: {
    fontSize: 20,
    fontWeight: '700',
    color: C.white,
    marginBottom: 8,
  },
  subtitle: {
    fontSize: 14,
    color: C.muted,
    textAlign: 'center',
    lineHeight: 20,
    marginBottom: 28,
  },
  actions: {
    flexDirection: 'row',
    gap: 12,
    alignItems: 'center',
  },
  cta: {
    borderRadius: 24,
    overflow: 'hidden',
  },
  ctaGrad: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    paddingHorizontal: 24,
    paddingVertical: 12,
  },
  ctaText: {
    fontSize: 15,
    fontWeight: '700',
    color: '#1a0f05',
  },
  ctaSecondary: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    paddingHorizontal: 20,
    paddingVertical: 12,
    borderRadius: 24,
    borderWidth: 1.5,
    borderColor: C.goldBorder,
    backgroundColor: C.goldDim,
  },
  ctaSecondaryText: {
    fontSize: 15,
    fontWeight: '700',
    color: C.gold,
  },
});

// ─── Page styles ─────────────────────────────────────────────────────
const s = StyleSheet.create({
  root: {
    flex: 1,
  },
  glow1: {
    position: 'absolute',
    width: 300,
    height: 300,
    borderRadius: 150,
    backgroundColor: 'rgba(226, 185, 108, 0.05)',
    top: -100,
    right: -80,
  },
  glow2: {
    position: 'absolute',
    width: 220,
    height: 220,
    borderRadius: 110,
    backgroundColor: 'rgba(100, 70, 200, 0.08)',
    bottom: 180,
    left: -60,
  },
  // header
  headerRow: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 20,
    paddingTop: 60,
    paddingBottom: 16,
    gap: 12,
  },
  headerCenter: {
    flex: 1,
    alignItems: 'center',
    gap: 4,
  },
  occasionPill: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 5,
    paddingHorizontal: 10,
    paddingVertical: 3,
    borderRadius: 20,
    backgroundColor: C.goldDim,
    borderWidth: 1,
    borderColor: C.goldBorder,
  },
  occasionPillText: {
    fontSize: 11,
    fontWeight: '600',
    color: C.gold,
    letterSpacing: 0.3,
  },
  headerTitle: {
    fontSize: 20,
    fontWeight: '800',
    color: C.white,
    letterSpacing: -0.3,
  },
  circleBtn: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: C.surfaceMid,
    borderWidth: 1,
    borderColor: C.border,
    alignItems: 'center',
    justifyContent: 'center',
  },
  // stats
  statsStrip: {
    flexDirection: 'row',
    marginHorizontal: 20,
    marginBottom: 8,
    paddingVertical: 14,
    paddingHorizontal: 8,
    borderRadius: 16,
    backgroundColor: C.surface,
    borderWidth: 1,
    borderColor: C.border,
    alignItems: 'center',
    justifyContent: 'space-around',
  },
  statCell: {
    flex: 1,
    alignItems: 'center',
  },
  statNum: {
    fontSize: 22,
    fontWeight: '800',
    color: C.white,
    letterSpacing: -0.5,
  },
  statLabel: {
    fontSize: 11,
    color: C.muted,
    marginTop: 2,
    letterSpacing: 0.3,
    textTransform: 'uppercase',
  },
  statSep: {
    width: 1,
    height: 32,
    backgroundColor: C.border,
  },
  // description
  descStrip: {
    marginHorizontal: 20,
    marginBottom: 12,
  },
  descText: {
    fontSize: 13,
    color: C.muted,
    lineHeight: 19,
  },
  // scroll
  scroll: {
    flex: 1,
  },
  scrollContent: {
    paddingHorizontal: 20,
    paddingTop: 8,
    paddingBottom: 100,
  },
  // loading/error
  center: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    gap: 12,
    padding: 32,
  },
  loadingText: {
    fontSize: 15,
    color: C.muted,
  },
  errorTitle: {
    fontSize: 18,
    fontWeight: '700',
    color: '#F87171',
    textAlign: 'center',
  },
  backBtn: {
    marginTop: 8,
    paddingHorizontal: 24,
    paddingVertical: 12,
    borderRadius: 14,
    backgroundColor: C.surfaceMid,
    borderWidth: 1,
    borderColor: C.border,
  },
  backBtnText: {
    fontSize: 15,
    fontWeight: '600',
    color: C.white,
  },
  // fab group
  fabGroup: {
    position: 'absolute',
    bottom: 28,
    right: 24,
    flexDirection: 'row',
    alignItems: 'center',
    gap: 12,
  },
  fab: {
    borderRadius: 30,
    shadowColor: C.gold,
    shadowOffset: { width: 0, height: 6 },
    shadowOpacity: 0.35,
    shadowRadius: 14,
    elevation: 10,
  },
  fabGrad: {
    width: 58,
    height: 58,
    borderRadius: 29,
    alignItems: 'center',
    justifyContent: 'center',
  },
  fabSecondary: {
    borderRadius: 26,
    shadowColor: C.gold,
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.2,
    shadowRadius: 8,
    elevation: 6,
  },
  fabSecondaryInner: {
    width: 50,
    height: 50,
    borderRadius: 25,
    backgroundColor: C.bg2,
    borderWidth: 1.5,
    borderColor: C.goldBorder,
    alignItems: 'center',
    justifyContent: 'center',
  },
});
