/**
 * pages/login.js
 *
 * SPA version of frontend/html/pages/login.html.
 * Supports login by nickname (username) or email. Sends POST to
 * /api/v1/login/username or /api/v1/login/email using the shared api client.
 */

import { api } from '../api.js';
import { navigate } from '../router.js';
import { setUser } from '../auth.js';
import { escapeHTML } from '../helpers.js';

export async function renderLoginPage(user) {
  const root = document.getElementById('app-root');
  if (!root) return;

  root.innerHTML = buildLoginHTML();

  initLoginBehaviour();
}

function buildLoginHTML() {
  return /* html */ `
    <header>
      <h1>Welcome Back</h1>
    </header>
    <main>
      <div class="signup-container">
        <div class="signup-wrapper">
          <h2 class="signup-title">Sign In</h2>
          <div class="text-base">
            Don't have an account?
            <a href="/register" data-link>Sign Up</a>
          </div>
          <div class="btn-box">
            <a class="signup-provider-btn google" href="/auth/google/login">
              <img src="/static/images/icons/google-logo.png" alt="Google Logo" />
              <p>Continue with Google</p>
            </a>
            <a class="signup-provider-btn github" href="/auth/github/login">
              <img src="/static/images/icons/github-white-logo.png" alt="Github Logo" />
              <p>Continue with Github</p>
            </a>
          </div>

          <div class="border">
            <span class="border-text">or</span>
          </div>

          <form class="signup" method="post" action="/login">
            <div class="input-wrapper">
              <div class="login-type-selector">
                <p class="login-type-title">Choose Type of Login</p>
                <label class="login-radio-label">
                  <input
                    type="radio"
                    name="loginType"
                    id="loginTypeNickname"
                    value="username"
                    class="login-type-radio"
                    checked
                  />
                  <span>Nickname</span>
                </label>
                <label class="login-radio-label">
                  <input
                    type="radio"
                    name="loginType"
                    id="loginTypeEmail"
                    value="email"
                    class="login-type-radio"
                  />
                  <span>Email</span>
                </label>
              </div>

              <div class="input-box" id="nicknameBox">
                <label for="nickname">Nickname</label>
                <input
                  type="text"
                  name="nickname"
                  id="nickname"
                  class="form-input"
                  placeholder="Enter your nickname"
                />
                <span class="error-message" id="nickname-error"></span>
              </div>

              <div class="input-box" id="emailBox" style="display:none">
                <label for="email">Email address</label>
                <input
                  type="email"
                  name="email"
                  id="email"
                  class="form-input"
                  placeholder="Enter your email"
                />
                <span class="error-message" id="email-error"></span>
              </div>

              <div class="input-box">
                <div class="password-wrapper">
                  <label for="password">Password</label>
                  <input
                    type="password"
                    name="password"
                    id="password"
                    class="form-input"
                    placeholder="Enter your password"
                  />
                  <label for="togglePassword" class="togglePassword">
                    <img src="/static/images/icons/eye.png" alt="Toggle Password Icon" id="eye-icon" class="eye-icon" />
                    <img src="/static/images/icons/hidden.png" alt="Toggle Password Icon" id="hidden-icon" class="hidden-icon" />
                  </label>
                  <input type="checkbox" id="togglePassword" class="hidden-checkbox" aria-label="show/hide password" />
                </div>
                <span class="error-message" id="password-error"></span>
              </div>
            </div>

            <div class="btn-box">
              <button type="reset" class="btn-reset-form">Reset Form</button>
              <button type="submit" class="btn-signup">Sign In</button>
            </div>
          </form>
        </div>
        <div class="home-link-container">
          <a href="/" class="home-link" data-link>Go to Homepage</a>
        </div>
      </div>
    </main>
  `;
}

function initLoginBehaviour() {
  const form = document.querySelector('.signup');
  if (!form) return;

  const loginTypeNickname = document.getElementById('loginTypeNickname');
  const loginTypeEmail = document.getElementById('loginTypeEmail');
  const nicknameBox = document.getElementById('nicknameBox');
  const emailBox = document.getElementById('emailBox');

  const nicknameInput = document.getElementById('nickname');
  const emailInput = document.getElementById('email');
  const passwordInput = document.getElementById('password');

  const nicknameError = document.getElementById('nickname-error');
  const emailError = document.getElementById('email-error');
  const passwordError = document.getElementById('password-error');

  function updateVisibility() {
    if (loginTypeEmail.checked) {
      nicknameBox.style.display = 'none';
      emailBox.style.display = '';
    } else {
      nicknameBox.style.display = '';
      emailBox.style.display = 'none';
    }
  }

  loginTypeNickname.addEventListener('change', updateVisibility);
  loginTypeEmail.addEventListener('change', updateVisibility);

  const toggle = document.getElementById('togglePassword');
  toggle?.addEventListener('change', () => {
    passwordInput.type = toggle.checked ? 'text' : 'password';
  });

  form.addEventListener('submit', async (e) => {
    e.preventDefault();

    // clear errors
    nicknameError.textContent = '';
    emailError.textContent = '';
    passwordError.textContent = '';

    const useEmail = loginTypeEmail.checked;
    const nickname = nicknameInput.value.trim();
    const email = emailInput.value.trim();
    const password = passwordInput.value.trim();

    if (!password) {
      passwordError.textContent = 'Please enter your password.';
      passwordInput.focus();
      return;
    }

    if (useEmail) {
      if (!email) {
        emailError.textContent = 'Please enter your email address.';
        emailInput.focus();
        return;
      }
    } else {
      if (!nickname) {
        nicknameError.textContent = 'Please enter your nickname.';
        nicknameInput.focus();
        return;
      }
    }

    const submitBtn = form.querySelector('button[type="submit"]');
    submitBtn.disabled = true;
    submitBtn.textContent = 'Signing in...';

    try {
      let res;
      if (useEmail) {
        res = await api.post('/login/email', { email, password });
      } else {
        // backend expects field `username` for username login handler
        res = await api.post('/login/username', { username: nickname, password });
      }

      // Successful login returns tokens and username; set in-memory user
      const user = {
        id: res.userId || res.UserID || null,
        username: res.username || res.Username || res.username || null,
        email: res.email || null,
      };
      try {
        setUser(user);
      } catch (err) {
        // ignore if setUser not available
      }

      navigate('/');
    } catch (err) {
      const msg = err?.message || String(err);
      const lower = String(msg).toLowerCase();
      if (lower.includes('email')) {
        emailError.textContent = escapeHTML(msg);
        emailInput.focus();
      } else if (
        lower.includes('username') ||
        lower.includes('nick') ||
        lower.includes('nickname')
      ) {
        nicknameError.textContent = escapeHTML(msg);
        nicknameInput.focus();
      } else {
        passwordError.textContent = escapeHTML(msg);
        passwordInput.focus();
      }
    } finally {
      submitBtn.disabled = false;
      submitBtn.textContent = 'Sign In';
    }
  });

  const resetBtn = form.querySelector('button[type="reset"]');
  resetBtn?.addEventListener('click', () => {
    nicknameError.textContent = '';
    emailError.textContent = '';
    passwordError.textContent = '';
  });

  // ensure initial visibility
  updateVisibility();
}
