<script lang="ts">
  import {
    CalendarDays,
    Clock,
    LogOut,
    Monitor,
    Waypoints,
  } from 'lucide-svelte';
  import type { Breakdown, Dashboard } from './api';
  import PageContainer from './page-container.svelte';
  import InstallBanner from './install-banner.svelte';
  import KpiCard from './kpi-card.svelte';
  import ChartCard from './chart-card.svelte';
  import CountsList from './counts-list.svelte';

  let {
    dashboard,
    breakdowns,
    loading,
    period,
    installSnippet,
    copied,
    onPeriodChange,
    onCopy,
  }: {
    dashboard: Dashboard | null;
    breakdowns: Record<string, Breakdown[]>;
    loading: boolean;
    period: string;
    installSnippet: string;
    copied: boolean;
    onPeriodChange: (period: string) => void;
    onCopy: () => void;
  } = $props();

  const duration = (seconds: number) =>
    `${Math.floor(seconds / 60) ? `${Math.floor(seconds / 60)}m ` : ''}${Math.floor(seconds % 60)}s`;
</script>

<PageContainer title="dashboard">
  {#snippet actions()}
    <div class="flex">
      <span
        class="flex size-9 items-center justify-center rounded-l-md border border-r-0 border-dashed border-gray-200"
        ><CalendarDays class="size-4 stroke-[1.6px]" /></span
      >
      <select
        class="bg-elevated h-9 rounded-r-md border px-3 text-sm font-semibold outline-none"
        value={period}
        onchange={(event) => onPeriodChange(event.currentTarget.value)}
        aria-label="time range"
      >
        <option value="7d">last 7 days</option>
        <option value="30d">last 30 days</option>
        <option value="90d">last 90 days</option>
      </select>
    </div>
  {/snippet}

  <main class="relative flex flex-col gap-3 pb-2 pt-3 sm:pt-0">
    {#if dashboard && !dashboard.has_received_event}<InstallBanner
        snippet={installSnippet}
        {copied}
        {onCopy}
      />{/if}

    <div
      class="grid w-full grid-cols-2 justify-between gap-2 gap-y-4 rounded-md px-3 pb-2 sm:grid-cols-3 md:grid-cols-4"
    >
      <KpiCard
        title="visits"
        icon={Monitor}
        value={dashboard?.visits.value ?? 0}
        change={dashboard?.visits.change ?? 0}
        format={(value) => Intl.NumberFormat().format(value)}
        {loading}
      />
      <KpiCard
        title="views per visit"
        icon={Waypoints}
        value={dashboard?.views_per_visit.value ?? 0}
        change={dashboard?.views_per_visit.change ?? 0}
        format={(value) => value.toFixed(2)}
        {loading}
      />
      <KpiCard
        title="bounce rate"
        icon={LogOut}
        value={dashboard?.bounce_rate.value ?? 0}
        change={dashboard?.bounce_rate.change ?? 0}
        format={(value) => `${value.toFixed(1)}%`}
        inverse
        {loading}
      />
      <KpiCard
        title="avg. visit duration"
        icon={Clock}
        value={dashboard?.average_duration.value ?? 0}
        change={dashboard?.average_duration.change ?? 0}
        format={duration}
        {loading}
      />
    </div>

    <div class="grid w-full flex-1 grid-cols-2 gap-3 md:grid-cols-4">
      <ChartCard
        title="visits"
        total={dashboard?.visits.value ?? 0}
        change={dashboard?.visits.change ?? 0}
        data={dashboard?.visit_series ?? []}
        {loading}
      />
      <ChartCard
        title="unique visitors"
        total={dashboard?.unique_visitors.value ?? 0}
        change={dashboard?.unique_visitors.change ?? 0}
        data={dashboard?.visitor_series ?? []}
        online={dashboard?.online_visitors ?? 0}
        {loading}
      />
    </div>

    <div class="grid min-w-0 grid-cols-1 gap-3 sm:grid-cols-2">
      <CountsList title="top pages" items={breakdowns.pages ?? []} {loading} />
      <CountsList
        title="sources"
        items={breakdowns.referrers ?? []}
        {loading}
      />
    </div>
    <div class="grid min-w-0 grid-cols-1 gap-3 sm:grid-cols-2">
      <CountsList
        title="location"
        items={breakdowns.countries ?? []}
        {loading}
      />
      <CountsList title="devices" items={breakdowns.devices ?? []} {loading} />
    </div>
  </main>
</PageContainer>
