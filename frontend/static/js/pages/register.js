/**
 * pages/register.js
 *
 * SPA version of frontend/html/pages/register.html.
 * Renders the registration page into #app-root and performs the
 * API POST to /api/v1/register via the shared `api` client.
 */

import { api } from '../api.js';
import { navigate } from '../router.js';
import { escapeHTML } from '../helpers.js';

export async function renderRegisterPage(user) {
  const root = document.getElementById('app-root');
  if (!root) return;

  root.innerHTML = buildRegisterHTML();

  initFormBehaviour();
}

function buildRegisterHTML() {
  return /* html */ `
    <header>
      <h1>Welcome to Forum</h1>
    </header>
    <main>
      <div class="signup-container">
        <div class="signup-wrapper">
          <h2 class="signup-title">Sign Up</h2>
          <div class="text-base">
            Already a member?
            <a href="/login" data-link>Sign In</a>
          </div>
          <div class="btn-box">
            <a class="signup-provider-btn google" href="/auth/google/login">
              <img src="/static/images/icons/google-logo.png" alt="Google Logo" />
              <p>Sign up with Google</p>
            </a>
            <a class="signup-provider-btn github" href="/auth/github/login">
              <img src="/static/images/icons/github-white-logo.png" alt="Github Logo" />
              <p>Sign up with Github</p>
            </a>
          </div>

          <div class="border">
            <span class="border-text">or</span>
          </div>

          <form class="signup" method="post" action="/register">
            <div class="input-wrapper">
              <div class="input-box">
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

              <div class="input-box">
                <label for="firstname">First Name</label>
                <input
                  type="text"
                  name="firstname"
                  id="firstname"
                  class="form-input"
                  placeholder="Enter your first name"
                />
                <span class="error-message" id="firstname-error"></span>
              </div>

              <div class="input-box">
                <label for="lastname">Last Name</label>
                <input
                  type="text"
                  name="lastname"
                  id="lastname"
                  class="form-input"
                  placeholder="Enter your last name"
                />
                <span class="error-message" id="lastname-error"></span>
              </div>

              <div class="input-box">
                <label for="age">Age</label>
                <input
                  type="number"
                  name="age"
                  id="age"
                  class="form-input"
                  placeholder="Enter your age"
                  min="0"
                />
                <span class="error-message" id="age-error"></span>
              </div>

              <div class="input-box">
                <label for="gender">Gender</label>
                <select name="gender" id="gender" class="form-input">
                  <option value="">Select gender</option>
                  <option value="male">Male</option>
                  <option value="female">Female</option>
                  <option value="other">Other</option>
                  <option value="prefer_not_to_say">Prefer not to say</option>
                </select>
                <span class="error-message" id="gender-error"></span>
              </div>

              <div class="input-box">
                <label for="email">E-mail</label>
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
                    <img
                      src="/static/images/icons/eye.png"
                      alt="Toggle Password Icon"
                      id="eye-icon"
                      class="eye-icon"
                    />
                    <img
                      src="/static/images/icons/hidden.png"
                      alt="Toggle Password Icon"
                      id="hidden-icon"
                      class="hidden-icon"
                    />
                  </label>
                  <input
                    type="checkbox"
                    id="togglePassword"
                    class="hidden-checkbox"
                    aria-label="show/hide password"
                  />
                </div>
                <span class="error-message" id="password-error"></span>
              </div>
            </div>

            <div class="btn-box">
              <button type="reset" class="btn-reset-form">Reset Form</button>
              <button type="submit" class="btn-signup">Sign Up</button>
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

function initFormBehaviour() {
  const form = document.querySelector('.signup');
  if (!form) return;

  const nicknameInput = form.querySelector('#nickname');
  const firstnameInput = form.querySelector('#firstname');
  const lastnameInput = form.querySelector('#lastname');
  const ageInput = form.querySelector('#age');
  const genderInput = form.querySelector('#gender');
  const emailInput = form.querySelector('#email');
  const passwordInput = form.querySelector('#password');

  const nicknameError = document.getElementById('nickname-error');
  const firstnameError = document.getElementById('firstname-error');
  const lastnameError = document.getElementById('lastname-error');
  const ageError = document.getElementById('age-error');
  const genderError = document.getElementById('gender-error');
  const emailError = document.getElementById('email-error');
  const passwordError = document.getElementById('password-error');

  const toggle = document.getElementById('togglePassword');
  toggle?.addEventListener('change', (e) => {
    if (toggle.checked) {
      passwordInput.type = 'text';
    } else {
      passwordInput.type = 'password';
    }
  });

  form.addEventListener('submit', async (e) => {
    e.preventDefault();

    // Clear previous errors
    // usernameError.textContent = '';
    // emailError.textContent = '';
    // passwordError.textContent = '';

    const nickname = String(nicknameInput.value || '').trim();
    const firstname = String(firstnameInput.value || '').trim();
    const lastname = String(lastnameInput.value || '').trim();
    const age = String(ageInput.value || '').trim();
    const gender = String(genderInput.value || '').trim();
    const email = String(emailInput.value || '').trim();
    const password = String(passwordInput.value || '').trim();

    // Basic client-side validation
    if (!nickname) {
      nicknameError.textContent = 'Please enter a nickname.';
      nicknameInput.focus();
      return;
    }
    if (!email) {
      emailError.textContent = 'Please enter an email address.';
      emailInput.focus();
      return;
    }
    if (!password) {
      passwordError.textContent = 'Please enter a password.';
      passwordInput.focus();
      return;
    }

    // Disable form during request
    const submitBtn = form.querySelector('button[type="submit"]');
    submitBtn.disabled = true;
    submitBtn.textContent = 'Signing up...';

    try {
      // Backend expects username, email, password
      // Build payload to match backend RegisterUserReguestModel
      const payload = {
        nickname,
        password,
        email,
        firstname,
        lastname,
        age: age ? Number(age) : 0,
        gender,
      };

      await api.post('/register', payload);

      // On success, navigate to login page
      navigate('/login');
    } catch (err) {
      // Show friendly error messages, try to map to fields when possible
      const msg = err?.message || String(err);

      // Heuristic mapping based on message content
      const lower = String(msg).toLowerCase();
      // Map probable field errors heuristically
      if (lower.includes('nickname') || lower.includes('nick')) {
        nicknameError.textContent = escapeHTML(msg);
        nicknameInput.focus();
      } else if (lower.includes('first') || lower.includes('firstname')) {
        firstnameError.textContent = escapeHTML(msg);
        firstnameInput.focus();
      } else if (lower.includes('last') || lower.includes('lastname')) {
        lastnameError.textContent = escapeHTML(msg);
        lastnameInput.focus();
      } else if (lower.includes('age')) {
        ageError.textContent = escapeHTML(msg);
        ageInput.focus();
      } else if (lower.includes('gender')) {
        genderError.textContent = escapeHTML(msg);
        genderInput.focus();
      } else if (lower.includes('email')) {
        emailError.textContent = escapeHTML(msg);
        emailInput.focus();
      } else {
        // Fallback to password area
        passwordError.textContent = escapeHTML(msg);
        passwordInput.focus();
      }
    } finally {
      submitBtn.disabled = false;
      submitBtn.textContent = 'Sign Up';
    }
  });

  // Reset button should clear errors too
  const resetBtn = form.querySelector('button[type="reset"]');
  resetBtn?.addEventListener('click', () => {
    nicknameError.textContent = '';
    firstnameError.textContent = '';
    lastnameError.textContent = '';
    ageError.textContent = '';
    genderError.textContent = '';
    emailError.textContent = '';
    passwordError.textContent = '';
  });
}
