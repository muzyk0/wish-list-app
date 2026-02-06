import { useRouter } from 'expo-router';
import { useState } from 'react';
import { Alert } from 'react-native';
import { setTokens } from '@/lib/api/auth';
import {
  startAppleOAuth,
  startFacebookOAuth,
  startGoogleOAuth,
} from '@/lib/oauth-service';

type OAuthProvider = 'google' | 'facebook' | 'apple';

export function useOAuthHandler() {
  const router = useRouter();
  const [oauthLoading, setOauthLoading] = useState<OAuthProvider | null>(null);

  const handleOAuth = async (provider: OAuthProvider) => {
    setOauthLoading(provider);

    try {
      let result: {
        success: boolean;
        accessToken?: string;
        refreshToken?: string;
        error?: string;
      };

      switch (provider) {
        case 'google':
          result = await startGoogleOAuth();
          break;
        case 'facebook':
          result = await startFacebookOAuth();
          break;
        case 'apple':
          result = await startAppleOAuth();
          break;
        default:
          throw new Error('Invalid provider');
      }

      if (result.success && result.accessToken && result.refreshToken) {
        try {
          await setTokens(result.accessToken, result.refreshToken);
          router.replace('/(tabs)');
        } catch (error) {
          console.error('Error storing OAuth tokens:', error);
          Alert.alert(
            'Error',
            'Failed to save authentication. Please try again.',
          );
        }
      } else if (result.error) {
        Alert.alert('OAuth Error', result.error);
      }
    } catch (error: any) {
      Alert.alert('Error', error.message || 'An error occurred during OAuth');
    } finally {
      setOauthLoading(null);
    }
  };

  return {
    oauthLoading,
    handleOAuth,
  };
}
