/**
 * pages/logout.js
 *
 * Performs logout by calling the backend POST /api/v1/logout endpoint
 * via the shared api client, clears client auth state, and redirects.
 */

import { api } from '../api.js';
import { clearUser } from '../auth.js';
import { navigate } from '../router.js';

export async function renderLogoutPage(user) {
  const root = document.getElementById('app-root');
  if (!root) return;

  // If not logged in, redirect to login
  if (!user) {
    navigate('/login');
    return;
  }

  root.innerHTML = buildHTML('Signing out...');

  try {
    // Backend logout expects POST /api/v1/logout with no body
    await api.post('/logout');

    // Clear client-side auth state and navigate home
    try {
      clearUser();
    } catch (err) {
      // ignore
    }

    navigate('/');
  } catch (err) {
    const msg = err?.message || String(err);
    root.innerHTML = buildHTML('Failed to sign out: ' + msg, true);

    // Attach retry handler
    const retryBtn = document.getElementById('logout-retry');
    retryBtn?.addEventListener('click', () => {
      renderLogoutPage(user);
    });
  }
}

function buildHTML(message, showRetry = false) {
  return /* html */ `
    <div class="main-container" style="text-align:center;padding:4rem 1rem">
      <h2>${message}</h2>
      ${showRetry ? '<div style="margin-top:1.5rem"><button id="logout-retry" class="btn-signup">Retry</button></div>' : ''}
    </div>
  `;
}
