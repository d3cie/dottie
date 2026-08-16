<script lang="ts">
  import type { Breakdown } from './api';
  let { title, items, loading = false }: { title: string; items: Breakdown[]; loading?: boolean } = $props();
</script>

<section class="card min-h-88 p-3">
  <header class="mb-3 flex items-center justify-between">
    <h3 class="text-sm font-semibold">{title}</h3>
  </header>
  <div class:opacity-40={loading} class="space-y-1.5 transition-opacity">
    {#if !items.length && !loading}
      <div class="flex h-64 items-center justify-center text-sm text-muted-foreground">no data.</div>
    {/if}
    {#each items as item (item.key)}
      <div class="relative flex h-7 items-center justify-between overflow-hidden rounded-sm px-2 text-sm">
        <div class="absolute inset-y-0 left-0 rounded-r-sm border-l-[3px] border-primary/25 bg-secondary" style={`width: ${Math.max(item.percent, 2)}%`}></div>
        <span class="relative min-w-0 truncate pr-3">{item.key || 'unknown'}</span>
        <span class="relative shrink-0 tabular-nums">{item.count} <span class="text-muted-foreground">({item.percent.toFixed(1)}%)</span></span>
      </div>
    {/each}
  </div>
</section>

