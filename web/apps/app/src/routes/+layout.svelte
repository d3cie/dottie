<script lang="ts">
  import '../app.css';
  let { children } = $props();
</script>

<svelte:head>
  <title>Dottie</title>
  <meta name="description" content="Private, self-hosted web analytics." />
</svelte:head>

{@render children()}

