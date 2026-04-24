/**
 * auth.js
 *
 * Client-side auth state.
 *
 * The Go BFF held the logged-in user in the HTTP request context after the
 * AuthMiddleware called /me. Here we replicate that pattern: on app startup
 * we call /me once, store the result in memory, and expose it to the rest of
 * the app. Every module that needs to know "is anyone logged in?" imports from
 * here instead of fetching /me independently.
 *
 * Later, when we add protected routes (mirrors RequireAuth middleware), the
 * router will call `requireAuth()` before rendering a page and redirect to
 * /login if the user is null.
 */

import { fetchCurrentUser } from './api.js';

// ─── Internal state ───────────────────────────────────────────────────────────

/** @type {{ id: string, username: string, email: string, avatar_url: string } | null} */
let _currentUser = null;

/** Whether we have already resolved the user at least once this session. */
let _resolved = false;

// ─── Public API ───────────────────────────────────────────────────────────────

/**
 * Initialises auth state by calling /me.
 * Called once at app startup before the first render.
 * Safe to call multiple times — resolves immediately after the first call.
 *
 * @returns {Promise<object|null>} The logged-in user, or null.
 */
export async function initAuth() {
  if (_resolved) return _currentUser;

  _currentUser = await fetchCurrentUser();
  _resolved = true;

  return _currentUser;
}

/**
 * Returns the currently logged-in user synchronously.
 * Returns null if not authenticated or if initAuth() hasn't been called yet.
 *
 * @returns {object|null}
 */
export function getUser() {
  return _currentUser;
}

/**
 * Returns true if a user is currently logged in.
 *
 * @returns {boolean}
 */
export function isLoggedIn() {
  return _currentUser !== null;
}

/**
 * Clears the in-memory user (called after logout).
 */
export function clearUser() {
  _currentUser = null;
  _resolved = false;
}

/**
 * Updates the in-memory user (called after login / register).
 *
 * @param {object} user
 */
export function setUser(user) {
  _currentUser = user;
  _resolved = true;
}

/**
 * Guard for protected routes. Mirrors the Go BFF RequireAuth middleware.
 * If no user is logged in, redirects to /login using the SPA router and
 * returns false. Returns true if the user is authenticated.
 *
 * Usage inside a page's render() function:
 *   if (!requireAuth()) return;
 *
 * @returns {boolean}
 */
export function requireAuth() {
  if (_currentUser !== null) return true;

  // Lazy-import the router to avoid circular dependency.
  import('./router.js').then(({ navigate }) => navigate('/login'));
  return false;
}
