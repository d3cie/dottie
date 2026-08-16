<script lang="ts">
  import { Search } from '@lucide/svelte';
  import type { Breakdown } from './api';
  import { Button } from '$lib/ui/button';
  import { Card } from '$lib/ui/card';
  import { Input } from '$lib/ui/input';
  import { cn } from '$lib/ui/cn';

  let {
    title,
    items,
    loading = false,
    class: className,
  }: {
    title: string;
    items: Breakdown[];
    loading?: boolean;
    class?: string;
  } = $props();
  let searching = $state(false);
  let query = $state('');
  const visible = $derived(
    items.filter((item) =>
      item.key.toLowerCase().includes(query.toLowerCase()),
    ),
  );
</script>

<Card
  class={cn(
    'bg-elevated relative flex min-h-80 w-full flex-col overflow-hidden rounded-md border p-3 shadow-sm',
    className,
  )}
>
  <header
    class="flex min-h-6 items-center justify-between text-sm font-semibold transition-opacity"
    class:opacity-50={loading}
  >
    <span>{title}</span>
    <Button
      variant="ghost"
      size="icon-xs"
      class="bg-background text-muted-foreground hover:text-foreground"
      aria-label={`search ${title}`}
      onclick={() => {
        searching = !searching;
        if (!searching) query = '';
      }}
    >
      <Search class="size-3.5" />
    </Button>
  </header>
  {#if searching}
    <Input
      class="mt-1 h-8 bg-elevated"
      bind:value={query}
      placeholder={`search ${title}`}
    />
  {/if}
  <div
    class="relative mt-2 flex min-h-0 flex-1 flex-col gap-0.5 transition-opacity"
    class:opacity-50={loading}
  >
    {#if visible.length === 0 && !loading}
      <div class="flex flex-1 items-center justify-center text-sm">
        no data.
      </div>
    {/if}
    {#each visible.slice(0, 9) as item (item.key)}
      <div
        class="bg-background/50 hover:bg-secondary/60 group relative flex h-[1.625rem] min-w-0 items-center justify-between overflow-hidden rounded-l-[2px] py-0.5 pr-2 text-sm transition-colors"
      >
        <div
          class="from-muted/60 to-muted pointer-events-none absolute inset-y-0 left-0 z-0 rounded-r-[2px] border-l-[3px] border-secondary bg-gradient-to-r"
          style={`width: ${Math.max(item.percent, 1)}%`}
        ></div>
        <span class="z-10 min-w-0 flex-1 truncate pl-2 pr-3 text-left"
          >{item.key || 'unknown'}</span
        >
        <span class="z-10 shrink-0 text-right tabular-nums"
          >{item.count}
          <span class="text-muted-foreground">({item.percent.toFixed(1)}%)</span
          ></span
        >
      </div>
    {/each}
  </div>
  <footer
    class="mt-2 flex items-center justify-between text-xs text-muted-foreground"
  >
    <span>{visible.length || '-'} results</span>
  </footer>
</Card>
