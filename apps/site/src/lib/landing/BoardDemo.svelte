<script lang="ts">
	import { type Act, Bar, Cursor, cursorAt, grow, Roll, Stage, Window } from '@facile/mise-en-scene';
	import { Icon, icons } from '@facile/muse';
	import { tick } from 'svelte';
	import { fade } from 'svelte/transition';
	import { t } from '$lib/i18n';

	type Phase = 'idle' | 'lifted' | 'dragging' | 'done';

	const KEY = 'WEB-14';

	let phase = $state<Phase>('idle');
	let offset = $state({ x: 0, y: 0 });
	let cursor = $state(cursorAt());
	let card = $state<HTMLElement>();
	let slot = $state<HTMLElement>();
	let slotHeight = $state(98);

	const moved = $derived(phase === 'done');
	const lifted = $derived(phase === 'lifted' || phase === 'dragging');

	function reset() {
		phase = 'idle';
		offset = { x: 0, y: 0 };
		cursor.visible = false;
		cursor.down = false;
	}

	function poster() {
		phase = 'done';
	}

	async function script(act: Act) {
		await act.wait(900);
		const grip = act.point(card, 0.35, 0.45);
		await act.move(cursor, { x: grip.x + 70, y: grip.y + 90 }, 400);
		cursor.visible = true;
		await act.move(cursor, grip);
		cursor.down = true;
		slotHeight = card?.offsetHeight ?? slotHeight;
		phase = 'lifted';
		await tick();
		await act.wait(250);
		const from = act.point(card, 0, 0);
		const to = act.point(slot, 0, 0);
		offset = { x: to.x - from.x, y: to.y - from.y };
		phase = 'dragging';
		await act.move(cursor, { x: grip.x + offset.x, y: grip.y + offset.y });
		cursor.down = false;
		phase = 'done';
		await act.wait(300);
		await act.move(cursor, { x: grip.x + offset.x + 60, y: grip.y + offset.y + 110 }, 600);
		cursor.visible = false;
		await act.wait(2000);
		reset();
		await act.wait(700);
	}
</script>

{#snippet focal(done: boolean)}
	<div class="bd-key">{KEY}</div>
	<p class="bd-title">{t('demo.task')}</p>
	<div class="bd-meta">
		{#if done}
			<span class="bd-check" in:fade={{ duration: 250 }}><Icon icon={icons.check} size={14} /></span>
		{:else}
			<Bar w={46} h={14} r={4} />
		{/if}
		<span class="bd-avatar bd-avatar-me"></span>
	</div>
{/snippet}

{#snippet filler(widths: [number, number])}
	<div class="bd-card">
		<Bar w={38} h={6} />
		<div class="bd-lines">
			<Bar w="{widths[0]}%" />
			<Bar w="{widths[1]}%" />
		</div>
		<div class="bd-meta">
			<Bar w={40} h={14} r={4} />
			<span class="bd-avatar"></span>
		</div>
	</div>
{/snippet}

{#snippet column(name: string, count: number, tone: string)}
	<div class="bd-head">
		<span class="bd-dot {tone}"></span>
		<span class="bd-name">{name}</span>
		<span class="bd-count"><Roll value={String(count)} /></span>
	</div>
{/snippet}

{#snippet side()}
	<div class="bd-side">
		<div class="bd-space"><span class="bd-logo"></span><Bar w="55%" /></div>
		<div class="bd-nav">
			<Bar w="62%" />
			<Bar w="48%" />
		</div>
		<div class="bd-nav">
			<Bar w={56} h={6} />
			<span class="bd-project"><span class="bd-dot bd-blue"></span>{t('demo.project')}</span>
			<span class="bd-row"><span class="bd-dot"></span><Bar w="58%" /></span>
			<span class="bd-row"><span class="bd-dot"></span><Bar w="44%" /></span>
		</div>
	</div>
{/snippet}

<Stage
	width={960}
	height={580}
	compact={{ below: 640, width: 358, height: 520 }}
	label={t('demo.label')}
	{script}
	{reset}
	{poster}
>
	{#snippet children({ compact })}
		<Window
			sidebar={compact ? undefined : side}
			style={compact ? 'top:16px;left:16px;width:326px;height:488px' : 'top:24px;left:40px;width:880px;height:532px'}
		>
			<div class="bd-top">
				<span class="bd-heading">{t('demo.project')}</span>
				<span class="bd-tab bd-tab-on">{t('demo.board')}</span>
				{#if !compact}
					<span class="bd-tab"><Bar w={40} /></span>
					<span class="bd-tab"><Bar w={34} /></span>
				{/if}
			</div>
			<div class="bd-board" class:bd-two={compact}>
				{#if !compact}
					<div class="bd-col">
						{@render column(t('demo.todo'), 3, '')}
						<div class="bd-list">
							{@render filler([82, 54])}
							{@render filler([70, 40])}
							{@render filler([64, 48])}
						</div>
					</div>
				{/if}
				<div class="bd-col">
					{@render column(t('demo.doing'), moved ? 1 : 2, 'bd-blue')}
					<div class="bd-list">
						{#if !moved}
							<div
								bind:this={card}
								class="bd-card bd-focal"
								class:bd-lifted={lifted}
								style="translate:{offset.x}px {offset.y}px"
								in:grow
								out:fade={{ duration: 120 }}
							>
								{@render focal(false)}
							</div>
						{/if}
						{@render filler([76, 50])}
					</div>
				</div>
				<div class="bd-col">
					{@render column(t('demo.done'), moved ? 3 : 2, 'bd-green')}
					<div class="bd-list">
						{#if phase === 'lifted' || phase === 'dragging'}
							<div bind:this={slot} class="bd-slot" style="height:{slotHeight}px" in:grow={{ duration: 400 }}></div>
						{/if}
						{#if moved}
							<div class="bd-card bd-focal" out:fade={{ duration: 250 }}>
								{@render focal(true)}
							</div>
						{/if}
						{@render filler([60, 44])}
						{@render filler([72, 38])}
					</div>
				</div>
			</div>
		</Window>
		<Cursor {cursor} />
	{/snippet}
</Stage>

<style>
	.bd-side {
		display: flex;
		flex-direction: column;
		gap: 22px;
		padding: 10px 16px;
	}
	.bd-space {
		display: flex;
		align-items: center;
		gap: 10px;
	}
	.bd-logo {
		width: 22px;
		height: 22px;
		border-radius: 6px;
		background: var(--color-fc-field);
	}
	.bd-nav {
		display: grid;
		gap: 14px;
	}
	.bd-row,
	.bd-project {
		display: flex;
		align-items: center;
		gap: 8px;
	}
	.bd-project {
		margin: -6px -8px;
		padding: 6px 8px;
		border-radius: 8px;
		background: var(--color-fc-field);
		font-size: 13px;
		font-weight: 700;
	}
	.bd-top {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 14px 18px 10px;
	}
	.bd-heading {
		margin-right: 8px;
		font-family: var(--font-fc-title);
		font-size: 16px;
		font-weight: 700;
	}
	.bd-tab {
		display: flex;
		align-items: center;
		height: 26px;
		padding: 0 10px;
		border-radius: 999px;
		font-size: 12px;
	}
	.bd-tab-on {
		background: var(--color-fc-surface);
		font-weight: 700;
	}
	.bd-board {
		display: grid;
		flex: 1;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 12px;
		min-height: 0;
		padding: 4px 16px 16px;
	}
	.bd-two {
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 8px;
		padding: 4px 10px 12px;
	}
	.bd-col {
		display: flex;
		flex-direction: column;
		min-width: 0;
		border-radius: 12px;
		background: var(--color-fc-surface);
	}
	.bd-head {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 10px 12px 4px;
		font-size: 13px;
		font-weight: 700;
	}
	.bd-name {
		overflow: hidden;
		white-space: nowrap;
		text-overflow: ellipsis;
	}
	.bd-count {
		display: inline-flex;
		justify-content: center;
		min-width: 22px;
		padding: 1px 6px;
		border-radius: 6px;
		background: var(--color-fc-field);
		font-size: 11px;
		color: var(--color-fc-fg-muted);
	}
	.bd-dot {
		flex: none;
		width: 8px;
		height: 8px;
		border-radius: 999px;
		background: var(--color-fc-fg-muted);
	}
	.bd-blue {
		background: var(--color-fc-info);
	}
	.bd-green {
		background: var(--color-fc-success);
	}
	.bd-list {
		display: flex;
		flex-direction: column;
		padding: 4px 8px 8px;
	}
	.bd-list > :global(*) {
		margin-bottom: 8px;
	}
	.bd-card {
		display: grid;
		gap: 10px;
		padding: 12px;
		border-radius: 8px;
		background: var(--mes-w-body);
		box-shadow: 0 1px 2px rgb(0 0 0 / 0.08);
	}
	.bd-lines {
		display: grid;
		gap: 7px;
	}
	.bd-focal {
		position: relative;
		z-index: 2;
		gap: 8px;
		transition:
			translate 0.9s cubic-bezier(0.45, 0, 0.1, 1),
			scale 0.2s ease,
			box-shadow 0.2s ease,
			rotate 0.2s ease;
	}
	.bd-lifted {
		scale: 1.03;
		rotate: -1.5deg;
		box-shadow: 0 14px 30px -10px rgb(0 0 0 / 0.3);
	}
	.bd-slot {
		border-radius: 8px;
		background: var(--color-fc-field);
	}
	.bd-key {
		font-family: var(--font-fc-mono);
		font-size: 10px;
		color: var(--color-fc-fg-muted);
	}
	.bd-title {
		margin: 0;
		font-size: 13px;
		font-weight: 700;
		line-height: 1.35;
	}
	.bd-meta {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 8px;
	}
	.bd-check {
		display: inline-flex;
		color: var(--color-fc-success);
	}
	.bd-avatar {
		flex: none;
		width: 20px;
		height: 20px;
		border-radius: 999px;
		background: var(--color-fc-field);
	}
	.bd-avatar-me {
		background: var(--color-fc-accent);
		opacity: 0.85;
	}
	@media (prefers-reduced-motion: reduce) {
		.bd-focal {
			transition: none;
		}
	}
</style>
