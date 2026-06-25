const loader = document.getElementById('loader');
const result = document.getElementById('result');

document.getElementById('tab-shorten').addEventListener('click', () => switchTab('shorten'));
document.getElementById('tab-expand').addEventListener('click',  () => switchTab('expand'));

// Show/hide the 4-dot loader around HTMX requests
document.body.addEventListener('htmx:beforeRequest', () => {
  loader.style.display = 'flex';
  result.innerHTML = '';
});
document.body.addEventListener('htmx:afterRequest', () => {
  loader.style.display = 'none';
});

// Tab toggle: switches between Shorten and Expand modes
function switchTab(mode) {
  const isShorten = mode === 'shorten';

  document.getElementById('form-shorten').hidden = !isShorten;
  document.getElementById('form-expand').hidden  =  isShorten;

  const tabShorten = document.getElementById('tab-shorten');
  const tabExpand  = document.getElementById('tab-expand');
  tabShorten.classList.toggle('active', isShorten);
  tabShorten.setAttribute('aria-selected', String(isShorten));
  tabExpand.classList.toggle('active', !isShorten);
  tabExpand.setAttribute('aria-selected', String(!isShorten));

  result.innerHTML = '';

  const input = document.getElementById(isShorten ? 'input-shorten' : 'input-expand');
  if (input) { input.value = ''; input.focus(); }
}

// Event delegation for result card buttons (avoids inline onclick — required for strict CSP)
result.addEventListener('click', (e) => {
  const copyBtn = e.target.closest('.btn-copy');
  if (copyBtn) { copyURL(copyBtn); return; }

  const openBtn = e.target.closest('.btn-open');
  if (openBtn) { openURL(openBtn); }
});

function copyURL(btn) {
  const url = btn.dataset.url;
  if (!url) return;
  navigator.clipboard.writeText(url).then(() => {
    const orig = btn.textContent;
    btn.textContent = 'Copied!';
    setTimeout(() => { btn.textContent = orig; }, 2000);
  });
}

function openURL(btn) {
  const url = btn.dataset.url;
  // Guard against non-http(s) schemes as a client-side defence-in-depth measure.
  // The server already validates this; this is a secondary check.
  if (url && (url.startsWith('https://') || url.startsWith('http://'))) {
    window.open(url, '_blank', 'noopener,noreferrer');
  }
}
