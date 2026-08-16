<script lang="ts">
  import { cn, type WithElementRef } from '$lib/ui/cn.js';
  import type { HTMLAttributes } from 'svelte/elements';

  let {
    ref = $bindable(null),
    class: className,
    children,
    size = 'default',
    ...restProps
  }: WithElementRef<HTMLAttributes<HTMLDivElement>> & {
    size?: 'default' | 'sm';
  } = $props();
</script>

<div
  bind:this={ref}
  data-slot="card"
  data-size={size}
  class={cn(
    'flex flex-col rounded-lg border bg-card text-sm text-card-foreground shadow-sm',
    className,
  )}
  {...restProps}
>
  {@render children?.()}
</div>
