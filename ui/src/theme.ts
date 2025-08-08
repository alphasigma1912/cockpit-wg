const setTheme = (isDark: boolean) => {
  const root = document.documentElement;
  root.classList.remove('pf-v5-theme-light', 'pf-v5-theme-dark');
  root.classList.add(isDark ? 'pf-v5-theme-dark' : 'pf-v5-theme-light');
};

const mq = window.matchMedia('(prefers-color-scheme: dark)');
setTheme(mq.matches);
mq.addEventListener('change', (e) => setTheme(e.matches));
