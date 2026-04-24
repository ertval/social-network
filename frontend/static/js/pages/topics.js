/**
 * pages/topics.js
 *
 * Renders the /topics page into #app-root.
 *
 * Replaces:
 *   - Go BFF TopicsPage handler (cmd/client/server/topics_handler.go)
 *   - frontend/html/pages/all_topics.html
 *
 * What the Go handler did → what we do here:
 *   1. Parse query params (page, search, order_by, order, category, page_size)
 *   2. Call backend /topics/all with those params
 *   3. Decode JSON envelope                    → handled in api.js
 *   4. Normalize CategoryColors on each topic  → normalizeTopicColors()
 *   5. Inject .User from context               → user param from router
 *   6. Execute template                        → buildPageHTML() + inject into DOM
 *
 * Template feature carried over:
 *   - truncate helper ({{ truncate .Content 100 }}) → truncate() in helpers
 *   - CategoryColors slice with matching CategoryNames (multi-category badges)
 *   - Single CategoryColor fallback for backward compatibility
 */

import { fetchTopics, fetchCategories } from '../api.js';
import { navigate } from '../router.js';
import { prepareCategories, normalizeColor, escapeHTML, formatRelativeDate } from '../helpers.js';
import { buildPaginationHTML } from '../components/pagination.js';

const DEFAULT_PAGE_SIZE = 10;

// ─── Public API ───────────────────────────────────────────────────────────────

export async function renderTopicsPage(user) {
  const root = document.getElementById('app-root');
  if (!root) return;

  const filters = parseFiltersFromURL();

  root.innerHTML = buildSkeletonHTML();

  try {
    // Fetch topics and categories in parallel — categories are needed
    // to populate the category filter <select>, just like the Go handler did.
    const [topicsData, categoriesData] = await Promise.all([
      fetchTopics({
        order_by: filters.order_by,
        order: filters.order,
        search: filters.search,
        category: filters.category || undefined,
        page: filters.page,
        page_size: filters.page_size,
      }),
      fetchCategories({ order_by: 'name', order: 'asc', page: 1, page_size: 100 }),
    ]);

    // ── Topics ──────────────────────────────────────────────────────────────
    const rawTopics = Array.isArray(topicsData)
      ? topicsData
      : (topicsData?.topics ?? topicsData?.Topics ?? []);

    const topics = normalizeTopicColors(rawTopics);

    // ── Pagination ──────────────────────────────────────────────────────────
    const pagination = extractPagination(topicsData, filters.page, rawTopics.length);

    // ── Categories (for the filter dropdown) ────────────────────────────────
    const rawCats = Array.isArray(categoriesData)
      ? categoriesData
      : (categoriesData?.categories ?? categoriesData?.Categories ?? []);

    const categories = prepareCategories(rawCats);

    root.innerHTML = buildPageHTML(topics, categories, pagination, filters);

    initFilterForm(filters);
    highlightActiveCategoryBtn(filters.category);
  } catch (err) {
    console.error('Failed to load topics:', err);
    root.innerHTML = buildErrorHTML(err.message || 'Failed to load topics');
  }
}

// ─── Filter parsing ───────────────────────────────────────────────────────────

function parseFiltersFromURL() {
  const p = new URLSearchParams(window.location.search);
  return {
    search: p.get('search') || '',
    order_by: p.get('order_by') || 'created_at',
    order: p.get('order') || 'desc',
    category: parseInt(p.get('category') || '0', 10),
    page: parseInt(p.get('page') || '1', 10),
    page_size: parseInt(p.get('page_size') || String(DEFAULT_PAGE_SIZE), 10),
  };
}

// ─── Color normalisation (mirrors the Go BFF loop over CategoryColors) ────────

/**
 * Normalises CategoryColors on every topic.
 * Go: for j := range pageData.Topics[i].CategoryColors { NormalizeColor(...) }
 */
function normalizeTopicColors(topics) {
  return topics.map((t) => {
    const colors = Array.isArray(t.CategoryColors ?? t.category_colors)
      ? (t.CategoryColors ?? t.category_colors).map(normalizeColor)
      : [];

    return {
      ...t,
      CategoryColors: colors,
      // Keep single-color fallback for backward compat
      CategoryColor: normalizeColor(t.CategoryColor ?? t.category_color ?? ''),
    };
  });
}

// ─── Pagination extraction ────────────────────────────────────────────────────

function extractPagination(data, currentPage, itemCount) {
  return {
    page: data?.pagination?.page ?? data?.Pagination?.Page ?? currentPage,
    total_pages: data?.pagination?.total_pages ?? data?.Pagination?.TotalPages ?? 1,
    total_items: data?.pagination?.total_items ?? data?.Pagination?.TotalItems ?? itemCount,
    prev_page: data?.pagination?.prev_page ?? data?.Pagination?.PrevPage ?? null,
    next_page: data?.pagination?.next_page ?? data?.Pagination?.NextPage ?? null,
  };
}

// ─── HTML builders ────────────────────────────────────────────────────────────

function buildPageHTML(topics, categories, pagination, filters) {
  const hasTopics = topics.length > 0;
  const moreThanOne = pagination.total_pages > 1;

  return /* html */ `
    <div class="main-container">
      <div class="category-container">
        <div class="category-content">
          <h1 class="forum-title">All Topics</h1>

          ${buildFilterSectionHTML(categories, filters)}

          <div class="topic-list">
            ${buildTableHeaderHTML()}
            ${
              hasTopics
                ? topics.map(buildTopicRowHTML).join('')
                : `<div class="no-topics-message"><p>No topics found matching your criteria.</p></div>`
            }
          </div>

          ${
            moreThanOne
              ? buildPaginationHTML(pagination, filters, '/topics')
              : `<div class="out-of-topics">
                 <p>${hasTopics ? "You've seen all the topics." : 'No topics available.'}</p>
               </div>`
          }
        </div>
      </div>
    </div>
  `;
}

// ── Filter section ────────────────────────────────────────────────────────────

function buildFilterSectionHTML(categories, filters) {
  const categoryOptions = categories
    .map((cat) => {
      const id = cat.ID ?? cat.id ?? 0;
      const name = escapeHTML(cat.Name ?? cat.name ?? '');
      const selected = filters.category === id ? 'selected' : '';
      return `<option value="${id}" ${selected}>${name}</option>`;
    })
    .join('');

  return /* html */ `
    <div class="topics-filter-section">
      <form class="topics-filter-form">

        <div class="search-wrapper">
          <input
            type="text"
            name="search"
            placeholder="Search topics..."
            value="${escapeHTML(filters.search)}"
            maxlength="100"
            class="search-input"
          />
          <button type="submit" class="search-btn" aria-label="Search">
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20"
                 viewBox="0 0 24 24" fill="none" stroke="currentColor"
                 stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="11" cy="11" r="8"></circle>
              <path d="m21 21-4.35-4.35"></path>
            </svg>
          </button>
        </div>

        <div class="filter-controls">

          <div class="filter-group">
            <label for="category">Category:</label>
            <select name="category" id="category" class="filter-select">
              <option value="0" ${filters.category === 0 ? 'selected' : ''}>All Categories</option>
              ${categoryOptions}
            </select>
          </div>

          <div class="filter-group">
            <label for="order_by">Sort by:</label>
            <select name="order_by" id="order_by" class="filter-select">
              <option value="created_at" ${filters.order_by === 'created_at' ? 'selected' : ''}>Date Created</option>
              <option value="updated_at" ${filters.order_by === 'updated_at' ? 'selected' : ''}>Last Updated</option>
              <option value="title"      ${filters.order_by === 'title' ? 'selected' : ''}>Title</option>
              <option value="vote_score" ${filters.order_by === 'vote_score' ? 'selected' : ''}>Most Popular</option>
            </select>
          </div>

          <div class="filter-group">
            <label for="order">Order:</label>
            <select name="order" id="order" class="filter-select">
              <option value="desc" ${filters.order === 'desc' ? 'selected' : ''}>Descending</option>
              <option value="asc"  ${filters.order === 'asc' ? 'selected' : ''}>Ascending</option>
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

// ── Table header ──────────────────────────────────────────────────────────────

function buildTableHeaderHTML() {
  return /* html */ `
    <div class="topic-row">
      <div class="head-wrapper">
        <span class="topic-title-head">TOPIC</span>
        <div class="other-heads">
          <span class="topic-poster-head">AUTHOR</span>
          <span class="topic-replies-head">SCORE</span>
          <span class="topic-replies-head">UPDATED</span>
        </div>
      </div>
    </div>
  `;
}

// ── Single topic row ──────────────────────────────────────────────────────────

/**
 * Mirrors the {{ range .Topics }} block in all_topics.html.
 *
 * Handles both multi-category (CategoryColors + CategoryNames arrays)
 * and single-category (CategoryColor + CategoryName) shapes from the backend.
 */
function buildTopicRowHTML(topic) {
  const id = escapeHTML(String(topic.ID ?? topic.id ?? ''));
  const title = escapeHTML(topic.Title ?? topic.title ?? 'Untitled');
  const content = escapeHTML(topic.Content ?? topic.content ?? '');
  const author = escapeHTML(topic.OwnerUsername ?? topic.owner_username ?? topic.Username ?? '');
  const upvotes = topic.UpvoteCount ?? topic.upvote_count ?? 0;
  const downvotes = topic.DownvoteCount ?? topic.downvote_count ?? 0;
  const updatedAt = formatRelativeDate(topic.UpdatedAt ?? topic.updated_at ?? '');

  // Category badges — multi-category takes priority, fall back to single
  const categoryBadgesHTML = buildCategoryBadgesHTML(topic);

  return /* html */ `
    <div class="topic-row">
      <div class="topic-content-wrapper">

        <div class="topic-text">
          <div class="color-category">
            ${categoryBadgesHTML}
          </div>
          <div class="topic-title">
            <a href="/topic/${id}" data-link>${title}</a>
            <p class="topic-preview">${truncate(content, 100)}</p>
          </div>
        </div>

        <div class="topic-meta">
          <span class="topic-author">${author}</span>

          <span class="topic-score">
            <span class="upvotes">
              <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 32 32">
                <path d="M27 11h-8.52L19 9.8A6.42 6.42 0 0 0 13 1a1 1 0 0 0-.93.63L8.32 11H5a3 3 0 0 0-3 3v14a3 3 0 0 0 3 3h18.17a3 3 0 0 0 2.12-.88l3.83-3.83a3 3 0 0 0 .88-2.12V14a3 3 0 0 0-3-3zM4 28V14a1 1 0 0 1 1-1h3v16H5a1 1 0 0 1-1-1zm24-3.83a1 1 0 0 1-.29.71l-3.83 3.83a1.05 1.05 0 0 1-.71.29H10V12.19l3.66-9.14a4.31 4.31 0 0 1 3 1.89 4.38 4.38 0 0 1 .44 4.12l-1 2.57A1 1 0 0 0 17 13h10a1 1 0 0 1 1 1z"/>
              </svg>
              <span class="like-count">${upvotes}</span>
            </span>
            <span class="downvotes">
              <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 32 32">
                <path d="m29.12 5.71-3.83-3.83A3 3 0 0 0 23.17 1H5a3 3 0 0 0-3 3v14a3 3 0 0 0 3 3h3.32l3.75 9.37A1 1 0 0 0 13 31a6.42 6.42 0 0 0 6-8.8l-.52-1.2H27a3 3 0 0 0 3-3V7.83a3 3 0 0 0-.88-2.12zM4 18V4a1 1 0 0 1 1-1h3v16H5a1 1 0 0 1-1-1zm24 0a1 1 0 0 1-1 1H17a1 1 0 0 0-.93 1.37l1 2.57a4.38 4.38 0 0 1-.44 4.12 4.31 4.31 0 0 1-3 1.89L10 19.81V3h13.17a1 1 0 0 1 .71.29l3.83 3.83a1 1 0 0 1 .29.71z"/>
              </svg>
              <span class="dislike-count">${downvotes}</span>
            </span>
          </span>

          <span class="topic-date">${updatedAt}</span>
        </div>

      </div>
    </div>
  `;
}

/**
 * Builds the category badge(s) for a topic row.
 * Mirrors the Go template's {{ if .CategoryColors }} / {{ else }} branching.
 */
function buildCategoryBadgesHTML(topic) {
  const colors = topic.CategoryColors ?? topic.category_colors ?? [];
  const names = topic.CategoryNames ?? topic.category_names ?? [];

  // Multi-category path
  if (Array.isArray(colors) && colors.length > 0) {
    return colors
      .map((color, i) => {
        const name = escapeHTML(names[i] ?? '');
        return /* html */ `
          <div class="category-badge">
            <span class="category-color" style="background-color: ${color}"></span>
            <span class="category-name">${name}</span>
          </div>`;
      })
      .join('');
  }

  // Single-category fallback
  const color = topic.CategoryColor ?? topic.category_color ?? '';
  const name = escapeHTML(topic.CategoryName ?? topic.category_name ?? '');
  if (!color && !name) return '';

  return /* html */ `
    <div class="category-badge">
      <span class="category-color" style="background-color: ${color}"></span>
      <span class="category-name">${name}</span>
    </div>`;
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

/**
 * Mirrors the Go template's {{ truncate .Content 100 }} custom function.
 */
function truncate(str, maxLength) {
  if (!str) return '';
  if (str.length <= maxLength) return str;
  return str.slice(0, maxLength) + '...';
}

function initFilterForm(currentFilters) {
  const form = document.querySelector('.topics-filter-form');
  if (!form) return;

  form.addEventListener('submit', (e) => {
    e.preventDefault();
    const formData = new FormData(form);
    const params = new URLSearchParams();

    for (const [key, value] of formData.entries()) {
      const trimmed = String(value).trim();
      // Always include category even if "0" so the select resets correctly
      if (trimmed && (trimmed !== '0' || key === 'category')) {
        params.set(key, trimmed);
      }
    }

    navigate(`/topics?${params.toString()}`);
  });

  const clearBtn = form.querySelector('.clear-btn');
  clearBtn?.addEventListener('click', () => {
    form.reset();
    navigate('/topics');
  });
}

/**
 * Adds the "active" class to the nav-categories-btn that matches the
 * current page, mirroring script.js's DOMContentLoaded handler.
 */
function highlightActiveCategoryBtn(categoryId) {
  const path = window.location.pathname;
  document.querySelectorAll('.nav-categories-btn').forEach((btn) => {
    btn.classList.toggle('active', btn.getAttribute('href') === path);
  });
}

// ─── Skeleton / Error states ──────────────────────────────────────────────────

function buildSkeletonHTML() {
  const skeletonRow = /* html */ `
    <div class="topic-row">
      <div class="topic-content-wrapper">
        <div class="topic-text" style="flex:1">
          <div class="skeleton" style="width:80px;height:14px;margin-bottom:8px"></div>
          <div class="skeleton" style="width:260px;height:18px;margin-bottom:6px"></div>
          <div class="skeleton" style="width:200px;height:13px"></div>
        </div>
        <div class="topic-meta" style="gap:1.5rem">
          <div class="skeleton" style="width:70px;height:14px"></div>
          <div class="skeleton" style="width:50px;height:14px"></div>
          <div class="skeleton" style="width:60px;height:14px"></div>
        </div>
      </div>
    </div>`;

  return /* html */ `
    <div class="main-container">
      <div class="category-container">
        <div class="category-content">
          <h1 class="forum-title">All Topics</h1>
          <div class="topics-filter-section">
            <div class="skeleton" style="height:50px;border-radius:4px"></div>
          </div>
          <div class="topic-list">
            ${buildTableHeaderHTML()}
            ${skeletonRow.repeat(DEFAULT_PAGE_SIZE)}
          </div>
        </div>
      </div>
    </div>`;
}

function buildErrorHTML(message) {
  return /* html */ `
    <div class="main-container">
      <div class="category-container">
        <h1 class="forum-title">All Topics</h1>
        <div class="topic-list" style="padding:2rem;text-align:center">
          <p style="color:#e53e3e;font-size:1.1rem">⚠️ ${escapeHTML(message)}</p>
          <p style="margin-top:1rem;color:var(--grey-color)">
            Could not load topics. Please try refreshing the page.
          </p>
        </div>
      </div>
    </div>`;
}
