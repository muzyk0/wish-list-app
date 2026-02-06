import { MaterialCommunityIcons } from '@expo/vector-icons';
import { BlurView } from 'expo-blur';
import { LinearGradient } from 'expo-linear-gradient';
import { useEffect, useRef } from 'react';
import { Animated, Pressable, StyleSheet, View } from 'react-native';
import { Text } from 'react-native-paper';
import type { GiftItem } from '@/lib/api/types';

interface GiftItemCardProps {
  item: GiftItem;
  listTitle: string;
  onPress: () => void;
  index: number;
}

export function GiftItemCard({
  item,
  listTitle,
  onPress,
  index,
}: GiftItemCardProps) {
  const scaleAnim = useRef(new Animated.Value(0.9)).current;
  const opacityAnim = useRef(new Animated.Value(0)).current;
  const pressScale = useRef(new Animated.Value(1)).current;

  useEffect(() => {
    Animated.sequence([
      Animated.delay(index * 100),
      Animated.parallel([
        Animated.spring(scaleAnim, {
          toValue: 1,
          tension: 50,
          friction: 7,
          useNativeDriver: true,
        }),
        Animated.timing(opacityAnim, {
          toValue: 1,
          duration: 400,
          useNativeDriver: true,
        }),
      ]),
    ]).start();
  }, [index, scaleAnim, opacityAnim]);

  const handlePressIn = () => {
    Animated.spring(pressScale, {
      toValue: 0.97,
      useNativeDriver: true,
    }).start();
  };

  const handlePressOut = () => {
    Animated.spring(pressScale, {
      toValue: 1,
      useNativeDriver: true,
    }).start();
  };

  const isReserved = !!item.reserved_by_user_id;
  const isPurchased = !!item.purchased_by_user_id;

  return (
    <Animated.View
      style={{
        opacity: opacityAnim,
        transform: [{ scale: scaleAnim }],
      }}
    >
      <Animated.View style={{ transform: [{ scale: pressScale }] }}>
        <Pressable
          onPress={onPress}
          onPressIn={handlePressIn}
          onPressOut={handlePressOut}
        >
          <BlurView intensity={20} style={styles.giftCard}>
            <View style={styles.giftCardContent}>
              {/* Gift Icon with Gradient */}
              <LinearGradient
                colors={['#FFD700', '#FFA500']}
                style={styles.giftIconContainer}
              >
                <MaterialCommunityIcons name="gift" size={24} color="#000000" />
              </LinearGradient>

              {/* Gift Info */}
              <View style={styles.giftInfo}>
                <Text style={styles.giftName} numberOfLines={1}>
                  {item.name}
                </Text>
                <View style={styles.giftMeta}>
                  <MaterialCommunityIcons
                    name="format-list-bulleted"
                    size={12}
                    color="rgba(255, 255, 255, 0.5)"
                  />
                  <Text style={styles.giftList} numberOfLines={1}>
                    {listTitle}
                  </Text>
                </View>
              </View>

              {/* Price & Status */}
              <View style={styles.giftRight}>
                {item.price !== undefined && item.price !== null && (
                  <View style={styles.priceTag}>
                    <Text style={styles.priceText}>
                      ${item.price.toFixed(0)}
                    </Text>
                  </View>
                )}
                {isPurchased ? (
                  <View style={[styles.statusBadge, styles.purchasedBadge]}>
                    <MaterialCommunityIcons
                      name="check-circle"
                      size={12}
                      color="#4CAF50"
                    />
                  </View>
                ) : isReserved ? (
                  <View style={[styles.statusBadge, styles.reservedBadge]}>
                    <MaterialCommunityIcons
                      name="lock"
                      size={12}
                      color="#FF9800"
                    />
                  </View>
                ) : (
                  <MaterialCommunityIcons
                    name="chevron-right"
                    size={20}
                    color="rgba(255, 255, 255, 0.3)"
                  />
                )}
              </View>
            </View>
          </BlurView>
        </Pressable>
      </Animated.View>
    </Animated.View>
  );
}

const styles = StyleSheet.create({
  giftCard: {
    borderRadius: 14,
    overflow: 'hidden',
    backgroundColor: 'rgba(255, 255, 255, 0.08)',
    borderWidth: 1,
    borderColor: 'rgba(255, 255, 255, 0.1)',
    marginBottom: 10,
  },
  giftCardContent: {
    flexDirection: 'row',
    alignItems: 'center',
    padding: 14,
  },
  giftIconContainer: {
    width: 48,
    height: 48,
    borderRadius: 12,
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 12,
  },
  giftInfo: {
    flex: 1,
    marginRight: 8,
  },
  giftName: {
    fontSize: 15,
    fontWeight: '600',
    color: '#ffffff',
    marginBottom: 4,
  },
  giftMeta: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
  },
  giftList: {
    fontSize: 12,
    color: 'rgba(255, 255, 255, 0.5)',
    flex: 1,
  },
  giftRight: {
    alignItems: 'flex-end',
    gap: 6,
  },
  priceTag: {
    backgroundColor: 'rgba(255, 215, 0, 0.15)',
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: 8,
  },
  priceText: {
    fontSize: 13,
    fontWeight: '700',
    color: '#FFD700',
  },
  statusBadge: {
    width: 24,
    height: 24,
    borderRadius: 12,
    justifyContent: 'center',
    alignItems: 'center',
  },
  reservedBadge: {
    backgroundColor: 'rgba(255, 152, 0, 0.15)',
  },
  purchasedBadge: {
    backgroundColor: 'rgba(76, 175, 80, 0.15)',
  },
});
