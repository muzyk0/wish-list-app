import { MaterialCommunityIcons } from '@expo/vector-icons';
import { BlurView } from 'expo-blur';
import { useEffect, useRef } from 'react';
import { Animated, Pressable, StyleSheet, View } from 'react-native';
import { Text } from 'react-native-paper';

interface CompactListCardProps {
  title: string;
  itemCount: number;
  onPress: () => void;
  index: number;
}

export function CompactListCard({
  title,
  itemCount,
  onPress,
  index,
}: CompactListCardProps) {
  const translateX = useRef(new Animated.Value(30)).current;
  const opacity = useRef(new Animated.Value(0)).current;

  useEffect(() => {
    Animated.sequence([
      Animated.delay(100 + index * 80),
      Animated.parallel([
        Animated.timing(translateX, {
          toValue: 0,
          duration: 300,
          useNativeDriver: true,
        }),
        Animated.timing(opacity, {
          toValue: 1,
          duration: 300,
          useNativeDriver: true,
        }),
      ]),
    ]).start();
  }, [index, translateX, opacity]);

  return (
    <Animated.View
      style={{
        opacity,
        transform: [{ translateX }],
      }}
    >
      <Pressable onPress={onPress}>
        <BlurView intensity={15} style={styles.compactListCard}>
          <View style={styles.compactListContent}>
            <View style={styles.compactListIcon}>
              <MaterialCommunityIcons
                name="gift-outline"
                size={16}
                color="#6B4EE6"
              />
            </View>
            <View style={styles.compactListInfo}>
              <Text style={styles.compactListTitle} numberOfLines={1}>
                {title}
              </Text>
              <Text style={styles.compactListMeta}>
                {itemCount} {itemCount === 1 ? 'item' : 'items'}
              </Text>
            </View>
            <MaterialCommunityIcons
              name="chevron-right"
              size={18}
              color="rgba(255, 255, 255, 0.4)"
            />
          </View>
        </BlurView>
      </Pressable>
    </Animated.View>
  );
}

const styles = StyleSheet.create({
  compactListCard: {
    borderRadius: 12,
    overflow: 'hidden',
    backgroundColor: 'rgba(255, 255, 255, 0.05)',
    borderWidth: 1,
    borderColor: 'rgba(255, 255, 255, 0.08)',
  },
  compactListContent: {
    flexDirection: 'row',
    alignItems: 'center',
    padding: 12,
  },
  compactListIcon: {
    width: 32,
    height: 32,
    borderRadius: 8,
    backgroundColor: 'rgba(107, 78, 230, 0.15)',
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 10,
  },
  compactListInfo: {
    flex: 1,
  },
  compactListTitle: {
    fontSize: 14,
    fontWeight: '600',
    color: '#ffffff',
    marginBottom: 2,
  },
  compactListMeta: {
    fontSize: 11,
    color: 'rgba(255, 255, 255, 0.5)',
  },
});
