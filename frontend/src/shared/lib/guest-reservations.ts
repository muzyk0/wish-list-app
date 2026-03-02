/**
 * Guest reservation localStorage utility
 * Manages guest reservation tokens stored in the browser for unauthenticated users.
 * Key: 'guest_reservations' — array of StoredReservation entries.
 */

const STORAGE_KEY = 'guest_reservations';

export interface StoredReservation {
  /** Gift item ID */
  itemId: string;
  /** Gift item name (cached for display) */
  itemName: string;
  /** Reservation token returned by the backend */
  reservationToken: string;
  /** ISO timestamp of when the reservation was made */
  reservedAt: string;
  /** Guest name provided during reservation */
  guestName: string;
  /** Optional: wishlist ID for cancel operations */
  wishlistId?: string;
}

function readFromStorage(): StoredReservation[] {
  if (typeof window === 'undefined') return [];
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw);
    return Array.isArray(parsed) ? parsed : [];
  } catch {
    return [];
  }
}

function writeToStorage(reservations: StoredReservation[]): void {
  if (typeof window === 'undefined') return;
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(reservations));
  } catch {
    // Silently fail if localStorage is unavailable (e.g., private mode quota)
  }
}

/** Returns all stored guest reservations, filtering out invalid tokens */
export function getStoredReservations(): StoredReservation[] {
  const NIL_UUID = '00000000-0000-0000-0000-000000000000';
  const all = readFromStorage();
  const valid = all.filter(
    (r) => r.reservationToken &&
           r.reservationToken !== '' &&
           r.reservationToken !== NIL_UUID
  );
  // Auto-clean stale entries if any were filtered out
  if (valid.length !== all.length) {
    writeToStorage(valid);
  }
  return valid;
}

/** Returns all reservation tokens (for bulk API queries) */
export function getAllTokens(): string[] {
  return readFromStorage().map((r) => r.reservationToken);
}

/** Adds a new reservation entry; deduplicates by reservationToken */
export function addReservation(reservation: StoredReservation): void {
  const current = readFromStorage();
  const exists = current.some(
    (r) => r.reservationToken === reservation.reservationToken,
  );
  if (!exists) {
    current.push(reservation);
    writeToStorage(current);
  }
}

/** Removes a reservation entry by its token */
export function removeReservation(reservationToken: string): void {
  const current = readFromStorage();
  const updated = current.filter(
    (r) => r.reservationToken !== reservationToken,
  );
  writeToStorage(updated);
}
