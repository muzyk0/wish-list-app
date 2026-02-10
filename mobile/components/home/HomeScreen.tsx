import { useRouter } from 'expo-router';
import { ScrollView, StyleSheet } from 'react-native';
import {
  QuickActionsRow,
  RecentItemsSection,
  StatsRow,
  YourListsSection,
} from '@/components/home';
import { TabsLayout } from '@/components/TabsLayout';

export function HomeScreen() {
  const router = useRouter();

  return (
    <TabsLayout title="My Gifts" subtitle="Your wishlist items">
      <ScrollView
        showsVerticalScrollIndicator={false}
        contentContainerStyle={styles.scrollContent}
      >
        <QuickActionsRow
          onAddGift={() => router.push('/lists')}
          onNewList={() => router.push('/lists/create')}
          onReserved={() => router.push('/reservations')}
        />

        <StatsRow />

        <RecentItemsSection
          onSeeAll={() => router.push('/lists')}
          onItemPress={(listId) => router.push(`/lists/${listId}`)}
        />

        <YourListsSection
          onAddPress={() => router.push('/lists')}
          onListPress={(listId) => router.push(`/lists/${listId}`)}
        />
      </ScrollView>
    </TabsLayout>
  );
}

const styles = StyleSheet.create({
  scrollContent: {
    paddingHorizontal: 20,
    paddingBottom: 100,
  },
});
