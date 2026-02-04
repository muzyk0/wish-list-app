import { useMutation } from "@tanstack/react-query";
import { useRouter } from "expo-router";
import { useState } from "react";
import { Alert, StyleSheet, View } from "react-native";
import {
  Appbar,
  Button,
  Card,
  Divider,
  Text,
  TextInput,
  useTheme,
} from "react-native-paper";
import OAuthButton from "@/components/OAuthButton";
import { loginUser } from "@/lib/api";
import { setTokens } from "@/lib/api/auth";
import {
  startAppleOAuth,
  startFacebookOAuth,
  startGoogleOAuth,
} from "@/lib/oauth-service";

export default function LoginScreen() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const router = useRouter();
  const { colors } = useTheme();

  const [oauthLoading, setOauthLoading] = useState<
    "google" | "facebook" | "apple" | null
  >(null);

  const mutation = useMutation({
    mutationFn: ({ email, password }: { email: string; password: string }) =>
      loginUser({ email, password }),
    onSuccess: async () => {
      // Tokens are already stored by the loginUser API call
      router.push("/(tabs)");
    },
    onError: (error: Error) => {
      Alert.alert("Error", error.message || "Login failed. Please try again.");
    },
  });

  const handleLogin = () => {
    if (!email || !password) {
      Alert.alert("Error", "Please fill in all required fields.");
      return;
    }

    mutation.mutate({ email, password });
  };

  const handleOAuth = async (provider: "google" | "facebook" | "apple") => {
    setOauthLoading(provider);

    try {
      let result: {
        success: boolean;
        accessToken?: string;
        refreshToken?: string;
        error?: string;
      };
      switch (provider) {
        case "google":
          result = await startGoogleOAuth();
          break;
        case "facebook":
          result = await startFacebookOAuth();
          break;
        case "apple":
          result = await startAppleOAuth();
          break;
        default:
          throw new Error("Invalid provider");
      }

      if (result.success && result.accessToken && result.refreshToken) {
        // Handle successful OAuth login
        try {
          // Store both tokens securely
          await setTokens(result.accessToken, result.refreshToken);
          Alert.alert(
            "Success",
            `${provider.charAt(0).toUpperCase() + provider.slice(1)} login successful!`,
          );
          router.push("/(tabs)"); // Navigate to main app
        } catch (error) {
          console.error("Error storing OAuth tokens:", error);
          Alert.alert(
            "Error",
            "Failed to save authentication. Please try again.",
          );
        }
      } else if (result.error) {
        Alert.alert("OAuth Error", result.error);
      }
      // biome-ignore lint/suspicious/noExplicitAny: Error type
    } catch (error: any) {
      Alert.alert(
        "Error",
        error.message || "An error occurred during OAuth login",
      );
    } finally {
      setOauthLoading(null);
    }
  };

  return (
    <View style={{ flex: 1, backgroundColor: colors.background }}>
      <Appbar.Header style={{ backgroundColor: colors.primary }}>
        <Appbar.BackAction
          onPress={() => router.back()}
          color={colors.onPrimary}
        />
        <Appbar.Content
          title="Sign In"
          titleStyle={{ color: colors.onPrimary }}
        />
      </Appbar.Header>

      <View style={styles.container}>
        <Card style={styles.card}>
          <Card.Content style={styles.cardContent}>
            <View style={styles.header}>
              <Text
                variant="displaySmall"
                style={[styles.title, { color: colors.onSurface }]}
              >
                Welcome Back
              </Text>
              <Text
                variant="bodyLarge"
                style={[styles.subtitle, { color: colors.outline }]}
              >
                Sign in to your account
              </Text>
            </View>

            <TextInput
              label="Email"
              value={email}
              onChangeText={setEmail}
              keyboardType="email-address"
              autoCapitalize="none"
              mode="outlined"
              style={styles.input}
              left={<TextInput.Icon icon="email" />}
            />

            <TextInput
              label="Password"
              value={password}
              onChangeText={setPassword}
              secureTextEntry
              mode="outlined"
              style={styles.input}
              left={<TextInput.Icon icon="lock" />}
            />

            <Button
              mode="contained"
              onPress={handleLogin}
              loading={mutation.isPending}
              disabled={mutation.isPending}
              style={styles.button}
              labelStyle={styles.buttonLabel}
            >
              Sign In
            </Button>

            <View style={styles.dividerContainer}>
              <Divider style={{ flex: 1 }} />
              <Text style={[styles.orText, { color: colors.outline }]}>OR</Text>
              <Divider style={{ flex: 1 }} />
            </View>

            <OAuthButton
              provider="google"
              onPress={() => handleOAuth("google")}
              loading={oauthLoading === "google"}
            />

            <OAuthButton
              provider="facebook"
              onPress={() => handleOAuth("facebook")}
              loading={oauthLoading === "facebook"}
            />

            <OAuthButton
              provider="apple"
              onPress={() => handleOAuth("apple")}
              loading={oauthLoading === "apple"}
            />

            <Text
              variant="bodyMedium"
              style={[styles.footerText, { color: colors.outline }]}
            >
              Don't have an account?{" "}
              <Button
                compact
                mode="text"
                onPress={() => router.push("/auth/register")}
                style={styles.linkButton}
                labelStyle={styles.linkLabel}
              >
                Sign up
              </Button>
            </Text>
          </Card.Content>
        </Card>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    justifyContent: "center",
    padding: 20,
  },
  card: {
    borderRadius: 16,
    elevation: 8,
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.1,
    shadowRadius: 8,
  },
  cardContent: {
    padding: 24,
  },
  header: {
    alignItems: "center",
    marginBottom: 32,
  },
  title: {
    fontSize: 28,
    fontWeight: "bold",
    marginBottom: 8,
  },
  subtitle: {
    fontSize: 16,
    opacity: 0.7,
  },
  input: {
    marginBottom: 16,
  },
  button: {
    marginTop: 8,
    borderRadius: 8,
    paddingVertical: 6,
  },
  buttonLabel: {
    fontWeight: "600",
    fontSize: 16,
  },
  dividerContainer: {
    flexDirection: "row",
    alignItems: "center",
    marginVertical: 20,
  },
  orText: {
    marginHorizontal: 12,
    fontSize: 14,
    opacity: 0.7,
  },
  footerText: {
    textAlign: "center",
    marginTop: 24,
  },
  linkButton: {
    marginLeft: 4,
  },
  linkLabel: {
    fontWeight: "600",
  },
});
