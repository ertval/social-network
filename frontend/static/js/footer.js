/**
 * footer.js
 *
 * Renders the footer into #footer-root.
 * Mirrors frontend/html/partials/footer.html.
 */

export function renderFooter() {
  const root = document.getElementById('footer-root');
  if (!root) return;

  root.innerHTML = /* html */ `
    <footer>
      <div class="main-container">
        <div class="footer-container">
          <p>Made with dedication and passion by</p>
          <div class="authors">
            <span class="author">geoikonomou,</span>
            <span class="author">epapamic,</span>
            <span class="author">agkiata,</span>
            <span class="author">sos247</span>
          </div>
          <span>- Forum © 2025 -</span>
        </div>
      </div>
    </footer>
  `;
}
