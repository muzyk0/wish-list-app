import { useRouter } from 'expo-router';
import { useState } from 'react';
import { setTokens } from '@/lib/api/auth';
import {
  startAppleOAuth,
  startFacebookOAuth,
  startGoogleOAuth,
} from '@/lib/oauth-service';
import { dialog } from '@/stores/dialogStore';

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
          dialog.error('Failed to save authentication. Please try again.');
        }
      } else if (result.error) {
        dialog.error(result.error, 'OAuth Error');
      }
    } catch (error: any) {
      dialog.error(error.message || 'An error occurred during OAuth');
    } finally {
      setOauthLoading(null);
    }
  };

  return {
    oauthLoading,
    handleOAuth,
  };
}
