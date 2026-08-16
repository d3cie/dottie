<script lang="ts">
  let { class: className = 'size-9' }: { class?: string } = $props();
</script>

<svg class={className} viewBox="0 0 40 40" role="img" aria-label="Dottie">
  <rect width="40" height="40" rx="11" fill="currentColor" />
  <circle cx="14" cy="20" r="3.25" fill="white" />
  <circle cx="26" cy="20" r="3.25" fill="white" />
</svg>

