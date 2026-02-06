import { MaterialCommunityIcons } from '@expo/vector-icons';
import { BlurView } from 'expo-blur';
import { LinearGradient } from 'expo-linear-gradient';
import { Pressable, StyleSheet, View } from 'react-native';
import { Text } from 'react-native-paper';

interface QuickActionsRowProps {
  onAddGift: () => void;
  onNewList: () => void;
  onReserved: () => void;
}

export function QuickActionsRow({
  onAddGift,
  onNewList,
  onReserved,
}: QuickActionsRowProps) {
  return (
    <View style={styles.quickActionsRow}>
      <Pressable onPress={onAddGift} style={styles.quickActionMain}>
        <LinearGradient
          colors={['#FFD700', '#FFA500']}
          start={{ x: 0, y: 0 }}
          end={{ x: 1, y: 1 }}
          style={styles.quickActionGradient}
        >
          <MaterialCommunityIcons name="plus" size={26} color="#000000" />
          <Text style={styles.quickActionMainText}>Add Gift</Text>
        </LinearGradient>
      </Pressable>

      <Pressable onPress={onNewList} style={styles.quickActionSecondary}>
        <BlurView intensity={20} style={styles.quickActionBlur}>
          <MaterialCommunityIcons
            name="playlist-plus"
            size={24}
            color="#6B4EE6"
          />
          <Text style={styles.quickActionSecondaryText}>New List</Text>
        </BlurView>
      </Pressable>

      <Pressable onPress={onReserved} style={styles.quickActionSecondary}>
        <BlurView intensity={20} style={styles.quickActionBlur}>
          <MaterialCommunityIcons
            name="bookmark-outline"
            size={24}
            color="#00A8CC"
          />
          <Text style={styles.quickActionSecondaryText}>Reserved</Text>
        </BlurView>
      </Pressable>
    </View>
  );
}

const styles = StyleSheet.create({
  quickActionsRow: {
    flexDirection: 'row',
    gap: 10,
    marginBottom: 20,
  },
  quickActionMain: {
    flex: 2,
    borderRadius: 16,
    overflow: 'hidden',
    minHeight: 70,
  },
  quickActionGradient: {
    flex: 1,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 20,
    paddingHorizontal: 16,
    gap: 10,
  },
  quickActionMainText: {
    fontSize: 17,
    fontWeight: '700',
    color: '#000000',
    lineHeight: 20,
  },
  quickActionSecondary: {
    flex: 1,
    borderRadius: 16,
    overflow: 'hidden',
    minHeight: 70,
  },
  quickActionBlur: {
    flex: 1,
    backgroundColor: 'rgba(255, 255, 255, 0.08)',
    borderWidth: 1,
    borderColor: 'rgba(255, 255, 255, 0.1)',
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 16,
    paddingHorizontal: 8,
    gap: 6,
  },
  quickActionSecondaryText: {
    fontSize: 12,
    fontWeight: '600',
    color: 'rgba(255, 255, 255, 0.7)',
    lineHeight: 14,
  },
});
