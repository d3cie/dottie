<script lang="ts">
  import { CircleGauge, Globe, LogOut, Settings, Users } from '@lucide/svelte';
  import type { User, Website } from './api';
  import Logo from '$lib/ui/logo.svelte';
  import { Button } from '$lib/ui/button';
  import * as Avatar from '$lib/ui/avatar';
  import * as Select from '$lib/ui/select';
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
    <div class="relative">
      <Globe
        class="text-muted-foreground pointer-events-none absolute left-2.5 top-2.5 size-4"
      />
      <Select.Root
        type="single"
        value={website.id}
        onValueChange={(value) => value && onWebsiteChange(value)}
      >
        <Select.Trigger
          aria-label="website"
          class="h-9 w-56 border-0 bg-transparent pl-9 text-sm font-semibold shadow-none hover:bg-elevated"
          >{website.name}</Select.Trigger
        >
        <Select.Content>
          {#each websites as item}<Select.Item value={item.id} label={item.name}
              >{item.name}</Select.Item
            >{/each}
          <Select.Item value="__add__" label="add website"
            >+ add website</Select.Item
          >
        </Select.Content>
      </Select.Root>
    </div>
  </div>

  <nav class="flex w-full flex-col items-center gap-0.5">
    {#each navigation as item}
      {@const Icon = item.icon}
      <Button
        variant="ghost"
        class={cn(
          'text-muted-foreground group h-9 w-9 justify-start rounded-md px-0 text-sm font-semibold lg:w-56 lg:px-3',
          section === item.id && 'bg-accent text-foreground',
        )}
        onclick={() => onNavigate(item.id)}
      >
        <Icon class="mx-auto size-4 stroke-[2.2px] opacity-80 lg:mx-0" />
        <span class="ml-2.5 mt-0.5 hidden flex-1 text-left leading-none lg:flex"
          >{item.label}</span
        >
      </Button>
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
    <Button
      variant="ghost"
      class="text-muted-foreground flex h-9 w-9 items-center rounded-md px-0 text-sm font-semibold opacity-50 lg:w-56 lg:px-3"
      disabled
    >
      <Settings class="mx-auto size-4 opacity-80 lg:mx-0" />
      <span class="ml-2.5 hidden lg:flex">settings</span>
    </Button>
    <Button
      variant="ghost"
      class="text-muted-foreground flex h-9 w-9 items-center rounded-md px-0 text-sm font-semibold hover:bg-elevated hover:text-foreground lg:w-56 lg:px-3"
      onclick={onLogout}
    >
      <LogOut class="mx-auto size-4 opacity-80 lg:mx-0" />
      <span class="ml-2.5 hidden lg:flex">sign out</span>
    </Button>
  </nav>

  <div
    class="hidden w-56 items-center gap-2 px-3 text-xs text-muted-foreground lg:flex"
  >
    <Avatar.Root size="sm">
      <Avatar.Fallback class="bg-secondary font-bold text-foreground"
        >{user?.email.slice(0, 1).toUpperCase()}</Avatar.Fallback
      >
    </Avatar.Root>
    <span class="min-w-0 truncate">{user?.email}</span>
  </div>
</aside>
