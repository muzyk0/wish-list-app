import { MaterialCommunityIcons } from '@expo/vector-icons';
import { BlurView } from 'expo-blur';
import { Redirect, Tabs } from 'expo-router';
import { useEffect, useState } from 'react';
import { Platform, StyleSheet, View } from 'react-native';
import { isAuthenticated } from '@/lib/api/auth';

export default function TabLayout() {
  const [authState, setAuthState] = useState<
    'loading' | 'authenticated' | 'unauthenticated'
  >('loading');

  useEffect(() => {
    isAuthenticated().then((ok) =>
      setAuthState(ok ? 'authenticated' : 'unauthenticated'),
    );
  }, []);

  if (authState === 'loading') {
    // Blank screen matching app background — no flash of content
    return <View style={{ flex: 1, backgroundColor: '#1a0a2e' }} />;
  }

  if (authState === 'unauthenticated') {
    return <Redirect href="/auth/login" />;
  }

  return (
    <Tabs
      screenOptions={{
        tabBarActiveTintColor: '#FFD700',
        tabBarInactiveTintColor: 'rgba(255, 255, 255, 0.5)',
        headerShown: false,
        tabBarStyle: {
          position: 'absolute',
          backgroundColor:
            Platform.OS === 'ios' ? 'transparent' : 'rgba(26, 10, 46, 0.95)',
          borderTopWidth: 0,
          elevation: 0,
          height: 85,
          paddingBottom: 25,
          paddingTop: 10,
        },
        tabBarBackground: () =>
          Platform.OS === 'ios' ? (
            <BlurView
              intensity={80}
              style={StyleSheet.absoluteFill}
              tint="dark"
            />
          ) : null,
        tabBarLabelStyle: {
          fontSize: 12,
          fontWeight: '600',
          marginTop: 4,
        },
      }}
    >
      <Tabs.Screen
        name="index"
        options={{
          title: 'Home',
          tabBarIcon: ({ color, focused }) => (
            <MaterialCommunityIcons
              name={focused ? 'home' : 'home-outline'}
              size={focused ? 28 : 24}
              color={color}
            />
          ),
        }}
      />
      <Tabs.Screen
        name="gifts"
        options={{
          title: 'Gifts',
          tabBarIcon: ({ color, focused }) => (
            <MaterialCommunityIcons
              name={focused ? 'gift' : 'gift-outline'}
              size={focused ? 28 : 24}
              color={color}
            />
          ),
        }}
      />
      <Tabs.Screen
        name="lists"
        options={{
          title: 'Lists',
          tabBarIcon: ({ color, focused }) => (
            <MaterialCommunityIcons
              name={focused ? 'format-list-bulleted' : 'format-list-bulleted'}
              size={focused ? 28 : 24}
              color={color}
            />
          ),
        }}
      />
      <Tabs.Screen
        name="reservations"
        options={{
          title: 'Reservations',
          tabBarIcon: ({ color, focused }) => (
            <MaterialCommunityIcons
              name={focused ? 'bookmark' : 'bookmark-outline'}
              size={focused ? 28 : 24}
              color={color}
            />
          ),
        }}
      />
      <Tabs.Screen
        name="profile"
        options={{
          title: 'Profile',
          tabBarIcon: ({ color, focused }) => (
            <MaterialCommunityIcons
              name={focused ? 'account' : 'account-outline'}
              size={focused ? 28 : 24}
              color={color}
            />
          ),
        }}
      />
    </Tabs>
  );
}
