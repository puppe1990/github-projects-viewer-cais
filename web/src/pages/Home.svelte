<script>
  import { router } from "@inertiajs/svelte";
  import ThemeToggle from "../components/ThemeToggle.svelte";
  import RepoCard from "../components/RepoCard.svelte";
  import SortModal from "../components/SortModal.svelte";
  import { applyFilters, languageOptions } from "../lib/filter.js";
  import { absoluteUrl, formatCount } from "../lib/format.js";

  export let profile = {};
  export let orgs = [];
  export let repos = [];
  export let source = { type: "user", login: "" };
  export let error = "";
  export let populated = false;
  export let lookup = "";

  let username = lookup || "";
  let language = "";
  let minStars = "";
  let query = "";
  let includeForks = true;
  let hasHomepage = false;
  let sortBy = "stars";
  let sortOrder = "desc";
  let sortThen = "";
  let sortThenOrder = "desc";
  let sortOpen = false;
  let loading = false;
  let preferTraffic = false;
  $: if (!preferTraffic && (repos || []).some((repo) => repo.traffic?.available)) {
    sortBy = "traffic";
    preferTraffic = true;
  }

  $: langs = languageOptions(repos);
  $: visible = applyFilters(repos, { language, minStars, query, includeForks, hasHomepage, sortBy, sortOrder, sortThen, sortThenOrder });
  $: if (typeof document !== "undefined") {
    document.body.classList.toggle("is-populated", !!populated);
  }

  function go(path) {
    loading = true;
    router.get(path, {}, { onFinish: () => { loading = false } });
  }

  function submit() {
    const login = username.trim().replace(/^@/, "");
    if (!login) return;
    go(`/u/${encodeURIComponent(login)}`);
  }

  function hint(name) {
    username = name;
    go(`/u/${encodeURIComponent(name)}`);
  }

  function selectSource(type, login) {
    if (type === "org") {
      go(`/u/${encodeURIComponent(profile.login)}/orgs/${encodeURIComponent(login)}`);
      return;
    }
    go(`/u/${encodeURIComponent(profile.login)}`);
  }

  $: resultsMeta = (() => {
    if (!populated) return "";
    const shown = formatCount(visible.length);
    const loaded = formatCount((repos || []).length);
    if (source?.type === "org") return `${shown} of ${loaded} in ${source.login}`;
    const publicCount = profile?.public_repos;
    if (publicCount && publicCount > (repos || []).length) {
      return `${shown} of ${loaded} loaded · ${formatCount(publicCount)} public`;
    }
    return `${shown} of ${loaded} repositories`;
  })();

  $: blogUrl = absoluteUrl(profile?.blog);
</script>

<svelte:head>
  <title>{populated && profile?.login ? `${profile.name || profile.login} — Projects Viewer` : "Projects Viewer — GitHub"}</title>
</svelte:head>

<div class="atmosphere" aria-hidden="true">
  <div class="glow glow-a"></div>
  <div class="glow glow-b"></div>
  <div class="grid-lines"></div>
  <div class="grain"></div>
</div>

<ThemeToggle />

<div class="app">
  <header class="masthead">
    <p class="kicker">GitHub public atlas</p>
    <h1>Projects<em>.</em> Viewer</h1>
    <p class="lede">Look up a username and read the public work like a catalog — stars, language, and the things that actually shipped.</p>
  </header>

  <form class="search-form" autocomplete="off" on:submit|preventDefault={submit}>
    <label class="sr-only" for="username">GitHub username</label>
    <div class="search-shell">
      <span class="search-prefix" aria-hidden="true">@</span>
      <input id="username" name="username" type="text" spellcheck="false" autocapitalize="off" autocorrect="off" placeholder="username" bind:value={username}>
      <button type="submit" id="search-btn">Look up</button>
    </div>
    <p class="hints">
      <span>Try</span>
      <button type="button" class="hint" on:click={() => hint("sindresorhus")}>sindresorhus</button>
      <button type="button" class="hint" on:click={() => hint("gaearon")}>gaearon</button>
      <button type="button" class="hint" on:click={() => hint("torvalds")}>torvalds</button>
      <button type="button" class="hint" on:click={() => hint("octocat")}>octocat</button>
    </p>
  </form>

  <div id="status" class="status" class:error={!!error} role="status" aria-live="polite">{error}</div>

  {#if populated && profile?.login}
    <section id="profile" class="profile">
      <img class="profile-avatar" src={profile.avatar_url} alt={`${profile.login} avatar`} width="76" height="76">
      <div class="profile-copy">
        <h2><a href={profile.html_url} target="_blank" rel="noopener noreferrer">{profile.name || profile.login}</a></h2>
        <p class="handle">@{profile.login}</p>
        {#if profile.bio}<p class="bio">{profile.bio}</p>{/if}
      </div>
      <div class="profile-stats">
        <span class="stat">Repos <b>{formatCount(profile.public_repos)}</b></span>
        <span class="stat">Followers <b>{formatCount(profile.followers)}</b></span>
        <span class="stat">Following <b>{formatCount(profile.following)}</b></span>
        {#if orgs.length}<span class="stat">Orgs <b>{formatCount(orgs.length)}</b></span>{/if}
        {#if profile.location}<span class="stat">From <b>{profile.location}</b></span>{/if}
        {#if profile.company}<span class="stat">Works <b>{profile.company}</b></span>{/if}
        {#if blogUrl}
          <a class="stat" href={blogUrl} target="_blank" rel="noopener noreferrer">Site <b>{(profile.blog || "").replace(/^https?:\/\//, "")}</b></a>
        {/if}
      </div>
    </section>
  {/if}

  {#if populated && orgs.length}
    <section id="catalog-nav" class="catalog-nav">
      <p class="catalog-kicker">Catalog</p>
      <div id="catalog-sources" class="catalog-sources">
        <button type="button" class="org-chip" class:is-active={source?.type === "user"} aria-pressed={source?.type === "user"} on:click={() => selectSource("user", profile.login)}>
          <img src={profile.avatar_url} alt="" width="22" height="22">
          <span>Personal</span>
          {#if source?.type === "user"}<span class="org-chip-count">{formatCount(repos.length)}</span>{/if}
        </button>
        {#each orgs as org}
          <button type="button" class="org-chip" class:is-active={source?.type === "org" && source?.login === org.login} title={org.description || org.login} aria-pressed={source?.type === "org" && source?.login === org.login} on:click={() => selectSource("org", org.login)}>
            <img src={org.avatar_url} alt="" width="22" height="22">
            <span>{org.login}</span>
          </button>
        {/each}
      </div>
    </section>
  {/if}

  {#if populated}
    <section id="filters" class="filters">
      <div class="filters-head">
        <p id="results-meta" class="results-meta">{resultsMeta}</p>
        <div class="toggles">
          <label class="switch">
            <input id="filter-include-forks" type="checkbox" bind:checked={includeForks}>
            <span class="switch-ui" aria-hidden="true"></span>
            <span>Include forks</span>
          </label>
          <label class="switch">
            <input id="filter-has-homepage" type="checkbox" bind:checked={hasHomepage}>
            <span class="switch-ui" aria-hidden="true"></span>
            <span>Has homepage</span>
          </label>
        </div>
      </div>
      <div class="filters-grid">
        <label class="field">
          <span>Language</span>
          <select id="filter-language" bind:value={language}>
            <option value="">All</option>
            {#each langs as lang}
              <option value={lang}>{lang}</option>
            {/each}
          </select>
        </label>
        <label class="field field-stars">
          <span>Min stars</span>
          <input id="filter-stars-min" type="number" min="0" placeholder="Any" inputmode="numeric" bind:value={minStars}>
        </label>
        <label class="field field-grow">
          <span>In name / description</span>
          <input id="filter-query" type="search" placeholder="Filter this catalog" bind:value={query}>
        </label>
        <div class="field field-sort">
          <span>Sort</span>
          <div class="split split-sort">
            <select id="sort-by" aria-label="Sort by" bind:value={sortBy}>
              <option value="stars">Stars</option>
              <option value="forks">Forks</option>
              <option value="updated">Updated</option>
              <option value="name">Name</option>
              <option value="traffic">Traffic</option>
              <option value="clones">Clones</option>
            </select>
            <select id="sort-order" aria-label="Sort order" bind:value={sortOrder}>
              <option value="desc">Desc</option>
              <option value="asc">Asc</option>
            </select>
            <button type="button" class="sort-more" aria-haspopup="dialog" aria-expanded={sortOpen} aria-label="More sort options" on:click={() => (sortOpen = true)}>
              More
            </button>
          </div>
        </div>
      </div>
    </section>
  {/if}

  {#if loading}
    <div id="loader" class="loader-wrap">
      <div class="loader" aria-hidden="true"></div>
      <p id="loader-copy">Reading the public graph…</p>
    </div>
  {/if}

  {#if populated && !loading}
    <div id="projects" class="projects">
      {#if visible.length === 0}
        <div class="empty">
          {#if source?.type === "org"}
            <strong>Nothing in {source.login} matches.</strong>Loosen language, stars, or the homepage toggle.
          {:else}
            <strong>Nothing matches these filters.</strong>Loosen language, stars, or the homepage toggle.
          {/if}
        </div>
      {:else}
        {#each visible as repo, index}
          <RepoCard {repo} {index} />
        {/each}
      {/if}
    </div>
  {/if}
</div>

<SortModal bind:open={sortOpen} bind:sortBy bind:sortOrder bind:sortThen bind:sortThenOrder />
