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
  }: {
    title: string;
    total: number;
    change: number;
    data: Point[];
    loading?: boolean;
  } = $props();

  const chartData = $derived(
    data.map((point) => ({ timestamp: new Date(point.timestamp), value: point.value })),
  );
</script>

<section class="card relative col-span-1 min-h-72 overflow-hidden p-3 md:col-span-2">
  <div class="flex items-start justify-between">
    <div>
      <p class="text-sm font-semibold">{title}</p>
      <p class="mt-1 text-2xl font-bold tabular-nums">{Intl.NumberFormat().format(total)}</p>
    </div>
    <span class:!text-red-600={change < 0} class="flex items-center gap-1 text-xs font-semibold text-green-600">
      {#if change < 0}<TrendingDown class="size-3.5" />{:else}<TrendingUp class="size-3.5" />{/if}
      {Math.abs(change).toFixed(1)}%
    </span>
  </div>
  <div class="mt-3 h-48 w-full transition-opacity" class:opacity-40={loading} style="--color-value: var(--chart)">
    {#if chartData.length}
      <BarChart
        data={chartData}
        x="timestamp"
        y="value"
        axis="x"
        grid={{ x: false, y: true }}
        padding={{ top: 8, right: 8, bottom: 20, left: 8 }}
        series={[{ key: 'value', label: title, value: (row: { value: number }) => row.value, color: 'var(--color-value)' }]}
        props={{ bars: { radius: 2 } }}
      />
    {:else}
      <div class="flex h-full items-center justify-center text-sm text-muted-foreground">no data yet.</div>
    {/if}
  </div>
</section>

