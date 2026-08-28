<script>
  const THEME_KEY = "gpv-theme";

  function currentTheme() {
    if (typeof document === "undefined") return "dark";
    return document.documentElement.dataset.theme === "light" ? "light" : "dark";
  }

  let theme = currentTheme();

  function apply(next) {
    theme = next === "light" ? "light" : "dark";
    if (typeof document === "undefined") return;
    document.documentElement.dataset.theme = theme;
    localStorage.setItem(THEME_KEY, theme);
  }

  function toggle() {
    apply(theme === "light" ? "dark" : "light");
  }
</script>

<button type="button" id="theme-toggle" class="theme-toggle" aria-pressed={theme === "light"} on:click={toggle}>
  <svg class="icon-sun" viewBox="0 0 24 24" width="16" height="16" aria-hidden="true">
    <circle cx="12" cy="12" r="4" fill="none" stroke="currentColor" stroke-width="1.8"/>
    <path d="M12 3v2.2M12 18.8V21M4.9 4.9l1.6 1.6M17.5 17.5l1.6 1.6M3 12h2.2M18.8 12H21M4.9 19.1l1.6-1.6M17.5 6.5l1.6-1.6" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"/>
  </svg>
  <svg class="icon-moon" viewBox="0 0 24 24" width="16" height="16" aria-hidden="true">
    <path d="M16.2 13.4A6.4 6.4 0 0 1 10.6 4.8 7 7 0 1 0 19.2 13.4a6.3 6.3 0 0 1-3 0z" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"/>
  </svg>
  <span class="theme-toggle-label">{theme === "light" ? "Dark mode" : "Light mode"}</span>
</button>
