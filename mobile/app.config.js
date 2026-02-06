// mobile/app.config.js
// Expo configuration with environment variable support

const WEB_DOMAIN = process.env.EXPO_PUBLIC_WEB_DOMAIN || 'wishlist.com';
const WWW_DOMAIN = process.env.EXPO_PUBLIC_WWW_DOMAIN || `www.${WEB_DOMAIN}`;

module.exports = {
  expo: {
    name: 'WishList',
    slug: 'wishlist',
    version: '1.0.0',
    orientation: 'portrait',
    icon: './assets/images/icon.png',
    scheme: 'wishlistapp',
    userInterfaceStyle: 'automatic',
    newArchEnabled: true,
    ios: {
      supportsTablet: true,
      bundleIdentifier: 'com.anonymous.mobile',
      associatedDomains: [`applinks:${WEB_DOMAIN}`, `applinks:${WWW_DOMAIN}`],
    },
    android: {
      adaptiveIcon: {
        backgroundColor: '#E6F4FE',
        foregroundImage: './assets/images/android-icon-foreground.png',
        backgroundImage: './assets/images/android-icon-background.png',
        monochromeImage: './assets/images/android-icon-monochrome.png',
      },
      edgeToEdgeEnabled: true,
      predictiveBackGestureEnabled: false,
      package: 'com.anonymous.mobile',
      intentFilters: [
        {
          action: 'VIEW',
          autoVerify: true,
          data: [
            {
              scheme: 'https',
              host: WEB_DOMAIN,
            },
            {
              scheme: 'https',
              host: WWW_DOMAIN,
            },
            {
              scheme: 'wishlistapp',
              host: '*',
            },
          ],
          category: ['BROWSABLE', 'DEFAULT'],
        },
      ],
    },
    web: {
      output: 'single',
      bundler: 'metro',
      favicon: './assets/images/favicon.png',
    },
    plugins: [
      'expo-router',
      [
        'expo-splash-screen',
        {
          image: './assets/images/splash-icon.png',
          imageWidth: 200,
          resizeMode: 'contain',
          backgroundColor: '#ffffff',
          dark: {
            backgroundColor: '#000000',
          },
        },
      ],
    ],
    experiments: {
      typedRoutes: true,
      reactCompiler: true,
    },
  },
};
