import { MaterialCommunityIcons } from '@expo/vector-icons';
import { BlurView } from 'expo-blur';
import { LinearGradient } from 'expo-linear-gradient';
import { useRouter } from 'expo-router';
import {
  Image,
  Linking,
  Modal,
  Pressable,
  ScrollView,
  StyleSheet,
  View,
} from 'react-native';
import { Text } from 'react-native-paper';
import type { GiftItem } from '@/lib/api/types';

interface GiftItemDetailModalProps {
  item: GiftItem | null;
  visible: boolean;
  onClose: () => void;
}

export default function GiftItemDetailModal({
  item,
  visible,
  onClose,
}: GiftItemDetailModalProps) {
  const router = useRouter();

  if (!item) return null;

  const isAttached = (item.wishlist_ids?.length ?? 0) > 0;
  const primaryWishlistId = item.wishlist_ids?.[0] || '';

  const handleEdit = () => {
    onClose();
    router.push(
      `/gift-items/${item.id}/edit?wishlistId=${primaryWishlistId}`,
    );
  };

  return (
    <Modal
      visible={visible}
      transparent
      animationType="slide"
      onRequestClose={onClose}
    >
      <View style={styles.modalOverlay}>
        <Pressable style={styles.backdrop} onPress={onClose} />

        <View style={styles.modalContainer}>
          <LinearGradient
            colors={['#2d1b4e', '#3d2a6e', '#1a0a2e']}
            style={StyleSheet.absoluteFill}
          />

          {/* Header */}
          <View style={styles.header}>
            <Text style={styles.headerTitle}>Gift Details</Text>
            <Pressable onPress={onClose} style={styles.closeButton}>
              <MaterialCommunityIcons name="close" size={24} color="#ffffff" />
            </Pressable>
          </View>

          <ScrollView
            style={styles.scrollView}
            showsVerticalScrollIndicator={false}
          >
            {/* Image */}
            {item.image_url ? (
              <Image
                source={{ uri: item.image_url }}
                style={styles.image}
                resizeMode="cover"
              />
            ) : (
              <View style={styles.placeholderImage}>
                <MaterialCommunityIcons
                  name="gift"
                  size={64}
                  color="rgba(255, 215, 0, 0.3)"
                />
              </View>
            )}

            <View style={styles.content}>
              {/* Title and Price */}
              <View style={styles.titleSection}>
                <Text style={styles.title}>{item.title}</Text>
                {item.price !== undefined && item.price !== null && (
                  <LinearGradient
                    colors={['#FFD700', '#FFA500']}
                    style={styles.priceTag}
                  >
                    <Text style={styles.price}>${item.price.toFixed(2)}</Text>
                  </LinearGradient>
                )}
              </View>

              {/* Status and Priority Badges */}
              <View style={styles.badgesRow}>
                {isAttached ? (
                  <View style={styles.statusBadge}>
                    <MaterialCommunityIcons
                      name="link"
                      size={16}
                      color="#4CAF50"
                    />
                    <Text style={styles.statusText}>Attached to List</Text>
                  </View>
                ) : (
                  <View style={styles.statusBadge}>
                    <MaterialCommunityIcons
                      name="link-off"
                      size={16}
                      color="#FF9800"
                    />
                    <Text style={styles.statusText}>Standalone</Text>
                  </View>
                )}

                {item.priority && item.priority > 0 && (
                  <View style={styles.priorityBadge}>
                    <MaterialCommunityIcons
                      name="star"
                      size={16}
                      color="#FFD700"
                    />
                    <Text style={styles.priorityText}>
                      Priority: {item.priority}/10
                    </Text>
                  </View>
                )}
              </View>

              {/* Description */}
              {item.description && (
                <BlurView intensity={20} style={styles.section}>
                  <View style={styles.sectionHeader}>
                    <MaterialCommunityIcons
                      name="text"
                      size={20}
                      color="#FFD700"
                    />
                    <Text style={styles.sectionTitle}>Description</Text>
                  </View>
                  <Text style={styles.description}>{item.description}</Text>
                </BlurView>
              )}

              {/* Link */}
              {item.link && (
                <BlurView intensity={20} style={styles.section}>
                  <View style={styles.sectionHeader}>
                    <MaterialCommunityIcons
                      name="link"
                      size={20}
                      color="#FFD700"
                    />
                    <Text style={styles.sectionTitle}>Link</Text>
                  </View>
                  <Pressable onPress={() => Linking.openURL(item.link || '')}>
                    <View style={styles.linkContainer}>
                      <Text style={styles.linkText} numberOfLines={1}>
                        {item.link}
                      </Text>
                      <MaterialCommunityIcons
                        name="open-in-new"
                        size={16}
                        color="#FFD700"
                      />
                    </View>
                  </Pressable>
                </BlurView>
              )}

              {/* Notes */}
              {item.notes && (
                <BlurView intensity={20} style={styles.section}>
                  <View style={styles.sectionHeader}>
                    <MaterialCommunityIcons
                      name="note-text"
                      size={20}
                      color="#FFD700"
                    />
                    <Text style={styles.sectionTitle}>Notes</Text>
                  </View>
                  <Text style={styles.notes}>{item.notes}</Text>
                </BlurView>
              )}

              {/* Edit Button */}
              <Pressable
                onPress={handleEdit}
                style={styles.editButtonContainer}
              >
                <LinearGradient
                  colors={['#FFD700', '#FFA500']}
                  style={styles.editButton}
                >
                  <MaterialCommunityIcons
                    name="pencil"
                    size={20}
                    color="#000000"
                  />
                  <Text style={styles.editButtonText}>Edit Gift</Text>
                </LinearGradient>
              </Pressable>
            </View>
          </ScrollView>
        </View>
      </View>
    </Modal>
  );
}

const styles = StyleSheet.create({
  modalOverlay: {
    flex: 1,
    justifyContent: 'flex-end',
  },
  backdrop: {
    ...StyleSheet.absoluteFillObject,
    backgroundColor: 'rgba(0, 0, 0, 0.5)',
  },
  modalContainer: {
    height: '90%',
    borderTopLeftRadius: 24,
    borderTopRightRadius: 24,
    overflow: 'hidden',
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: 20,
    paddingTop: 20,
    paddingBottom: 16,
    borderBottomWidth: 1,
    borderBottomColor: 'rgba(255, 255, 255, 0.1)',
  },
  headerTitle: {
    fontSize: 20,
    fontWeight: '700',
    color: '#ffffff',
  },
  closeButton: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: 'rgba(255, 255, 255, 0.1)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  scrollView: {
    flex: 1,
  },
  image: {
    width: '100%',
    height: 300,
    backgroundColor: 'rgba(255, 255, 255, 0.05)',
  },
  placeholderImage: {
    width: '100%',
    height: 300,
    backgroundColor: 'rgba(255, 255, 255, 0.05)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  content: {
    padding: 20,
    paddingBottom: 40,
  },
  titleSection: {
    marginBottom: 16,
  },
  title: {
    fontSize: 24,
    fontWeight: '700',
    color: '#ffffff',
    marginBottom: 12,
    lineHeight: 32,
  },
  priceTag: {
    alignSelf: 'flex-start',
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderRadius: 12,
  },
  price: {
    fontSize: 20,
    fontWeight: '700',
    color: '#000000',
  },
  badgesRow: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 8,
    marginBottom: 20,
  },
  statusBadge: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 12,
    backgroundColor: 'rgba(255, 255, 255, 0.1)',
  },
  statusText: {
    fontSize: 13,
    color: 'rgba(255, 255, 255, 0.9)',
    fontWeight: '600',
  },
  priorityBadge: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 12,
    backgroundColor: 'rgba(255, 215, 0, 0.2)',
  },
  priorityText: {
    fontSize: 13,
    color: '#FFD700',
    fontWeight: '600',
  },
  section: {
    borderRadius: 16,
    backgroundColor: 'rgba(255, 255, 255, 0.08)',
    borderWidth: 1,
    borderColor: 'rgba(255, 255, 255, 0.1)',
    padding: 16,
    marginBottom: 12,
  },
  sectionHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    marginBottom: 12,
  },
  sectionTitle: {
    fontSize: 16,
    fontWeight: '600',
    color: '#ffffff',
  },
  description: {
    fontSize: 15,
    color: 'rgba(255, 255, 255, 0.8)',
    lineHeight: 22,
  },
  linkContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    backgroundColor: 'rgba(255, 215, 0, 0.1)',
    padding: 12,
    borderRadius: 12,
  },
  linkText: {
    flex: 1,
    fontSize: 14,
    color: '#FFD700',
    fontWeight: '500',
  },
  notes: {
    fontSize: 14,
    color: 'rgba(255, 255, 255, 0.7)',
    lineHeight: 20,
    fontStyle: 'italic',
  },
  editButtonContainer: {
    marginTop: 8,
  },
  editButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    gap: 8,
    paddingVertical: 16,
    borderRadius: 16,
  },
  editButtonText: {
    fontSize: 16,
    fontWeight: '700',
    color: '#000000',
  },
});
