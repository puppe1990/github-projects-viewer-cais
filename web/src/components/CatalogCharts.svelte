<script>
  import { formatCount } from "../lib/format.js";
  import {
    catalogTrafficTotals,
    dailyCatalogPulse,
    pulseBars,
    quietStarsRank,
    rankTraffic,
    trafficRepos,
    uniqueVisitorsByLanguage,
  } from "../lib/trafficCharts.js";
  import RankChart from "./RankChart.svelte";

  export let repos = [];

  $: tracked = trafficRepos(repos);
  $: totals = catalogTrafficTotals(repos);
  $: visitors = rankTraffic(repos, "view_uniques", 10);
  $: cloners = rankTraffic(repos, "clone_uniques", 10);
  $: pulse = pulseBars(dailyCatalogPulse(repos));
  $: languages = uniqueVisitorsByLanguage(repos);
  $: quiet = quietStarsRank(repos).map((row) => ({
    ...row,
    value: Math.round(row.score * 100) / 100,
  }));
</script>

{#if tracked.length === 0}
  <div class="empty">
    <strong>No traffic snapshots in this catalog yet.</strong>
    Charts fill in after traffic jobs run for repositories you can push to.
  </div>
{:else}
  <div class="catalog-charts">
    <div class="chart-totals">
      <div class="traffic-stat">
        <span>Unique visitors</span>
        <b>{formatCount(totals.viewUniques)}</b>
        <small>{formatCount(totals.tracked)} of {formatCount(totals.loaded)} repos</small>
      </div>
      <div class="traffic-stat">
        <span>Unique cloners</span>
        <b>{formatCount(totals.cloneUniques)}</b>
        <small>Summed GitHub windows</small>
      </div>
      <div class="traffic-stat">
        <span>Views</span>
        <b>{formatCount(totals.views)}</b>
        <small>Last 14 days</small>
      </div>
      <div class="traffic-stat">
        <span>Clones</span>
        <b>{formatCount(totals.clones)}</b>
        <small>Last 14 days</small>
      </div>
    </div>

    <RankChart title="Top unique visitors" headingId="chart-visitors" caption="GitHub unique visitors in the current window." rows={visitors} />
    <RankChart title="Top unique cloners" headingId="chart-cloners" caption="GitHub unique cloners in the current window." rows={cloners} tone="clones" />

    {#if pulse.length}
      <section class="chart-panel chart-panel-wide" aria-labelledby="chart-pulse">
        <header class="chart-head">
          <h3 id="chart-pulse">Last 14 days</h3>
          <p>Unique visitors and cloners summed across this catalog.</p>
        </header>
        <div class="pulse-chart">
          {#each pulse as day}
            <div class="pulse-col">
              <p class="pulse-tip">
                <span>{formatCount(day.viewUniques)}</span>
                <span class="is-clone">{formatCount(day.cloneUniques)}</span>
              </p>
              <div class="pulse-pair">
                <div class="traffic-bar" style={`height:${day.viewHeight}`} title={`${formatCount(day.viewUniques)} unique visitor${day.viewUniques === 1 ? "" : "s"}`}></div>
                <div class="traffic-bar traffic-bar-clone" style={`height:${day.cloneHeight}`} title={`${formatCount(day.cloneUniques)} unique cloner${day.cloneUniques === 1 ? "" : "s"}`}></div>
              </div>
              <span>{day.label}</span>
            </div>
          {/each}
        </div>
        <p class="chart-legend">
          <span class="chart-swatch"></span> Unique visitors
          <span class="chart-swatch is-clone"></span> Unique cloners
        </p>
      </section>
    {/if}

    <RankChart title="Unique visitors by language" headingId="chart-languages" caption="Where unique visitors land in this catalog." rows={languages} nameKey="language" />
    <RankChart title="Quiet stars" headingId="chart-quiet" caption="Unique visitors per GitHub star." rows={quiet} />
  </div>
{/if}
