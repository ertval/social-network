/**
 * components/pagination.js
 *
 * Shared pagination HTML builder.
 * Used by pages/categories.js, pages/topics.js, and any future list page.
 *
 * @param {object} pagination   - { page, total_pages, total_items, prev_page, next_page }
 * @param {object} filters      - current filter state (search, order_by, order, etc.)
 * @param {string} basePath     - e.g. "/topics" or "/categories"
 * @param {string} [itemLabel]  - label for the total count line, default "items"
 * @returns {string} HTML string, empty string if only one page
 */
export function buildPaginationHTML(pagination, filters, basePath, itemLabel = 'items') {
  if (!pagination.total_pages || pagination.total_pages <= 1) return '';

  const prevHTML = pagination.prev_page
    ? `<a href="${buildPageUrl(basePath, pagination.prev_page, filters)}"
          class="pagination-btn prev-btn" data-link>
         ${arrowLeft()} Previous
       </a>`
    : `<span class="pagination-btn prev-btn disabled">
         ${arrowLeft()} Previous
       </span>`;

  const nextHTML = pagination.next_page
    ? `<a href="${buildPageUrl(basePath, pagination.next_page, filters)}"
          class="pagination-btn next-btn" data-link>
         Next ${arrowRight()}
       </a>`
    : `<span class="pagination-btn next-btn disabled">
         Next ${arrowRight()}
       </span>`;

  return /* html */ `
    <div class="pagination-container">
      <div class="pagination">
        ${prevHTML}
        <span class="page-number active">${pagination.page}</span>
        ${nextHTML}
      </div>
      <div class="pagination-info">
        Page ${pagination.page} of ${pagination.total_pages}
        (${pagination.total_items} ${itemLabel} total)
      </div>
    </div>
  `;
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

function buildPageUrl(basePath, page, filters) {
  const params = new URLSearchParams();
  params.set('page', page);

  // Carry over every active filter — skip falsy / zero values so the URL stays clean
  const carry = ['search', 'order_by', 'order', 'category'];
  carry.forEach((key) => {
    const val = filters[key];
    if (val !== undefined && val !== null && val !== '' && val !== 0) {
      params.set(key, val);
    }
  });

  return `${basePath}?${params.toString()}`;
}

function arrowLeft() {
  return /* html */ `
    <span class="pagination-arrow previous-arrow">
      <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24">
        <path d="M12 2a10 10 0 1 0 10 10A10.011 10.011 0 0 0 12 2zm0 18a8 8 0 1 1 8-8 8.009 8.009 0 0 1-8 8z"/>
        <path d="M13.293 7.293 8.586 12l4.707 4.707 1.414-1.414L11.414 12l3.293-3.293-1.414-1.414z"/>
      </svg>
    </span>`;
}

function arrowRight() {
  return /* html */ `
    <span class="pagination-arrow next-arrow">
      <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24">
        <path d="M12 2a10 10 0 1 0 10 10A10.011 10.011 0 0 0 12 2zm0 18a8 8 0 1 1 8-8 8.009 8.009 0 0 1-8 8z"/>
        <path d="M9.293 8.707 12.586 12l-3.293 3.293 1.414 1.414L15.414 12l-4.707-4.707-1.414 1.414z"/>
      </svg>
    </span>`;
}
