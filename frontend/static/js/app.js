/**
 * app.js
 *
 * SPA entry point. Loaded as <script type="module"> from index.html.
 *
 * Boot sequence:
 *   1. initAuth()   — call /me to find out if the browser session is valid
 *                     (mirrors the Go BFF AuthMiddleware running on every request)
 *   2. initRouter() — render the current URL's page and set up link interception
 *
 * Everything else (navbar, footer, page modules) is lazy-loaded by the router.
 */

import { initAuth } from './auth.js';
import { initRouter } from './router.js';

async function boot() {
  try {
    // Resolve auth state before the first render so every page knows the user.
    // NOTE: Right now, even if /me fails (not logged in), we still render the
    // homepage. Later, protected routes will redirect to /login automatically.
    await initAuth();
  } catch (err) {
    // Don't block the app if /me is unreachable — guest mode is fine.
    console.warn('Auth init failed, continuing as guest:', err.message);
  }

  // Start the router — this renders the first page.
  initRouter();
}

boot();
