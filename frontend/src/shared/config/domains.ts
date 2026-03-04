/**
 * Domain constants for the Wish List application
 */

// Base domain for mobile deep linking
const MOBILE_APP_DOMAIN =
  process.env.NEXT_PUBLIC_MOBILE_APP_DOMAIN || 'lk.domain.com';

// Base URL for mobile app deep linking
const MOBILE_APP_BASE_URL = `https://${MOBILE_APP_DOMAIN}`;

export const DOMAIN_CONSTANTS = {
  MOBILE_APP_DOMAIN,
  MOBILE_APP_BASE_URL,
  MOBILE_AUTH_LOGIN_PATH: 'auth/login',
  MOBILE_AUTH_REGISTER_PATH: 'auth/register',
  MOBILE_MY_RESERVATIONS_PATH: 'my/reservations',
} as const;

export const MOBILE_APP_URLS = {
  LOGIN: `${MOBILE_APP_BASE_URL}/${DOMAIN_CONSTANTS.MOBILE_AUTH_LOGIN_PATH}`,
  REGISTER: `${MOBILE_APP_BASE_URL}/${DOMAIN_CONSTANTS.MOBILE_AUTH_REGISTER_PATH}`,
  MY_RESERVATIONS: `${MOBILE_APP_BASE_URL}/${DOMAIN_CONSTANTS.MOBILE_MY_RESERVATIONS_PATH}`,
} as const;

export const MOBILE_APP_REDIRECT_PATHS = {
  HOME: 'home',
  AUTH_LOGIN: 'auth/login',
  AUTH_REGISTER: 'auth/register',
  MY_RESERVATIONS: 'my/reservations',
} as const;
