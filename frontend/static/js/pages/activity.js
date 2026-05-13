/**
 * pages/activity.js
 *
 * Renders the /activity page.
 * Mirrors the Go BFF's ActivityPage handler + activity.html template.
 *
 * Data shape from GET /api/v1/user/activity (after apiFetch unwraps "data"):
 * {
 *   CreatedTopics:    [{ ID, Title, CreatedAt }]
 *   LikedTopics:      [{ ID, Title, CreatedAt }]
 *   DislikedTopics:   [{ ID, Title, CreatedAt }]
 *   LikedComments:    [{ CommentID, TopicID, TopicTitle, CreatedAt }]
 *   DislikedComments: [{ CommentID, TopicID, TopicTitle, CreatedAt }]
 *   UserComments:     [{ ID, TopicID, TopicTitle, Content, CreatedAt }]
 * }
 */

import { fetchUserActivity } from '../api.js';
import { escapeHTML } from '../helpers.js';

export async function renderActivityPage(user) {
  const root = document.getElementById('app-root');
  if (!root) return;

  try {
    const activity = await fetchUserActivity();
    root.innerHTML = buildPageHTML(activity, user);
  } catch (err) {
    console.error('Failed to load activity:', err);
    root.innerHTML = buildErrorHTML();
  }
}

// ─── Page builder ─────────────────────────────────────────────────────────────

function buildPageHTML(activity, user) {
  const {
    CreatedTopics = [],
    LikedTopics = [],
    DislikedTopics = [],
    LikedComments = [],
    DislikedComments = [],
    UserComments = [],
  } = activity ?? {};

  const isEmpty =
    !CreatedTopics.length &&
    !LikedTopics.length &&
    !DislikedTopics.length &&
    !LikedComments.length &&
    !DislikedComments.length &&
    !UserComments.length;

  return /* html */ `
    <div class="main-container">
      ${buildProfileHeader(user, CreatedTopics.length)}
      <div class="activity-container">
        <div class="activity-header">
          <h2 class="activity-page-title">Activity</h2>
        </div>
        ${isEmpty
      ? buildEmptyHTML()
      : buildSectionsHTML({
        CreatedTopics,
        LikedTopics,
        DislikedTopics,
        LikedComments,
        DislikedComments,
        UserComments,
      })
    }
      </div>
    </div>
  `;
}

function buildProfileHeader(user, postsCount) {
  const avatarSrc = escapeHTML(
    user?.avatar_url || user?.AvatarURL || '/static/images/user-avatar.png'
  );
  const username = escapeHTML(user?.username || user?.Username || '');

  return /* html */ `
    <section class="profile-card">
      <div class="profile-card-top">
        <div class="profile-avatar-shell">
          <div class="profile-avatar">
            <img src="${avatarSrc}" alt="User Avatar" />
          </div>
        </div>

        <div class="profile-main">
          <h1 class="profile-name">${username}</h1>
          <div class="profile-stats">
            <div class="profile-stat">
              <span class="profile-stat-value">-</span>
              <span class="profile-stat-label">Followers</span>
            </div>
            <div class="profile-stat">
              <span class="profile-stat-value">-</span>
              <span class="profile-stat-label">Following</span>
            </div>
            <div class="profile-stat">
              <span class="profile-stat-value">${escapeHTML(String(postsCount))}</span>
              <span class="profile-stat-label">Posts</span>
            </div>
          </div>
          <div class="profile-bio">
            <p class="profile-bio-text">NVIM BTW</p>
          </div>
        </div>
      </div>
    </section>
  `;
}

function buildSectionsHTML(activity) {
  const sections = [
    activity.CreatedTopics.length
      ? buildTopicSection('Posts You Created', activity.CreatedTopics, 'You created')
      : '',
    activity.LikedTopics.length
      ? buildTopicSection('Topics You Liked', activity.LikedTopics, 'You liked')
      : '',
    activity.DislikedTopics.length
      ? buildTopicSection('Topics You Disliked', activity.DislikedTopics, 'You disliked')
      : '',
    activity.LikedComments.length
      ? buildCommentVoteSection(
        'Comments You Liked',
        activity.LikedComments,
        'You liked a comment in'
      )
      : '',
    activity.DislikedComments.length
      ? buildCommentVoteSection(
        'Comments You Disliked',
        activity.DislikedComments,
        'You disliked a comment in'
      )
      : '',
    activity.UserComments.length ? buildUserCommentsSection(activity.UserComments) : '',
  ];

  return sections.filter(Boolean).join('');
}

// ─── Section builders ─────────────────────────────────────────────────────────

function buildTopicSection(title, topics, verb) {
  return /* html */ `
    <div class="activity-section">
      <h3 class="activity-section-title">${title}</h3>
      ${topics
      .map(
        (t) => /* html */ `
        <div class="activity-row">
          <div class="activity-content">
            <p class="activity-text">
              ${verb}:
              <a href="/topic/${t.ID}" class="activity-link" data-link>
                ${escapeHTML(t.Title)}
              </a>
            </p>
            <span class="activity-date">${escapeHTML(t.CreatedAt)}</span>
          </div>
        </div>
      `
      )
      .join('')}
    </div>
  `;
}

function buildCommentVoteSection(title, comments, verb) {
  return /* html */ `
    <div class="activity-section">
      <h3 class="activity-section-title">${title}</h3>
      ${comments
      .map(
        (c) => /* html */ `
        <div class="activity-row">
          <div class="activity-content">
            <p class="activity-text">
              ${verb}:
              <a href="/topic/${c.TopicID}" class="activity-link" data-link>
                ${escapeHTML(c.TopicTitle)}
              </a>
            </p>
            <span class="activity-date">${escapeHTML(c.CreatedAt)}</span>
          </div>
        </div>
      `
      )
      .join('')}
    </div>
  `;
}

function buildUserCommentsSection(comments) {
  return /* html */ `
    <div class="activity-section">
      <h3 class="activity-section-title">Your Comments</h3>
      ${comments
      .map(
        (c) => /* html */ `
        <div class="activity-row activity-row-comment">
          <div class="activity-content">
            <p class="activity-text">
              You commented in:
              <a href="/topic/${c.TopicID}" class="activity-link" data-link>
                ${escapeHTML(c.TopicTitle)}
              </a>
            </p>
            <div class="activity-comment-preview">
              <p class="comment-preview-text">"${escapeHTML(c.Content)}"</p>
            </div>
            <span class="activity-date">${escapeHTML(c.CreatedAt)}</span>
          </div>
        </div>
      `
      )
      .join('')}
    </div>
  `;
}

// ─── State helpers ────────────────────────────────────────────────────────────

function buildEmptyHTML() {
  return /* html */ `
    <div class="activity-empty">
      <p class="activity-empty-text">
        No activity yet. Start by creating a post or commenting!
      </p>
    </div>
  `;
}

function buildErrorHTML() {
  return /* html */ `
    <div class="main-container">
      <div class="activity-empty">
        <p class="activity-empty-text">Failed to load activity. Please try again.</p>
      </div>
    </div>
  `;
}
