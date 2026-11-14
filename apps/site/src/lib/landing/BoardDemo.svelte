<script lang="ts">
	import { type Act, Bar, Cursor, cursorAt, type Point, Roll, Stage, Window } from '@facile/mise-en-scene';
	import { tick } from 'svelte';
	import { fade, fly } from 'svelte/transition';
	import { t } from '$lib/i18n';
	import AgendaPeek from './AgendaPeek.svelte';
	import BoardCard from './BoardCard.svelte';
	import BoardSide from './BoardSide.svelte';
	import { drag, idleMove, LEA_COLOR, leave, reach } from './board';
	import LiveCursor from './LiveCursor.svelte';

	const ME = 'var(--color-fc-accent)';

	let mine = $state(idleMove(98));
	let lea = $state(idleMove(110));
	let me = $state(cursorAt());
	let her = $state(cursorAt());
	let peek = $state(false);
	let myCard = $state<HTMLElement>();
	let mySlot = $state<HTMLElement>();
	let herCard = $state<HTMLElement>();
	let herSlot = $state<HTMLElement>();
	let herLanded = $state<HTMLElement>();

	const busy = (phase: string) => phase === 'lifted' || phase === 'dragging';
	const todo = $derived(lea.phase === 'done' ? 2 : 3);
	const doing = $derived(1 + (mine.phase === 'done' ? 0 : 1) + (lea.phase === 'done' ? 1 : 0));
	const done = $derived(mine.phase === 'done' ? 3 : 2);

	function reset() {
		mine = idleMove(mine.height);
		lea = idleMove(lea.height);
		peek = false;
		for (const who of [me, her]) {
			who.visible = false;
			who.down = false;
		}
	}

	function poster() {
		mine.phase = 'done';
		lea.phase = 'done';
		peek = true;
	}

	async function arrive(act: Act) {
		lea.phase = 'done';
		await tick();
		const grip = act.point(herLanded, 0.35, 0.45);
		her.visible = true;
		her.down = true;
		await act.move(her, grip);
		her.down = false;
	}

	async function script(act: Act) {
		await act.wait(800);
		const wide = Boolean(herCard);
		const myDrop = reach(act, me, myCard).then((grip) =>
			drag(act, me, mine, grip, () => myCard, () => mySlot)
		);
		const herGrip = act.wait(700).then(async (): Promise<Point | null> => {
			if (wide) return reach(act, her, herCard);
			await act.move(her, { ...act.point(myCard, 0, 0.45), x: 8 }, 0);
			return null;
		});
		const [, grip] = await Promise.all([myDrop, herGrip]);
		await act.wait(200);
		await Promise.all([
			leave(act, me),
			grip ? drag(act, her, lea, grip, () => herCard, () => herSlot) : arrive(act)
		]);
		peek = true;
		await act.wait(300);
		await leave(act, her);
		await act.wait(2400);
		peek = false;
		lea = idleMove(lea.height);
		await act.wait(350);
		mine = idleMove(mine.height);
		await act.wait(900);
	}
</script>

{#snippet column(name: string, count: number, tone: string)}
	<div class="bd-head">
		<span class="bd-dot {tone}"></span>
		<span class="bd-name">{name}</span>
		<span class="bd-count"><Roll value={String(count)} /></span>
	</div>
{/snippet}

{#snippet filler(widths: [number, number], who?: string)}
	<div class="bd-card"><BoardCard {widths} {who} /></div>
{/snippet}

{#snippet slot(height: number)}
	<div class="bd-slot" style="height:{height}px" in:fade={{ duration: 200 }}></div>
{/snippet}

<Stage
	width={960}
	height={580}
	compact={{ below: 640, width: 358, height: 560 }}
	label={t('demo.label')}
	{script}
	{reset}
	{poster}
>
	{#snippet children({ compact })}
		<Window
			sidebar={compact ? undefined : side}
			style={compact ? 'top:16px;left:16px;width:326px;height:528px' : 'top:24px;left:40px;width:880px;height:532px'}
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
						{@render column(t('demo.todo'), todo, '')}
						<div class="bd-list">
							{@render filler([82, 54])}
							{@render filler([64, 48], LEA_COLOR)}
							{#if lea.phase !== 'done'}
								<div
									bind:this={herCard}
									class="bd-card bd-focal"
									class:bd-lifted={busy(lea.phase)}
									style="translate:{lea.offset.x}px {lea.offset.y}px"
									in:fade={{ duration: 250 }}
									out:fade={{ duration: 120 }}
								>
									<BoardCard key="WEB-17" title={t('demo.leaTask')} due={t('demo.due')} who={LEA_COLOR} />
								</div>
							{/if}
						</div>
					</div>
				{/if}
				<div class="bd-col">
					{@render column(t('demo.doing'), doing, 'bd-blue')}
					<div class="bd-list">
						{@render filler([76, 50], LEA_COLOR)}
						{#if mine.phase !== 'done'}
							<div
								bind:this={myCard}
								class="bd-card bd-focal"
								class:bd-lifted={busy(mine.phase)}
								style="translate:{mine.offset.x}px {mine.offset.y}px"
								in:fade={{ duration: 250 }}
								out:fade={{ duration: 120 }}
							>
								<BoardCard key="WEB-14" title={t('demo.task')} who={ME} />
							</div>
						{/if}
						{#if busy(lea.phase)}
							<div bind:this={herSlot}>{@render slot(lea.height)}</div>
						{/if}
						{#if lea.phase === 'done'}
							<div
								bind:this={herLanded}
								class="bd-card bd-focal"
								in:fly={{ x: compact ? -120 : 0, duration: compact ? 900 : 0, opacity: compact ? 0 : 1 }}
								out:fade={{ duration: 250 }}
							>
								<BoardCard key="WEB-17" title={t('demo.leaTask')} due={t('demo.due')} who={LEA_COLOR} />
							</div>
						{/if}
					</div>
				</div>
				<div class="bd-col">
					{@render column(t('demo.done'), done, 'bd-green')}
					<div class="bd-list">
						{@render filler([60, 44])}
						{@render filler([72, 38], LEA_COLOR)}
						{#if busy(mine.phase)}
							<div bind:this={mySlot}>{@render slot(mine.height)}</div>
						{/if}
						{#if mine.phase === 'done'}
							<div class="bd-card bd-focal" out:fade={{ duration: 250 }}>
								<BoardCard key="WEB-14" title={t('demo.task')} who={ME} done />
							</div>
						{/if}
					</div>
				</div>
			</div>
		</Window>
		<AgendaPeek
			shown={peek}
			style={compact ? 'left:28px;right:28px;bottom:10px' : 'right:58px;bottom:40px;width:230px'}
		/>
		<LiveCursor cursor={her} color={LEA_COLOR} name="Léa" />
		<Cursor cursor={me} />
	{/snippet}
</Stage>

{#snippet side()}<BoardSide />{/snippet}

<style>
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
	@media (prefers-reduced-motion: reduce) {
		.bd-focal {
			transition: none;
		}
	}
</style>
