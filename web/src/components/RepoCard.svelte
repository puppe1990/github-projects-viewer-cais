<script>
  import { absoluteUrl, formatCount, formatRelative, languageColor } from "../lib/format.js";
  import TrafficModal from "./TrafficModal.svelte";

  export let repo = {};
  export let index = 0;

  const FORK_ICON = `<svg class="count-icon" viewBox="0 0 16 16" width="12" height="12" aria-hidden="true"><path fill="currentColor" d="M5 5.372v.878c0 .414.336.75.75.75h4.5a.75.75 0 0 0 .75-.75v-.878a2.25 2.25 0 1 1 1.5 0v.878a2.25 2.25 0 0 1-2.25 2.25h-1.5v2.128a2.251 2.251 0 1 1-1.5 0V8.5h-1.5A2.25 2.25 0 0 1 3.5 6.25v-.878a2.25 2.25 0 1 1 1.5 0ZM5 3.25a.75.75 0 1 0-1.5 0 .75.75 0 0 0 1.5 0Zm6.75.75a.75.75 0 1 0 0-1.5.75.75 0 0 0 0 1.5Zm-3 8.75a.75.75 0 1 0-1.5 0 .75.75 0 0 0 1.5 0Z"/></svg>`;

  $: language = repo.language || "Not specified";
  $: color = languageColor(repo.language);
  $: homepageUrl = absoluteUrl(repo.homepage);
  $: traffic = repo.traffic || {};
  let trafficOpen = false;
  let liveTraffic = null;

  $: modalTraffic = liveTraffic || traffic;

  async function openTraffic() {
    trafficOpen = true;
    const owner = repo.owner_login;
    const name = repo.name;
    if (!owner || !name) return;
    try {
      const response = await fetch(`/traffic/${encodeURIComponent(owner)}/${encodeURIComponent(name)}`);
      if (!response.ok) return;
      liveTraffic = await response.json();
    } catch {
      /* keep snapshot already on the card */
    }
  }
</script>

<article class="card" style={`animation-delay: ${Math.min(index, 12) * 40}ms`}>
  <div class="card-top">
    <span class="lang"><i style={`background:${color}; box-shadow: 0 0 0 3px ${color}22`}></i>{language}</span>
    <span class="card-counts">
      <span class="stars" title={`${formatCount(repo.stargazers_count)} stars`}>★ {formatCount(repo.stargazers_count)}</span>
      <span class="forks" title={`${formatCount(repo.forks_count)} forks`}>{@html FORK_ICON} {formatCount(repo.forks_count)}</span>
    </span>
  </div>
  {#if traffic.available}
    <div class="card-traffic">
      <span class="views" title={`${formatCount(traffic.views)} views · ${formatCount(traffic.view_uniques)} unique`}>↗ {formatCount(traffic.views)} <span class="metric-uniques">· {formatCount(traffic.view_uniques)} unique</span></span>
      <span class="clones-metric" title={`${formatCount(traffic.clones)} clones · ${formatCount(traffic.clone_uniques)} unique`}>↓ {formatCount(traffic.clones)} <span class="metric-uniques">· {formatCount(traffic.clone_uniques)} unique</span></span>
    </div>
  {/if}
  <h2><a href={repo.html_url} target="_blank" rel="noopener noreferrer">{repo.name}</a></h2>
  {#if repo.fork || repo.archived || repo.private}
    <div class="badges">
      {#if repo.private}<span class="badge private">Private</span>{/if}
      {#if repo.fork}<span class="badge fork">Fork</span>{/if}
      {#if repo.archived}<span class="badge archived">Archived</span>{/if}
    </div>
  {/if}
  <p class="desc">{repo.description || "No description provided."}</p>
  <p class="meta"><span>Updated {formatRelative(repo.updated_at)}</span></p>
  <div class="actions">
    <a class="btn btn-primary" href={repo.html_url} target="_blank" rel="noopener noreferrer">Open repo</a>
    {#if homepageUrl}
      <a class="btn btn-ghost" href={homepageUrl} target="_blank" rel="noopener noreferrer">Homepage</a>
    {/if}
    {#if traffic.available}
      <button type="button" class="btn btn-ghost" aria-label="View traffic" on:click={openTraffic}>Traffic</button>
    {/if}
  </div>
</article>

<TrafficModal bind:open={trafficOpen} {repo} traffic={modalTraffic} />
