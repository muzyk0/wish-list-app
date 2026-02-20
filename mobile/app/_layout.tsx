import * as Linking from 'expo-linking';
import { Stack, useRouter } from 'expo-router';
import { StatusBar } from 'expo-status-bar';
import { useEffect } from 'react';
import 'react-native-reanimated';

import Providers from '@/app/providers';
import { exchangeCodeForTokens } from '@/lib/api/auth';

export default function RootLayout() {
  const router = useRouter();

  useEffect(() => {
    // Handle native deep links only (e.g. wishlistapp://auth?code=xxx).
    // Web URL routing is handled by Expo Router natively — no mapping needed.
    const handleDeepLink = async (event: { url: string }) => {
      const { scheme, path, queryParams } = Linking.parse(event.url);

      // Only handle custom scheme deep links, not http/https web URLs
      if (scheme === 'http' || scheme === 'https') return;

      if (path === 'auth' && queryParams?.code) {
        try {
          await exchangeCodeForTokens(queryParams.code as string);
          router.replace('/(tabs)');
        } catch (error) {
          console.error('Failed to exchange auth code:', error);
          router.replace('/auth/login');
        }
      }
    };

    const subscription = Linking.addEventListener('url', handleDeepLink);
    return () => subscription.remove();
  }, [router]);

  return (
    <Providers>
      <Stack screenOptions={{ headerShown: false }}>
        <Stack.Screen name="splash" options={{ animation: 'fade' }} />
        <Stack.Screen
          name="modal"
          options={{ presentation: 'modal', headerShown: true, title: 'Modal' }}
        />
      </Stack>
      <StatusBar style="auto" />
    </Providers>
  );
}
