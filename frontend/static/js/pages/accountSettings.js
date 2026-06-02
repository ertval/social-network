export async function renderAccountSettingsPage() {
  const root = document.getElementById('app-root');
  if (!root) return;

  root.innerHTML = /* html */ `
    <div class="main-container">
      <section class="settings-card">
        <div class="settings-card-header">
          <h1 class="settings-title">Account Settings</h1>
          <p class="settings-subtitle">Manage the providers connected to your forum account.</p>
        </div>

        <div class="settings-list">
          <a href="/api/v1/auth/github/link" class="settings-item-link">
            <div class="settings-item-copy">
              <span class="settings-item-title">Link To Github Provider</span>
              <span class="settings-item-description">Connect your GitHub account to use it with this profile.</span>
            </div>
            <img src="/static/images/icons/github-white-logo.png" alt="GitHub" class="settings-provider-icon settings-provider-icon-github" />
          </a>

          <a href="/api/v1/auth/google/link" class="settings-item-link">
            <div class="settings-item-copy">
              <span class="settings-item-title">Link To Google Provider</span>
              <span class="settings-item-description">Connect your Google account to use it with this profile.</span>
            </div>
            <img src="/static/images/icons/google-logo.png" alt="Google" class="settings-provider-icon" />
          </a>
        </div>
      </section>
    </div>
  `;
}
