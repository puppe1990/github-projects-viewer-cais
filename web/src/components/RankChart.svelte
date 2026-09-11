<script>
  import { formatCount } from "../lib/format.js";

  export let title = "";
  export let caption = "";
  export let headingId = "chart-rank";
  export let rows = [];
  export let tone = "views";
  export let nameKey = "name";
</script>

<section class="chart-panel" aria-labelledby={headingId}>
  <header class="chart-head">
    <h3 id={headingId}>{title}</h3>
    {#if caption}<p>{caption}</p>{/if}
  </header>
  {#if rows.length}
    <ol class="rank-list">
      {#each rows as row}
        <li class="rank-row">
          <span class="rank-name">{row[nameKey]}</span>
          <div class="rank-track" aria-hidden="true">
            <i class="rank-fill" class:is-clone={tone === "clones"} style={`--bar:${row.share}%`}></i>
          </div>
          <b class="rank-value">{formatCount(row.value)}</b>
        </li>
      {/each}
    </ol>
  {:else}
    <p class="chart-empty">Nothing with traffic in this slice.</p>
  {/if}
</section>
