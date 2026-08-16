<script lang="ts">
  import { Check, Clipboard } from '@lucide/svelte';
  import { Button } from '$lib/ui/button';
  import { Card } from '$lib/ui/card';

  let {
    snippet,
    copied,
    onCopy,
  }: { snippet: string; copied: boolean; onCopy: () => void } = $props();
</script>

<Card
  class="from-elevated via-elevated mb-3 w-full gap-2 rounded-xl bg-gradient-to-tl to-green-50 p-3 shadow-sm"
>
  <h2 class="mt-1 font-serif text-xl font-semibold">
    install dottie on your website
  </h2>
  <p class="text-secondary-foreground max-w-2xl text-sm leading-relaxed">
    it takes about a minute. add this script to your website's
    <code
      class="bg-elevated text-foreground rounded-sm border px-1 py-0.5 font-mono shadow-sm"
      >&lt;head&gt;</code
    >
    section to start seeing insights here.
  </p>
  <code
    class="bg-elevated mt-2 overflow-x-auto rounded-md border p-2.5 text-xs shadow-sm"
    >{snippet}</code
  >
  <div class="mt-1 flex justify-end">
    <Button size="sm" onclick={onCopy}>
      {#if copied}<Check class="size-3.5" /> copied{:else}<Clipboard
          class="size-3.5"
        /> copy script{/if}
    </Button>
  </div>
</Card>
