import { MaterialCommunityIcons } from '@expo/vector-icons';
import { useMutation } from '@tanstack/react-query';
import { BlurView } from 'expo-blur';
import { LinearGradient } from 'expo-linear-gradient';
import { useRouter } from 'expo-router';
import { useState } from 'react';
import {
  Alert,
  Pressable,
  ScrollView,
  StyleSheet,
  Switch,
  View,
} from 'react-native';
import { Text, TextInput } from 'react-native-paper';
import { apiClient } from '@/lib/api';

export default function CreateWishListScreen() {
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [occasion, setOccasion] = useState('');
  const [isPublic, setIsPublic] = useState(false);
  const router = useRouter();

  const mutation = useMutation({
    mutationFn: (data: {
      title: string;
      description?: string;
      occasion?: string;
      is_public: boolean;
    }) =>
      apiClient.createWishList({
        title: data.title,
        description: data.description || '',
        occasion: data.occasion || '',
        is_public: data.is_public,
        template_id: 'default',
      }),
    onSuccess: () => {
      Alert.alert('Success', 'Wishlist created successfully!', [
        { text: 'OK', onPress: () => router.push('/lists') },
      ]);
    },
    onError: (error: Error) => {
      Alert.alert(
        'Error',
        error.message || 'Failed to create wishlist. Please try again.',
      );
    },
  });

  const handleCreate = () => {
    if (!title.trim()) {
      Alert.alert('Error', 'Please enter a title for your wishlist.');
      return;
    }

    mutation.mutate({
      title: title.trim(),
      description: description.trim(),
      occasion: occasion.trim(),
      is_public: isPublic,
    });
  };

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
        <Text style={styles.headerTitle}>New Wishlist</Text>
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
                <MaterialCommunityIcons name="gift" size={32} color="#000000" />
              </LinearGradient>
            </View>

            <TextInput
              label="Title *"
              value={title}
              onChangeText={setTitle}
              maxLength={200}
              style={styles.input}
              textColor="#ffffff"
              underlineColor="transparent"
              activeUnderlineColor="#FFD700"
              placeholderTextColor="rgba(255, 255, 255, 0.4)"
              theme={{
                colors: {
                  primary: '#FFD700',
                  onSurfaceVariant: 'rgba(255, 255, 255, 0.5)',
                },
              }}
            />

            <View style={styles.textAreaWrapper}>
              <TextInput
                label="Description"
                value={description}
                onChangeText={setDescription}
                multiline
                numberOfLines={3}
                mode="flat"
                style={styles.textArea}
                textColor="#ffffff"
                underlineColor="transparent"
                activeUnderlineColor="#FFD700"
                placeholderTextColor="rgba(255, 255, 255, 0.4)"
                contentStyle={{
                  backgroundColor: 'transparent',
                }}
                theme={{
                  colors: {
                    primary: '#FFD700',
                    onSurfaceVariant: 'rgba(255, 255, 255, 0.5)',
                    background: 'transparent',
                    surface: 'transparent',
                    surfaceVariant: 'transparent',
                    elevation: {
                      level0: 'transparent',
                      level1: 'transparent',
                      level2: 'transparent',
                      level3: 'transparent',
                      level4: 'transparent',
                      level5: 'transparent',
                    },
                  },
                }}
              />
            </View>

            <TextInput
              label="Occasion (e.g., Birthday, Wedding)"
              value={occasion}
              onChangeText={setOccasion}
              maxLength={100}
              style={styles.input}
              textColor="#ffffff"
              underlineColor="transparent"
              activeUnderlineColor="#FFD700"
              placeholderTextColor="rgba(255, 255, 255, 0.4)"
              theme={{
                colors: {
                  primary: '#FFD700',
                  onSurfaceVariant: 'rgba(255, 255, 255, 0.5)',
                },
              }}
            />

            {/* Public Toggle */}
            <View style={styles.toggleContainer}>
              <View style={styles.toggleLeft}>
                <MaterialCommunityIcons
                  name="earth"
                  size={20}
                  color="#FFD700"
                />
                <Text style={styles.toggleLabel}>Make Public</Text>
              </View>
              <Switch
                value={isPublic}
                onValueChange={setIsPublic}
                trackColor={{ false: '#767577', true: '#FFD700' }}
                thumbColor={isPublic ? '#FFA500' : '#f4f3f4'}
              />
            </View>

            {/* Create Button */}
            <Pressable onPress={handleCreate} disabled={mutation.isPending}>
              <LinearGradient
                colors={['#FFD700', '#FFA500']}
                style={styles.createButton}
              >
                {mutation.isPending ? (
                  <Text style={styles.createButtonText}>Creating...</Text>
                ) : (
                  <>
                    <MaterialCommunityIcons
                      name="plus"
                      size={20}
                      color="#000000"
                    />
                    <Text style={styles.createButtonText}>Create Wishlist</Text>
                  </>
                )}
              </LinearGradient>
            </Pressable>
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
            Public wishlists can be shared with anyone via link
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
  input: {
    backgroundColor: 'rgba(255, 255, 255, 0.08)',
    marginBottom: 16,
    borderRadius: 12,
  },
  textAreaWrapper: {
    backgroundColor: 'rgba(255, 255, 255, 0.08)',
    borderTopLeftRadius: 4,
    borderTopRightRadius: 4,
    borderBottomRightRadius: 12,
    borderBottomLeftRadius: 12,
    marginBottom: 16,
  },
  textArea: {
    backgroundColor: 'transparent',
    minHeight: 100,
  },
  toggleContainer: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingVertical: 16,
    paddingHorizontal: 16,
    backgroundColor: 'rgba(255, 255, 255, 0.05)',
    borderRadius: 12,
    marginBottom: 24,
  },
  toggleLeft: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 12,
  },
  toggleLabel: {
    fontSize: 16,
    color: '#ffffff',
    fontWeight: '500',
  },
  createButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 16,
    borderRadius: 12,
    gap: 8,
  },
  createButtonText: {
    fontSize: 16,
    fontWeight: '700',
    color: '#000000',
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
});
