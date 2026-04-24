/**
 * navbar.js
 *
 * Renders the navbar into #navbar-root and manages the notification bell.
 *
 * This replaces:
 *   - frontend/html/partials/navbar.html  (the template)
 *   - frontend/static/js/notifications.js (the notification bell JS)
 *   - The Go BFF's AuthMiddleware passing .User into templates
 *
 * Call renderNavbar(user) on every route change so the nav reflects the
 * current auth state without a full page reload.
 */

import { escapeHTML } from './helpers.js';
import {
  fetchNotifications,
  fetchUnreadCount,
  markNotificationRead,
  markAllNotificationsRead,
} from './api.js';

// ─── Public API ───────────────────────────────────────────────────────────────

/**
 * Renders the navbar for the given user (or null for guests).
 * Attaches notification bell logic if a user is present.
 *
 * @param {object|null} user
 */
export function renderNavbar(user) {
  const root = document.getElementById('navbar-root');
  if (!root) return;

  root.innerHTML = buildNavbarHTML(user);

  if (user) {
    initNotifications(user);
    initActiveNavLink();
  }
}

// ─── HTML builders ────────────────────────────────────────────────────────────

function buildNavbarHTML(user) {
  return /* html */ `
    <header>
      <nav class="navbar">
        <div class="nav-container">
          <div class="logo">
            <a class="logo-link" href="/" data-link>
              <div class="logo-icon">
                <img src="/static/images/icons/logo-icon.png" alt="Logo Icon" />
              </div>
              <span class="logo-title">Forum</span>
            </a>
          </div>

          ${user ? buildLoggedInNav(user) : buildGuestNav()}
        </div>
      </nav>
    </header>
  `;
}

function buildLoggedInNav(user) {
  const avatarSrc =
    user.avatar_url || user.AvatarURL
      ? escapeHTML(user.avatar_url || user.AvatarURL)
      : '/static/images/user-avatar.png';

  const username = escapeHTML(user.username || user.Username || '');

  return /* html */ `
    <div class="welcome-box">
      <div class="welcome-user-box">
        <div class="welcome-user-text">
          <span class="welcome-user">Welcome,</span>
          <span class="welcome-user-name">${username}</span>
        </div>
        <a href="/activity" class="user-avatar-link" data-link>
          <span class="user-avatar">
            <img src="${avatarSrc}" alt="User Avatar" />
          </span>
        </a>
      </div>

      <ul class="nav-links">
        <!-- Notification bell -->
        <li class="nav-icon notification-wrapper">
          <button
            class="nav-icon-link notification-bell"
            id="notificationBell"
            aria-label="Notifications"
          >
            <svg width="24" height="24" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
              <path
                d="M12 22C13.1 22 14 21.1 14 20H10C10 21.1 10.9 22 12 22ZM18 16V11C18 7.93 16.37 5.36 13.5 4.68V4C13.5 3.17 12.83 2.5 12 2.5C11.17 2.5 10.5 3.17 10.5 4V4.68C7.64 5.36 6 7.92 6 11V16L4 18V19H20V18L18 16Z"
                fill="currentColor"
              />
            </svg>
            <span class="notification-badge" id="notificationBadge" style="display:none">0</span>
          </button>

          <div class="notification-dropdown" id="notificationDropdown" style="display:none">
            <div class="notification-header">
              <h3>Notifications</h3>
              <button class="mark-all-read-btn" id="markAllReadBtn">Mark all as read</button>
            </div>
            <div class="notification-list" id="notificationList">
              <p class="notification-empty">No notifications yet</p>
            </div>
          </div>
        </li>

        <li class="nav-link nav-link-create">
          <a href="/topics/create" data-link>New Post</a>
        </li>
        <li class="nav-link">
          <a href="/logout" data-link>Logout</a>
        </li>
      </ul>
    </div>
  `;
}

function buildGuestNav() {
  return /* html */ `
    <ul class="nav-links">
      <li class="nav-link">
        <a href="/login" data-link>Login</a>
      </li>
      <li class="nav-link">
        <a href="/register" data-link>Register</a>
      </li>
    </ul>
  `;
}

// ─── Active nav link highlighting ────────────────────────────────────────────

function initActiveNavLink() {
  const path = window.location.pathname;
  document.querySelectorAll('.nav-link a[data-link]').forEach((a) => {
    a.closest('.nav-link')?.classList.toggle('active', a.getAttribute('href') === path);
  });
}

// ─── Notification bell (mirrors frontend/static/js/notifications.js) ─────────

let _notificationSSE = null;

function initNotifications() {
  const bell = document.getElementById('notificationBell');
  const dropdown = document.getElementById('notificationDropdown');
  const markAllBtn = document.getElementById('markAllReadBtn');

  if (!bell || !dropdown) return;

  // Load initial unread count
  refreshUnreadCount();

  // Toggle dropdown on bell click
  bell.addEventListener('click', (e) => {
    e.stopPropagation();
    const isOpen = dropdown.style.display !== 'none';
    if (isOpen) {
      dropdown.style.display = 'none';
    } else {
      dropdown.style.display = 'flex';
      dropdown.style.flexDirection = 'column';
      loadNotifications();
    }
  });

  // Close dropdown on outside click
  document.addEventListener('click', (e) => {
    if (!bell.contains(e.target) && !dropdown.contains(e.target)) {
      dropdown.style.display = 'none';
    }
  });

  // Mark all as read
  markAllBtn?.addEventListener('click', async () => {
    try {
      await markAllNotificationsRead();
      refreshUnreadCount();
      loadNotifications();
    } catch (err) {
      console.error('Failed to mark all as read:', err);
    }
  });

  // Start SSE stream for live notification count updates
  startNotificationStream();
}

async function refreshUnreadCount() {
  try {
    const data = await fetchUnreadCount();
    const count = data?.count ?? data?.unread_count ?? 0;
    updateBadge(count);
  } catch {
    // Not logged in or network error — badge stays hidden
  }
}

function updateBadge(count) {
  const badge = document.getElementById('notificationBadge');
  if (!badge) return;
  if (count > 0) {
    badge.textContent = count > 99 ? '99+' : String(count);
    badge.style.display = 'block';
  } else {
    badge.style.display = 'none';
  }
}

async function loadNotifications() {
  const list = document.getElementById('notificationList');
  if (!list) return;

  list.innerHTML = `<p class="notification-empty">Loading…</p>`;

  try {
    const data = await fetchNotifications();
    const notifications = data?.notifications ?? data ?? [];

    if (!notifications.length) {
      list.innerHTML = `<p class="notification-empty">No notifications yet</p>`;
      return;
    }

    list.innerHTML = notifications.map(buildNotificationItemHTML).join('');

    // Click to mark as read
    list.querySelectorAll('.notification-item').forEach((item) => {
      item.addEventListener('click', async () => {
        const id = item.dataset.id;
        if (!id) return;
        try {
          await markNotificationRead(id);
          item.classList.remove('unread');
          item.querySelector('.notification-unread-dot')?.remove();
          refreshUnreadCount();
        } catch (err) {
          console.error('Failed to mark notification as read:', err);
        }
      });
    });
  } catch (err) {
    list.innerHTML = `<p class="notification-empty">Failed to load notifications</p>`;
    console.error(err);
  }
}

function buildNotificationItemHTML(n) {
  const isUnread = !n.read_at && !n.ReadAt;
  const icon = notificationIcon(n.type || n.Type);
  const title = escapeHTML(n.title || n.Title || 'Notification');
  const message = escapeHTML(n.message || n.Message || '');
  const time = formatNotificationTime(n.created_at || n.CreatedAt);

  return /* html */ `
    <div class="notification-item ${isUnread ? 'unread' : ''}" data-id="${escapeHTML(String(n.id || n.ID || ''))}">
      <div class="notification-icon ${escapeHTML(n.type || n.Type || '')}">
        ${icon}
      </div>
      <div class="notification-content">
        <div class="notification-title">${title}</div>
        <div class="notification-message">${message}</div>
        <div class="notification-time">${time}</div>
      </div>
      ${isUnread ? `<div class="notification-unread-dot"></div>` : ''}
    </div>
  `;
}

function notificationIcon(type) {
  switch (type) {
    case 'reply':
      return '💬';
    case 'like':
      return '❤️';
    case 'mention':
      return '📣';
    default:
      return '🔔';
  }
}

function formatNotificationTime(isoString) {
  if (!isoString) return '';
  const date = new Date(isoString);
  if (isNaN(date)) return '';
  const diffMin = Math.floor((Date.now() - date) / 60000);
  if (diffMin < 1) return 'just now';
  if (diffMin < 60) return `${diffMin}m ago`;
  const diffHr = Math.floor(diffMin / 60);
  if (diffHr < 24) return `${diffHr}h ago`;
  return date.toLocaleDateString('en-GB', { day: 'numeric', month: 'short' });
}

function startNotificationStream() {
  // Close any existing stream first (e.g. after a re-render / route change)
  if (_notificationSSE) {
    _notificationSSE.close();
    _notificationSSE = null;
  }

  try {
    const es = new EventSource('/api/v1/notifications/stream', {
      withCredentials: true,
    });

    es.addEventListener('notification', () => {
      refreshUnreadCount();
    });

    es.onerror = () => {
      // SSE reconnects automatically; we just close and let the browser retry
      es.close();
      _notificationSSE = null;
    };

    _notificationSSE = es;
  } catch (err) {
    console.warn('Could not start notification stream:', err);
  }
}

/**
 * Closes the SSE stream. Call this before logging out.
 */
export function closeNotificationStream() {
  if (_notificationSSE) {
    _notificationSSE.close();
    _notificationSSE = null;
  }
}
