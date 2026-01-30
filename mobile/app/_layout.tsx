import { Stack, useRouter } from 'expo-router';
import { StatusBar } from 'expo-status-bar';
import { useEffect } from 'react';
import * as Linking from 'expo-linking';
import 'react-native-reanimated';

import Providers from '@/app/providers';

export const unstable_settings = {
  initialRouteName: '(tabs)',
};

export default function RootLayout() {
  const router = useRouter();

  useEffect(() => {
    // Handle deep links
    const handleDeepLink = (event: { url: string }) => {
      const { path, queryParams } = Linking.parse(event.url);

      if (path) {
        // Map web paths to mobile routes
        const routeMap: { [key: string]: string } = {
          home: '/(tabs)',
          'auth/login': '/auth/login',
          'auth/register': '/auth/register',
          'my/reservations': '/(tabs)/reservations',
          lists: '/(tabs)/lists',
          explore: '/(tabs)/explore',
          profile: '/(tabs)/profile',
        };

        // Handle parameterized routes
        if (path.startsWith('lists/') && path.includes('/edit')) {
          const id = path.split('/')[1];
          router.push(`/lists/${id}/edit` as any);
          return;
        }

        if (path.startsWith('lists/')) {
          const id = path.split('/')[1];
          router.push(`/lists/${id}` as any);
          return;
        }

        if (path.startsWith('public/')) {
          const slug = path.split('/')[1];
          router.push(`/public/${slug}` as any);
          return;
        }

        if (path.startsWith('gift-items/') && path.includes('/edit')) {
          const match = path.match(/^gift-items\/([^\/]+)\/edit/);
          if (match && match[1]) {
            const id = match[1];
            router.push(`/gift-items/${id}/edit` as any);
          }
          return;
        }

        // Handle mapped routes
        const targetRoute = routeMap[path];
        if (targetRoute) {
          router.push(targetRoute as any);
        }
      }
    };

    // Get initial URL (cold start)
    Linking.getInitialURL().then((url) => {
      if (url) {
        handleDeepLink({ url });
      }
    });

    // Listen for deep link events (warm start)
    const subscription = Linking.addEventListener('url', handleDeepLink);

    return () => {
      subscription.remove();
    };
  }, [router]);

  return (
    <Providers>
      <Stack>
        <Stack.Screen
          name="(tabs)"
          options={{
            headerShown: false,
          }}
        />
        <Stack.Screen name="auth" options={{ headerShown: false }} />
        <Stack.Screen
          name="modal"
          options={{ presentation: 'modal', title: 'Modal' }}
        />
      </Stack>
      <StatusBar style="auto" />
    </Providers>
  );
}
