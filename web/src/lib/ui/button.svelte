<script lang="ts">
  import type {
    HTMLAnchorAttributes,
    HTMLButtonAttributes,
  } from 'svelte/elements';
  import { cn } from './cn';

  let {
    href,
    variant = 'default',
    size = 'default',
    class: className,
    children,
    ...props
  }: (HTMLButtonAttributes & HTMLAnchorAttributes) & {
    href?: string;
    variant?: 'default' | 'outline' | 'ghost' | 'secondary';
    size?: 'default' | 'sm' | 'icon';
  } = $props();

  const styles = $derived(
    cn(
      'inline-flex shrink-0 items-center justify-center gap-2 rounded-md text-sm font-semibold transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50',
      variant === 'default' &&
        'bg-primary text-primary-foreground hover:bg-primary/90',
      variant === 'outline' && 'border bg-elevated hover:bg-accent',
      variant === 'ghost' && 'hover:bg-accent',
      variant === 'secondary' && 'bg-secondary hover:bg-secondary/75',
      size === 'default' && 'h-9 px-4',
      size === 'sm' && 'h-8 px-3 text-xs',
      size === 'icon' && 'size-9',
      className,
    ),
  );
</script>

{#if href}
  <a {href} class={styles} {...props}>{@render children?.()}</a>
{:else}
  <button class={styles} {...props}>{@render children?.()}</button>
{/if}
