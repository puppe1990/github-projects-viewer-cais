<script>
  import { SORT_FIELDS } from "../lib/filter.js";

  export let open = false;
  export let sortBy = "stars";
  export let sortOrder = "desc";
  export let sortThen = "";
  export let sortThenOrder = "desc";
  export let hasTraffic = false;

  const catalogFields = SORT_FIELDS.filter((field) => field.group === "catalog");
  const trafficFields = SORT_FIELDS.filter((field) => field.group === "traffic");
  $: thenFields = [{ value: "", label: "None" }, ...SORT_FIELDS.filter((field) => hasTraffic || field.group !== "traffic")];

  function close() {
    open = false;
  }

  function onKey(event) {
    if (event.key === "Escape") close();
  }
</script>

<svelte:window on:keydown={onKey} />

{#if open}
  <div class="sort-modal-overlay">
    <button type="button" class="sort-modal-scrim" aria-label="Dismiss sort" on:click={close}></button>
    <div class="sort-modal" role="dialog" aria-modal="true" aria-labelledby="sort-modal-title" tabindex="-1">
      <header class="sort-modal-head">
        <div>
          <p class="kicker">Sort</p>
          <h2 id="sort-modal-title">Arrange this catalog</h2>
        </div>
        <button type="button" class="sort-modal-close" aria-label="Close sort" on:click={close}>×</button>
      </header>

      <section class="sort-modal-block">
        <h3>Catalog</h3>
        <div class="sort-chips">
          {#each catalogFields as field}
            <button type="button" class="sort-chip" class:is-active={sortBy === field.value} aria-pressed={sortBy === field.value} on:click={() => (sortBy = field.value)}>
              {field.label}
            </button>
          {/each}
        </div>
      </section>

      {#if hasTraffic}
        <section class="sort-modal-block">
          <h3>Traffic · 14 days</h3>
          <div class="sort-chips">
            {#each trafficFields as field}
              <button type="button" class="sort-chip" class:is-active={sortBy === field.value} aria-pressed={sortBy === field.value} on:click={() => (sortBy = field.value)}>
                {field.label}
              </button>
            {/each}
          </div>
        </section>
      {/if}

      <section class="sort-modal-block">
        <h3>Order</h3>
        <div class="sort-chips">
          <button type="button" class="sort-chip" class:is-active={sortOrder === "desc"} aria-pressed={sortOrder === "desc"} on:click={() => (sortOrder = "desc")}>High → low</button>
          <button type="button" class="sort-chip" class:is-active={sortOrder === "asc"} aria-pressed={sortOrder === "asc"} on:click={() => (sortOrder = "asc")}>Low → high</button>
        </div>
      </section>

      <section class="sort-modal-block">
        <h3>Then by</h3>
        <div class="sort-chips">
          {#each thenFields as field}
            <button type="button" class="sort-chip" class:is-active={sortThen === field.value} aria-pressed={sortThen === field.value} on:click={() => (sortThen = field.value)}>
              {field.label}
            </button>
          {/each}
        </div>
        {#if sortThen}
          <div class="sort-chips sort-chips-follow">
            <button type="button" class="sort-chip" class:is-active={sortThenOrder === "desc"} aria-pressed={sortThenOrder === "desc"} on:click={() => (sortThenOrder = "desc")}>Then high → low</button>
            <button type="button" class="sort-chip" class:is-active={sortThenOrder === "asc"} aria-pressed={sortThenOrder === "asc"} on:click={() => (sortThenOrder = "asc")}>Then low → high</button>
          </div>
        {/if}
      </section>

      <div class="sort-modal-actions">
        <button type="button" class="btn btn-primary" on:click={close}>Done</button>
      </div>
    </div>
  </div>
{/if}
