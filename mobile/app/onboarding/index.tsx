import { MaterialCommunityIcons } from '@expo/vector-icons';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { BlurView } from 'expo-blur';
import { LinearGradient } from 'expo-linear-gradient';
import { useRouter } from 'expo-router';
import { useEffect, useRef, useState } from 'react';
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

const { width, height } = Dimensions.get('window');
const ONBOARDING_KEY = 'hasSeenOnboarding';

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

// Animated floating orb
const FloatingOrb = ({
  color,
  size,
  initialX,
  initialY,
  delay,
}: {
  color: string;
  size: number;
  initialX: number;
  initialY: number;
  delay: number;
}) => {
  const translateX = useRef(new Animated.Value(0)).current;
  const translateY = useRef(new Animated.Value(0)).current;
  const scale = useRef(new Animated.Value(1)).current;

  useEffect(() => {
    const animate = () => {
      Animated.loop(
        Animated.sequence([
          Animated.delay(delay),
          Animated.parallel([
            Animated.sequence([
              Animated.timing(translateX, {
                toValue: 30,
                duration: 3000,
                useNativeDriver: true,
              }),
              Animated.timing(translateX, {
                toValue: -30,
                duration: 3000,
                useNativeDriver: true,
              }),
              Animated.timing(translateX, {
                toValue: 0,
                duration: 3000,
                useNativeDriver: true,
              }),
            ]),
            Animated.sequence([
              Animated.timing(translateY, {
                toValue: -20,
                duration: 2000,
                useNativeDriver: true,
              }),
              Animated.timing(translateY, {
                toValue: 20,
                duration: 2000,
                useNativeDriver: true,
              }),
              Animated.timing(translateY, {
                toValue: 0,
                duration: 2000,
                useNativeDriver: true,
              }),
            ]),
            Animated.sequence([
              Animated.timing(scale, {
                toValue: 1.2,
                duration: 4000,
                useNativeDriver: true,
              }),
              Animated.timing(scale, {
                toValue: 1,
                duration: 4000,
                useNativeDriver: true,
              }),
            ]),
          ]),
        ]),
      ).start();
    };
    animate();
  }, [delay, translateX, translateY, scale]);

  return (
    <Animated.View
      style={[
        styles.floatingOrb,
        {
          width: size,
          height: size,
          borderRadius: size / 2,
          backgroundColor: color,
          left: initialX,
          top: initialY,
          transform: [{ translateX }, { translateY }, { scale }],
        },
      ]}
    />
  );
};

export default function OnboardingScreen() {
  const router = useRouter();
  const [currentIndex, setCurrentIndex] = useState(0);
  const flatListRef = useRef<FlatList<OnboardingSlide>>(null);
  const scrollX = useRef(new Animated.Value(0)).current;
  const iconScale = useRef(new Animated.Value(1)).current;
  const buttonScale = useRef(new Animated.Value(1)).current;

  const viewableItemsChanged = useRef(
    ({ viewableItems }: { viewableItems: ViewToken[] }) => {
      if (viewableItems.length > 0 && viewableItems[0].index !== null) {
        setCurrentIndex(viewableItems[0].index);
        // Bounce animation on slide change
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

  const getItemLayout = (_: any, index: number) => ({
    length: width,
    offset: width * index,
    index,
  });

  const completeOnboarding = async () => {
    try {
      await AsyncStorage.setItem(ONBOARDING_KEY, 'true');
      router.replace('/auth/login');
    } catch (error) {
      console.error('Failed to save onboarding state:', error);
      router.replace('/auth/login');
    }
  };

  const handleNext = () => {
    // Button press animation
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
      completeOnboarding();
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
      <View style={styles.slide}>
        <LinearGradient
          colors={item.gradientColors}
          start={{ x: 0, y: 0 }}
          end={{ x: 1, y: 1 }}
          style={StyleSheet.absoluteFill}
        />

        {/* Floating orbs */}
        <FloatingOrb
          color={`${item.accentColor}20`}
          size={200}
          initialX={-50}
          initialY={100}
          delay={0}
        />
        <FloatingOrb
          color={`${item.accentColor}15`}
          size={150}
          initialX={width - 100}
          initialY={height - 300}
          delay={500}
        />
        <FloatingOrb
          color={`${item.accentColor}10`}
          size={100}
          initialX={width / 2}
          initialY={200}
          delay={1000}
        />

        {/* Content */}
        <View style={styles.slideContent}>
          {/* Icon container with glassmorphism */}
          <Animated.View
            style={[
              styles.iconWrapper,
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
                styles.iconOuter,
                { borderColor: `${item.accentColor}30` },
              ]}
            >
              <View
                style={[styles.iconInner, { backgroundColor: item.iconBg }]}
              >
                <MaterialCommunityIcons
                  name={item.icon}
                  size={80}
                  color={item.accentColor}
                />
              </View>
            </View>
            {/* Glow ring */}
            <View
              style={[
                styles.glowRing,
                {
                  borderColor: item.accentColor,
                  shadowColor: item.accentColor,
                },
              ]}
            />
          </Animated.View>

          {/* Text content */}
          <Animated.View
            style={[
              styles.textContainer,
              {
                opacity: textOpacity,
                transform: [{ translateY: textTranslateY }],
              },
            ]}
          >
            <Text style={[styles.title, { color: item.accentColor }]}>
              {item.title}
            </Text>
            <Text style={styles.subtitle}>{item.subtitle}</Text>
            <Text style={styles.description}>{item.description}</Text>
          </Animated.View>
        </View>
      </View>
    );
  };

  const currentSlide = slides[currentIndex];

  return (
    <View style={styles.container}>
      {/* Skip button with glassmorphism */}
      <View style={styles.skipContainer}>
        <Pressable onPress={completeOnboarding}>
          <BlurView intensity={20} style={styles.skipButton}>
            <Text style={styles.skipText}>Skip</Text>
          </BlurView>
        </Pressable>
      </View>

      {/* Progress indicator */}
      <View style={styles.progressContainer}>
        <Text style={styles.progressText}>
          {String(currentIndex + 1).padStart(2, '0')}
        </Text>
        <View style={styles.progressLine} />
        <Text style={styles.progressTotal}>
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

      {/* Bottom navigation */}
      <View style={styles.bottomContainer}>
        {/* Custom dots with animation */}
        <View style={styles.dotsContainer}>
          {slides.map((slide, index) => {
            const inputRange = [
              (index - 1) * width,
              index * width,
              (index + 1) * width,
            ];

            const dotWidth = scrollX.interpolate({
              inputRange,
              outputRange: [8, 32, 8],
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
                  styles.dot,
                  {
                    width: dotWidth,
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

        {/* Next button with glassmorphism */}
        <Animated.View style={{ transform: [{ scale: buttonScale }] }}>
          <Pressable onPress={handleNext}>
            <LinearGradient
              colors={[
                currentSlide.accentColor,
                `${currentSlide.accentColor}CC`,
              ]}
              start={{ x: 0, y: 0 }}
              end={{ x: 1, y: 1 }}
              style={styles.nextButton}
            >
              <Text style={styles.nextButtonText}>
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

const styles = StyleSheet.create({
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
  floatingOrb: {
    position: 'absolute',
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
