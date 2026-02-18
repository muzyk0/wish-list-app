import { LinearGradient } from 'expo-linear-gradient';
import { useRouter } from 'expo-router';
import { useEffect, useId, useRef } from 'react';
import { Animated, Dimensions, Image, StyleSheet, View } from 'react-native';
import { Text } from 'react-native-paper';

const { width, height } = Dimensions.get('window');
const SPLASH_DURATION = 2500;

// Floating particle component
const FloatingParticle = ({
  delay,
  startX,
  size,
}: {
  delay: number;
  startX: number;
  size: number;
}) => {
  const translateY = useRef(new Animated.Value(height + 50)).current;
  const opacity = useRef(new Animated.Value(0)).current;
  const scale = useRef(new Animated.Value(0.5)).current;

  useEffect(() => {
    const animate = () => {
      translateY.setValue(height + 50);
      opacity.setValue(0);
      scale.setValue(0.5);

      Animated.sequence([
        Animated.delay(delay),
        Animated.parallel([
          Animated.timing(translateY, {
            toValue: -100,
            duration: 4000 + Math.random() * 2000,
            useNativeDriver: true,
          }),
          Animated.sequence([
            Animated.timing(opacity, {
              toValue: 0.6,
              duration: 1000,
              useNativeDriver: true,
            }),
            Animated.delay(2000),
            Animated.timing(opacity, {
              toValue: 0,
              duration: 1000,
              useNativeDriver: true,
            }),
          ]),
          Animated.timing(scale, {
            toValue: 1.2,
            duration: 4000,
            useNativeDriver: true,
          }),
        ]),
      ]).start();
    };

    animate();
    const interval = setInterval(animate, 5000 + Math.random() * 2000);
    return () => clearInterval(interval);
  }, [delay, translateY, opacity, scale]);

  return (
    <Animated.View
      style={[
        styles.particle,
        {
          left: startX,
          width: size,
          height: size,
          borderRadius: size / 2,
          opacity,
          transform: [{ translateY }, { scale }],
        },
      ]}
    />
  );
};

export default function SplashScreen() {
  const id = useId();
  const router = useRouter();
  const fadeAnim = useRef(new Animated.Value(0)).current;
  const scaleAnim = useRef(new Animated.Value(0.6)).current;
  const rotateAnim = useRef(new Animated.Value(0)).current;
  const glowAnim = useRef(new Animated.Value(0)).current;
  const textFadeAnim = useRef(new Animated.Value(0)).current;
  const taglineFadeAnim = useRef(new Animated.Value(0)).current;

  useEffect(() => {
    // Orchestrated entrance animation
    Animated.sequence([
      // Logo entrance with rotation
      Animated.parallel([
        Animated.spring(scaleAnim, {
          toValue: 1,
          tension: 40,
          friction: 5,
          useNativeDriver: true,
        }),
        Animated.timing(fadeAnim, {
          toValue: 1,
          duration: 800,
          useNativeDriver: true,
        }),
        Animated.timing(rotateAnim, {
          toValue: 1,
          duration: 1000,
          useNativeDriver: true,
        }),
      ]),
      // Glow pulse
      Animated.loop(
        Animated.sequence([
          Animated.timing(glowAnim, {
            toValue: 1,
            duration: 1500,
            useNativeDriver: true,
          }),
          Animated.timing(glowAnim, {
            toValue: 0,
            duration: 1500,
            useNativeDriver: true,
          }),
        ]),
      ),
    ]).start();

    // Staggered text entrance
    Animated.sequence([
      Animated.delay(400),
      Animated.timing(textFadeAnim, {
        toValue: 1,
        duration: 600,
        useNativeDriver: true,
      }),
    ]).start();

    Animated.sequence([
      Animated.delay(700),
      Animated.timing(taglineFadeAnim, {
        toValue: 1,
        duration: 600,
        useNativeDriver: true,
      }),
    ]).start();

    // After animation — navigate away from splash.
    // Auth check and redirect are handled by (tabs)/_layout.tsx guard.
    const navigateAfterSplash = () => {
      setTimeout(() => {
        Animated.parallel([
          Animated.timing(fadeAnim, {
            toValue: 0,
            duration: 400,
            useNativeDriver: true,
          }),
          Animated.timing(scaleAnim, {
            toValue: 1.2,
            duration: 400,
            useNativeDriver: true,
          }),
        ]).start(() => {
          router.replace('/(tabs)');
        });
      }, SPLASH_DURATION);
    };

    navigateAfterSplash();
  }, [
    router,
    fadeAnim,
    scaleAnim,
    rotateAnim,
    glowAnim,
    textFadeAnim,
    taglineFadeAnim,
  ]);

  const spin = rotateAnim.interpolate({
    inputRange: [0, 1],
    outputRange: ['0deg', '360deg'],
  });

  const glowOpacity = glowAnim.interpolate({
    inputRange: [0, 1],
    outputRange: [0.3, 0.8],
  });

  // Generate particles
  const particles = Array.from({ length: 12 }, (_, i) => ({
    key: `splash-screen|${id}|${i}`,
    delay: i * 300,
    startX: Math.random() * width,
    size: 8 + Math.random() * 16,
  }));

  return (
    <View style={styles.container}>
      <LinearGradient
        colors={['#1a0a2e', '#3d1a5c', '#5c2d7a', '#7b3fa0']}
        start={{ x: 0, y: 0 }}
        end={{ x: 1, y: 1 }}
        style={StyleSheet.absoluteFill}
      />

      {/* Floating particles */}
      {particles.map(({ key, ...particle }) => (
        <FloatingParticle key={key} {...particle} />
      ))}

      {/* Decorative circles */}
      <View style={[styles.decorCircle, styles.decorCircle1]} />
      <View style={[styles.decorCircle, styles.decorCircle2]} />
      <View style={[styles.decorCircle, styles.decorCircle3]} />

      <Animated.View
        style={[
          styles.content,
          {
            opacity: fadeAnim,
            transform: [{ scale: scaleAnim }],
          },
        ]}
      >
        {/* Glow effect behind logo */}
        <Animated.View style={[styles.glowContainer, { opacity: glowOpacity }]}>
          <LinearGradient
            colors={[
              'rgba(255, 215, 0, 0.4)',
              'rgba(255, 107, 107, 0.2)',
              'transparent',
            ]}
            style={styles.glow}
          />
        </Animated.View>

        {/* Logo with rotation */}
        <Animated.View
          style={[styles.logoContainer, { transform: [{ rotate: spin }] }]}
        >
          <LinearGradient
            colors={['#FFD700', '#FFA500', '#FF6B6B']}
            start={{ x: 0, y: 0 }}
            end={{ x: 1, y: 1 }}
            style={styles.logoGradient}
          >
            <Image
              source={require('@/assets/images/icon.png')}
              style={styles.logo}
              resizeMode="contain"
            />
          </LinearGradient>
        </Animated.View>

        {/* App Name with shimmer effect */}
        <Animated.View style={{ opacity: textFadeAnim }}>
          <Text style={styles.appName}>Wish List</Text>
        </Animated.View>

        {/* Tagline */}
        <Animated.View style={{ opacity: taglineFadeAnim }}>
          <Text style={styles.tagline}>Make wishes come true</Text>
        </Animated.View>

        {/* Loading dots */}
        <View style={styles.loadingContainer}>
          {[0, 1, 2].map((i) => (
            <LoadingDot key={i} delay={i * 200} />
          ))}
        </View>
      </Animated.View>
    </View>
  );
}

// Animated loading dot
const LoadingDot = ({ delay }: { delay: number }) => {
  const scale = useRef(new Animated.Value(0.5)).current;
  const opacity = useRef(new Animated.Value(0.3)).current;

  useEffect(() => {
    Animated.loop(
      Animated.sequence([
        Animated.delay(delay),
        Animated.parallel([
          Animated.timing(scale, {
            toValue: 1,
            duration: 400,
            useNativeDriver: true,
          }),
          Animated.timing(opacity, {
            toValue: 1,
            duration: 400,
            useNativeDriver: true,
          }),
        ]),
        Animated.parallel([
          Animated.timing(scale, {
            toValue: 0.5,
            duration: 400,
            useNativeDriver: true,
          }),
          Animated.timing(opacity, {
            toValue: 0.3,
            duration: 400,
            useNativeDriver: true,
          }),
        ]),
      ]),
    ).start();
  }, [delay, scale, opacity]);

  return (
    <Animated.View
      style={[styles.loadingDot, { opacity, transform: [{ scale }] }]}
    />
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  particle: {
    position: 'absolute',
    backgroundColor: 'rgba(255, 215, 0, 0.6)',
    shadowColor: '#FFD700',
    shadowOffset: { width: 0, height: 0 },
    shadowOpacity: 0.8,
    shadowRadius: 10,
  },
  decorCircle: {
    position: 'absolute',
    borderRadius: 500,
    backgroundColor: 'rgba(255, 255, 255, 0.03)',
  },
  decorCircle1: {
    width: 400,
    height: 400,
    top: -100,
    right: -150,
  },
  decorCircle2: {
    width: 300,
    height: 300,
    bottom: 100,
    left: -100,
  },
  decorCircle3: {
    width: 200,
    height: 200,
    bottom: -50,
    right: 50,
  },
  content: {
    alignItems: 'center',
    justifyContent: 'center',
  },
  glowContainer: {
    position: 'absolute',
    width: 250,
    height: 250,
    top: -65,
  },
  glow: {
    width: '100%',
    height: '100%',
    borderRadius: 125,
  },
  logoContainer: {
    marginBottom: 32,
    shadowColor: '#FFD700',
    shadowOffset: { width: 0, height: 8 },
    shadowOpacity: 0.5,
    shadowRadius: 20,
    elevation: 20,
  },
  logoGradient: {
    width: 120,
    height: 120,
    borderRadius: 32,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 4,
  },
  logo: {
    width: 70,
    height: 70,
    tintColor: '#ffffff',
  },
  appName: {
    fontSize: 42,
    fontWeight: '800',
    color: '#ffffff',
    letterSpacing: 2,
    textShadowColor: 'rgba(255, 215, 0, 0.5)',
    textShadowOffset: { width: 0, height: 2 },
    textShadowRadius: 10,
    marginBottom: 8,
  },
  tagline: {
    fontSize: 18,
    color: 'rgba(255, 255, 255, 0.7)',
    letterSpacing: 4,
    textTransform: 'uppercase',
    fontWeight: '300',
  },
  loadingContainer: {
    flexDirection: 'row',
    marginTop: 60,
    gap: 12,
  },
  loadingDot: {
    width: 10,
    height: 10,
    borderRadius: 5,
    backgroundColor: '#FFD700',
  },
});
