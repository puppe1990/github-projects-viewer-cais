<script>
  export let page = 1;
  export let pages = 1;

  function go(next) {
    page = Math.min(pages, Math.max(1, next));
    document.getElementById("projects")?.scrollIntoView?.({ behavior: "smooth", block: "start" });
  }
</script>

{#if pages > 1}
  <nav class="catalog-pager" aria-label="Catalog pages">
    <button type="button" class="pager-btn" disabled={page <= 1} aria-label="Previous page" on:click={() => go(page - 1)}>Previous</button>
    <p class="pager-status">Page {page} of {pages}</p>
    <button type="button" class="pager-btn" disabled={page >= pages} aria-label="Next page" on:click={() => go(page + 1)}>Next</button>
  </nav>
{/if}
