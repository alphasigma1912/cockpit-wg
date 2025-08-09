(async function () {
  const root = document.getElementById('root');

  function renderSafeMode(code, detail) {
    if (!root) return;
    root.innerHTML = `
      <h1>Cockpit WireGuard Safe Mode</h1>
      <p><strong>Code:</strong> ${code}</p>
      <p>${detail || 'Check browser console or /etc/cockpit-wg/flags.json for next steps.'}</p>
    `;
  }
  window.__renderSafeMode = renderSafeMode;

  async function isDisabled() {
    try {
      const res = await fetch('/etc/cockpit-wg/flags.json', { cache: 'no-store' });
      if (!res.ok) return false;
      const data = await res.json();
      return data && data.disabled === true;
    } catch {
      return false;
    }
  }

  try {
    const params = new URLSearchParams(window.location.search);
    const forced = params.get('safe') === '1' || localStorage.getItem('cwm.safe') === '1';
    if (forced) {
      renderSafeMode('FORCED', 'Safe Mode was requested. Remove the flag to re-enable the app.');
      return;
    }
    if (await isDisabled()) {
      renderSafeMode('DISABLED', 'Plugin disabled by system flag. Remove the flag to load normally.');
      return;
    }
    const errors = await import('/src/globalErrorHandlers');
    errors.registerGlobalErrorHandlers();
    await import('/src/main.tsx');
  } catch (err) {
    console.error('bootstrap error', err);
    renderSafeMode('BOOT_ERR', 'Startup failed. See console for details.');
  }
})();
