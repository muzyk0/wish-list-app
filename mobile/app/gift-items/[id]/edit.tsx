import { MaterialCommunityIcons } from '@expo/vector-icons';
import { useQuery } from '@tanstack/react-query';
import { BlurView } from 'expo-blur';
import { LinearGradient } from 'expo-linear-gradient';
import { useLocalSearchParams, useRouter } from 'expo-router';
import { Pressable, ScrollView, StyleSheet, View } from 'react-native';
import { ActivityIndicator, Text } from 'react-native-paper';
import { apiClient } from '@/lib/api';
import GiftItemForm from '../../../components/wish-list/GiftItemForm';

export default function EditGiftItemScreen() {
  const { id, wishlistId } = useLocalSearchParams<{
    id: string;
    wishlistId?: string;
  }>();
  const router = useRouter();

  const {
    data: giftItem,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['giftItem', id],
    queryFn: () => apiClient.getGiftItemById(wishlistId || '', id as string),
    enabled: !!id,
  });

  if (isLoading) {
    return (
      <View style={styles.container}>
        <LinearGradient
          colors={['#1a0a2e', '#2d1b4e', '#3d2a6e']}
          style={StyleSheet.absoluteFill}
        />
        <View style={styles.centerContainer}>
          <ActivityIndicator size="large" color="#FFD700" />
          <Text style={styles.loadingText}>Loading gift item...</Text>
        </View>
      </View>
    );
  }

  if (error) {
    return (
      <View style={styles.container}>
        <LinearGradient
          colors={['#1a0a2e', '#2d1b4e', '#3d2a6e']}
          style={StyleSheet.absoluteFill}
        />
        <View style={styles.decorCircle1} />
        <View style={styles.decorCircle2} />

        <View style={styles.header}>
          <Pressable onPress={() => router.back()} style={styles.backButton}>
            <MaterialCommunityIcons
              name="arrow-left"
              size={24}
              color="#ffffff"
            />
          </Pressable>
          <Text style={styles.headerTitle}>Error</Text>
          <View style={{ width: 40 }} />
        </View>

        <View style={styles.centerContainer}>
          <BlurView intensity={20} style={styles.errorCard}>
            <MaterialCommunityIcons
              name="alert-circle"
              size={64}
              color="#FF6B6B"
            />
            <Text style={styles.errorTitle}>Error loading gift item</Text>
            <Text style={styles.errorMessage}>{error.message}</Text>
            <Pressable onPress={() => router.back()}>
              <LinearGradient
                colors={['#FFD700', '#FFA500']}
                style={styles.backHomeButton}
              >
                <Text style={styles.backHomeButtonText}>Go Back</Text>
              </LinearGradient>
            </Pressable>
          </BlurView>
        </View>
      </View>
    );
  }

  return (
    <View style={styles.container}>
      <LinearGradient
        colors={['#1a0a2e', '#2d1b4e', '#3d2a6e']}
        style={StyleSheet.absoluteFill}
      />

      {/* Decorative elements */}
      <View style={styles.decorCircle1} />
      <View style={styles.decorCircle2} />

      {/* Header */}
      <View style={styles.header}>
        <Pressable onPress={() => router.back()} style={styles.backButton}>
          <MaterialCommunityIcons name="arrow-left" size={24} color="#ffffff" />
        </Pressable>
        <Text style={styles.headerTitle}>Edit Gift</Text>
        <View style={{ width: 40 }} />
      </View>

      <ScrollView
        style={styles.scrollView}
        contentContainerStyle={styles.scrollContent}
        showsVerticalScrollIndicator={false}
      >
        {/* Form Card */}
        <BlurView intensity={20} style={styles.formCard}>
          <View style={styles.formContent}>
            <View style={styles.iconContainer}>
              <LinearGradient
                colors={['#FFD700', '#FFA500']}
                style={styles.iconGradient}
              >
                <MaterialCommunityIcons
                  name="pencil"
                  size={32}
                  color="#000000"
                />
              </LinearGradient>
            </View>

            {giftItem && (
              <GiftItemForm
                wishlistId={giftItem.wishlist_ids?.[0] || ''}
                existingItem={giftItem}
                onComplete={() => router.back()}
              />
            )}
          </View>
        </BlurView>

        {/* Help Text */}
        <View style={styles.helpContainer}>
          <MaterialCommunityIcons
            name="information"
            size={16}
            color="rgba(255, 255, 255, 0.5)"
          />
          <Text style={styles.helpText}>
            Update your gift item details and preferences
          </Text>
        </View>
      </ScrollView>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  decorCircle1: {
    position: 'absolute',
    width: 250,
    height: 250,
    borderRadius: 125,
    backgroundColor: 'rgba(255, 215, 0, 0.06)',
    top: -80,
    right: -60,
  },
  decorCircle2: {
    position: 'absolute',
    width: 180,
    height: 180,
    borderRadius: 90,
    backgroundColor: 'rgba(107, 78, 230, 0.12)',
    bottom: 200,
    left: -40,
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: 20,
    paddingTop: 60,
    paddingBottom: 20,
  },
  backButton: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: 'rgba(255, 255, 255, 0.1)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  headerTitle: {
    fontSize: 18,
    fontWeight: '700',
    color: '#ffffff',
  },
  scrollView: {
    flex: 1,
  },
  scrollContent: {
    padding: 24,
    paddingBottom: 100,
  },
  formCard: {
    borderRadius: 24,
    overflow: 'hidden',
    backgroundColor: 'rgba(255, 255, 255, 0.08)',
    borderWidth: 1,
    borderColor: 'rgba(255, 255, 255, 0.1)',
  },
  formContent: {
    padding: 24,
  },
  iconContainer: {
    alignItems: 'center',
    marginBottom: 24,
  },
  iconGradient: {
    width: 80,
    height: 80,
    borderRadius: 20,
    justifyContent: 'center',
    alignItems: 'center',
  },
  helpContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    marginTop: 16,
    paddingHorizontal: 4,
  },
  helpText: {
    fontSize: 13,
    color: 'rgba(255, 255, 255, 0.5)',
    flex: 1,
  },
  centerContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 20,
  },
  loadingText: {
    fontSize: 16,
    color: 'rgba(255, 255, 255, 0.7)',
    marginTop: 16,
  },
  errorCard: {
    borderRadius: 24,
    overflow: 'hidden',
    backgroundColor: 'rgba(255, 255, 255, 0.08)',
    borderWidth: 1,
    borderColor: 'rgba(255, 255, 255, 0.1)',
    padding: 32,
    alignItems: 'center',
    maxWidth: 400,
  },
  errorTitle: {
    fontSize: 20,
    fontWeight: '700',
    color: '#FF6B6B',
    marginTop: 16,
    marginBottom: 8,
    textAlign: 'center',
  },
  errorMessage: {
    fontSize: 14,
    color: 'rgba(255, 255, 255, 0.6)',
    textAlign: 'center',
    marginBottom: 24,
  },
  backHomeButton: {
    paddingVertical: 12,
    paddingHorizontal: 32,
    borderRadius: 12,
  },
  backHomeButtonText: {
    fontSize: 16,
    fontWeight: '700',
    color: '#000000',
  },
});
