<script lang="ts">
	import { Bar } from '@facile/mise-en-scene';
	import { Icon, icons } from '@facile/muse';
	import { fade } from 'svelte/transition';

	let {
		key,
		title,
		due,
		done = false,
		who,
		widths = [70, 45]
	}: {
		key?: string;
		title?: string;
		due?: string;
		done?: boolean;
		who?: string;
		widths?: [number, number];
	} = $props();
</script>

{#if title}
	<div class="bc-key">{key}</div>
	<p class="bc-title">{title}</p>
{:else}
	<Bar w={38} h={6} />
	<div class="bc-lines">
		<Bar w="{widths[0]}%" />
		<Bar w="{widths[1]}%" />
	</div>
{/if}
<div class="bc-meta">
	{#if done}
		<span class="bc-check" in:fade={{ duration: 250 }}><Icon icon={icons.check} size={14} /></span>
	{:else if due}
		<span class="bc-due"><Icon icon={icons.calendar} size={12} />{due}</span>
	{:else}
		<Bar w={40} h={14} r={4} />
	{/if}
	<span class="bc-avatar" style:background={who}></span>
</div>

<style>
	.bc-key {
		font-family: var(--font-fc-mono);
		font-size: 10px;
		color: var(--color-fc-fg-muted);
	}
	.bc-title {
		margin: 0;
		font-size: 13px;
		font-weight: 700;
		line-height: 1.35;
	}
	.bc-lines {
		display: grid;
		gap: 7px;
	}
	.bc-meta {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 8px;
	}
	.bc-check {
		display: inline-flex;
		color: var(--color-fc-success);
	}
	.bc-due {
		display: inline-flex;
		align-items: center;
		gap: 4px;
		height: 20px;
		padding: 0 6px;
		border-radius: 6px;
		background: var(--color-fc-field);
		font-size: 11px;
		font-weight: 600;
		white-space: nowrap;
		color: var(--color-fc-fg-muted);
	}
	.bc-avatar {
		flex: none;
		width: 20px;
		height: 20px;
		border-radius: 999px;
		background: var(--color-fc-field);
	}
</style>
