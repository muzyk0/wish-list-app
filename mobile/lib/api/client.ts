// mobile/lib/api/client.ts
// Base openapi-fetch client without middleware
// Used by both api.ts and auth.ts to avoid circular dependencies

import createClient from "openapi-fetch";
import type { paths } from "./schema";

const API_BASE_URL =
  process.env.EXPO_PUBLIC_API_URL || "http://10.0.2.2:8080/api";

/**
 * Base openapi-fetch client without any middleware
 * This is used in auth.ts for authentication operations
 *
 * Note: api.ts creates a separate client WITH middleware for protected routes
 */
export const baseClient = createClient<paths>({ baseUrl: API_BASE_URL });

/**
 * Export API_BASE_URL for consistency
 */
export { API_BASE_URL };
