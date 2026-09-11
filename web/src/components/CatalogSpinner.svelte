<script>
  export let size = "md";
</script>

<div class="atlas-spinner" class:is-sm={size === "sm"} aria-hidden="true">
  <span class="atlas-spinner-orbit"></span>
  <span class="atlas-spinner-arc"></span>
  <span class="atlas-spinner-core"></span>
</div>
