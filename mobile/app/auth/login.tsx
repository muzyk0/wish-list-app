import { MaterialCommunityIcons } from '@expo/vector-icons';
import { zodResolver } from '@hookform/resolvers/zod';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { useMutation } from '@tanstack/react-query';
import { BlurView } from 'expo-blur';
import { LinearGradient } from 'expo-linear-gradient';
import { useRouter } from 'expo-router';
import { useEffect, useRef, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import {
  Animated,
  Dimensions,
  type FlatList,
  Pressable,
  StyleSheet,
  View,
  type ViewToken,
} from 'react-native';
import { Text } from 'react-native-paper';
import { z } from 'zod';
import {
  AuthDivider,
  AuthFooter,
  AuthGradientButton,
  AuthInput,
  AuthLayout,
} from '@/components/auth';
import { OAuthButtonGroup } from '@/components/OAuthButton';
import { useOAuthHandler } from '@/hooks/useOAuthHandler';
import { loginUser } from '@/lib/api';
import { dialog } from '@/stores/dialogStore';

const ONBOARDING_KEY = 'hasSeenOnboarding';
const { width, height } = Dimensions.get('window');

// ─── Onboarding data ─────────────────────────────────────────────────────────

interface OnboardingSlide {
  id: string;
  icon: keyof typeof MaterialCommunityIcons.glyphMap;
  title: string;
  subtitle: string;
  description: string;
  gradientColors: readonly [string, string, string, ...string[]];
  accentColor: string;
  iconBg: string;
}

const slides: OnboardingSlide[] = [
  {
    id: '1',
    icon: 'gift-outline',
    title: 'Create',
    subtitle: 'Your Wish Lists',
    description:
      'Organize your dreams into beautiful collections for any occasion — birthdays, holidays, or just because.',
    gradientColors: ['#1a0a2e', '#2d1b4e', '#4a2c7a', '#6b3fa0'],
    accentColor: '#FFD700',
    iconBg: 'rgba(255, 215, 0, 0.15)',
  },
  {
    id: '2',
    icon: 'heart-multiple-outline',
    title: 'Share',
    subtitle: 'With Loved Ones',
    description:
      'Let friends and family know exactly what makes you happy. No more guessing games.',
    gradientColors: ['#0a1628', '#1a3a5c', '#2a5a8a', '#3a7ab8'],
    accentColor: '#FF6B9D',
    iconBg: 'rgba(255, 107, 157, 0.15)',
  },
  {
    id: '3',
    icon: 'magic-staff',
    title: 'Surprise',
    subtitle: 'Keep The Magic',
    description:
      'Reserve gifts secretly. Coordinate with others without spoiling the surprise.',
    gradientColors: ['#1a0a1a', '#3d1a3d', '#5c2a5c', '#8b3a8b'],
    accentColor: '#00D9FF',
    iconBg: 'rgba(0, 217, 255, 0.15)',
  },
];

// ─── Onboarding overlay ───────────────────────────────────────────────────────

function OnboardingOverlay({ onDone }: { onDone: () => void }) {
  const [currentIndex, setCurrentIndex] = useState(0);
  const flatListRef = useRef<FlatList<OnboardingSlide>>(null);
  const scrollX = useRef(new Animated.Value(0)).current;
  const iconScale = useRef(new Animated.Value(1)).current;
  const buttonScale = useRef(new Animated.Value(1)).current;

  const viewableItemsChanged = useRef(
    ({ viewableItems }: { viewableItems: ViewToken[] }) => {
      if (viewableItems.length > 0 && viewableItems[0].index !== null) {
        setCurrentIndex(viewableItems[0].index);
        Animated.sequence([
          Animated.timing(iconScale, {
            toValue: 0.8,
            duration: 100,
            useNativeDriver: true,
          }),
          Animated.spring(iconScale, {
            toValue: 1,
            tension: 100,
            friction: 5,
            useNativeDriver: true,
          }),
        ]).start();
      }
    },
  ).current;

  const viewConfig = useRef({ viewAreaCoveragePercentThreshold: 50 }).current;

  const getItemLayout = (_: unknown, index: number) => ({
    length: width,
    offset: width * index,
    index,
  });

  const handleNext = () => {
    Animated.sequence([
      Animated.timing(buttonScale, {
        toValue: 0.95,
        duration: 100,
        useNativeDriver: true,
      }),
      Animated.spring(buttonScale, {
        toValue: 1,
        tension: 100,
        friction: 5,
        useNativeDriver: true,
      }),
    ]).start();

    if (currentIndex < slides.length - 1) {
      flatListRef.current?.scrollToIndex({
        index: currentIndex + 1,
        animated: true,
      });
    } else {
      onDone();
    }
  };

  const renderSlide = ({
    item,
    index,
  }: {
    item: OnboardingSlide;
    index: number;
  }) => {
    const inputRange = [
      (index - 1) * width,
      index * width,
      (index + 1) * width,
    ];

    const iconTranslateY = scrollX.interpolate({
      inputRange,
      outputRange: [100, 0, -100],
      extrapolate: 'clamp',
    });

    const textOpacity = scrollX.interpolate({
      inputRange,
      outputRange: [0, 1, 0],
      extrapolate: 'clamp',
    });

    const textTranslateY = scrollX.interpolate({
      inputRange,
      outputRange: [50, 0, -50],
      extrapolate: 'clamp',
    });

    return (
      <View style={onboardingStyles.slide}>
        <LinearGradient
          colors={item.gradientColors}
          start={{ x: 0, y: 0 }}
          end={{ x: 1, y: 1 }}
          style={StyleSheet.absoluteFill}
        />

        <View style={onboardingStyles.slideContent}>
          <Animated.View
            style={[
              onboardingStyles.iconWrapper,
              {
                transform: [
                  { translateY: iconTranslateY },
                  { scale: iconScale },
                ],
              },
            ]}
          >
            <View
              style={[
                onboardingStyles.iconOuter,
                { borderColor: `${item.accentColor}30` },
              ]}
            >
              <View
                style={[
                  onboardingStyles.iconInner,
                  { backgroundColor: item.iconBg },
                ]}
              >
                <MaterialCommunityIcons
                  name={item.icon}
                  size={80}
                  color={item.accentColor}
                />
              </View>
            </View>
            <View
              style={[
                onboardingStyles.glowRing,
                {
                  borderColor: item.accentColor,
                  shadowColor: item.accentColor,
                },
              ]}
            />
          </Animated.View>

          <Animated.View
            style={[
              onboardingStyles.textContainer,
              {
                opacity: textOpacity,
                transform: [{ translateY: textTranslateY }],
              },
            ]}
          >
            <Text style={[onboardingStyles.title, { color: item.accentColor }]}>
              {item.title}
            </Text>
            <Text style={onboardingStyles.subtitle}>{item.subtitle}</Text>
            <Text style={onboardingStyles.description}>{item.description}</Text>
          </Animated.View>
        </View>
      </View>
    );
  };

  const currentSlide = slides[currentIndex];

  return (
    <View style={onboardingStyles.container}>
      {/* Skip */}
      <View style={onboardingStyles.skipContainer}>
        <Pressable onPress={onDone}>
          <BlurView intensity={20} style={onboardingStyles.skipButton}>
            <Text style={onboardingStyles.skipText}>Skip</Text>
          </BlurView>
        </Pressable>
      </View>

      {/* Progress */}
      <View style={onboardingStyles.progressContainer}>
        <Text style={onboardingStyles.progressText}>
          {String(currentIndex + 1).padStart(2, '0')}
        </Text>
        <View style={onboardingStyles.progressLine} />
        <Text style={onboardingStyles.progressTotal}>
          {String(slides.length).padStart(2, '0')}
        </Text>
      </View>

      {/* Slides */}
      <Animated.FlatList
        ref={flatListRef}
        data={slides}
        renderItem={renderSlide}
        horizontal
        pagingEnabled
        showsHorizontalScrollIndicator={false}
        bounces={false}
        keyExtractor={(item) => item.id}
        getItemLayout={getItemLayout}
        onScroll={Animated.event(
          [{ nativeEvent: { contentOffset: { x: scrollX } } }],
          { useNativeDriver: true },
        )}
        onViewableItemsChanged={viewableItemsChanged}
        viewabilityConfig={viewConfig}
        scrollEventThrottle={16}
      />

      {/* Bottom */}
      <View style={onboardingStyles.bottomContainer}>
        <View style={onboardingStyles.dotsContainer}>
          {slides.map((slide, index) => {
            const inputRange = [
              (index - 1) * width,
              index * width,
              (index + 1) * width,
            ];
            const dotScale = scrollX.interpolate({
              inputRange,
              outputRange: [1, 4, 1],
              extrapolate: 'clamp',
            });
            const dotOpacity = scrollX.interpolate({
              inputRange,
              outputRange: [0.3, 1, 0.3],
              extrapolate: 'clamp',
            });
            return (
              <Animated.View
                key={slide.id}
                style={[
                  onboardingStyles.dot,
                  {
                    transform: [{ scaleX: dotScale }],
                    opacity: dotOpacity,
                    backgroundColor:
                      index === currentIndex
                        ? currentSlide.accentColor
                        : 'rgba(255, 255, 255, 0.5)',
                  },
                ]}
              />
            );
          })}
        </View>

        <Animated.View style={{ transform: [{ scale: buttonScale }] }}>
          <Pressable onPress={handleNext}>
            <LinearGradient
              colors={[
                currentSlide.accentColor,
                `${currentSlide.accentColor}CC`,
              ]}
              start={{ x: 0, y: 0 }}
              end={{ x: 1, y: 1 }}
              style={onboardingStyles.nextButton}
            >
              <Text style={onboardingStyles.nextButtonText}>
                {currentIndex < slides.length - 1 ? 'Continue' : 'Get Started'}
              </Text>
              <MaterialCommunityIcons
                name={
                  currentIndex < slides.length - 1
                    ? 'arrow-right'
                    : 'rocket-launch'
                }
                size={24}
                color="#000000"
              />
            </LinearGradient>
          </Pressable>
        </Animated.View>
      </View>
    </View>
  );
}

// ─── Login form ───────────────────────────────────────────────────────────────

const loginSchema = z.object({
  email: z.string().min(1, 'Email is required').email('Invalid email address'),
  password: z
    .string()
    .min(6, 'Password must be at least 6 characters')
    .max(100, 'Password is too long'),
});

type LoginFormData = z.infer<typeof loginSchema>;

export default function LoginScreen() {
  const [showOnboarding, setShowOnboarding] = useState<boolean | null>(null);
  const overlayTranslateY = useRef(new Animated.Value(0)).current;

  const router = useRouter();
  const { oauthLoading, handleOAuth } = useOAuthHandler();

  const {
    control,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
    defaultValues: { email: '', password: '' },
  });

  const [showPassword, setShowPassword] = useState(false);

  const mutation = useMutation({
    mutationFn: ({ email, password }: { email: string; password: string }) =>
      loginUser({ email, password }),
    onSuccess: () => router.replace('/(tabs)'),
    onError: (error: Error) => {
      dialog.error(error.message || 'Login failed. Please try again.');
    },
  });

  useEffect(() => {
    AsyncStorage.getItem(ONBOARDING_KEY).then((val) => {
      setShowOnboarding(!val);
    });
  }, []);

  const handleOnboardingDone = async () => {
    await AsyncStorage.setItem(ONBOARDING_KEY, 'true');
    Animated.timing(overlayTranslateY, {
      toValue: -height,
      duration: 480,
      useNativeDriver: true,
    }).start(() => setShowOnboarding(false));
  };

  return (
    <View style={{ flex: 1 }}>
      {/* Login form — always mounted, visible once overlay leaves */}
      <AuthLayout title="Wish List" subtitle="Welcome back!">
        <Controller
          control={control}
          name="email"
          render={({ field: { onChange, onBlur, value } }) => (
            <AuthInput
              testID="login-email-input"
              placeholder="Email"
              value={value}
              onChangeText={onChange}
              onBlur={onBlur}
              icon="email-outline"
              keyboardType="email-address"
              error={errors.email?.message}
            />
          )}
        />

        <Controller
          control={control}
          name="password"
          render={({ field: { onChange, onBlur, value } }) => (
            <AuthInput
              testID="login-password-input"
              placeholder="Password"
              value={value}
              onChangeText={onChange}
              onBlur={onBlur}
              icon="lock-outline"
              secureTextEntry
              showPasswordToggle
              showPassword={showPassword}
              onTogglePassword={() => setShowPassword((p) => !p)}
              error={errors.password?.message}
            />
          )}
        />

        <AuthGradientButton
          testID="login-submit-button"
          label="Sign In"
          loadingLabel="Signing in..."
          loading={mutation.isPending}
          onPress={handleSubmit((data) => mutation.mutate(data))}
        />

        <AuthDivider />

        <OAuthButtonGroup
          onGooglePress={() => handleOAuth('google')}
          onApplePress={() => handleOAuth('apple')}
          onFacebookPress={() => handleOAuth('facebook')}
          loadingProvider={oauthLoading}
        />

        <AuthFooter
          text="Don't have an account? "
          linkText="Create one"
          onLinkPress={() => router.push('/auth/register')}
        />
      </AuthLayout>

      {/* Onboarding overlay — slides up to reveal login */}
      {showOnboarding && (
        <Animated.View
          style={[
            StyleSheet.absoluteFillObject,
            { transform: [{ translateY: overlayTranslateY }] },
          ]}
        >
          <OnboardingOverlay onDone={handleOnboardingDone} />
        </Animated.View>
      )}
    </View>
  );
}

// ─── Onboarding styles ────────────────────────────────────────────────────────

const onboardingStyles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#000000',
  },
  skipContainer: {
    position: 'absolute',
    top: 60,
    right: 20,
    zIndex: 10,
  },
  skipButton: {
    paddingHorizontal: 20,
    paddingVertical: 10,
    borderRadius: 20,
    overflow: 'hidden',
    backgroundColor: 'rgba(255, 255, 255, 0.1)',
  },
  skipText: {
    color: 'rgba(255, 255, 255, 0.7)',
    fontSize: 14,
    fontWeight: '600',
    letterSpacing: 1,
  },
  progressContainer: {
    position: 'absolute',
    top: 64,
    left: 24,
    zIndex: 10,
    flexDirection: 'row',
    alignItems: 'center',
  },
  progressText: {
    color: '#ffffff',
    fontSize: 24,
    fontWeight: '700',
  },
  progressLine: {
    width: 20,
    height: 2,
    backgroundColor: 'rgba(255, 255, 255, 0.3)',
    marginHorizontal: 8,
  },
  progressTotal: {
    color: 'rgba(255, 255, 255, 0.4)',
    fontSize: 14,
    fontWeight: '500',
  },
  slide: {
    width,
    height,
  },
  slideContent: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingHorizontal: 32,
    paddingBottom: 180,
  },
  iconWrapper: {
    marginBottom: 48,
    alignItems: 'center',
    justifyContent: 'center',
  },
  iconOuter: {
    width: 180,
    height: 180,
    borderRadius: 90,
    borderWidth: 2,
    justifyContent: 'center',
    alignItems: 'center',
  },
  iconInner: {
    width: 140,
    height: 140,
    borderRadius: 70,
    justifyContent: 'center',
    alignItems: 'center',
  },
  glowRing: {
    position: 'absolute',
    width: 200,
    height: 200,
    borderRadius: 100,
    borderWidth: 1,
    opacity: 0.3,
    shadowOffset: { width: 0, height: 0 },
    shadowOpacity: 0.5,
    shadowRadius: 20,
  },
  textContainer: {
    alignItems: 'center',
  },
  title: {
    fontSize: 48,
    fontWeight: '800',
    letterSpacing: 2,
    marginBottom: 4,
  },
  subtitle: {
    fontSize: 28,
    fontWeight: '300',
    color: '#ffffff',
    marginBottom: 24,
    letterSpacing: 1,
  },
  description: {
    fontSize: 16,
    color: 'rgba(255, 255, 255, 0.7)',
    textAlign: 'center',
    lineHeight: 26,
    maxWidth: 320,
  },
  bottomContainer: {
    position: 'absolute',
    bottom: 60,
    left: 0,
    right: 0,
    paddingHorizontal: 32,
  },
  dotsContainer: {
    flexDirection: 'row',
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: 32,
  },
  dot: {
    height: 8,
    borderRadius: 4,
    marginHorizontal: 4,
  },
  nextButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 18,
    paddingHorizontal: 32,
    borderRadius: 30,
    gap: 12,
  },
  nextButtonText: {
    fontSize: 18,
    fontWeight: '700',
    color: '#000000',
    letterSpacing: 1,
  },
});
