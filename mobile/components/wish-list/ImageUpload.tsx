import { Ionicons } from '@expo/vector-icons';
import * as FileSystem from 'expo-file-system';
import * as ImagePicker from 'expo-image-picker';
import { useState } from 'react';
import { Alert, View } from 'react-native';
import { Button, Card, Text, useTheme } from 'react-native-paper';
import { apiClient } from '@/lib/api';

interface ImageUploadProps {
  onImageUpload: (url: string) => void;
  currentImageUrl?: string;
  disabled?: boolean;
}

export default function ImageUpload({
  onImageUpload,
  currentImageUrl,
  disabled,
}: ImageUploadProps) {
  const [imageUri, setImageUri] = useState<string | null>(
    currentImageUrl || null,
  );
  const [uploading, setUploading] = useState(false);

  const pickImage = async () => {
    if (disabled) return;

    try {
      // Request media library permissions
      const { status } =
        await ImagePicker.requestMediaLibraryPermissionsAsync();
      if (status !== 'granted') {
        Alert.alert(
          'Permission Denied',
          'Sorry, we need camera roll permissions to upload images.',
        );
        return;
      }

      const result = await ImagePicker.launchImageLibraryAsync({
        mediaTypes: ImagePicker.MediaTypeOptions.Images,
        allowsEditing: true,
        aspect: [4, 3],
        quality: 1,
      });

      if (!result.canceled) {
        uploadImage(result.assets[0].uri);
      }
    } catch (error) {
      console.error('Error picking image:', error);
      Alert.alert('Error', 'Failed to pick image. Please try again.');
    }
  };

  const takePhoto = async () => {
    if (disabled) return;

    try {
      // Request camera permissions
      const { status } = await ImagePicker.requestCameraPermissionsAsync();
      if (status !== 'granted') {
        Alert.alert(
          'Permission Denied',
          'Sorry, we need camera permissions to take photos.',
        );
        return;
      }

      const result = await ImagePicker.launchCameraAsync({
        mediaTypes: ImagePicker.MediaTypeOptions.Images,
        allowsEditing: true,
        aspect: [4, 3],
        quality: 1,
      });

      if (!result.canceled) {
        uploadImage(result.assets[0].uri);
      }
    } catch (error) {
      console.error('Error taking photo:', error);
      Alert.alert('Error', 'Failed to take photo. Please try again.');
    }
  };

  const uploadImage = async (uri: string) => {
    setUploading(true);

    try {
      // Validate file type and size
      const fileInfo = await getFileSize(uri);

      if (fileInfo.size > 10 * 1024 * 1024) {
        // 10MB limit
        Alert.alert(
          'File Too Large',
          'Please select an image smaller than 10MB.',
        );
        setUploading(false);
        return;
      }

      // Check file extension
      const fileExtension = uri.split('.').pop()?.toLowerCase();
      const validExtensions = ['jpg', 'jpeg', 'png', 'gif', 'webp'];

      if (!fileExtension || !validExtensions.includes(fileExtension)) {
        Alert.alert(
          'Invalid File Type',
          'Please select a valid image file (JPG, PNG, GIF, WEBP).',
        );
        setUploading(false);
        return;
      }

      // Prepare form data
      const formData = new FormData();
      formData.append('image', {
        uri,
        type: `image/${fileExtension}`,
        name: `gift-image.${fileExtension}`,
      } as unknown as File);

      // Get auth token and upload with authentication
      const token = await (apiClient as any).getAuthToken?.();
      const headers: Record<string, string> = {};
      if (token) {
        headers.Authorization = `Bearer ${token}`;
      }

      const response = await fetch(
        `${process.env.EXPO_PUBLIC_API_URL}/images/upload`,
        {
          method: 'POST',
          headers,
          body: formData,
        },
      );

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`Upload failed: ${errorText}`);
      }

      const result = await response.json();
      setImageUri(result.url);
      onImageUpload(result.url);
      Alert.alert('Success', 'Image uploaded successfully!');
    } catch (error: unknown) {
      console.error('Upload error:', error);
      const errorMessage =
        error instanceof Error
          ? error.message
          : 'Failed to upload image. Please try again.';
      Alert.alert('Upload Failed', errorMessage);
    } finally {
      setUploading(false);
    }
  };

  const getFileSize = async (uri: string): Promise<{ size: number }> => {
    try {
      const info = await FileSystem.getInfoAsync(uri, { size: true });
      return { size: (info as any).size || 0 };
    } catch (error) {
      console.error('Error getting file size:', error);
      return { size: 0 };
    }
  };

  const removeImage = () => {
    setImageUri(null);
    onImageUpload('');
  };

  const theme = useTheme();

  return (
    <View style={{ marginBottom: 20 }}>
      <Text variant="titleMedium" style={{ marginBottom: 8 }}>
        Gift Image
      </Text>

      {imageUri ? (
        <Card style={{ marginBottom: 10 }}>
          <Card.Cover source={{ uri: imageUri }} style={{ height: 200 }} />
          <Card.Actions>
            <Button
              mode="contained-tonal"
              buttonColor={theme.colors.error}
              onPress={removeImage}
              disabled={disabled}
              icon="trash-can"
            >
              Remove
            </Button>
          </Card.Actions>
        </Card>
      ) : (
        <Card
          style={{
            height: 200,
            justifyContent: 'center',
            alignItems: 'center',
            marginBottom: 10,
          }}
        >
          <Ionicons
            name="image-outline"
            size={48}
            color={theme.colors.onSurfaceDisabled}
          />
          <Text
            variant="bodyLarge"
            style={{ color: theme.colors.onSurfaceDisabled, marginTop: 8 }}
          >
            No image selected
          </Text>
        </Card>
      )}

      <View style={{ flexDirection: 'row', justifyContent: 'space-between' }}>
        <Button
          mode="contained"
          onPress={pickImage}
          disabled={disabled || uploading}
          loading={uploading}
          icon="folder-image"
          style={{ flex: 1, marginHorizontal: 2 }}
        >
          Gallery
        </Button>

        <Button
          mode="contained"
          onPress={takePhoto}
          disabled={disabled || uploading}
          loading={uploading}
          icon="camera"
          style={{ flex: 1, marginHorizontal: 2 }}
        >
          Camera
        </Button>
      </View>
    </View>
  );
}
