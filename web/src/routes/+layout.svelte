<script lang="ts">
  import '@fontsource-variable/lora';
  import '@fontsource/nunito';
  import '@fontsource/nunito/200.css';
  import '@fontsource/nunito/300.css';
  import '@fontsource/nunito/400.css';
  import '@fontsource/nunito/500.css';
  import '@fontsource/nunito/600.css';
  import '@fontsource/nunito/700.css';
  import '@fontsource/nunito/800.css';
  import '@fontsource/nunito/900.css';
  import '../app.css';
  let { children } = $props();
</script>

<svelte:head>
  <title>Dottie</title>
  <meta name="description" content="Private, self-hosted web analytics." />
</svelte:head>

{@render children()}
