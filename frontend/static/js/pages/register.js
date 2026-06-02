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
    <div class="page-title-box">
      <h1>Welcome to Forum</h1>
    </div>
    <div class="signup-container">
      <div class="signup-wrapper">
        <h2 class="signup-title">Sign Up</h2>
        <div class="text-base">
          Already a member?
          <a href="/login" data-link>Sign In</a>
        </div>
 
        <div class="btn-box">
          <a class="signup-provider-btn google" href="api/v1/auth/google/login">
            <img src="/static/images/icons/google-logo.png" alt="Google Logo" />
            <p>Sign up with Google</p>
          </a>
          <a class="signup-provider-btn github" href="api/v1/auth/github/login">
            <img src="/static/images/icons/github-white-logo.png" alt="Github Logo" />
            <p>Sign up with Github</p>
          </a>
        </div>
 
        <div class="border">
          <span class="border-text">or</span>
        </div>
 
        <form class="signup">
          <div class="input-wrapper">
 
            <div class="input-box">
              <label for="nickname">Nickname</label>
              <input type="text" name="nickname" id="nickname"
                     class="form-input" placeholder="Enter your nickname" />
              <span class="error-message" id="nickname-error"></span>
            </div>
 
            <div class="input-box">
              <label for="firstname">First Name</label>
              <input type="text" name="firstname" id="firstname"
                     class="form-input" placeholder="Enter your first name" />
              <span class="error-message" id="firstname-error"></span>
            </div>
 
            <div class="input-box">
              <label for="lastname">Last Name</label>
              <input type="text" name="lastname" id="lastname"
                     class="form-input" placeholder="Enter your last name" />
              <span class="error-message" id="lastname-error"></span>
            </div>
 
            <div class="input-box">
              <label for="age">Age</label>
              <input type="number" name="age" id="age" min="0"
                     class="form-input" placeholder="Enter your age" />
              <span class="error-message" id="age-error"></span>
            </div>
 
            <div class="input-box">
              <label for="gender">Gender</label>
              <select name="gender" id="gender" class="form-input form-select">
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
              <input type="email" name="email" id="email"
                     class="form-input" placeholder="Enter your email" />
              <span class="error-message" id="email-error"></span>
            </div>
 
            <div class="input-box">
              <div class="password-wrapper">
                <label for="password">Password</label>
                <input type="password" name="password" id="password"
                       class="form-input" placeholder="Enter your password" />
                <label for="togglePassword" class="togglePassword">
                  <!-- hidden-icon starts hidden; JS toggles both on checkbox change -->
                  <img src="/static/images/icons/eye.png"    alt="Show password"
                       id="eye-icon"    class="eye-icon" />
                  <img src="/static/images/icons/hidden.png" alt="Hide password"
                       id="hidden-icon" class="hidden-icon" style="display:none" />
                </label>
                <input type="checkbox" id="togglePassword"
                       class="hidden-checkbox" aria-label="show/hide password" />
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
  `;
}

function initFormBehaviour() {
  const form = document.querySelector('.signup');
  if (!form) return;

  // Field refs
  const nicknameInput = form.querySelector('#nickname');
  const firstnameInput = form.querySelector('#firstname');
  const lastnameInput = form.querySelector('#lastname');
  const ageInput = form.querySelector('#age');
  const genderInput = form.querySelector('#gender');
  const emailInput = form.querySelector('#email');
  const passwordInput = form.querySelector('#password');

  // Error refs
  const nicknameError = document.getElementById('nickname-error');
  const firstnameError = document.getElementById('firstname-error');
  const lastnameError = document.getElementById('lastname-error');
  const ageError = document.getElementById('age-error');
  const genderError = document.getElementById('gender-error');
  const emailError = document.getElementById('email-error');
  const passwordError = document.getElementById('password-error');

  // ---Password toggle---
  const toggle = document.getElementById('togglePassword');
  const eyeIcon = document.getElementById('eye-icon');
  const hiddenIcon = document.getElementById('hidden-icon');

  toggle?.addEventListener('change', () => {
    if (toggle.checked) {
      passwordInput.type = 'text';
      eyeIcon.style.display = 'none';
      hiddenIcon.style.display = 'block';
    } else {
      passwordInput.type = 'password';
      eyeIcon.style.display = 'block';
      hiddenIcon.style.display = 'none';
    }
  });

  // ---Clear errors on input---
  const errorMap = [
    [nicknameInput, nicknameError],
    [firstnameInput, firstnameError],
    [lastnameInput, lastnameError],
    [ageInput, ageError],
    [genderInput, genderError],
    [emailInput, emailError],
    [passwordInput, passwordError],
  ];
  errorMap.forEach(([input, errEl]) => {
    input?.addEventListener('input', () => {
      if (errEl) errEl.textContent = '';
    });
  });

  // ---Reset button clears errors too---
  form.querySelector('button[type="reset"]')?.addEventListener('click', () => {
    errorMap.forEach(([, errEl]) => {
      if (errEl) errEl.textContent = '';
    });
    // Restore password visibility
    passwordInput.type = 'password';
    eyeIcon.style.display = 'block';
    hiddenIcon.style.display = 'none';
    if (toggle) toggle.checked = false;
  });

  // ---Submit---
  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    clearAllErrors(errorMap);

    const nickname = nicknameInput.value.trim();
    const firstname = firstnameInput.value.trim();
    const lastname = lastnameInput.value.trim();
    const age = ageInput.value.trim();
    const gender = genderInput.value.trim();
    const email = emailInput.value.trim();
    const password = passwordInput.value.trim();

    // Client-side validation
    let hasError = false;
    if (!nickname) {
      nicknameError.textContent = 'Please enter a nickname.';
      nicknameInput.focus();
      hasError = true;
    }
    if (!firstname && !hasError) {
      firstnameError.textContent = 'Please enter your first name.';
      firstnameInput.focus();
      hasError = true;
    }
    if (!lastname && !hasError) {
      lastnameError.textContent = 'Please enter your last name.';
      lastnameInput.focus();
      hasError = true;
    }
    if (!age && !hasError) {
      ageError.textContent = 'Please enter your age.';
      ageInput.focus();
      hasError = true;
    }
    if (!gender && !hasError) {
      genderError.textContent = 'Please select a gender.';
      genderInput.focus();
      hasError = true;
    }
    if (!email && !hasError) {
      emailError.textContent = 'Please enter an email address.';
      emailInput.focus();
      hasError = true;
    }
    if (!password && !hasError) {
      passwordError.textContent = 'Please enter a password.';
      passwordInput.focus();
      hasError = true;
    }
    if (hasError) return;

    const submitBtn = form.querySelector('button[type="submit"]');
    submitBtn.disabled = true;
    submitBtn.textContent = 'Signing up...';
    try {
      await api.post('/register', {
        nickname,
        email,
        password,
        firstname,
        lastname,
        age: age ? Number(age) : 0,
        gender,
      });

      navigate('/login');
    } catch (err) {
      mapErrorToField(
        err?.message || String(err),
        {
          nicknameError,
          firstnameError,
          lastnameError,
          ageError,
          genderError,
          emailError,
          passwordError,
        },
        {
          nicknameInput,
          firstnameInput,
          lastnameInput,
          ageInput,
          genderInput,
          emailInput,
          passwordInput,
        }
      );
    } finally {
      submitBtn.disabled = false;
      submitBtn.textContent = 'Sign Up';
    }
  });
}

// ─── Helpers ──────────────────────────────────────────────────────────────────
function clearAllErrors(errorMap) {
  errorMap.forEach(([, errEl]) => {
    if (errEl) errEl.textContent = '';
  });
}

function mapErrorToField(msg, errors, inputs) {
  const lower = msg.toLowerCase();
  if (lower.includes('nickname') || lower.includes('nick')) {
    errors.nicknameError.textContent = escapeHTML(msg);
    inputs.nicknameInput.focus();
  } else if (lower.includes('first')) {
    errors.firstnameError.textContent = escapeHTML(msg);
    inputs.firstnameInput.focus();
  } else if (lower.includes('last')) {
    errors.lastnameError.textContent = escapeHTML(msg);
    inputs.lastnameInput.focus();
  } else if (lower.includes('age')) {
    errors.ageError.textContent = escapeHTML(msg);
    inputs.ageInput.focus();
  } else if (lower.includes('gender')) {
    errors.genderError.textContent = escapeHTML(msg);
    inputs.genderInput.focus();
  } else if (lower.includes('email')) {
    errors.emailError.textContent = escapeHTML(msg);
    inputs.emailInput.focus();
  } else {
    errors.passwordError.textContent = escapeHTML(msg);
    inputs.passwordInput.focus();
  }
}
