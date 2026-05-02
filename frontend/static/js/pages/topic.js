/**
 * pages/topic.js
 *
 * Renders the /topic/:id page into #app-root.
 *
 * Replaces:
 *   - Go BFF TopicPage handler  (cmd/client/server/topic_handler.go)
 *   - frontend/html/pages/topic.html
 *   - frontend/static/js/topic.js      (edit/delete/comment form toggling)
 *   - frontend/static/js/topicVotes.js (vote casting & UI update)
 *
 * What the Go handler did → what we do here:
 *   1. Parse topic ID from URL path              → parseTopicIDFromURL()
 *   2. Fetch /topic?id=X (with cookies)          → api.get (credentials:"include")
 *   3. Fetch /categories/all for edit form       → fetchCategories()
 *   4. Normalize CategoryColors                  → normalizeColor() per color
 *   5. Inject .User from context                 → user param from router
 *   6. Execute template + hasID helper           → hasID() used in edit form
 *   7. topic.js     — form toggling, validation  → initTopicInteractions()
 *   8. topicVotes.js — vote cast/delete/UI       → initVoteButtons()
 */

import { api, fetchCategories, castVote, deleteVote, fetchVoteCounts } from '../api.js';
import { normalizeColor, escapeHTML, formatRelativeDate } from '../helpers.js';
import { prepareCategories } from '../helpers.js';

// ─── Public API ───────────────────────────────────────────────────────────────

export async function renderTopicPage(user) {
  const root = document.getElementById('app-root');
  if (!root) return;

  const topicID = parseTopicIDFromURL();
  if (!topicID) {
    root.innerHTML = buildNotFoundHTML('Invalid topic URL.');
    return;
  }

  root.innerHTML = buildSkeletonHTML();

  try {
    // Fetch topic + categories in parallel, mirrors the two HTTP calls in the Go handler.
    // Topic fetch must send cookies so the backend can resolve UserVote.
    const [topicData, categoriesData] = await Promise.all([
      api.get(`/topic`, { id: topicID }),
      fetchCategories({ order_by: 'name', order: 'asc', page: 1, page_size: 100 }),
    ]);

    if (!topicData) {
      root.innerHTML = buildNotFoundHTML('Topic not found.');
      return;
    }

    // Normalise colors — mirrors the Go loop:
    // for i, color := range topicData.CategoryColors { NormalizeColor(color) }
    const normalizedColors = (topicData.categoryColors ?? []).map(normalizeColor);

    const topic = {
      ...topicData,
      categoryColors: normalizedColors,
    };

    const rawCats = Array.isArray(categoriesData)
      ? categoriesData
      : (categoriesData?.categories ?? categoriesData?.Categories ?? []);

    const categories = prepareCategories(rawCats);

    root.innerHTML = buildPageHTML(topic, categories, user);

    // Wire up all interactive behaviour after DOM is ready
    initTopicInteractions(topic, user);
    initVoteButtons(topic, user);
  } catch (err) {
    console.error('Failed to load topic:', err);
    if (err.status === 404) {
      root.innerHTML = buildNotFoundHTML('Topic not found.');
    } else {
      root.innerHTML = buildErrorHTML(err.message || 'Failed to load topic.');
    }
  }
}

// ─── Initialize Global Listeners ──────────────────────────────────────────────────────────────

let listenersInitialized = false;

export function initGlobalListeners() {
  if (listenersInitialized) return;
  listenersInitialized = true;

  document.addEventListener('click', async (e) => {
    try {
      // ─── DELETE TOPIC ───
      const topicDeleteBtn = e.target.closest(".btn-delete[data-type='topic']");
      if (topicDeleteBtn) {
        if (!confirm('Delete topic?')) return;

        const topicID = topicDeleteBtn.dataset.topicId;
        await api.delete(`/topics/delete?id=${topicID}`);
        const { navigate } = await import('../router.js');
        navigate('/topics');
        return;
      }

      // ─── DELETE COMMENT ───
      const commentDeleteBtn = e.target.closest(".btn-delete[data-type='comment']");
      if (commentDeleteBtn) {
        if (!confirm('Delete comment?')) return;

        const commentID = commentDeleteBtn.dataset.commentId;
        await api.delete(`/comments/delete?id=${commentID}`);

        document.querySelector(`[data-comment-id="${commentID}"]`)?.remove();
        updateCommentCount(-1);
        return;
      }

      // ─── EDIT TOGGLE ───
      const editBtn = e.target.closest('.btn-edit');
      if (editBtn) {
        closeAllEditForms();
        clearAllErrors();
        document.querySelector('.add-comment')?.classList.remove('active');

        // Reset file inputs and labels when opening
        const topicForm = document.querySelector('.topic-edit-form');
        if (topicForm) {
          topicForm.querySelector('#topic-file-name').textContent = '';
          topicForm.querySelector('#topicImageUpload').value = '';
        }

        const type = editBtn.dataset.type;

        if (type === 'topic') {
          document.querySelector('.edit-topic-form')?.style.setProperty('display', 'block');
        }

        if (type === 'comment') {
          const id = editBtn.dataset.commentId;
          document
            .querySelector(`[data-comment-id="${id}"] .edit-comment-form`)
            ?.style.setProperty('display', 'block');
        }
        return;
      }

      // ─── CLOSE EDIT FORM ───
      if (e.target.classList.contains('close-edit-form')) {
        e.target.closest('.edit-form')?.style.setProperty('display', 'none');
        clearAllErrors();
        return;
      }
    } catch (err) {
      console.error('Global handler error:', err);
      alert('Something went wrong: ' + err.message);
    }
  });
}

initGlobalListeners();

// ─── URL parsing ──────────────────────────────────────────────────────────────

/**
 * Extracts the numeric topic ID from the URL path.
 * /topic/42  →  "42"
 * Mirrors: pathParts := strings.Split(...); topicID, err := strconv.Atoi(pathParts[1])
 */
function parseTopicIDFromURL() {
  const parts = window.location.pathname.replace(/^\/+/, '').split('/');
  // parts[0] = "topic", parts[1] = id
  const id = parts[1];
  if (!id || isNaN(Number(id)) || Number(id) <= 0) return null;
  return id;
}

// ─── hasID helper (mirrors Go hasID template func) ───────────────────────────

function hasID(ids, id) {
  if (!Array.isArray(ids)) return false;
  return ids.some((v) => Number(v) === Number(id));
}

// ─── Main page HTML ───────────────────────────────────────────────────────────

function buildPageHTML(topic, categories, user) {
  const isOwner = user && String(user.id ?? user.ID ?? '') === String(topic.userId ?? '');

  return /* html */ `
    <div class="main-container">
      <div class="topic-container">

        ${buildTopicHeaderHTML(topic)}
        ${buildTopicContentHTML(topic, user)}
        ${isOwner ? buildPostActionsHTML(topic) : ''}
        ${isOwner ? buildEditTopicFormHTML(topic, categories) : ''}
        ${user ? buildAddCommentSectionHTML(topic) : ''}
        ${buildCommentsSectionHTML(topic, user)}

      </div>
    </div>
  `;
}

// ── Topic header (category badges) ───────────────────────────────────────────

function buildTopicHeaderHTML(topic) {
  const colors = topic.categoryColors ?? [];
  const names = topic.categoryNames ?? [];
  const ids = topic.categoryIds ?? [];

  let badgesHTML;

  if (colors.length > 0) {
    badgesHTML = colors
      .map(
        (color, i) => /* html */ `
      <a href="/topics?category=${escapeHTML(String(ids[i] ?? ''))}" class="topic-category-link" data-link>
        <div class="topic-category">
          <span class="topic-category-color" style="background-color: ${normalizeColor(color)}"></span>
          <span class="topic-category-name">${escapeHTML(names[i] ?? '')}</span>
        </div>
      </a>`
      )
      .join('');
  } else {
    const color = normalizeColor(topic.categoryColor ?? '');
    const name = escapeHTML(topic.categoryName ?? '');
    const catID = topic.categoryId ?? '';
    badgesHTML = /* html */ `
      <a href="/topics?category=${escapeHTML(String(catID))}" class="topic-category-link" data-link>
        <div class="topic-category">
          <span class="topic-category-color" style="background-color: ${color}"></span>
          <span class="topic-category-name">${name}</span>
        </div>
      </a>`;
  }

  return /* html */ `
    <div class="topic-header">
      <div class="topic-categories">${badgesHTML}</div>
    </div>
  `;
}

// ── Topic body ────────────────────────────────────────────────────────────────

function buildTopicContentHTML(topic, user) {
  const title = escapeHTML(topic.title ?? '');
  const author = escapeHTML(topic.ownerUsername ?? '');
  const createdAt = topic.createdAt ?? '';
  const content = escapeHTML(topic.content ?? '');
  const imagePath = topic.imagePath ?? '';
  const upvotes = topic.upvotes ?? 0;
  const downvotes = topic.downvotes ?? 0;
  const score = topic.score ?? 0;
  const commentsLen = (topic.comments ?? []).length;
  const topicID = topic.topicId ?? '';

  const userVote = topic.userVote ?? null;
  const userVoteVal = userVote !== null ? String(userVote) : '';

  const likeDisabled = user ? '' : 'disabled';
  const likeActive = userVote === 1 ? 'active' : '';
  const dislikeActive = userVote === -1 ? 'active' : '';

  const imageHTML = imagePath
    ? /* html */ `
    <div class="img-box">
      <img src="${escapeHTML(imagePath)}" alt="Post Image" class="post-image" />
    </div>`
    : '';

  return /* html */ `
    <div class="topic-content">
      <p class="post-title">${title}</p>

      <div class="topic-head">
        <div class="post-meta">
          <div class="topic-user-box">
            <img src="/static/images/user-avatar.png" alt="Author Avatar" class="author-avatar" />
          </div>
          <span class="post-author-name">${author}</span>
        </div>
        <span class="post-date">${createdAt}</span>
      </div>

      <div class="topic-body-container" data-user-vote="${userVoteVal}" data-topic-id="${topicID}">
        <div class="topic-body">
          <p class="post-text">${content}</p>
          ${imageHTML}
        </div>

        <div class="reactions">
          <div class="reaction-box">
            <button class="btn like-btn ${likeActive}" ${likeDisabled}>
              <img class="like-icon" src="/static/images/icons/icon-like.png" alt="Like the post" />
            </button>
            <span class="like-count">${upvotes}</span>
          </div>

          <div class="reaction-box">
            <button class="btn dislike-btn ${dislikeActive}" ${likeDisabled}>
              <img class="dislike-icon" src="/static/images/icons/icon-dislike.png" alt="Dislike the post" />
            </button>
            <span class="dislike-count">${downvotes}</span>
          </div>

          <div class="topic-extra-info">
            <div class="views-box">
              <span class="topic-views">Vote Score</span>
              <span class="views-count">${score}</span>
            </div>
            <div class="comments-box">
              <span class="topic-comments">Comments</span>
              <span class="comments-count">${commentsLen}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  `;
}

// ── Post actions (edit / delete — owner only) ─────────────────────────────────

function buildPostActionsHTML(topic) {
  const topicID = escapeHTML(String(topic.topicId ?? ''));
  return /* html */ `
    <div class="post-actions">
      <div class="author-choices">
        <button class="action-btn btn-edit" data-type="topic">Edit Topic</button>
        <button class="action-btn btn-delete" data-type="topic" data-topic-id="${topicID}">
          Delete Topic
        </button>
      </div>
    </div>
  `;
}

// ── Edit topic form ───────────────────────────────────────────────────────────

function buildEditTopicFormHTML(topic, categories) {
  const topicID = escapeHTML(String(topic.topicId ?? ''));
  const title = escapeHTML(topic.title ?? '');
  const content = escapeHTML(topic.content ?? '');
  const imagePath = escapeHTML(topic.imagePath ?? '');
  const categoryIDs = topic.categoryIds ?? [];

  const categoryCheckboxes = categories
    .map((cat) => {
      const catID = cat.id ?? 0;
      const catName = escapeHTML(cat.name ?? '');
      const color = cat.Color ?? cat.color ?? '#ccc';
      const checked = hasID(categoryIDs, catID) ? 'checked' : '';
      return /* html */ `
      <div class="category-checkbox">
        <input type="checkbox" name="categories" value="${catID}"
               id="edit-category-${catID}" ${checked} />
        <label for="edit-category-${catID}" class="category-label">
          <span class="category-color" style="background-color: ${color}"></span>
          ${catName}
        </label>
      </div>`;
    })
    .join('');

  return /* html */ `
    <div class="edit-form edit-topic-form" style="display:none">
      <div class="comment-form-header">
        <h3>Edit Topic</h3>
        <button type="button" class="close-edit-form">✖</button>
      </div>
      <form class="topic-edit-form" data-topic-id="${topicID}">
        <input type="hidden" name="topic_id" value="${topicID}" />
        <input type="hidden" name="current_image_path" value="${imagePath}" />

        <div class="comment-form-field">
          <label>Categories:</label>
          <div class="categories-checkboxes">${categoryCheckboxes}</div>
          <div class="field-error" id="error-topic-categories"></div>
        </div>

        <div class="comment-form-field">
          <input class="input topic-title-input" name="title" type="text"
                 placeholder="Topic title..." value="${title}" required />
          <div class="field-error" id="error-topic-title"></div>
        </div>

        <div class="comment-form-field">
          <textarea class="input comment-textarea" name="content" rows="5"
                    placeholder="Edit your topic content..." required>${content}</textarea>
          <div class="field-error" id="error-topic-content"></div>
        </div>

        <div class="comment-form-field">
          <label class="label">Attach new image (optional)</label>

          <div class="upload-box" id="topicUploadBox">
            <input class="input" id="topicImageUpload"
                  name="image_path"
                  type="file"
                  accept="image/jpeg,image/png,image/gif"
                  hidden />

            <div class="upload-placeholder">
              <img src="/static/images/icons/upload-icon.png" class="upload-icon" />
              <p class="upload-text">Click to upload image</p>
              <p class="upload-subtext">JPEG, PNG, or GIF — Max 20MB</p>
            </div>
          </div>

          <div class="field-error" id="error-topic-image"></div>
          <div id="topic-file-name" class="file-name"></div>
        </div>

        <button class="post-comment" type="submit">Save Changes</button>
      </form>
    </div>
  `;
}

// ── Add comment section ───────────────────────────────────────────────────────

function buildAddCommentSectionHTML(topic) {
  const topicID = escapeHTML(String(topic.topicId ?? ''));
  return /* html */ `
    <div class="post-actions">
      <button class="action-btn btn-comment">Add a Comment</button>
    </div>
    <div class="add-comment">
      <div class="comment-form-header">
        <h3>Create Comment</h3>
        <button type="button" class="close-comment-form">✖</button>
      </div>
      <form class="comment-form">
        <input type="hidden" name="topic_id" value="${topicID}" />
        <div class="comment-form-field">
          <textarea class="input comment-textarea" name="content" rows="5"
                    placeholder="Write your comment..." required></textarea>
          <div class="field-error" id="error-comment-content"></div>
        </div>
        <button class="post-comment" type="submit">Post Comment</button>
      </form>
    </div>
  `;
}

// ── Comments section ──────────────────────────────────────────────────────────

function buildCommentsSectionHTML(topic, user) {
  const comments = topic.comments ?? [];
  if (!comments.length) return '';

  const topicID = escapeHTML(String(topic.topicId ?? ''));

  const commentItems = comments
    .map((comment) => buildCommentItemHTML(comment, topicID, user))
    .join('');

  return /* html */ `
    <div class="comments-section">
      ${commentItems}
    </div>
  `;
}

function buildCommentItemHTML(comment, topicID, user) {
  // Comments in the response use PascalCase (no JSON tags on the backend Comment struct)
  const commentID = escapeHTML(String(comment.ID ?? ''));
  const author = escapeHTML(comment.OwnerUsername ?? '');
  const content = escapeHTML(comment.Content ?? '');
  const createdAt = comment.CreatedAt ?? '';
  const upvotes = comment.UpvoteCount ?? 0;
  const downvotes = comment.DownvoteCount ?? 0;
  const userVote = comment.UserVote ?? null;
  const commentUserID = String(comment.UserID ?? '');

  const isOwner = user && String(user.id ?? user.ID ?? '') === commentUserID;
  const likeDisabled = user ? '' : 'disabled';
  const likeActive = userVote === 1 ? 'active' : '';
  const dislikeActive = userVote === -1 ? 'active' : '';

  const actionsHTML = isOwner
    ? /* html */ `
    <div class="comment-actions">
      <button class="action-btn btn-edit" data-type="comment" data-comment-id="${commentID}">Edit</button>
      <button class="action-btn btn-delete" data-type="comment"
              data-comment-id="${commentID}" data-topic-id="${topicID}">Delete</button>
    </div>`
    : '';

  const editFormHTML = isOwner
    ? /* html */ `
    <div class="edit-form edit-comment-form" style="display:none">
      <div class="comment-form-header">
        <h3>Edit Comment</h3>
        <button type="button" class="close-edit-form">✖</button>
      </div>
      <form class="comment-edit-form" data-comment-id="${commentID}">
        <input type="hidden" name="topic_id"   value="${topicID}" />
        <input type="hidden" name="comment_id" value="${commentID}" />
        <div class="comment-form-field">
          <textarea class="input comment-textarea" name="content" rows="5"
                    placeholder="Edit your comment..." required>${content}</textarea>
          <div class="field-error" id="error-edit-comment-${commentID}"></div>
        </div>
        <button class="post-comment" type="submit">Save Changes</button>
      </form>
    </div>`
    : '';

  return /* html */ `
    <div class="comment-content" data-comment-id="${commentID}"
         data-user-vote="${userVote !== null ? userVote : ''}">
      <div class="comment-head">
        <div class="comment-meta">
          <div class="comment-img-box">
            <img src="/static/images/user-avatar.png" alt="User Avatar" class="comment-avatar" />
          </div>
          <span class="comment-author">${author}</span>
        </div>
        <span class="comment-date">${createdAt}</span>
      </div>

      <div class="comment-body-container">
        <div class="comment-body">
          <p class="comment-text">${content}</p>

          <div class="reactions">
            <div class="reaction-box">
              <button class="btn like-btn ${likeActive}" ${likeDisabled}>
                <img class="like-icon" src="/static/images/icons/icon-like.png" alt="Like" />
              </button>
              <span class="like-count">${upvotes}</span>
            </div>
            <div class="reaction-box">
              <button class="btn dislike-btn ${dislikeActive}" ${likeDisabled}>
                <img class="dislike-icon" src="/static/images/icons/icon-dislike.png" alt="Dislike" />
              </button>
              <span class="dislike-count">${downvotes}</span>
            </div>
          </div>

          ${actionsHTML}
        </div>
      </div>

      ${editFormHTML}
    </div>
  `;
}

// ─── Interactive behaviour ────────────────────────────────────────────────────
// Mirrors frontend/static/js/topic.js

function initTopicInteractions(topic, user) {
  initCommentFormToggle();
  initTopicEditFormSubmit(topic);
  initCommentFormSubmit(topic, user);
  initEditCommentForms(topic);
}

// ── Comment form open/close ───────────────────────────────────────────────────

function initCommentFormToggle() {
  const addBtn = document.querySelector('.btn-comment');
  const form = document.querySelector('.add-comment');
  const closeBtn = document.querySelector('.close-comment-form');

  addBtn?.addEventListener('click', () => {
    closeAllEditForms();
    form?.classList.toggle('active');
  });

  closeBtn?.addEventListener('click', () => {
    form?.classList.remove('active');
    clearAllErrors();
  });
}

function closeAllEditForms() {
  document.querySelectorAll('.edit-form').forEach((f) => f.style.setProperty('display', 'none'));
}

// ── Topic edit form submit ────────────────────────────────────────────────────

function initTopicEditFormSubmit(topic) {
  const form = document.querySelector('.topic-edit-form');
  if (!form) return;

  const uploadBox = form.querySelector('#topicUploadBox');
  const fileInput = form.querySelector('#topicImageUpload');
  const fileNameEl = form.querySelector('#topic-file-name');

  uploadBox?.addEventListener('click', () => fileInput.click());

  fileInput?.addEventListener('change', () => {
    if (fileInput.files && fileInput.files[0]) {
      fileNameEl.textContent = `Selected: ${fileInput.files[0].name}`;
      fileNameEl.style.color = '#2ecc71'; // Give it a success color
    } else {
      fileNameEl.textContent = '';
    }
    clearError('error-topic-image');
  });

  // Clear errors on input
  form
    .querySelector('input[name="title"]')
    ?.addEventListener('input', () => clearError('error-topic-title'));
  form
    .querySelector('textarea[name="content"]')
    ?.addEventListener('input', () => clearError('error-topic-content'));
  form
    .querySelectorAll('input[name="categories"]')
    .forEach((cb) => cb.addEventListener('change', () => clearError('error-topic-categories')));

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    clearAllErrors();

    if (!validateTopicForm(form)) return;

    const formData = new FormData(form);
    const categories = [...formData.getAll('categories')].map(Number);
    const topicId = Number(formData.get('topic_id'));
    const title = formData.get('title')?.trim();
    const content = formData.get('content')?.trim();
    const currentPath = formData.get('current_image_path') || '';
    const file = fileInput?.files?.[0] ?? null;

    try {
      if (file) {
        // New image selected — send multipart
        const fd = new FormData();
        fd.append('topic_id', String(topicId));
        fd.append('title', title);
        fd.append('content', content);
        fd.append('current_image_path', currentPath);
        categories.forEach((id) => fd.append('categories', String(id)));
        fd.append('image_path', file);

        const response = await fetch('/api/v1/topics/update', {
          method: 'PUT',
          credentials: 'include',
          body: fd,
        });

        if (!response.ok) {
          const body = await response.json().catch(() => ({}));
          throw new Error(body?.error || body?.message || `HTTP ${response.status}`);
        }
      } else {
        // No new image — send JSON with updateTopicRequest shape
        // imagePath: keep currentPath (preserves existing image)
        // to remove an image entirely, set imagePath: ""
        await api.put('/topics/update', {
          topicId,
          title,
          content,
          categoryIds: categories,
          imagePath: currentPath,
        });
      }

      const { getUser } = await import('../auth.js');
      await renderTopicPage(getUser());
    } catch (err) {
      alert('Failed to update topic: ' + err.message);
    }
  });
}

// ── Create comment form submit ────────────────────────────────────────────────

function initCommentFormSubmit(topic, user) {
  const form = document.querySelector('.comment-form');
  if (!form) return;

  form
    .querySelector('textarea[name="content"]')
    ?.addEventListener('input', () => clearError('error-comment-content'));

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    clearAllErrors();

    const content = form.querySelector('textarea[name="content"]')?.value.trim();
    if (!content) {
      showError('error-comment-content', 'Comment is required');
      return;
    }
    if (content.length < 3) {
      showError('error-comment-content', 'Comment must be at least 3 characters');
      return;
    }

    const topicID = form.querySelector('input[name="topic_id"]')?.value;

    try {
      // BFF CreateCommentPost sent POST with { topicId, content }
      await api.post('/comments/create', {
        topicId: Number(topicID),
        content,
      });
      // Re-render the page to show the new comment
      const { getUser } = await import('../auth.js');
      await renderTopicPage(getUser());
    } catch (err) {
      showError('error-comment-content', 'Failed to post comment: ' + err.message);
    }
  });
}

// ── Edit comment forms submit ─────────────────────────────────────────────────

function initEditCommentForms(topic) {
  document.querySelectorAll('.comment-edit-form').forEach((form) => {
    const commentID = form.dataset.commentId;
    const errorID = `error-edit-comment-${commentID}`;

    form
      .querySelector('textarea[name="content"]')
      ?.addEventListener('input', () => clearError(errorID));

    form.addEventListener('submit', async (e) => {
      e.preventDefault();
      clearError(errorID);

      const content = form.querySelector('textarea[name="content"]')?.value.trim();
      if (!content) {
        showError(errorID, 'Comment is required');
        return;
      }
      if (content.length < 3) {
        showError(errorID, 'Comment must be at least 3 characters');
        return;
      }

      const topicID = form.querySelector('input[name="topic_id"]')?.value;

      try {
        // BFF UpdateCommentPost sent PUT with { id, content } — no topic_id in body
        await api.put('/comments/update', {
          id: Number(commentID),
          content,
        });
        // Re-render to show updated comment
        const { getUser } = await import('../auth.js');
        await renderTopicPage(getUser());
      } catch (err) {
        showError(errorID, 'Failed to update comment: ' + err.message);
      }
    });
  });
}

// ─── Vote system ──────────────────────────────────────────────────────────────
// Mirrors frontend/static/js/topicVotes.js
// Key change: fetch URLs are now /api/v1/* via api.js, not the old BFF /api/* paths.

function initVoteButtons(topic, user) {
  if (!user) return; // buttons are already disabled in the HTML for guests

  document.querySelectorAll('.like-btn, .dislike-btn').forEach((btn) => {
    btn.addEventListener('click', async function () {
      if (this.disabled) return;

      const isLike = this.classList.contains('like-btn');
      const commentEl = this.closest('.comment-content');
      const isComment = commentEl !== null;
      const reactionType = isLike ? 1 : -1;

      let targetID, targetType;
      if (isComment) {
        targetID = parseInt(commentEl.dataset.commentId);
        targetType = 'comment';
      } else {
        targetID = parseInt(
          this.closest('.topic-body-container')?.dataset.topicId || topic.topicId
        );
        targetType = 'topic';
      }

      const container = this.closest('.reactions');
      setVoteButtonsDisabled(container, true);

      try {
        const isActive = this.classList.contains('active');

        if (isActive) {
          // Toggle off — delete the vote
          const payload =
            targetType === 'comment' ? { commentId: targetID } : { topicId: targetID };
          await deleteVote(payload);
        } else {
          // Cast or switch vote
          const payload = { reactionType: reactionType };
          if (targetType === 'comment') payload.commentId = targetID;
          else payload.topicId = targetID;
          await castVote(payload);
        }

        await refreshVoteUI(targetID, targetType, this);
      } catch (err) {
        console.error('Vote error:', err);
        if (err.status === 401) {
          const { navigate } = await import('../router.js');
          navigate('/login');
        }
      } finally {
        setVoteButtonsDisabled(container, false);
      }
    });
  });
}

function setVoteButtonsDisabled(container, disabled) {
  container?.querySelectorAll('.like-btn, .dislike-btn').forEach((b) => {
    b.disabled = disabled;
  });
  container?.classList.toggle('loading', disabled);
}

async function refreshVoteUI(targetID, targetType, clickedBtn) {
  const params = targetType === 'comment' ? { comment_id: targetID } : { topic_id: targetID };

  const counts = await fetchVoteCounts(params);
  const container = clickedBtn.closest('.reactions');

  const likeCount = container.querySelector('.like-count');
  const dislikeCount = container.querySelector('.dislike-count');
  const likeBtn = container.querySelector('.like-btn');
  const dislikeBtn = container.querySelector('.dislike-btn');

  likeCount.classList.add('updating');
  dislikeCount.classList.add('updating');

  likeCount.textContent = counts.upvotes ?? counts.Upvotes ?? 0;
  dislikeCount.textContent = counts.downvotes ?? counts.Downvotes ?? 0;

  // Update vote score for topics
  const scoreEl = container.closest('.topic-body-container')?.querySelector('.views-count');
  if (scoreEl) scoreEl.textContent = counts.score ?? counts.Score ?? 0;

  // Toggle active state — if we just deleted a vote nothing is active
  const wasActive = clickedBtn.classList.contains('active');
  likeBtn.classList.remove('active');
  dislikeBtn.classList.remove('active');
  if (!wasActive) clickedBtn.classList.add('active');

  setTimeout(() => {
    likeCount.classList.remove('updating');
    dislikeCount.classList.remove('updating');
  }, 300);
}

// ─── Form validation helpers ──────────────────────────────────────────────────

function validateTopicForm(form) {
  let ok = true;
  const categories = form.querySelectorAll('input[name="categories"]:checked');
  const title = form.querySelector('input[name="title"]')?.value.trim();
  const content = form.querySelector('textarea[name="content"]')?.value.trim();

  if (!categories.length) {
    showError('error-topic-categories', 'At least one category is required');
    ok = false;
  }
  if (!title) {
    showError('error-topic-title', 'Title is required');
    ok = false;
  } else if (title.length < 3) {
    showError('error-topic-title', 'Title must be at least 3 characters');
    ok = false;
  } else if (title.length > 200) {
    showError('error-topic-title', 'Title must not exceed 200 characters');
    ok = false;
  }
  if (!content) {
    showError('error-topic-content', 'Content is required');
    ok = false;
  } else if (content.length < 10) {
    showError('error-topic-content', 'Content must be at least 10 characters');
    ok = false;
  }
  return ok;
}

function updateCommentCount(delta) {
  const el = document.querySelector('.comments-count');
  if (!el) return;
  el.textContent = Math.max(0, parseInt(el.textContent || '0') + delta);
}

function showError(id, msg) {
  const el = document.getElementById(id);
  if (el) el.textContent = msg;
}

function clearError(id) {
  const el = document.getElementById(id);
  if (el) el.textContent = '';
}

function clearAllErrors() {
  document.querySelectorAll('.field-error').forEach((el) => (el.textContent = ''));
}

// ─── Skeleton / error states ──────────────────────────────────────────────────

function buildSkeletonHTML() {
  return /* html */ `
    <div class="main-container">
      <div class="topic-container">
        <div class="topic-header">
          <div class="skeleton" style="width:120px;height:24px;border-radius:12px"></div>
        </div>
        <div class="topic-content">
          <div class="skeleton" style="width:70%;height:28px;margin-bottom:1rem"></div>
          <div class="skeleton" style="width:40%;height:16px;margin-bottom:2rem"></div>
          <div class="skeleton" style="width:100%;height:120px;border-radius:8px;margin-bottom:1rem"></div>
          <div class="skeleton" style="width:100%;height:60px;border-radius:8px"></div>
        </div>
      </div>
    </div>`;
}

function buildNotFoundHTML(message) {
  return /* html */ `
    <div class="main-container" style="text-align:center;padding:4rem 1rem">
      <h2 style="color:var(--dark-background)">404 — Not Found</h2>
      <p style="color:var(--grey-color);margin-top:1rem">${escapeHTML(message)}</p>
      <a href="/topics" data-link style="
        display:inline-block;margin-top:2rem;padding:0.8rem 1.5rem;
        background:var(--primary-color);color:#fff;border-radius:8px;text-decoration:none">
        Back to Topics
      </a>
    </div>`;
}

function buildErrorHTML(message) {
  return /* html */ `
    <div class="main-container" style="text-align:center;padding:4rem 1rem">
      <h2 style="color:#e53e3e">Something went wrong</h2>
      <p style="color:var(--grey-color);margin-top:1rem">${escapeHTML(message)}</p>
      <a href="/topics" data-link style="
        display:inline-block;margin-top:2rem;padding:0.8rem 1.5rem;
        background:var(--primary-color);color:#fff;border-radius:8px;text-decoration:none">
        Back to Topics
      </a>
    </div>`;
}
