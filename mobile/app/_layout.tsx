import * as Linking from 'expo-linking';
import { Stack, useRouter } from 'expo-router';
import { StatusBar } from 'expo-status-bar';
import { useEffect } from 'react';
import 'react-native-reanimated';

import Providers from '@/app/providers';
import { exchangeCodeForTokens } from '@/lib/api/auth';

export const unstable_settings = {
  initialRouteName: 'splash',
};

export default function RootLayout() {
  const router = useRouter();

  useEffect(() => {
    // Handle deep links
    const handleDeepLink = async (event: { url: string }) => {
      const { path, queryParams } = Linking.parse(event.url);

      // Handle auth deep link: wishlistapp://auth?code=xxx
      if (path === 'auth' && queryParams?.code) {
        try {
          await exchangeCodeForTokens(queryParams.code as string);
          // Navigate to home after successful authentication
          router.replace('/(tabs)');
        } catch (error) {
          console.error('Failed to exchange auth code:', error);
          // Navigate to login on error
          router.replace('/auth/login');
        }
        return;
      }

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

        // Handle parameterized routes with type-safe navigation
        if (path.startsWith('lists/') && path.includes('/edit')) {
          const match = path.match(/^lists\/([^/]+)\/edit/);
          if (match?.[1]) {
            router.navigate({
              pathname: '/lists/[id]/edit',
              params: { id: match[1] },
            });
          }
          return;
        }

        if (path.startsWith('lists/')) {
          const match = path.match(/^lists\/([^/]+)/);
          if (match?.[1]) {
            router.navigate({
              pathname: '/lists/[id]',
              params: { id: match[1] },
            });
          }
          return;
        }

        if (path.startsWith('public/')) {
          const match = path.match(/^public\/([^/]+)/);
          if (match?.[1]) {
            router.navigate({
              pathname: '/public/[slug]' as any,
              params: { slug: match[1] },
            });
          }
          return;
        }

        if (path.startsWith('gift-items/') && path.includes('/edit')) {
          const match = path.match(/^gift-items\/([^/]+)\/edit/);
          if (match?.[1]) {
            router.navigate({
              pathname: '/gift-items/[id]/edit',
              params: { id: match[1] },
            });
          }
          return;
        }

        // Handle mapped routes (static routes)
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
          name="splash"
          options={{
            headerShown: false,
            animation: 'fade',
          }}
        />
        <Stack.Screen
          name="onboarding"
          options={{
            headerShown: false,
            animation: 'slide_from_right',
          }}
        />
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
