<script lang="ts">
  import { Button } from '$lib/ui/button';
  import { Input } from '$lib/ui/input';
  import { Label } from '$lib/ui/label';
  import * as Select from '$lib/ui/select';
  import { cn } from '$lib/ui/cn';
  import Logo from '$lib/ui/logo.svelte';
  import { CircleGauge, LoaderCircle, Users } from '@lucide/svelte';
  import { onMount } from 'svelte';
  import DashboardPage from './dashboard-page.svelte';
  import Sidebar from './sidebar.svelte';
  import VisitorsPage from './visitors-page.svelte';
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
      user =
        mode === 'setup'
          ? await api.setup(email, password)
          : await api.login(email, password);
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
    selectedWebsite =
      websites.find((website) => website.id === requested) ?? websites[0];
    section = location.pathname.endsWith('/visitors')
      ? 'visitors'
      : 'dashboard';
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
    selectedWebsite =
      websites.find((website) => website.id === id) ?? selectedWebsite;
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

  const loadCurrentSection = () =>
    section === 'dashboard' ? loadDashboard() : loadVisitors();

  const loadDashboard = async () => {
    if (!selectedWebsite) return;
    analyticsLoading = true;
    error = '';
    try {
      const [summary, pages, referrers, countries, devices] = await Promise.all(
        [
          api.dashboard(selectedWebsite.id, period),
          api.breakdown(selectedWebsite.id, period, 'pages'),
          api.breakdown(selectedWebsite.id, period, 'referrers'),
          api.breakdown(selectedWebsite.id, period, 'countries'),
          api.breakdown(selectedWebsite.id, period, 'devices'),
        ],
      );
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
      const response = await api.visitors(
        selectedWebsite.id,
        visitorPage,
        visitorSearch,
      );
      visitorRows = response.visitors;
      visitorTotal = response.total;
    } catch (caught) {
      error = errorMessage(caught);
    } finally {
      visitorsLoading = false;
    }
  };

  const searchVisitors = async (query: string) => {
    visitorSearch = query;
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
</script>

{#if phase === 'booting'}
  <div class="flex min-h-svh items-center justify-center">
    <LoaderCircle class="size-6 animate-spin text-muted-foreground" />
  </div>
{:else if phase === 'setup' || phase === 'login'}
  <main class="flex min-h-svh items-center justify-center p-5">
    <form
      class="bg-elevated w-full max-w-sm rounded-lg border p-6 shadow-sm"
      onsubmit={(event) => {
        event.preventDefault();
        void authenticate(phase as 'setup' | 'login');
      }}
    >
      <div class="mb-5 flex items-center gap-2 text-lg font-bold">
        <Logo class="size-10" /><span>dottie</span>
      </div>
      <h1 class="font-serif text-2xl font-semibold">
        {phase === 'setup' ? 'welcome to dottie.' : 'welcome back.'}
      </h1>
      <p class="mt-1 text-sm text-muted-foreground">
        {phase === 'setup'
          ? 'Create the administrator for this installation.'
          : 'Sign in to your local analytics dashboard.'}
      </p>
      <div class="mt-6">
        <Label class="mb-1.5 font-semibold" for="email">Email</Label><Input
          class="h-10 bg-elevated"
          id="email"
          type="email"
          bind:value={email}
          autocomplete="email"
          required
        />
      </div>
      <div class="mt-4">
        <Label class="mb-1.5 font-semibold" for="password">Password</Label
        ><Input
          class="h-10 bg-elevated"
          id="password"
          type="password"
          bind:value={password}
          minlength={12}
          autocomplete={phase === 'setup' ? 'new-password' : 'current-password'}
          required
        />{#if phase === 'setup'}<p
            class="mt-1.5 text-xs text-muted-foreground"
          >
            Use at least 12 characters.
          </p>{/if}
      </div>
      {#if error}<p class="mt-4 rounded-md bg-red-50 p-3 text-sm text-red-700">
          {error}
        </p>{/if}
      <Button class="mt-5 w-full" type="submit" disabled={busy}
        >{#if busy}<LoaderCircle class="size-4 animate-spin" />{/if}{phase ===
        'setup'
          ? 'Create administrator'
          : 'Sign in'}</Button
      >
    </form>
  </main>
{:else if phase === 'create-website'}
  <main class="flex min-h-svh items-center justify-center p-5">
    <form
      class="bg-elevated w-full max-w-md rounded-lg border p-6 shadow-sm"
      onsubmit={(event) => {
        event.preventDefault();
        void createWebsite();
      }}
    >
      <div class="mb-5 flex items-center gap-2 text-lg font-bold">
        <Logo class="size-10" /><span>dottie</span>
      </div>
      <h1 class="font-serif text-2xl font-semibold">add a website.</h1>
      <p class="mt-1 text-sm text-muted-foreground">
        Dottie uses the domain to accept analytics only from your website.
      </p>
      <div class="mt-6">
        <Label class="mb-1.5 font-semibold" for="website-name"
          >Website name</Label
        ><Input
          class="h-10 bg-elevated"
          id="website-name"
          bind:value={websiteName}
          placeholder="Acme"
          required
        />
      </div>
      <div class="mt-4">
        <Label class="mb-1.5 font-semibold" for="website-domain">Domain</Label
        ><Input
          class="h-10 bg-elevated"
          id="website-domain"
          bind:value={websiteDomain}
          placeholder="example.com"
          required
        />
      </div>
      {#if error}<p class="mt-4 rounded-md bg-red-50 p-3 text-sm text-red-700">
          {error}
        </p>{/if}
      <Button class="mt-5 w-full" type="submit" disabled={busy}
        >{#if busy}<LoaderCircle class="size-4 animate-spin" />{/if}Add website</Button
      >
    </form>
  </main>
{:else if selectedWebsite}
  <div class="flex h-[100dvh] w-full flex-col">
    <div
      id="dashboard-layout"
      class="w-full justify-center overflow-y-scroll pt-2 md:flex md:h-full md:pt-0"
    >
      <Sidebar
        {user}
        {websites}
        website={selectedWebsite}
        {section}
        onWebsiteChange={(id) => void chooseWebsite(id)}
        onNavigate={(next) => void navigate(next)}
        onLogout={() => void logout()}
      />

      <header
        class="bg-background sticky top-0 z-40 flex h-14 items-center gap-3 border-b px-4 md:hidden"
      >
        <Logo class="size-8" /><span class="font-bold">dottie</span><span
          class="text-muted-foreground">/</span
        >
        <Select.Root
          type="single"
          value={selectedWebsite.id}
          onValueChange={(value) => value && void chooseWebsite(value)}
        >
          <Select.Trigger
            class="min-w-0 flex-1 border-0 bg-transparent px-2 text-sm font-semibold shadow-none"
            >{selectedWebsite.name}</Select.Trigger
          >
          <Select.Content>
            {#each websites as website}<Select.Item
                value={website.id}
                label={website.name}>{website.name}</Select.Item
              >{/each}
            <Select.Item value="__add__" label="add website"
              >+ add website</Select.Item
            >
          </Select.Content>
        </Select.Root>
      </header>

      {#if error}<p
          class="fixed left-1/2 top-3 z-[100] -translate-x-1/2 rounded-md border border-red-200 bg-red-50 px-4 py-2 text-sm text-red-700 shadow-sm"
        >
          {error}
        </p>{/if}

      {#if section === 'dashboard'}
        <DashboardPage
          {dashboard}
          {breakdowns}
          loading={analyticsLoading}
          {period}
          {installSnippet}
          {copied}
          onPeriodChange={(next) => void changePeriod(next)}
          onCopy={() => void copySnippet()}
        />
      {:else}
        <VisitorsPage
          visitors={visitorRows}
          total={visitorTotal}
          page={visitorPage}
          search={visitorSearch}
          loading={visitorsLoading}
          onSearch={(query) => void searchVisitors(query)}
          onPageChange={(next) => void changeVisitorPage(next)}
        />
      {/if}

      <nav
        class="bg-elevated fixed inset-x-3 bottom-3 z-50 flex h-12 items-center justify-around rounded-xl border p-1 shadow-lg md:hidden"
      >
        <Button
          variant="ghost"
          class={cn(
            'flex h-10 flex-1 items-center justify-center rounded-lg',
            section === 'dashboard' && 'bg-accent',
          )}
          onclick={() => void navigate('dashboard')}
          aria-label="dashboard"><CircleGauge class="size-4" /></Button
        >
        <Button
          variant="ghost"
          class={cn(
            'flex h-10 flex-1 items-center justify-center rounded-lg',
            section === 'visitors' && 'bg-accent',
          )}
          onclick={() => void navigate('visitors')}
          aria-label="visitors"><Users class="size-4" /></Button
        >
      </nav>
    </div>
  </div>
{/if}
