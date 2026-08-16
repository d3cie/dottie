<script lang="ts">
  import { TrendingDown, TrendingUp } from '@lucide/svelte';
  import { Skeleton } from '$lib/ui/skeleton';
  import { cn } from '$lib/ui/cn';

  let {
    title,
    icon: Icon,
    value,
    change,
    format,
    inverse = false,
    loading = false,
  }: {
    title: string;
    icon: any;
    value: number;
    change: number;
    format: (value: number) => string;
    inverse?: boolean;
    loading?: boolean;
  } = $props();

  const isNegative = $derived(inverse ? change > 0 : change < 0);
</script>

<div
  class={cn(
    'flex items-center justify-start gap-3 rounded-md text-left transition-opacity',
    loading && 'opacity-50',
  )}
>
  <div
    class="border-border bg-muted/60 hidden aspect-square h-10 w-10 items-center justify-center rounded-md sm:flex"
  >
    <Icon class="size-4" />
  </div>
  <div class="min-w-0">
    <p class="flex gap-2 truncate pb-0.5 text-sm font-semibold leading-4">
      <Icon class="size-4 sm:hidden" />{title}
    </p>
    <span class="flex items-start gap-1.5">
      {#if loading}
        <Skeleton class="mt-0.5 h-5 w-16" />
      {:else}
        <span class="text-lg font-bold leading-6 tabular-nums sm:text-xl"
          >{format(value)}</span
        >
        <span
          class={cn(
            'flex h-[1.4rem] items-center gap-0.5 text-xs font-semibold',
            isNegative ? 'text-red-600' : 'text-green-600',
          )}
        >
          {#if isNegative}<TrendingDown class="size-3" />{:else}<TrendingUp
              class="size-3"
            />{/if}
          {Math.abs(change).toFixed(1)}%
        </span>
      {/if}
    </span>
  </div>
</div>
