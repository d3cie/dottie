<script lang="ts">
  import {
    Dot,
    Laptop,
    ListFilter,
    Monitor,
    Search,
    Smartphone,
  } from '@lucide/svelte';
  import { onMount } from 'svelte';
  import type { Visitor } from './api';
  import PageContainer from './page-container.svelte';
  import { Button } from '$lib/ui/button';
  import { Input } from '$lib/ui/input';
  import * as Avatar from '$lib/ui/avatar';
  import { Kbd } from '$lib/ui/kbd';
  import * as Table from '$lib/ui/table';
  import { cn } from '$lib/ui/cn';

  let {
    visitors,
    total,
    page,
    search,
    loading,
    onSearch,
    onPageChange,
  }: {
    visitors: Visitor[];
    total: number;
    page: number;
    search: string;
    loading: boolean;
    onSearch: (query: string) => void;
    onPageChange: (page: number) => void;
  } = $props();

  let query = $state('');
  const pages = $derived(Math.max(1, Math.ceil(total / 15)));

  onMount(() => {
    query = search;
    const shortcut = (event: KeyboardEvent) => {
      if (event.target instanceof HTMLInputElement) return;
      if (event.key.toLowerCase() === 'z' && page > 1) onPageChange(page - 1);
      if (event.key.toLowerCase() === 'x' && page < pages)
        onPageChange(page + 1);
    };
    document.addEventListener('keydown', shortcut);
    return () => document.removeEventListener('keydown', shortcut);
  });

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

  const deviceIcon = (device: string) =>
    device === 'mobile' ? Smartphone : device === 'desktop' ? Monitor : Laptop;
</script>

<PageContainer title="visitors">
  {#snippet actions()}
    <form
      class="flex gap-2"
      onsubmit={(event) => {
        event.preventDefault();
        onSearch(query);
      }}
    >
      <div class="relative hidden sm:block">
        <Search
          class="pointer-events-none absolute left-3 top-2.5 size-4 text-muted-foreground"
        />
        <Input
          class="h-9 w-56 bg-elevated pl-9"
          bind:value={query}
          placeholder="search visitors"
        />
      </div>
      <Button
        type="submit"
        size="sm"
        variant="outline"
        class="group h-9 gap-2 border-dashed border-gray-200 px-3"
      >
        <ListFilter
          class="size-4 stroke-[1.6px] opacity-60 group-hover:opacity-100"
        /><span>filter</span>
      </Button>
    </form>
  {/snippet}

  <main
    class="relative flex h-full min-h-0 w-full flex-col items-center gap-3 overflow-hidden pb-2"
  >
    <div
      class="bg-elevated relative flex min-h-40 w-full flex-1 flex-col overflow-auto rounded-lg border text-center shadow-sm"
    >
      <Table.Root
        class={cn(
          'w-full min-w-[900px] text-sm transition-opacity duration-300',
          loading && 'opacity-50',
        )}
      >
        <Table.Header>
          <Table.Row class="border-b hover:bg-elevated">
            <Table.Head class="h-10 w-14 px-4"></Table.Head>
            <Table.Head
              class="h-10 whitespace-nowrap px-4 text-left text-xs font-semibold uppercase text-muted-foreground"
              >name</Table.Head
            >
            <Table.Head
              class="h-10 whitespace-nowrap px-4 text-center text-xs font-semibold uppercase text-muted-foreground"
              >country</Table.Head
            >
            <Table.Head
              class="h-10 whitespace-nowrap px-4 text-left text-xs font-semibold uppercase text-muted-foreground"
              >referrer</Table.Head
            >
            <Table.Head
              class="h-10 whitespace-nowrap px-4 text-left text-xs font-semibold uppercase text-muted-foreground"
              >first seen</Table.Head
            >
            <Table.Head
              class="h-10 whitespace-nowrap px-4 text-left text-xs font-semibold uppercase text-muted-foreground"
              >last activity</Table.Head
            >
            <Table.Head
              class="h-10 whitespace-nowrap px-4 text-left text-xs font-semibold uppercase text-muted-foreground"
              >device info</Table.Head
            >
          </Table.Row>
        </Table.Header>
        <Table.Body>
          {#each visitors as visitor (visitor.id)}
            {@const DeviceIcon = deviceIcon(visitor.device)}
            <Table.Row class="h-14 border-b last:border-0 hover:bg-muted/50">
              <Table.Cell class="px-4 pr-0">
                <Avatar.Root>
                  <Avatar.Fallback
                    class="bg-secondary text-xs font-bold text-foreground"
                    >{visitor.name.slice(-2).toUpperCase()}</Avatar.Fallback
                  >
                  {#if new Date(visitor.last_active_at).getTime() > Date.now() - 300000}<span
                      class="absolute bottom-0 right-0 size-2.5 rounded-full border-2 border-elevated bg-green-500"
                    ></span>{/if}
                </Avatar.Root>
              </Table.Cell>
              <Table.Cell
                class="w-fit whitespace-nowrap px-4 text-left font-semibold"
                ><span
                  class="text-primary underline decoration-primary/40 underline-offset-4"
                  >{visitor.name}</span
                ></Table.Cell
              >
              <Table.Cell class="px-4 text-center font-medium"
                >{visitor.country || 'unknown'}{#if visitor.city}<span
                    class="block text-xs text-muted-foreground"
                    >{visitor.city}</span
                  >{/if}</Table.Cell
              >
              <Table.Cell class="max-w-48 truncate px-4 text-left font-medium"
                >{visitor.referrer || 'direct/unknown'}</Table.Cell
              >
              <Table.Cell class="whitespace-nowrap px-4 text-left font-medium"
                >{new Intl.DateTimeFormat('en-GB', {
                  day: '2-digit',
                  month: 'short',
                  year: '2-digit',
                }).format(new Date(visitor.first_seen_at))}</Table.Cell
              >
              <Table.Cell class="whitespace-nowrap px-4 text-left font-medium"
                >{relativeTime(visitor.last_active_at)}</Table.Cell
              >
              <Table.Cell class="whitespace-nowrap px-4 text-left font-medium">
                <span class="flex items-center text-muted-foreground"
                  ><DeviceIcon class="mr-1 size-4 opacity-80" /><Dot /><span
                    class="text-foreground">{visitor.os || 'other'}</span
                  ><Dot /><span class="text-foreground"
                    >{visitor.browser || 'other'}</span
                  ></span
                >
              </Table.Cell>
            </Table.Row>
          {:else}
            <Table.Row
              ><Table.Cell colspan={7} class="h-64 text-center text-sm"
                >{loading
                  ? 'loading visitors…'
                  : 'no visitors found'}</Table.Cell
              ></Table.Row
            >
          {/each}
        </Table.Body>
      </Table.Root>
    </div>

    <div class="flex w-full justify-between px-2">
      <div class="flex w-full items-center gap-1 text-sm font-medium">
        page <b>{page}</b> of <b>{total ? pages : '-'}</b>
      </div>
      <div class="flex w-full justify-end gap-2">
        <Button
          size="sm"
          variant="outline"
          class="gap-2"
          disabled={page === 1 || loading}
          onclick={() => onPageChange(page - 1)}
          >Previous <Kbd class="hidden bg-elevated sm:flex">z</Kbd></Button
        >
        <Button
          size="sm"
          variant="outline"
          class="gap-2"
          disabled={page >= pages || loading}
          onclick={() => onPageChange(page + 1)}
          >Next <Kbd class="hidden bg-elevated sm:flex">x</Kbd></Button
        >
      </div>
    </div>
  </main>
</PageContainer>
