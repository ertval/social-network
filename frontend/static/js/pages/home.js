/**
 * pages/home.js
 *
 * Renders the homepage into #app-root.
 *
 * This replaces the Go BFF's HomePage handler and the three templates it used:
 *   - frontend/html/pages/home.html
 *   - frontend/html/partials/categories.html
 *   - frontend/html/partials/category_details.html
 *
 * What the Go handler did → what we do here:
 *   1. Build query params (defaultCategoriesOptions) → fetchCategories()
 *   2. Call backend /categories/all                  → api.get (inside fetchCategories)
 *   3. Decode JSON envelope                          → handled in api.js
 *   4. PrepareCategories (color normalisation)       → prepareCategories() from helpers.js
 *   5. Inject .User from context                     → getUser() from auth.js
 *   6. Execute template                              → buildHomeHTML() + inject into DOM
 */

import { fetchCategories } from '../api.js';
import { navigate } from '../router.js';
import { prepareCategories, escapeHTML } from '../helpers.js';
import {
  buildCategoriesListHTML,
  buildCategoryDetailsHTML,
  buildCategoriesSkeletonHTML,
} from '../components/categoryCard.js';

// ─── Public API ───────────────────────────────────────────────────────────────

/**
 * Entry point called by the router when navigating to "/".
 *
 * @param {object|null} user - The currently logged-in user (or null for guests)
 */
export async function renderHomePage(user) {
  const root = document.getElementById('app-root');
  if (!root) return;

  // If not logged in, redirect to login
  if (!user) {
    navigate('/login');
    return;
  }

  // Show skeleton while loading
  root.innerHTML = buildSkeletonHTML();

  let categories = [];
  let error = null;

  try {
    const data = await fetchCategories();

    // The backend wraps category data; handle both array and { categories: [] } shapes
    const raw = Array.isArray(data) ? data : (data?.categories ?? data?.Categories ?? []);

    categories = prepareCategories(raw);
  } catch (err) {
    console.error('Failed to load categories:', err);
    error = err.message || 'Failed to load categories';
  }

  root.innerHTML = error ? buildErrorHTML(error) : buildHomeHTML(categories);

  // Attach interactive behaviour after the DOM is ready
  if (!error) {
    initCategoryDetails();
  }
}

// ─── HTML builders (mirror the Go templates exactly) ─────────────────────────

/**
 * Mirrors: home.html {{ define "content" }}
 */
function buildHomeHTML(categories) {
  return /* html */ `
    <h1 class="forum-title">Welcome to Forum</h1>
    <div class="main-container">
      ${buildCategoryDetailsHTML(categories)}
      ${buildCategoriesListHTML(categories, 'No categories found. Check back later!')}
    </div>
  `;
}

// ─── Loading skeleton ─────────────────────────────────────────────────────────

function buildSkeletonHTML() {
  return /* html */ `
    <h1 class="forum-title">Welcome to Forum</h1>
    <div class="main-container">
      <div class="nav-categories">
        <div class="skeleton skeleton-btn" style="width:110px;height:40px;border-radius:4px"></div>
        <div class="skeleton skeleton-btn" style="width:100px;height:40px;border-radius:4px"></div>
        <div class="skeleton skeleton-btn" style="width:80px;height:40px;border-radius:4px"></div>
      </div>
      ${buildCategoriesSkeletonHTML(3)}
    </div>
  `;
}

// ─── Error state ──────────────────────────────────────────────────────────────

function buildErrorHTML(message) {
  return /* html */ `
    <h1 class="forum-title">Welcome to Forum</h1>
    <div class="main-container">
      <div class="categories-container" style="padding:2rem;text-align:center">
        <p style="color:#e53e3e;font-size:1.1rem">⚠️ ${escapeHTML(message)}</p>
        <p style="margin-top:1rem;color:var(--grey-color)">
          Could not load categories. Please try refreshing the page.
        </p>
      </div>
    </div>
  `;
}

// ─── Interactive behaviour ────────────────────────────────────────────────────

/**
 * Closes the category <details> dropdown when the user clicks outside it.
 * The native <details> element already handles toggle on summary click.
 */
function initCategoryDetails() {
  const details = document.querySelector('.category-details');
  if (!details) return;

  document.addEventListener(
    'click',
    (e) => {
      if (!details.contains(e.target)) {
        details.removeAttribute('open');
      }
    },
    { passive: true }
  );
}
