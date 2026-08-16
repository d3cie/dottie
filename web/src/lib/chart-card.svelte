<script lang="ts">
  import { BarChart } from 'layerchart';
  import { TrendingDown, TrendingUp } from 'lucide-svelte';
  import type { Point } from './api';

  let {
    title,
    total,
    change,
    data,
    loading = false,
    online = 0,
  }: {
    title: string;
    total: number;
    change: number;
    data: Point[];
    loading?: boolean;
    online?: number;
  } = $props();
  const chartData = $derived(
    data.map((point) => ({
      timestamp: new Date(point.timestamp),
      value: point.value,
    })),
  );
</script>

<section
  class="bg-elevated relative col-span-2 flex min-h-72 w-full flex-col overflow-hidden rounded-md border p-3 shadow-sm"
>
  <header
    class="flex items-start justify-between text-sm font-semibold transition-opacity"
    class:opacity-50={loading}
  >
    <div>
      <p>{title}</p>
      <div class="mt-1 flex items-center gap-2">
        <span class="text-2xl font-bold tabular-nums"
          >{Intl.NumberFormat().format(total)}</span
        >
        <span
          class:!text-red-600={change < 0}
          class="flex items-center gap-0.5 text-xs text-green-600"
        >
          {#if change < 0}<TrendingDown class="size-3.5" />{:else}<TrendingUp
              class="size-3.5"
            />{/if}
          {Math.abs(change).toFixed(1)}%
        </span>
      </div>
      {#if online > 0}
        <div
          class="mt-1.5 flex items-center gap-2 text-xs font-medium text-muted-foreground"
        >
          <span class="relative flex size-2"
            ><span
              class="absolute inline-flex size-full animate-ping rounded-full bg-blue-500 opacity-60"
            ></span><span
              class="relative inline-flex size-2 rounded-full bg-blue-500"
            ></span></span
          >
          {online} online
        </div>
      {/if}
    </div>
  </header>
  <div
    class="mt-2 min-h-0 flex-1 transition-opacity"
    class:opacity-50={loading}
    style="--color-value: #3b82f6"
  >
    {#if chartData.length}
      <BarChart
        data={chartData}
        x="timestamp"
        y="value"
        axis="x"
        grid={{ x: false, y: true }}
        padding={{ top: 8, right: 6, bottom: 20, left: 6 }}
        series={[
          {
            key: 'value',
            label: title,
            value: (row: { value: number }) => row.value,
            color: 'var(--color-value)',
          },
        ]}
        props={{ bars: { radius: 2 } }}
      />
    {:else}
      <div
        class="flex h-full min-h-40 items-center justify-center text-sm text-muted-foreground"
      >
        no data yet.
      </div>
    {/if}
  </div>
</section>
