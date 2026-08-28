<script>
  import { formatCount } from "../lib/format.js";

  export let open = false;
  export let repo = {};
  export let traffic = {};

  $: title = `Traffic · ${repo.name || "repository"}`;
  $: viewsDays = traffic.views_by_day || [];
  $: clonesDays = traffic.clones_by_day || [];

  function close() {
    open = false;
  }

  function onKey(event) {
    if (event.key === "Escape") close();
  }

  function dayLabel(iso) {
    const date = new Date(iso);
    if (Number.isNaN(date.getTime())) return "";
    return date.toLocaleDateString("en-US", { month: "short", day: "numeric" });
  }

  function bars(days, key) {
    const max = Math.max(1, ...days.map((day) => day[key] || 0));
    return days.map((day) => ({
      label: dayLabel(day.timestamp),
      value: day[key] || 0,
      height: `${Math.max(6, Math.round(((day[key] || 0) / max) * 100))}%`,
    }));
  }
</script>

<svelte:window on:keydown={onKey} />

{#if open}
  <div class="sort-modal-overlay">
    <button type="button" class="sort-modal-scrim" aria-label="Dismiss traffic" on:click={close}></button>
    <div class="sort-modal traffic-modal" role="dialog" aria-modal="true" aria-labelledby="traffic-modal-title" tabindex="-1">
      <header class="sort-modal-head">
        <div>
          <p class="kicker">Last 14 days</p>
          <h2 id="traffic-modal-title">{title}</h2>
        </div>
        <button type="button" class="sort-modal-close" aria-label="Close traffic" on:click={close}>×</button>
      </header>

      <div class="traffic-stats">
        <div class="traffic-stat">
          <span>Views</span>
          <b>{formatCount(traffic.views)}</b>
          <small>{formatCount(traffic.view_uniques)} unique</small>
        </div>
        <div class="traffic-stat">
          <span>Clones</span>
          <b>{formatCount(traffic.clones)}</b>
          <small>{formatCount(traffic.clone_uniques)} unique</small>
        </div>
      </div>

      {#if viewsDays.length}
        <section class="sort-modal-block">
          <h3>Views by day</h3>
          <div class="traffic-chart" aria-hidden="true">
            {#each bars(viewsDays, "count") as bar}
              <div class="traffic-col">
                <div class="traffic-bar" style={`height:${bar.height}`}></div>
                <span>{bar.label}</span>
              </div>
            {/each}
          </div>
        </section>
      {/if}

      {#if clonesDays.length}
        <section class="sort-modal-block">
          <h3>Clones by day</h3>
          <div class="traffic-chart" aria-hidden="true">
            {#each bars(clonesDays, "count") as bar}
              <div class="traffic-col">
                <div class="traffic-bar traffic-bar-clone" style={`height:${bar.height}`}></div>
                <span>{bar.label}</span>
              </div>
            {/each}
          </div>
        </section>
      {/if}

      <div class="sort-modal-actions">
        <button type="button" class="btn btn-primary" on:click={close}>Done</button>
      </div>
    </div>
  </div>
{/if}
