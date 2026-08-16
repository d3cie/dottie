<script lang="ts">
  import { Button, Logo, cn } from '@dottie/ui';
  import {
    ArrowLeft,
    ArrowRight,
    Check,
    CircleGauge,
    Clipboard,
    Filter,
    Globe,
    LoaderCircle,
    LogOut,
    Monitor,
    Search,
    Settings,
    Smartphone,
    Users,
  } from 'lucide-svelte';
  import { onMount } from 'svelte';
  import AnalyticsChart from './analytics-chart.svelte';
  import CountsCard from './counts-card.svelte';
  import {
    api,
    errorMessage,
    type Breakdown,
    type Dashboard,
    type User,
    type Visitor,
    type Website,
  } from './api';

  type Phase = 'booting' | 'setup' | 'login' | 'create-website' | 'ready';
  type Section = 'dashboard' | 'visitors';

  let phase: Phase = $state('booting');
  let user: User | null = $state(null);
  let websites: Website[] = $state([]);
  let selectedWebsite: Website | null = $state(null);
  let section: Section = $state('dashboard');
  let error = $state('');
  let busy = $state(false);
  let email = $state('');
  let password = $state('');
  let websiteName = $state('');
  let websiteDomain = $state('');
  let period = $state('30d');
  let dashboard: Dashboard | null = $state(null);
  let breakdowns: Record<string, Breakdown[]> = $state({});
  let analyticsLoading = $state(false);
  let visitorRows: Visitor[] = $state([]);
  let visitorTotal = $state(0);
  let visitorPage = $state(1);
  let visitorSearch = $state('');
  let visitorsLoading = $state(false);
  let copied = $state(false);

  const installSnippet = $derived.by(() => {
    const website = selectedWebsite;
    if (!website) return '';
    return `<script defer src="${window.location.origin}/tracker.js" data-website-id="${website.id}"><\/script>`;
  });

  onMount(async () => {
    try {
      const bootstrap = await api.bootstrap();
      if (bootstrap.setup_required) {
        phase = 'setup';
        return;
      }
      try {
        user = await api.me();
        await loadWebsites();
      } catch {
        phase = 'login';
      }
    } catch (caught) {
      error = errorMessage(caught);
      phase = 'login';
    }
  });

  const authenticate = async (mode: 'setup' | 'login') => {
    busy = true;
    error = '';
    try {
      user = mode === 'setup' ? await api.setup(email, password) : await api.login(email, password);
      password = '';
      await loadWebsites();
    } catch (caught) {
      error = errorMessage(caught);
    } finally {
      busy = false;
    }
  };

  const loadWebsites = async () => {
    const response = await api.websites();
    websites = response.websites;
    if (!websites.length) {
      phase = 'create-website';
      return;
    }
    const requested = location.pathname.match(/\/w\/([^/]+)/)?.[1];
    selectedWebsite = websites.find((website) => website.id === requested) ?? websites[0];
    section = location.pathname.endsWith('/visitors') ? 'visitors' : 'dashboard';
    phase = 'ready';
    updatePath(false);
    await loadCurrentSection();
  };

  const createWebsite = async () => {
    busy = true;
    error = '';
    try {
      const website = await api.createWebsite(
        websiteName,
        websiteDomain,
        Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC',
      );
      websites = [...websites, website];
      selectedWebsite = website;
      phase = 'ready';
      updatePath(true);
      await loadDashboard();
    } catch (caught) {
      error = errorMessage(caught);
    } finally {
      busy = false;
    }
  };

  const chooseWebsite = async (id: string) => {
    if (id === '__add__') {
      websiteName = '';
      websiteDomain = '';
      phase = 'create-website';
      return;
    }
    selectedWebsite = websites.find((website) => website.id === id) ?? selectedWebsite;
    visitorPage = 1;
    updatePath(true);
    await loadCurrentSection();
  };

  const navigate = async (next: Section) => {
    section = next;
    updatePath(true);
    await loadCurrentSection();
  };

  const updatePath = (push: boolean) => {
    if (!selectedWebsite) return;
    const next = `/w/${selectedWebsite.id}/${section}`;
    if (push) history.pushState({}, '', next);
    else history.replaceState({}, '', next);
  };

  const loadCurrentSection = () => (section === 'dashboard' ? loadDashboard() : loadVisitors());

  const loadDashboard = async () => {
    if (!selectedWebsite) return;
    analyticsLoading = true;
    error = '';
    try {
      const [summary, pages, referrers, countries, devices] = await Promise.all([
        api.dashboard(selectedWebsite.id, period),
        api.breakdown(selectedWebsite.id, period, 'pages'),
        api.breakdown(selectedWebsite.id, period, 'referrers'),
        api.breakdown(selectedWebsite.id, period, 'countries'),
        api.breakdown(selectedWebsite.id, period, 'devices'),
      ]);
      dashboard = summary;
      breakdowns = {
        pages: pages.items,
        referrers: referrers.items,
        countries: countries.items,
        devices: devices.items,
      };
    } catch (caught) {
      error = errorMessage(caught);
    } finally {
      analyticsLoading = false;
    }
  };

  const changePeriod = async (next: string) => {
    period = next;
    const query = new URLSearchParams(location.search);
    query.set('period', period);
    history.replaceState({}, '', `${location.pathname}?${query}`);
    await loadDashboard();
  };

  const loadVisitors = async () => {
    if (!selectedWebsite) return;
    visitorsLoading = true;
    error = '';
    try {
      const response = await api.visitors(selectedWebsite.id, visitorPage, visitorSearch);
      visitorRows = response.visitors;
      visitorTotal = response.total;
    } catch (caught) {
      error = errorMessage(caught);
    } finally {
      visitorsLoading = false;
    }
  };

  const searchVisitors = async () => {
    visitorPage = 1;
    await loadVisitors();
  };

  const changeVisitorPage = async (next: number) => {
    visitorPage = next;
    await loadVisitors();
  };

  const copySnippet = async () => {
    await navigator.clipboard.writeText(installSnippet);
    copied = true;
    setTimeout(() => (copied = false), 1500);
  };

  const logout = async () => {
    await api.logout();
    user = null;
    websites = [];
    selectedWebsite = null;
    phase = 'login';
    history.replaceState({}, '', '/');
  };

  const relativeTime = (value: string) => {
    const seconds = Math.round((new Date(value).getTime() - Date.now()) / 1000);
    const formatter = new Intl.RelativeTimeFormat('en', { numeric: 'auto' });
    if (Math.abs(seconds) < 60) return formatter.format(seconds, 'second');
    const minutes = Math.round(seconds / 60);
    if (Math.abs(minutes) < 60) return formatter.format(minutes, 'minute');
    const hours = Math.round(minutes / 60);
    if (Math.abs(hours) < 24) return formatter.format(hours, 'hour');
    return formatter.format(Math.round(hours / 24), 'day');
  };

  const formatDuration = (seconds: number) => {
    const minutes = Math.floor(seconds / 60);
    return `${minutes ? `${minutes}m ` : ''}${Math.floor(seconds % 60)}s`;
  };
</script>

{#if phase === 'booting'}
  <div class="flex min-h-svh items-center justify-center"><LoaderCircle class="size-6 animate-spin text-muted-foreground" /></div>
{:else if phase === 'setup' || phase === 'login'}
  <main class="flex min-h-svh items-center justify-center p-5">
    <form class="card w-full max-w-sm p-6" onsubmit={(event) => { event.preventDefault(); void authenticate(phase as 'setup' | 'login'); }}>
      <Logo class="mb-5 size-10 text-foreground" />
      <h1 class="font-serif text-2xl font-semibold">{phase === 'setup' ? 'welcome to dottie.' : 'welcome back.'}</h1>
      <p class="mt-1 text-sm text-muted-foreground">{phase === 'setup' ? 'Create the administrator for this installation.' : 'Sign in to your local analytics dashboard.'}</p>
      <div class="mt-6">
        <label class="label" for="email">Email</label>
        <input class="input" id="email" type="email" bind:value={email} autocomplete="email" required />
      </div>
      <div class="mt-4">
        <label class="label" for="password">Password</label>
        <input class="input" id="password" type="password" bind:value={password} minlength="12" autocomplete={phase === 'setup' ? 'new-password' : 'current-password'} required />
        {#if phase === 'setup'}<p class="mt-1.5 text-xs text-muted-foreground">Use at least 12 characters.</p>{/if}
      </div>
      {#if error}<p class="mt-4 rounded-md bg-red-50 p-3 text-sm text-red-700">{error}</p>{/if}
      <Button class="mt-5 w-full" type="submit" disabled={busy}>
        {#if busy}<LoaderCircle class="size-4 animate-spin" />{/if}
        {phase === 'setup' ? 'Create administrator' : 'Sign in'}
      </Button>
    </form>
  </main>
{:else if phase === 'create-website'}
  <main class="flex min-h-svh items-center justify-center p-5">
    <form class="card w-full max-w-md p-6" onsubmit={(event) => { event.preventDefault(); void createWebsite(); }}>
      <div class="mb-5 flex items-center gap-3"><Logo class="size-9" /><span class="text-lg font-bold">dottie</span></div>
      <h1 class="font-serif text-2xl font-semibold">add a website.</h1>
      <p class="mt-1 text-sm text-muted-foreground">Dottie uses the domain to accept analytics only from your website.</p>
      <div class="mt-6"><label class="label" for="website-name">Website name</label><input class="input" id="website-name" bind:value={websiteName} placeholder="Acme" required /></div>
      <div class="mt-4"><label class="label" for="website-domain">Domain</label><input class="input" id="website-domain" bind:value={websiteDomain} placeholder="example.com" required /></div>
      {#if error}<p class="mt-4 rounded-md bg-red-50 p-3 text-sm text-red-700">{error}</p>{/if}
      <Button class="mt-5 w-full" type="submit" disabled={busy}>{#if busy}<LoaderCircle class="size-4 animate-spin" />{/if}Add website</Button>
    </form>
  </main>
{:else if selectedWebsite}
  <div class="flex min-h-svh w-full">
    <aside class="sticky top-0 hidden h-svh w-60 shrink-0 flex-col gap-6 p-5 md:flex">
      <div class="flex items-center gap-2 text-lg font-bold"><Logo class="size-9" /><span>dottie</span></div>
      <select class="input border-none bg-card font-semibold" value={selectedWebsite.id} onchange={(event) => void chooseWebsite(event.currentTarget.value)} aria-label="Website">
        {#each websites as website}<option value={website.id}>{website.name}</option>{/each}
        <option value="__add__">+ add website</option>
      </select>
      <nav class="space-y-1">
        <button class={cn('flex h-9 w-full items-center gap-3 rounded-md px-3 text-sm font-semibold text-muted-foreground hover:bg-card hover:text-foreground', section === 'dashboard' && 'bg-card text-foreground shadow-sm')} onclick={() => void navigate('dashboard')}><CircleGauge class="size-4" /> dashboard</button>
        <button class={cn('flex h-9 w-full items-center gap-3 rounded-md px-3 text-sm font-semibold text-muted-foreground hover:bg-card hover:text-foreground', section === 'visitors' && 'bg-card text-foreground shadow-sm')} onclick={() => void navigate('visitors')}><Users class="size-4" /> visitors</button>
      </nav>
      <div class="flex-1"></div>
      <nav class="space-y-1">
        <button class="flex h-9 w-full items-center gap-3 rounded-md px-3 text-sm font-semibold text-muted-foreground opacity-60" disabled><Settings class="size-4" /> settings</button>
        <button class="flex h-9 w-full items-center gap-3 rounded-md px-3 text-sm font-semibold text-muted-foreground hover:bg-card hover:text-foreground" onclick={() => void logout()}><LogOut class="size-4" /> sign out</button>
      </nav>
      <p class="truncate px-3 text-xs text-muted-foreground" title={user?.email}>{user?.email}</p>
    </aside>

    <div class="min-w-0 flex-1">
      <header class="sticky top-0 z-40 flex h-14 items-center gap-3 border-b bg-background/90 px-4 backdrop-blur md:hidden">
        <Logo class="size-8" /><span class="font-bold">dottie</span><span class="text-muted-foreground">/</span>
        <select class="min-w-0 flex-1 bg-transparent text-sm font-semibold" value={selectedWebsite.id} onchange={(event) => void chooseWebsite(event.currentTarget.value)}>{#each websites as website}<option value={website.id}>{website.name}</option>{/each}<option value="__add__">+ add website</option></select>
        <Button variant="ghost" size="icon" onclick={() => void navigate(section === 'dashboard' ? 'visitors' : 'dashboard')}>{#if section === 'dashboard'}<Users />{:else}<CircleGauge />{/if}</Button>
      </header>

      <main class="mx-auto w-full max-w-6xl p-4 pb-10 lg:p-6">
        <div class="mb-5 flex min-h-9 items-center justify-between gap-4">
          <div><h1 class="text-lg font-bold">{section}</h1><p class="text-xs text-muted-foreground">{selectedWebsite.domain}</p></div>
          {#if section === 'dashboard'}
            <select class="input w-auto bg-card" value={period} onchange={(event) => void changePeriod(event.currentTarget.value)} aria-label="Time range"><option value="7d">last 7 days</option><option value="30d">last 30 days</option><option value="90d">last 90 days</option></select>
          {/if}
        </div>

        {#if error}<p class="mb-4 rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700">{error}</p>{/if}

        {#if section === 'dashboard'}
          {#if dashboard && !dashboard.has_received_event}
            <section class="card mb-3 p-4">
              <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between"><div><h2 class="font-semibold">install dottie on your website.</h2><p class="text-sm text-muted-foreground">Paste this before the closing <code>&lt;/head&gt;</code> tag.</p></div><Button variant="outline" onclick={() => void copySnippet()}>{#if copied}<Check class="size-4" /> copied{:else}<Clipboard class="size-4" /> copy script{/if}</Button></div>
              <code class="mt-3 block overflow-x-auto rounded-md bg-foreground p-3 text-xs text-background">{installSnippet}</code>
            </section>
          {/if}
          <div class="grid grid-cols-2 gap-4 rounded-md px-2 py-3 md:grid-cols-4">
            {#each [
              { title: 'visits', value: dashboard?.visits.value ?? 0, change: dashboard?.visits.change ?? 0, icon: Monitor, format: (v: number) => Intl.NumberFormat().format(v) },
              { title: 'views per visit', value: dashboard?.views_per_visit.value ?? 0, change: dashboard?.views_per_visit.change ?? 0, icon: Globe, format: (v: number) => v.toFixed(2) },
              { title: 'bounce rate', value: dashboard?.bounce_rate.value ?? 0, change: -(dashboard?.bounce_rate.change ?? 0), icon: LogOut, format: (v: number) => `${v.toFixed(1)}%` },
              { title: 'avg. visit duration', value: dashboard?.average_duration.value ?? 0, change: dashboard?.average_duration.change ?? 0, icon: CircleGauge, format: formatDuration },
            ] as kpi}
              {@const Icon = kpi.icon}
              <div class="flex items-center gap-3 transition-opacity" class:opacity-40={analyticsLoading}>
                <div class="hidden size-10 items-center justify-center rounded-md border bg-muted sm:flex"><Icon class="size-4" /></div>
                <div class="min-w-0"><p class="truncate text-sm font-semibold">{kpi.title}</p><p class="text-xl font-bold tabular-nums">{kpi.format(kpi.value)}</p><p class={cn('text-xs font-semibold', kpi.change < 0 ? 'text-red-600' : 'text-green-600')}>{kpi.change >= 0 ? '+' : ''}{kpi.change.toFixed(1)}%</p></div>
              </div>
            {/each}
          </div>
          <div class="mt-3 grid grid-cols-1 gap-3 md:grid-cols-4">
            <AnalyticsChart title="visits" total={dashboard?.visits.value ?? 0} change={dashboard?.visits.change ?? 0} data={dashboard?.visit_series ?? []} loading={analyticsLoading} />
            <AnalyticsChart title="unique visitors" total={dashboard?.unique_visitors.value ?? 0} change={dashboard?.unique_visitors.change ?? 0} data={dashboard?.visitor_series ?? []} loading={analyticsLoading} />
          </div>
          <div class="mt-3 grid grid-cols-1 gap-3 sm:grid-cols-2"><CountsCard title="top pages" items={breakdowns.pages ?? []} loading={analyticsLoading} /><CountsCard title="sources" items={breakdowns.referrers ?? []} loading={analyticsLoading} /><CountsCard title="countries" items={breakdowns.countries ?? []} loading={analyticsLoading} /><CountsCard title="devices" items={breakdowns.devices ?? []} loading={analyticsLoading} /></div>
        {:else}
          <form class="mb-3 flex justify-end gap-2" onsubmit={(event) => { event.preventDefault(); void searchVisitors(); }}><div class="relative w-full max-w-xs"><Search class="absolute left-3 top-2.5 size-4 text-muted-foreground" /><input class="input pl-9" bind:value={visitorSearch} placeholder="search visitors" /></div><Button type="submit" variant="outline"><Filter class="size-4" /> filter</Button></form>
          <section class="card overflow-hidden">
            <div class="overflow-x-auto transition-opacity" class:opacity-40={visitorsLoading}>
              <table class="w-full min-w-[850px] text-left text-sm">
                <thead><tr class="border-b text-xs font-bold uppercase text-muted-foreground"><th class="h-10 px-4">name</th><th class="px-4 text-center">country</th><th class="px-4">referrer</th><th class="px-4">first seen</th><th class="px-4">last activity</th><th class="px-4">device info</th><th class="px-4 text-right">views</th></tr></thead>
                <tbody>
                  {#each visitorRows as visitor (visitor.id)}
                    <tr class="h-14 border-b last:border-0 hover:bg-muted/50"><td class="px-4 font-semibold"><div class="flex items-center gap-3"><span class="flex size-8 items-center justify-center rounded-full bg-secondary text-xs">{visitor.name.slice(-2)}</span>{visitor.name}</div></td><td class="px-4 text-center">{visitor.country}</td><td class="max-w-48 truncate px-4">{visitor.referrer}</td><td class="px-4">{new Intl.DateTimeFormat('en-GB', { day: '2-digit', month: 'short', year: '2-digit' }).format(new Date(visitor.first_seen_at))}</td><td class="px-4">{relativeTime(visitor.last_active_at)}</td><td class="px-4"><span class="flex items-center gap-2">{#if visitor.device === 'mobile'}<Smartphone class="size-4" />{:else}<Monitor class="size-4" />{/if}{visitor.browser} · {visitor.os}</span></td><td class="px-4 text-right tabular-nums">{visitor.views}</td></tr>
                  {:else}
                    <tr><td colspan="7" class="h-64 text-center text-muted-foreground">{visitorsLoading ? 'loading visitors…' : 'no visitors found.'}</td></tr>
                  {/each}
                </tbody>
              </table>
            </div>
          </section>
          <div class="mt-3 flex items-center justify-between px-2 text-sm"><span>page <b>{visitorPage}</b> of <b>{Math.max(1, Math.ceil(visitorTotal / 15))}</b></span><div class="flex gap-2"><Button size="sm" variant="outline" disabled={visitorPage === 1 || visitorsLoading} onclick={() => void changeVisitorPage(visitorPage - 1)}><ArrowLeft class="size-4" /> previous</Button><Button size="sm" variant="outline" disabled={visitorPage * 15 >= visitorTotal || visitorsLoading} onclick={() => void changeVisitorPage(visitorPage + 1)}>next <ArrowRight class="size-4" /></Button></div></div>
        {/if}
      </main>
    </div>
  </div>
{/if}
