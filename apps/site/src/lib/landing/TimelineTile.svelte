<script lang="ts">
	import { type Act, Bar, Cursor, cursorAt, Stage, Tile } from '@facile/mise-en-scene';
	import { t } from '$lib/i18n';

	const COL = 38;
	const DAYS = [10, 11, 12, 13, 14, 15, 16];
	const TODAY = 2;
	const FROM = 1;
	const SHORT = 3;
	const LONG = 5;

	let span = $state(SHORT);
	let held = $state(false);
	let cursor = $state(cursorAt());
	let handle = $state<HTMLElement>();

	function reset() {
		span = SHORT;
		held = false;
		cursor.visible = false;
		cursor.down = false;
	}

	function poster() {
		span = LONG;
	}

	async function script(act: Act) {
		await act.wait(800);
		const grip = act.point(handle, 0.5, 0.5);
		await act.move(cursor, { x: grip.x + 40, y: grip.y + 70 }, 400);
		cursor.visible = true;
		await act.move(cursor, grip);
		cursor.down = true;
		held = true;
		await act.wait(200);
		span = LONG;
		await act.move(cursor, { x: grip.x + COL * (LONG - SHORT), y: grip.y });
		cursor.down = false;
		held = false;
		await act.wait(200);
		await act.move(cursor, { x: grip.x + COL * (LONG - SHORT) + 30, y: grip.y + 60 }, 500);
		cursor.visible = false;
		await act.wait(2000);
		reset();
		await act.wait(600);
	}
</script>

{#snippet bar(from: number, width: number, top: number, focal: boolean)}
	<div class="gt-bar" class:gt-focal={focal} style="left:{from * COL + 3}px;width:{width * COL - 6}px;top:{top}px">
		<span class="gt-handle"></span>
		<span class="gt-label">
			{#if focal}{t('tiles.timeline.task')}{:else}<Bar w="70%" />{/if}
		</span>
		{#if focal}
			<span bind:this={handle} class="gt-handle" class:gt-held={held}></span>
		{:else}
			<span class="gt-handle"></span>
		{/if}
	</div>
{/snippet}

<Tile title={t('tiles.timeline.title')} body={t('tiles.timeline.body')}>
	<Stage width={320} height={240} {script} {reset} {poster}>
		{#snippet children()}
			<div class="panel absolute top-5 left-[19px] h-[200px] w-[282px] overflow-hidden rounded-fc-lg bg-fc-page p-2">
				<div class="relative h-full">
					<div class="gt-grid">
						{#each DAYS as day, i (day)}
							<div class="gt-day" class:gt-today={i === TODAY}>{day}</div>
						{/each}
					</div>
					{@render bar(0, 2, 32, false)}
					{@render bar(FROM, span, 72, true)}
					{@render bar(3, 3, 112, false)}
					{@render bar(4, 3, 152, false)}
				</div>
			</div>
			<Cursor {cursor} />
		{/snippet}
	</Stage>
</Tile>

<style>
	.panel {
		box-shadow: 0 10px 24px -14px rgb(0 0 0 / 0.3);
	}
	.gt-grid {
		position: absolute;
		inset: 0;
		display: grid;
		grid-template-columns: repeat(7, 38px);
	}
	.gt-day {
		padding-top: 6px;
		border-radius: 8px;
		font-family: var(--font-fc-mono);
		font-size: 10px;
		text-align: center;
		color: var(--color-fc-fg-muted);
	}
	.gt-today {
		background: var(--color-fc-surface);
		font-weight: 700;
		color: var(--color-fc-fg);
	}
	.gt-bar {
		position: absolute;
		display: flex;
		align-items: stretch;
		height: 30px;
		overflow: hidden;
		border-radius: 6px;
		background: color-mix(in srgb, var(--color-fc-info) 14%, var(--color-fc-page));
		box-shadow: 0 1px 2px rgb(0 0 0 / 0.08);
		font-size: 11px;
		font-weight: 600;
		color: var(--color-fc-fg);
	}
	.gt-focal {
		z-index: 2;
		transition: width 0.9s cubic-bezier(0.45, 0, 0.1, 1);
	}
	.gt-handle {
		flex: none;
		width: 6px;
		background: color-mix(in srgb, var(--color-fc-info) 22%, var(--color-fc-page));
		transition: background 0.2s ease;
	}
	.gt-held {
		background: color-mix(in srgb, var(--color-fc-info) 55%, var(--color-fc-page));
	}
	.gt-label {
		display: flex;
		flex: 1;
		align-items: center;
		min-width: 0;
		padding: 0 8px;
		overflow: hidden;
		white-space: nowrap;
		text-overflow: ellipsis;
	}
	@media (prefers-reduced-motion: reduce) {
		.gt-focal {
			transition: none;
		}
	}
</style>
