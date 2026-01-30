import { useMutation, useQuery } from '@tanstack/react-query';
import { useState } from 'react';
import {
  ActivityIndicator,
  Alert,
  FlatList,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';
import { apiClient } from '@/lib/api';

interface TemplateSelectorProps {
  wishlistId: string;
  currentTemplateId: string;
  onTemplateChange?: (templateId: string) => void;
}

export default function TemplateSelector({
  wishlistId,
  currentTemplateId,
  onTemplateChange,
}: TemplateSelectorProps) {
  const [selectedTemplateId, setSelectedTemplateId] =
    useState(currentTemplateId);

  const {
    data: templates,
    isLoading,
    isError,
    refetch,
    // @ts-expect-error
  } = useQuery<Template[]>({
    queryKey: ['templates'],
    // @ts-expect-error
    queryFn: () => apiClient.getTemplates(),
  });

  const updateTemplateMutation = useMutation({
    mutationFn: (templateId: string) =>
      // @ts-expect-error
      apiClient.updateWishListTemplate(wishlistId, templateId),
    onSuccess: (_data, templateId) => {
      Alert.alert('Success', 'Template updated successfully!');
      if (onTemplateChange) {
        onTemplateChange(templateId);
      }
    },
    onError: (error) => {
      Alert.alert(
        'Error',
        error.message || 'Failed to update template. Please try again.',
      );
    },
  });

  const handleSelectTemplate = (templateId: string) => {
    setSelectedTemplateId(templateId);
    updateTemplateMutation.mutate(templateId);
  };

  // @ts-expect-error
  const renderTemplate = ({ item }: { item: Template }) => (
    <TouchableOpacity
      style={[
        styles.templateItem,
        item.id === selectedTemplateId && styles.selectedTemplate,
      ]}
      onPress={() => handleSelectTemplate(item.id)}
      disabled={updateTemplateMutation.isPending}
    >
      <View style={styles.templateContent}>
        <Text style={styles.templateName}>{item.name}</Text>
        <Text style={styles.templateDescription} numberOfLines={2}>
          {item.description}
        </Text>
        {item.is_default && <Text style={styles.defaultLabel}>Default</Text>}
      </View>
      {updateTemplateMutation.isPending && item.id === selectedTemplateId && (
        <ActivityIndicator size="small" color="#007AFF" />
      )}
    </TouchableOpacity>
  );

  if (isLoading) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color="#007AFF" />
        <Text>Loading templates...</Text>
      </View>
    );
  }

  if (isError) {
    return (
      <View style={styles.errorContainer}>
        <Text>Error loading templates</Text>
        <TouchableOpacity onPress={() => refetch()}>
          <Text style={styles.retryText}>Retry</Text>
        </TouchableOpacity>
      </View>
    );
  }

  return (
    <View style={styles.container}>
      <Text style={styles.title}>Select Template</Text>

      <FlatList
        data={templates}
        renderItem={renderTemplate}
        keyExtractor={(item) => item.id}
        horizontal={false}
        showsVerticalScrollIndicator={false}
        contentContainerStyle={styles.listContainer}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    marginVertical: 10,
  },
  title: {
    fontSize: 18,
    fontWeight: 'bold',
    marginBottom: 10,
  },
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 20,
  },
  errorContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 20,
  },
  retryText: {
    color: '#007AFF',
    marginTop: 10,
  },
  listContainer: {
    paddingBottom: 10,
  },
  templateItem: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 15,
    marginVertical: 5,
    backgroundColor: '#f9f9f9',
    borderRadius: 8,
    borderWidth: 1,
    borderColor: '#ddd',
  },
  selectedTemplate: {
    borderColor: '#007AFF',
    backgroundColor: '#e6f0ff',
  },
  templateContent: {
    flex: 1,
  },
  templateName: {
    fontSize: 16,
    fontWeight: '600',
    marginBottom: 5,
  },
  templateDescription: {
    fontSize: 14,
    color: '#666',
    marginBottom: 5,
  },
  defaultLabel: {
    fontSize: 12,
    color: '#007AFF',
    alignSelf: 'flex-start',
    backgroundColor: '#e6f0ff',
    paddingHorizontal: 6,
    paddingVertical: 2,
    borderRadius: 4,
  },
});
