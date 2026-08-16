<script lang="ts">
  import { cn, type WithElementRef } from '$lib/ui/cn.js';
  import type {
    HTMLInputAttributes,
    HTMLInputTypeAttribute,
  } from 'svelte/elements';

  type InputType = Exclude<HTMLInputTypeAttribute, 'file'>;

  type Props = WithElementRef<
    Omit<HTMLInputAttributes, 'type'> &
      (
        | { type: 'file'; files?: FileList }
        | { type?: InputType; files?: undefined }
      )
  >;

  let {
    ref = $bindable(null),
    value = $bindable(),
    type,
    files = $bindable(),
    class: className,
    'data-slot': dataSlot = 'input',
    ...restProps
  }: Props = $props();
</script>

{#if type === 'file'}
  <input
    bind:this={ref}
    data-slot={dataSlot}
    class={cn(
      'flex h-10 w-full min-w-0 rounded-md border border-input bg-background px-3 py-2 text-base outline-3 outline-transparent transition-[outline-color] duration-200 file:inline-flex file:h-6 file:border-0 file:bg-transparent file:text-sm file:font-medium file:text-foreground placeholder:text-muted-foreground focus-visible:outline-primary/60 disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50 md:text-sm',
      className,
    )}
    type="file"
    bind:files
    bind:value
    {...restProps}
  />
{:else}
  <input
    bind:this={ref}
    data-slot={dataSlot}
    class={cn(
      'flex h-10 w-full min-w-0 rounded-md border border-input bg-background px-3 py-2 text-base outline-3 outline-transparent transition-[outline-color] duration-200 file:inline-flex file:h-6 file:border-0 file:bg-transparent file:text-sm file:font-medium file:text-foreground placeholder:text-muted-foreground focus-visible:outline-primary/60 disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50 md:text-sm',
      className,
    )}
    {type}
    bind:value
    {...restProps}
  />
{/if}
