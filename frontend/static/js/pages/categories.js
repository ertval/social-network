/**
 * pages/categories.js
 */

import { fetchCategories } from '../api.js';
import { navigate } from '../router.js';
import { prepareCategories, escapeHTML } from '../helpers.js';
import {
  buildCategoriesListHTML,
  buildCategoriesSkeletonHTML,
} from '../components/categoryCard.js';
import { buildPaginationHTML } from '../components/pagination.js';

export async function renderCategoriesPage(user) {
  const root = document.getElementById('app-root');
  if (!root) return;

  const urlParams = new URLSearchParams(window.location.search);
  const filters = {
    search: urlParams.get('search') || '',
    order_by: urlParams.get('order_by') || 'created_at',
    order: urlParams.get('order') || 'desc',
    page: parseInt(urlParams.get('page') || '1', 10),
  };

  root.innerHTML = buildSkeletonHTML();

  try {
    const data = await fetchCategories(filters);

    // 1. Robust data extraction
    const rawCategories = Array.isArray(data) ? data : (data?.categories ?? data?.Categories ?? []);

    // 2. Fix "undefined" by providing strict defaults for pagination
    // Matches Go JSON naming (TotalPages) and JS naming (total_pages)
    const pagination = {
      page: data?.pagination?.page ?? data?.Pagination?.Page ?? filters.page,
      total_pages: data?.pagination?.total_pages ?? data?.Pagination?.TotalPages ?? 1,
      total_items:
        data?.pagination?.total_items ?? data?.Pagination?.TotalItems ?? rawCategories.length,
      prev_page: data?.pagination?.prev_page ?? data?.Pagination?.PrevPage ?? null,
      next_page: data?.pagination?.next_page ?? data?.Pagination?.NextPage ?? null,
    };

    const categories = prepareCategories(rawCategories);

    root.innerHTML = buildPageHTML(categories, pagination, filters);

    // Attach listeners
    initFilterForm();
  } catch (err) {
    console.error('Failed to load categories:', err);
    root.innerHTML = buildErrorHTML(err.message || 'Failed to load categories');
  }
}

// ─── Updated HTML Builders ──────────────────────────────────────────────────

function buildPageHTML(categories, pagination, filters) {
  return /* html */ `
    <h1 class="forum-title">Categories</h1>
    <div class="main-container">
      ${buildFilterSectionHTML(filters)}
      ${buildCategoriesListHTML(categories, 'No categories found matching your search.')}
      ${buildPaginationHTML(pagination, filters)}
    </div>
  `;
}

function buildFilterSectionHTML(filters) {
  return /* html */ `
    <div class="topics-filter-section">
      <form class="topics-filter-form">
        <div class="search-wrapper">
          <input 
            type="text" 
            name="search" 
            placeholder="Search categories..." 
            value="${escapeHTML(filters.search)}"
            maxlength="100" 
            class="search-input"
          />
          <button type="submit" class="search-btn">
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="11" cy="11" r="8"></circle>
              <path d="m21 21-4.35-4.35"></path>
            </svg>
          </button>
        </div>

        <div class="filter-controls">
          <div class="filter-group">
            <label for="order_by">Sort by:</label>
            <select name="order_by" id="order_by" class="filter-select">
              <option value="created_at" ${filters.order_by === 'created_at' ? 'selected' : ''}>Date Created</option>
              <option value="name" ${filters.order_by === 'name' ? 'selected' : ''}>Name</option>
            </select>
          </div>

          <div class="filter-group">
            <label for="order">Order:</label>
            <select name="order" id="order" class="filter-select">
              <option value="desc" ${filters.order === 'desc' ? 'selected' : ''}>Descending</option>
              <option value="asc" ${filters.order === 'asc' ? 'selected' : ''}>Ascending</option>
            </select>
          </div>

          <div class="filter-buttons">
            <button type="submit" class="apply-btn">Apply</button>
            <button type="button" class="clear-btn">Clear</button>
          </div>
        </div>
      </form>
    </div>
  `;
}

// ─── Helpers ────────────────────────────────────────────────────────────────

function initFilterForm() {
  const form = document.querySelector('.topics-filter-form');
  if (!form) return;

  // 1. Handle search/apply
  form.addEventListener('submit', (e) => {
    e.preventDefault();
    const formData = new FormData(form);
    const params = new URLSearchParams();

    for (const [key, value] of formData.entries()) {
      const trimmed = value.trim();
      if (trimmed) params.set(key, trimmed);
    }

    navigate(`/categories?${params.toString()}`);
  });

  // 2. Direct reset on the Clear button
  const clearBtn = form.querySelector('.clear-btn');
  if (clearBtn) {
    clearBtn.addEventListener('click', () => {
      form.reset(); // Clears all inputs in the form instantly
      navigate('/categories'); // Wipes the URL and triggers a clean fetch
    });
  }
}

function buildSkeletonHTML() {
  return /* html */ `
    <h1 class="forum-title">Categories</h1>
    <div class="main-container">
      <div class="topics-filter-section">
        <div class="skeleton" style="height: 50px; border-radius: 4px;"></div>
      </div>
      <div class="categories-container">
        ${buildCategoriesSkeletonHTML(3)}
      </div>
    </div>
  `;
}

function buildErrorHTML(message) {
  return /* html */ `
    <h1 class="forum-title">Categories</h1>
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
