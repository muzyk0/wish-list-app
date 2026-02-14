import { MaterialCommunityIcons } from '@expo/vector-icons';
import { BlurView } from 'expo-blur';
import { useState } from 'react';
import { Pressable, StyleSheet, View } from 'react-native';
import { Text } from 'react-native-paper';
import { TabsLayout } from '@/components/TabsLayout';
import { MyReservations } from '@/components/wish-list/MyReservations';
import { MyWishesReserved } from '@/components/wish-list/MyWishesReserved';

type TabType = 'wishes' | 'reservations';

export default function ReservationsScreen() {
  const [activeTab, setActiveTab] = useState<TabType>('wishes');

  return (
    <TabsLayout title="Reservations" subtitle="Track your reserved items">
      {/* Tab Switcher */}
      <View style={styles.tabContainer}>
        <BlurView intensity={20} style={styles.tabsWrapper}>
          <View style={styles.tabs}>
            <Pressable
              onPress={() => setActiveTab('wishes')}
              style={[styles.tab, activeTab === 'wishes' && styles.activeTab]}
            >
              <MaterialCommunityIcons
                name="gift"
                size={20}
                color={
                  activeTab === 'wishes'
                    ? '#000000'
                    : 'rgba(255, 255, 255, 0.6)'
                }
              />
              <Text
                style={[
                  styles.tabText,
                  activeTab === 'wishes' && styles.activeTabText,
                ]}
              >
                My Wishes
              </Text>
            </Pressable>

            <Pressable
              onPress={() => setActiveTab('reservations')}
              style={[
                styles.tab,
                activeTab === 'reservations' && styles.activeTab,
              ]}
            >
              <MaterialCommunityIcons
                name="bookmark"
                size={20}
                color={
                  activeTab === 'reservations'
                    ? '#000000'
                    : 'rgba(255, 255, 255, 0.6)'
                }
              />
              <Text
                style={[
                  styles.tabText,
                  activeTab === 'reservations' && styles.activeTabText,
                ]}
              >
                I Reserved
              </Text>
            </Pressable>
          </View>
        </BlurView>
      </View>

      {/* Content */}
      <View style={styles.content}>
        {activeTab === 'wishes' ? <MyWishesReserved /> : <MyReservations />}
      </View>
    </TabsLayout>
  );
}

const styles = StyleSheet.create({
  tabContainer: {
    paddingHorizontal: 20,
    marginBottom: 20,
  },
  tabsWrapper: {
    borderRadius: 16,
    overflow: 'hidden',
    backgroundColor: 'rgba(255, 255, 255, 0.08)',
    borderWidth: 1,
    borderColor: 'rgba(255, 255, 255, 0.1)',
  },
  tabs: {
    flexDirection: 'row',
    padding: 4,
  },
  tab: {
    flex: 1,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 12,
    paddingHorizontal: 16,
    borderRadius: 12,
    gap: 8,
  },
  activeTab: {
    backgroundColor: '#FFD700',
  },
  tabText: {
    fontSize: 14,
    fontWeight: '600',
    color: 'rgba(255, 255, 255, 0.6)',
  },
  activeTabText: {
    color: '#000000',
  },
  content: {
    flex: 1,
    paddingHorizontal: 20,
  },
});
