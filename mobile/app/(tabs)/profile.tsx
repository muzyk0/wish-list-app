import { MaterialCommunityIcons } from '@expo/vector-icons';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { BlurView } from 'expo-blur';
import { LinearGradient } from 'expo-linear-gradient';
import { useRouter } from 'expo-router';
import { Alert, Pressable, StyleSheet, Switch, View } from 'react-native';
import { ActivityIndicator, Avatar, Text } from 'react-native-paper';
import { TabsLayout } from '@/components/TabsLayout';
import { useThemeContext } from '@/contexts/ThemeContext';
import { apiClient } from '@/lib/api';
import { clearTokens } from '@/lib/api/auth';

type MenuItemIcon = keyof typeof MaterialCommunityIcons.glyphMap;

interface MenuItem {
  icon: MenuItemIcon;
  label: string;
  onPress?: () => void;
  showChevron?: boolean;
  value?: string;
  rightElement?: React.ReactNode;
}

interface MenuSection {
  title: string;
  titleIcon?: MenuItemIcon;
  items: MenuItem[];
}

const MenuItemComponent = ({
  icon,
  label,
  onPress,
  showChevron = true,
  value,
  rightElement,
}: MenuItem) => {
  return (
    <Pressable onPress={onPress} disabled={!onPress}>
      <View style={styles.menuItem}>
        <View style={styles.menuItemLeft}>
          <View style={styles.menuIconContainer}>
            <MaterialCommunityIcons name={icon} size={20} color="#FFD700" />
          </View>
          <Text style={styles.menuItemLabel}>{label}</Text>
        </View>
        <View style={styles.menuItemRight}>
          {value && <Text style={styles.menuItemValue}>{value}</Text>}
          {rightElement}
          {showChevron && !rightElement && (
            <MaterialCommunityIcons
              name="chevron-right"
              size={20}
              color="rgba(255, 255, 255, 0.3)"
            />
          )}
        </View>
      </View>
    </Pressable>
  );
};

const MenuSectionComponent = ({ title, titleIcon, items }: MenuSection) => {
  return (
    <BlurView intensity={20} style={styles.section}>
      <View style={styles.sectionInner}>
        {title && (
          <View style={styles.sectionHeader}>
            {titleIcon && (
              <MaterialCommunityIcons
                name={titleIcon}
                size={20}
                color="#FFD700"
              />
            )}
            <Text style={styles.sectionTitle}>{title}</Text>
          </View>
        )}
        <View style={styles.menuItems}>
          {items.map((item, index) => (
            <View key={item.label}>
              <MenuItemComponent {...item} />
              {index < items.length - 1 && <View style={styles.divider} />}
            </View>
          ))}
        </View>
      </View>
    </BlurView>
  );
};

export default function ProfileScreen() {
  const _queryClient = useQueryClient();
  const router = useRouter();
  const { isDark, toggleTheme } = useThemeContext();

  const {
    data: user,
    isLoading,
    refetch,
  } = useQuery({
    queryKey: ['profile'],
    queryFn: () => apiClient.getProfile(),
    retry: 1,
  });

  const deleteAccountMutation = useMutation({
    mutationFn: () => apiClient.deleteAccount(),
    onSuccess: async () => {
      Alert.alert('Account Deleted', 'Your account has been deleted.');
      await clearTokens();
      router.replace('/auth/login');
    },
    onError: (error: Error) => {
      Alert.alert('Error', error.message || 'Failed to delete account');
    },
  });

  const handleLogout = async () => {
    Alert.alert('Confirm Logout', 'Are you sure you want to log out?', [
      { text: 'Cancel', style: 'cancel' },
      {
        text: 'Logout',
        style: 'destructive',
        onPress: async () => {
          await clearTokens();
          router.replace('/auth/login');
        },
      },
    ]);
  };

  const handleDeleteAccount = () => {
    Alert.alert(
      'Delete Account',
      'Are you sure you want to delete your account? This action cannot be undone and will permanently delete all your data.',
      [
        { text: 'Cancel', style: 'cancel' },
        {
          text: 'Delete',
          style: 'destructive',
          onPress: () => deleteAccountMutation.mutate(),
        },
      ],
    );
  };

  if (isLoading) {
    return (
      <TabsLayout title="Profile" subtitle="Your account">
        <View style={styles.loadingContainer}>
          <ActivityIndicator size="large" color="#FFD700" />
        </View>
      </TabsLayout>
    );
  }

  const profileSection: MenuSection = {
    title: 'Profile',
    titleIcon: 'account-circle',
    items: [
      {
        icon: 'account-edit',
        label: 'Edit Profile',
        onPress: () => router.push('/profile/edit'),
      },
      {
        icon: 'card-account-details',
        label: 'Personal Information',
        value:
          `${user?.first_name || ''} ${user?.last_name || ''}`.trim() ||
          'Not set',
        onPress: () => router.push('/profile/edit'),
      },
      {
        icon: 'email',
        label: 'Email',
        value: user?.email || '',
        showChevron: false,
      },
    ],
  };

  const securitySection: MenuSection = {
    title: 'Security',
    titleIcon: 'shield-lock',
    items: [
      {
        icon: 'email-edit',
        label: 'Change Email',
        onPress: () => router.push('/profile/change-email'),
      },
      {
        icon: 'lock-reset',
        label: 'Change Password',
        onPress: () => router.push('/profile/change-password'),
      },
      {
        icon: 'two-factor-authentication',
        label: 'Two-Factor Authentication',
        value: 'Disabled',
        onPress: () =>
          Alert.alert('Coming Soon', 'This feature is coming soon!'),
      },
    ],
  };

  const preferencesSection: MenuSection = {
    title: 'Preferences',
    titleIcon: 'cog',
    items: [
      {
        icon: 'theme-light-dark',
        label: 'Dark Mode',
        showChevron: false,
        rightElement: (
          <Switch
            value={isDark}
            onValueChange={toggleTheme}
            trackColor={{ false: '#767577', true: '#FFD700' }}
            thumbColor={isDark ? '#FFA500' : '#f4f3f4'}
          />
        ),
      },
      {
        icon: 'bell',
        label: 'Notifications',
        onPress: () =>
          Alert.alert('Coming Soon', 'This feature is coming soon!'),
      },
      {
        icon: 'translate',
        label: 'Language',
        value: 'English',
        onPress: () =>
          Alert.alert('Coming Soon', 'This feature is coming soon!'),
      },
    ],
  };

  const aboutSection: MenuSection = {
    title: 'About',
    titleIcon: 'information',
    items: [
      {
        icon: 'help-circle',
        label: 'Help & Support',
        onPress: () => Alert.alert('Help', 'Contact support@wishlist.app'),
      },
      {
        icon: 'file-document',
        label: 'Terms of Service',
        onPress: () =>
          Alert.alert('Coming Soon', 'This feature is coming soon!'),
      },
      {
        icon: 'shield-check',
        label: 'Privacy Policy',
        onPress: () =>
          Alert.alert('Coming Soon', 'This feature is coming soon!'),
      },
      {
        icon: 'information-variant',
        label: 'App Version',
        value: '1.0.0',
        showChevron: false,
      },
    ],
  };

  const _dangerSection: MenuSection = {
    title: 'Danger Zone',
    titleIcon: 'alert-circle',
    items: [
      {
        icon: 'delete-forever',
        label: 'Delete Account',
        onPress: handleDeleteAccount,
      },
    ],
  };

  return (
    <TabsLayout
      title="Profile"
      subtitle="Your account"
      refreshing={isLoading}
      onRefresh={refetch}
    >
      {/* Avatar Section */}
      <View style={styles.avatarSection}>
        <LinearGradient
          colors={['#FFD700', '#FFA500']}
          style={styles.avatarGradient}
        >
          <Avatar.Text
            size={90}
            label={`${user?.first_name?.[0] || ''}${user?.last_name?.[0] || ''}`.toUpperCase()}
            color="#000000"
            style={styles.avatar}
          />
        </LinearGradient>
        <Text style={styles.userName}>
          {user?.first_name} {user?.last_name}
        </Text>
        <Text style={styles.userEmail}>{user?.email}</Text>
      </View>

      {/* Menu Sections */}
      <MenuSectionComponent {...profileSection} />
      <MenuSectionComponent {...securitySection} />
      <MenuSectionComponent {...preferencesSection} />
      <MenuSectionComponent {...aboutSection} />

      {/* Danger Zone */}
      <BlurView intensity={20} style={[styles.section, styles.dangerSection]}>
        <View style={styles.sectionInner}>
          <View style={styles.sectionHeader}>
            <MaterialCommunityIcons
              name="alert-circle"
              size={20}
              color="#FF6B6B"
            />
            <Text style={[styles.sectionTitle, { color: '#FF6B6B' }]}>
              Danger Zone
            </Text>
          </View>
          <Pressable onPress={handleDeleteAccount}>
            <View style={styles.dangerMenuItem}>
              <View style={styles.menuItemLeft}>
                <View
                  style={[styles.menuIconContainer, styles.dangerIconContainer]}
                >
                  <MaterialCommunityIcons
                    name="delete-forever"
                    size={20}
                    color="#FF6B6B"
                  />
                </View>
                <Text style={styles.dangerMenuItemLabel}>Delete Account</Text>
              </View>
              <MaterialCommunityIcons
                name="chevron-right"
                size={20}
                color="rgba(255, 107, 107, 0.5)"
              />
            </View>
          </Pressable>
        </View>
      </BlurView>

      {/* Logout Button */}
      <Pressable onPress={handleLogout} style={{ marginTop: 24 }}>
        <LinearGradient
          colors={['#FFD700', '#FFA500']}
          style={styles.logoutButton}
        >
          <MaterialCommunityIcons name="logout" size={20} color="#000000" />
          <Text style={styles.logoutText}>Logout</Text>
        </LinearGradient>
      </Pressable>
    </TabsLayout>
  );
}

const styles = StyleSheet.create({
  loadingContainer: {
    paddingVertical: 60,
    alignItems: 'center',
  },
  avatarSection: {
    alignItems: 'center',
    marginBottom: 32,
  },
  avatarGradient: {
    width: 110,
    height: 110,
    borderRadius: 55,
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: 16,
    shadowColor: '#FFD700',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.3,
    shadowRadius: 8,
    elevation: 8,
  },
  avatar: {
    backgroundColor: 'transparent',
  },
  userName: {
    fontSize: 24,
    fontWeight: '700',
    color: '#ffffff',
    marginBottom: 4,
  },
  userEmail: {
    fontSize: 14,
    color: 'rgba(255, 255, 255, 0.6)',
  },
  section: {
    borderRadius: 20,
    overflow: 'hidden',
    backgroundColor: 'rgba(255, 255, 255, 0.08)',
    borderWidth: 1,
    borderColor: 'rgba(255, 255, 255, 0.1)',
    marginBottom: 16,
  },
  sectionInner: {
    padding: 20,
  },
  sectionHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 10,
    marginBottom: 16,
  },
  sectionTitle: {
    fontSize: 16,
    fontWeight: '700',
    color: '#ffffff',
    textTransform: 'uppercase',
    letterSpacing: 0.5,
  },
  menuItems: {
    gap: 0,
  },
  menuItem: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingVertical: 14,
  },
  menuItemLeft: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 12,
    flex: 1,
  },
  menuIconContainer: {
    width: 36,
    height: 36,
    borderRadius: 10,
    backgroundColor: 'rgba(255, 215, 0, 0.15)',
    justifyContent: 'center',
    alignItems: 'center',
  },
  menuItemLabel: {
    fontSize: 15,
    color: '#ffffff',
    fontWeight: '500',
    flex: 1,
  },
  menuItemRight: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
  },
  menuItemValue: {
    fontSize: 14,
    color: 'rgba(255, 255, 255, 0.5)',
    maxWidth: 140,
  },
  divider: {
    height: 1,
    backgroundColor: 'rgba(255, 255, 255, 0.08)',
    marginLeft: 48,
  },
  dangerSection: {
    backgroundColor: 'rgba(255, 107, 107, 0.08)',
    borderColor: 'rgba(255, 107, 107, 0.2)',
  },
  dangerMenuItem: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingVertical: 14,
  },
  dangerIconContainer: {
    backgroundColor: 'rgba(255, 107, 107, 0.15)',
  },
  dangerMenuItemLabel: {
    fontSize: 15,
    color: '#FF6B6B',
    fontWeight: '600',
    flex: 1,
  },
  logoutButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 16,
    borderRadius: 16,
    gap: 8,
  },
  logoutText: {
    fontSize: 16,
    fontWeight: '700',
    color: '#000000',
  },
});
