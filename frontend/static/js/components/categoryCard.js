/**
 * components/categoryCard.js
 *
 * Shared rendering logic for category rows/cards.
 * Used by pages/home.js and pages/categories.js so the markup
 * stays in sync when one of them changes.
 */

import { escapeHTML, formatRelativeDate } from '../helpers.js';

/**
 * Renders a single category row.
 * Mirrors the {{ range . }} block from categories.html.
 *
 * @param {object} cat - A prepared category object (Color already normalised)
 * @returns {string} HTML string
 */
export function buildCategoryRowHTML(cat) {
  const id = escapeHTML(String(cat.ID ?? cat.id ?? ''));
  const name = escapeHTML(cat.Name ?? cat.name ?? '');
  const description = escapeHTML(cat.Description ?? cat.description ?? '');
  const color = cat.Color ?? cat.color ?? '#00C6FF';
  const imagePath = escapeHTML(
    cat.ImagePath ?? cat.image_path ?? '/static/images/categories/default_category.png'
  );

  const topics = Array.isArray(cat.Topics)
    ? cat.Topics
    : Array.isArray(cat.topics)
      ? cat.topics
      : [];

  const topicRows = topics
    .slice(0, 3)
    .map((topic) => {
      const topicID = escapeHTML(String(topic.ID ?? topic.id ?? ''));
      const topicTitle = escapeHTML(topic.Title ?? topic.title ?? 'Untitled');
      const topicDate = formatRelativeDate(topic.CreatedAt ?? topic.created_at ?? '');

      return /* html */ `
        <div class="category-post">
          <a href="/topic/${topicID}" class="topic-link" data-link>
            <span class="category-post-title">
              <span class="left-arrow">&#10147;</span> ${topicTitle}
            </span>
          </a>
          <span class="category-post-date">${topicDate}</span>
        </div>`;
    })
    .join('');

  return /* html */ `
    <div class="category">
      <div class="category-wrapper">
        <div class="category-img-box">
          <a href="/topics?search=&category=${id}" class="category-link" data-link>
            <img class="category-img" src="${imagePath}" alt="${name}" />
          </a>
        </div>
        <div class="category-info">
          <div class="category-info-box">
            <a href="/topics?search=&category=${id}" class="category-link" data-link>
              <div class="category-title-box">
                <span class="category-title-color" style="background-color: ${color}"></span>
                <span class="category-title">${name}</span>
              </div>
            </a>
            <p class="category-description">${description}</p>
          </div>
        </div>
      </div>
      <div class="category-posts">
        ${topicRows || `<span class="category-post-date">No posts yet</span>`}
      </div>
    </div>`;
}

/**
 * Renders the full categories container from an array of prepared categories.
 *
 * @param {Array}  categories
 * @param {string} emptyMessage - shown when the array is empty
 * @returns {string} HTML string
 */
export function buildCategoriesListHTML(categories, emptyMessage = 'No categories found.') {
  if (!categories.length) {
    return /* html */ `
      <div class="categories-container">
        <p class="no-topics-message">${escapeHTML(emptyMessage)}</p>
      </div>`;
  }

  return /* html */ `
    <div class="categories-container">
      ${categories.map(buildCategoryRowHTML).join('')}
    </div>`;
}

/**
 * Renders the category details dropdown + nav buttons.
 * Mirrors category_details.html.
 * Used by home.js (and any future page that needs the same dropdown).
 *
 * @param {Array} categories
 * @returns {string} HTML string
 */
export function buildCategoryDetailsHTML(categories) {
  const links = categories
    .map(
      (cat) => /* html */ `
        <a href="/topics?search=&category=${escapeHTML(String(cat.ID ?? cat.id ?? ''))}"
           class="details-category-link" data-link>
          <div class="details-text-box">
            <span class="category-title-color"
                  style="background-color: ${cat.Color ?? cat.color ?? '#00C6FF'}"></span>
            <span class="details-category-title">
              ${escapeHTML(cat.Name ?? cat.name ?? '')}
            </span>
          </div>
          <span class="category-count">
            ${escapeHTML(String(cat.TopicCount ?? cat.topic_count ?? 0))}
          </span>
        </a>`
    )
    .join('');

  return /* html */ `
    <div class="nav-categories">
      <details class="category-details">
        <summary>Categories</summary>
        <div class="details-content">
          ${links || `<span class="details-category-title">No categories yet</span>`}
        </div>
      </details>
      <a href="/categories" class="nav-categories-btn" data-link>Categories</a>
      <a href="/topics"     class="nav-categories-btn" data-link>Topics</a>
    </div>`;
}

/**
 * Skeleton placeholder for a list of category rows.
 *
 * @param {number} count - how many skeleton rows to show
 * @returns {string} HTML string
 */
export function buildCategoriesSkeletonHTML(count = 3) {
  const row = /* html */ `
    <div class="category category--skeleton">
      <div class="category-wrapper">
        <div class="skeleton skeleton-img"></div>
        <div class="category-info-box">
          <div class="skeleton skeleton-title"></div>
          <div class="skeleton skeleton-desc"></div>
        </div>
      </div>
      <div class="category-posts">
        <div class="skeleton skeleton-post"></div>
        <div class="skeleton skeleton-post"></div>
        <div class="skeleton skeleton-post"></div>
      </div>
    </div>`;

  return /* html */ `
    <div class="categories-container">
      ${row.repeat(count)}
    </div>`;
}
