<script lang="ts">
  import type { Snippet } from 'svelte';
  import { cn } from '$lib/ui/cn';

  let {
    title,
    actions,
    children,
    class: className,
  }: {
    title: string;
    actions?: Snippet;
    children: Snippet;
    class?: string;
  } = $props();
</script>

<div
  class={cn(
    'flex h-[calc(100dvh-4rem)] w-full max-w-screen-lg flex-col gap-5 p-3 md:h-full md:p-6',
    className,
  )}
>
  <header class="flex min-h-9 w-full items-center justify-between gap-3">
    <h1 class="flex h-6 items-center text-xl font-bold leading-none">
      {title}
    </h1>
    {@render actions?.()}
  </header>
  {@render children()}
</div>
