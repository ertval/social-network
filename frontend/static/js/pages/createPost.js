/**
 * pages/createPost.js
 *
 * Renders the /topics/create page into #app-root.
 *
 * Replaces:
 *   - Go BFF CreateTopicPage handler  (GET  /topics/create)
 *   - Go BFF CreateTopicPost handler  (POST /topics/create)
 *   - frontend/html/pages/create_post.html
 *   - frontend/static/js/create-post.js
 *
 * What the Go handler did → what we do here:
 *   1. Fetch /categories/all for the multi-select dropdown → fetchCategories()
 *   2. Normalize category colors                          → prepareCategories()
 *   3. Render the form                                    → buildPageHTML()
 *   4. Parse multipart form (title, content, categories,
 *      image_path)                                        → FormData POST to backend
 *   5. Validate client-side before sending               → validateForm()
 *   6. On success redirect to /topics                    → navigate('/topics')
 *
 * Image upload note:
 *   The BFF saved the file to the server filesystem and sent the path to the
 *   backend. In CSR the browser cannot touch the filesystem, so we send the
 *   file directly to the backend as multipart/form-data. The backend must
 *   accept multipart on POST /api/v1/topics/create and handle the file save.
 *   Until that backend change lands, image upload is wired up and ready —
 *   it just won't persist if the backend still expects JSON.
 */

import { api, fetchCategories } from '../api.js';
import { navigate } from '../router.js';
import { prepareCategories, escapeHTML } from '../helpers.js';

// ─── Public API ───────────────────────────────────────────────────────────────

export async function renderCreatePostPage(user) {
  const root = document.getElementById('app-root');
  if (!root) return;

  root.innerHTML = buildSkeletonHTML();

  try {
    const data = await fetchCategories({ order_by: 'name', order: 'asc', page: 1, page_size: 100 });

    const raw = Array.isArray(data) ? data : (data?.categories ?? data?.Categories ?? []);

    const categories = prepareCategories(raw);

    root.innerHTML = buildPageHTML(categories);

    initMultiSelect();
    initFileUpload();
    initFormSubmit(categories);
  } catch (err) {
    console.error('Failed to load create post page:', err);
    root.innerHTML = buildErrorHTML(err.message || 'Failed to load page.');
  }
}

// ─── HTML ─────────────────────────────────────────────────────────────────────

function buildPageHTML(categories) {
  const categoryOptions = categories
    .map((cat) => {
      const id = cat.id ?? cat.ID ?? 0;
      const name = escapeHTML(cat.name ?? cat.Name ?? '');
      const color = cat.Color ?? cat.color ?? '#ccc';
      return /* html */ `
      <li class="option">
        <label>
          <input type="checkbox" name="categories" value="${id}" />
          <span class="option-label">${name}</span>
          <span class="option-color" style="background-color: ${color}"></span>
        </label>
      </li>`;
    })
    .join('');

  return /* html */ `
    <div class="create-post-container">
      <form class="form" id="createPostForm">
        <div class="form-wrapper">
          <h2 class="post-title">Create Topic</h2>

          <!-- Category multi-select -->
          <div class="field">
            <label class="label">Select categories</label>
            <div class="multi-select" id="categorySelect" tabindex="0"
                 aria-haspopup="listbox" aria-expanded="false" aria-multiselectable="true">
              <div class="multi-selected" id="multiSelected">
                <span class="placeholder">Select one or more categories</span>
                <div class="chips" id="chips"></div>
                <button class="chev" type="button" aria-hidden="true">▾</button>
              </div>
              <ul class="options" id="multiOptions" role="listbox"
                  aria-multiselectable="true" tabindex="-1">
                ${categoryOptions}
              </ul>
            </div>
            <div class="field-error" id="error-categories"></div>
          </div>

          <!-- Title -->
          <div class="field">
            <label class="label" for="title">Title</label>
            <input class="input" id="title" name="title" type="text"
                   placeholder="Enter topic title..." required />
            <div class="field-error" id="error-title"></div>
          </div>

          <!-- Content -->
          <div class="field">
            <label class="label" for="content">Content</label>
            <textarea class="input textarea" id="content" name="content"
                      rows="8" placeholder="Write your topic content..." required></textarea>
            <div class="field-error" id="error-content"></div>
          </div>

          <!-- Image upload (optional) -->
          <div class="field">
            <label class="label">Attach an image (optional)</label>
            <div class="upload-box" id="uploadBox">
              <input class="input" id="image-upload" name="image_path"
                     type="file" accept="image/jpeg,image/png,image/gif" hidden />
              <div class="upload-placeholder">
                <img src="/static/images/icons/upload-icon.png"
                     alt="Upload Icon" class="upload-icon" />
                <p class="upload-text">Click to upload image</p>
                <p class="upload-subtext">JPEG, PNG, or GIF — Max 20MB</p>
              </div>
            </div>
            <div class="field-error" id="error-image"></div>
            <div id="file-name" class="file-name"></div>
          </div>

          <!-- Actions -->
          <div class="actions">
            <button type="reset" class="btn btn-reset">Reset Form</button>
            <button type="submit" class="btn btn-submit">Create Topic</button>
          </div>
        </div>
      </form>
    </div>
    <div class="home-link-container">
      <a href="/" class="home-link" data-link>Go to Homepage</a>
    </div>
  `;
}

// ─── Multi-select dropdown (mirrors create-post.js IIFE) ─────────────────────

function initMultiSelect() {
  const ms = document.getElementById('categorySelect');
  const multiOptions = document.getElementById('multiOptions');
  const chips = document.getElementById('chips');
  const placeholder = ms?.querySelector('.placeholder');

  if (!ms) return;

  const open = () => {
    ms.classList.add('open');
    ms.setAttribute('aria-expanded', 'true');
    multiOptions.focus();
  };
  const close = () => {
    ms.classList.remove('open');
    ms.setAttribute('aria-expanded', 'false');
  };

  ms.addEventListener('click', (e) => {
    if (e.target.closest('.options')) return;
    ms.classList.contains('open') ? close() : open();
  });

  document.addEventListener('click', (e) => {
    if (!ms.contains(e.target)) close();
  });

  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') close();
  });

  ms.addEventListener('keydown', (e) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      ms.classList.contains('open') ? close() : open();
    }
  });

  function rebuildChips() {
    chips.innerHTML = '';
    const checked = ms.querySelectorAll('input[type="checkbox"]:checked');

    if (checked.length === 0) {
      placeholder.style.display = 'inline';
    } else {
      placeholder.style.display = 'none';
    }

    checked.forEach((cb) => {
      const label = cb.parentElement.querySelector('.option-label')?.textContent || cb.value;
      const chip = document.createElement('span');
      chip.className = 'chip';
      chip.textContent = label;

      const removeBtn = document.createElement('button');
      removeBtn.type = 'button';
      removeBtn.setAttribute('aria-label', `Remove ${label}`);
      removeBtn.style.cssText = 'background:transparent;border:none;margin-left:8px;cursor:pointer';
      removeBtn.textContent = '✕';
      removeBtn.addEventListener('click', (ev) => {
        ev.stopPropagation();
        cb.checked = false;
        cb.dispatchEvent(new Event('change', { bubbles: true }));
      });

      chip.appendChild(removeBtn);
      chips.appendChild(chip);
    });

    // Clear category error when user selects something
    if (checked.length > 0) {
      document.getElementById('error-categories').textContent = '';
    }
  }

  ms.querySelectorAll('input[type="checkbox"]').forEach((cb) => {
    cb.addEventListener('change', rebuildChips);
  });

  rebuildChips();
}

// ─── File upload (mirrors create-post.js IIFE) ───────────────────────────────

function initFileUpload() {
  const fileInput = document.getElementById('image-upload');
  const uploadBox = document.getElementById('uploadBox');
  const fileNameDisplay = document.getElementById('file-name');
  const errorEl = document.getElementById('error-image');

  if (!fileInput || !uploadBox) return;

  // Clicking the upload box triggers the hidden file input
  uploadBox.addEventListener('click', () => fileInput.click());

  fileInput.addEventListener('change', () => {
    errorEl.textContent = '';
    fileNameDisplay.textContent = '';

    const file = fileInput.files[0];
    if (!file) return;

    const allowedTypes = ['image/jpeg', 'image/png', 'image/gif'];
    const maxSizeMB = 20;

    if (!allowedTypes.includes(file.type)) {
      errorEl.textContent = 'Only JPEG, PNG, or GIF images are allowed.';
      fileInput.value = '';
      return;
    }

    if (file.size > maxSizeMB * 1024 * 1024) {
      errorEl.textContent = 'Image is too large. Maximum size is 20 MB.';
      fileInput.value = '';
      return;
    }

    fileNameDisplay.textContent = `Selected: ${file.name}`;
    fileNameDisplay.style.color = '#068f56';
  });
}

// ─── Form submit ──────────────────────────────────────────────────────────────

function initFormSubmit() {
  const form = document.getElementById('createPostForm');
  if (!form) return;

  // Clear individual field errors on input
  document.getElementById('title')?.addEventListener('input', () => {
    document.getElementById('error-title').textContent = '';
  });
  document.getElementById('content')?.addEventListener('input', () => {
    document.getElementById('error-content').textContent = '';
  });

  // Reset button clears everything including chips and errors
  form.addEventListener('reset', () => {
    document.getElementById('error-categories').textContent = '';
    document.getElementById('error-title').textContent = '';
    document.getElementById('error-content').textContent = '';
    document.getElementById('error-image').textContent = '';
    document.getElementById('file-name').textContent = '';

    const fileInput = document.getElementById('image-upload');
    if (fileInput) fileInput.value = '';

    // Uncheck all category checkboxes and rebuild chips
    form.querySelectorAll('input[name="categories"]').forEach((cb) => {
      cb.checked = false;
    });
    // Trigger chip rebuild via change event on first checkbox
    const first = form.querySelector('input[name="categories"]');
    first?.dispatchEvent(new Event('change', { bubbles: true }));
  });

  form.addEventListener('submit', async (e) => {
    e.preventDefault();

    // Clear all errors
    ['error-categories', 'error-title', 'error-content', 'error-image'].forEach((id) => {
      document.getElementById(id).textContent = '';
    });

    if (!validateForm(form)) return;

    const submitBtn = form.querySelector('.btn-submit');
    submitBtn.disabled = true;
    submitBtn.textContent = 'Creating...';

    const title = document.getElementById('title').value.trim();
    const content = document.getElementById('content').value.trim();
    const categories = [...form.querySelectorAll('input[name="categories"]:checked')].map((cb) =>
      Number(cb.value)
    );
    const fileInput = document.getElementById('image-upload');
    const file = fileInput?.files?.[0] ?? null;

    try {
      // Image selected — send multipart so the backend can receive the file.
      // Requires the backend createTopic handler to accept multipart/form-data.
      const formData = new FormData();
      formData.append('title', title);
      formData.append('content', content);
      categories.forEach((id) => formData.append('categories', String(id)));
      if (file) {
        formData.append('image_path', file);
      }

      const response = await fetch('/api/v1/topics/create', {
        method: 'POST',
        credentials: 'include',
        body: formData,
        // Do NOT set Content-Type — browser sets the multipart boundary automatically
      });

      if (!response.ok) {
        const body = await response.json().catch(() => ({}));
        throw new Error(body?.error || body?.message || `HTTP ${response.status}`);
      }
      // } else {
      //   // No image — send JSON directly, matching createTopicRequest struct:
      //   // { title, content, imagePath, categoryIds }
      //   await api.post('/topics/create', {
      //     title,
      //     content,
      //     imagePath: '',
      //     categoryIds: categories,
      //   });
      // }

      navigate('/topics');
    } catch (err) {
      console.error('Failed to create topic:', err);
      document.getElementById('error-content').textContent =
        'Failed to create topic: ' + (err.message || 'Unknown error');
    } finally {
      submitBtn.disabled = false;
      submitBtn.textContent = 'Create Topic';
    }
  });
}

// ─── Client-side validation (mirrors create-post.js submit listener) ─────────

function validateForm(form) {
  let ok = true;

  const selected = form.querySelectorAll('input[name="categories"]:checked');
  if (selected.length === 0) {
    document.getElementById('error-categories').textContent =
      'Please select at least one category.';
    ok = false;
  }

  const title = document.getElementById('title')?.value.trim() ?? '';
  if (!title) {
    document.getElementById('error-title').textContent = 'Title is required.';
    ok = false;
  } else if (title.length < 3) {
    document.getElementById('error-title').textContent = 'Title must be at least 3 characters.';
    ok = false;
  } else if (title.length > 200) {
    document.getElementById('error-title').textContent = 'Title must not exceed 200 characters.';
    ok = false;
  }

  const content = document.getElementById('content')?.value.trim() ?? '';
  if (!content) {
    document.getElementById('error-content').textContent = 'Content is required.';
    ok = false;
  } else if (content.length < 10) {
    document.getElementById('error-content').textContent =
      'Content must be at least 10 characters.';
    ok = false;
  } else if (content.length > 5000) {
    document.getElementById('error-content').textContent =
      'Content must not exceed 5000 characters.';
    ok = false;
  }

  return ok;
}

// ─── Skeleton / error states ──────────────────────────────────────────────────

function buildSkeletonHTML() {
  return /* html */ `
    <div class="create-post-container">
      <div class="form">
        <div class="form-wrapper">
          <div class="skeleton" style="width:140px;height:28px;margin-bottom:1rem;border-radius:4px"></div>
          <div class="skeleton" style="width:100%;height:40px;margin-bottom:1rem;border-radius:8px"></div>
          <div class="skeleton" style="width:100%;height:40px;margin-bottom:1rem;border-radius:8px"></div>
          <div class="skeleton" style="width:100%;height:120px;margin-bottom:1rem;border-radius:8px"></div>
          <div class="skeleton" style="width:100%;height:80px;border-radius:8px"></div>
        </div>
      </div>
    </div>`;
}

function buildErrorHTML(message) {
  return /* html */ `
    <div class="main-container" style="text-align:center;padding:4rem 1rem">
      <h2 style="color:#e53e3e">Something went wrong</h2>
      <p style="color:var(--grey-color);margin-top:1rem">${escapeHTML(message)}</p>
      <a href="/" data-link style="
        display:inline-block;margin-top:2rem;padding:0.8rem 1.5rem;
        background:var(--primary-color);color:#fff;border-radius:8px;text-decoration:none">
        Go Home
      </a>
    </div>`;
}
