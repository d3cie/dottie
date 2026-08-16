<script lang="ts">
  import { CircleGauge, Globe, LogOut, Settings, Users } from 'lucide-svelte';
  import type { User, Website } from './api';
  import Logo from '$lib/ui/logo.svelte';
  import { cn } from '$lib/ui/cn';

  let {
    user,
    websites,
    website,
    section,
    onWebsiteChange,
    onNavigate,
    onLogout,
  }: {
    user: User | null;
    websites: Website[];
    website: Website;
    section: 'dashboard' | 'visitors';
    onWebsiteChange: (id: string) => void;
    onNavigate: (section: 'dashboard' | 'visitors') => void;
    onLogout: () => void;
  } = $props();

  const navigation = [
    { id: 'dashboard' as const, label: 'dashboard', icon: CircleGauge },
    { id: 'visitors' as const, label: 'visitors', icon: Users },
  ];
</script>

<aside
  class="text-foreground sticky left-0 top-0 z-50 hidden h-full w-fit flex-col items-center gap-6 p-1.5 transition-[width] duration-300 md:flex lg:p-2 lg:py-6"
>
  <div class="hidden w-full items-center gap-2 text-lg font-bold md:flex">
    <Logo class="h-9 w-9 shrink-0" />
    <span class="hidden lg:block">dottie</span>
  </div>

  <div class="hidden w-full pb-0.5 lg:block">
    <label class="sr-only" for="website-switcher">website</label>
    <div class="relative">
      <Globe
        class="text-muted-foreground pointer-events-none absolute left-2.5 top-2.5 size-4"
      />
      <select
        id="website-switcher"
        class="h-9 w-56 appearance-none rounded-md border-0 bg-transparent pl-9 pr-8 text-sm font-semibold outline-none hover:bg-elevated"
        value={website.id}
        onchange={(event) => onWebsiteChange(event.currentTarget.value)}
      >
        {#each websites as item}<option value={item.id}>{item.name}</option
          >{/each}
        <option value="__add__">+ add website</option>
      </select>
    </div>
  </div>

  <nav class="flex w-full flex-col items-center gap-0.5">
    {#each navigation as item}
      {@const Icon = item.icon}
      <button
        class={cn(
          'text-muted-foreground group flex h-9 w-9 items-center rounded-md px-0 text-sm font-semibold transition-colors lg:w-56 lg:px-3',
          section === item.id
            ? 'bg-accent text-foreground'
            : 'hover:bg-elevated hover:text-foreground',
        )}
        onclick={() => onNavigate(item.id)}
      >
        <Icon class="mx-auto size-4 stroke-[2.2px] opacity-80 lg:mx-0" />
        <span class="ml-2.5 mt-0.5 hidden flex-1 text-left leading-none lg:flex"
          >{item.label}</span
        >
      </button>
    {/each}
  </nav>

  <div class="flex-1"></div>

  <div
    class="hidden w-56 rounded-md border p-3 text-sm text-muted-foreground lg:block"
  >
    analytics stays on <span class="font-semibold text-foreground"
      >your server.</span
    >
  </div>

  <nav class="flex w-full flex-col items-center gap-0.5">
    <button
      class="text-muted-foreground flex h-9 w-9 items-center rounded-md px-0 text-sm font-semibold opacity-50 lg:w-56 lg:px-3"
      disabled
    >
      <Settings class="mx-auto size-4 opacity-80 lg:mx-0" />
      <span class="ml-2.5 hidden lg:flex">settings</span>
    </button>
    <button
      class="text-muted-foreground flex h-9 w-9 items-center rounded-md px-0 text-sm font-semibold hover:bg-elevated hover:text-foreground lg:w-56 lg:px-3"
      onclick={onLogout}
    >
      <LogOut class="mx-auto size-4 opacity-80 lg:mx-0" />
      <span class="ml-2.5 hidden lg:flex">sign out</span>
    </button>
  </nav>

  <div
    class="hidden w-56 items-center gap-2 px-3 text-xs text-muted-foreground lg:flex"
  >
    <span
      class="flex size-7 items-center justify-center rounded-full bg-secondary font-bold text-foreground"
      >{user?.email.slice(0, 1).toUpperCase()}</span
    >
    <span class="min-w-0 truncate">{user?.email}</span>
  </div>
</aside>
