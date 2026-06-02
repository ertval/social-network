import { navigate } from '../router.js';
import { escapeHTML } from '../helpers.js';

export async function renderOauthCallback() {
  const root = document.getElementById('app-root');
  if (!root) return;

  const params = new URLSearchParams(window.location.search);
  const error = params.get('error');
  const provider = params.get('provider');
  const success = params.get('success');

  if (isSuccess(success) && !error) {
    navigate('/');
    return;
  }

  root.innerHTML = buildErrorHTML(getFriendlyErrorMessage(error, provider));

  const loginButton = document.getElementById('oauth-callback-login');
  loginButton?.addEventListener('click', () => {
    navigate('/login');
  });
}

function isSuccess(value) {
  if (value === null) return false;
  return value === '' || value === '1' || value === 'true' || value === 'ok';
}

function getFriendlyErrorMessage(errorCode, provider) {
  switch (errorCode) {
    case 'access_denied':
      return 'You cancelled the sign-in request. Please try again if you still want to continue.';
    case 'email_exists':
      return `An account with this email already exists. Please log in with your usual method and link you account with ${provider} in the settings in order to be able to login with this provider.`;
    case 'state_invalid':
      return 'Your sign-in session expired or became invalid. Please start again.';
    case 'oauth_failed':
      return 'We could not complete the OAuth sign-in. Please try again.';
    case 'session_not_found':
      return 'We could not confirm your session after sign-in. Please try again.';
    case null:
    case '':
      return 'The OAuth sign-in could not be completed.';
    default:
      return `OAuth sign-in failed: ${escapeHTML(errorCode)}`;
  }
}

function buildErrorHTML(message) {
  return /* html */ `
    <div class="main-container" style="text-align:center;padding:4rem 1rem">
      <h1 style="font-size:2rem;color:var(--dark-background)">OAuth Sign-In Failed</h1>
      <p style="margin-top:1rem;color:var(--grey-color);font-size:1.05rem;max-width:36rem;margin-left:auto;margin-right:auto">
        ${message}
      </p>
      <div style="margin-top:2rem">
        <button id="oauth-callback-login" type="button" class="btn-signup">Go to Login</button>
      </div>
    </div>
  `;
}
