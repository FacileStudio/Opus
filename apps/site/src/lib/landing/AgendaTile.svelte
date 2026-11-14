<script lang="ts">
	import { type Act, Bar, Cursor, cursorAt, Roll, Stage, Tile } from '@facile/mise-en-scene';
	import { Icon, icons } from '@facile/muse';
	import { t } from '$lib/i18n';
	import { AGENDA_ICON, TASK_ICON } from '$lib/icons';

	const DATES = [18, 19, 20];

	let dated = $state(false);
	let shown = $state(false);
	let cursor = $state(cursorAt());
	let chip = $state<HTMLElement>();

	const days = $derived(t('tiles.agenda.days').split(','));

	function reset() {
		dated = false;
		shown = false;
		cursor.visible = false;
		cursor.down = false;
	}

	function poster() {
		dated = true;
		shown = true;
	}

	async function script(act: Act) {
		await act.wait(800);
		const target = act.point(chip, 0.5, 0.55);
		await act.move(cursor, { x: target.x + 40, y: target.y + 60 }, 400);
		cursor.visible = true;
		await act.move(cursor, target);
		await act.click(cursor, () => (dated = true));
		await act.wait(500);
		shown = true;
		await act.move(cursor, { x: target.x + 30, y: target.y + 70 }, 500);
		cursor.visible = false;
		await act.wait(2400);
		shown = false;
		await act.wait(300);
		dated = false;
		await act.wait(700);
	}
</script>

<Tile title={t('tiles.agenda.title')} body={t('tiles.agenda.body')}>
	<Stage width={320} height={240} {script} {reset} {poster}>
		{#snippet children()}
			<div class="panel absolute top-4 right-5 left-5 flex h-[68px] flex-col justify-center gap-2 rounded-fc-lg bg-fc-surface px-3">
				<div class="flex items-center gap-2 text-fc-xs font-bold text-fc-fg">
					<Icon icon={TASK_ICON} size={13} class="shrink-0 text-fc-fg-muted" />
					<span class="truncate">{t('tiles.agenda.task')}</span>
				</div>
				<div class="flex items-center gap-3 text-[11px] text-fc-fg-muted">
					<span class="w-16 shrink-0">{t('tiles.agenda.due')}</span>
					<span bind:this={chip} class="due" class:due-on={dated}>
						<Icon icon={icons.calendar} size={12} />
						{#if dated}<Roll value={t('tiles.agenda.date')} />{:else}<Bar w={34} h={6} />{/if}
					</span>
				</div>
			</div>
			<div class="panel absolute top-[100px] right-5 bottom-4 left-5 flex flex-col gap-2 rounded-fc-lg bg-fc-page p-3">
				<div class="flex items-center gap-2 text-fc-xs font-bold text-fc-fg">
					<Icon icon={AGENDA_ICON} size={14} />Agenda
				</div>
				<div class="grid flex-1 grid-cols-3 gap-1.5">
					{#each DATES as date, i (date)}
						<div class="day">
							<span class="text-[10px] text-fc-fg-muted">{days[i]} {date}</span>
							{#if i === DATES.length - 1}
								<span class="event" class:event-on={shown}>{t('tiles.agenda.task')}</span>
							{:else}
								<Bar w="70%" h={6} />
							{/if}
						</div>
					{/each}
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
	.due {
		display: inline-flex;
		align-items: center;
		gap: 5px;
		height: 22px;
		padding: 0 8px;
		border-radius: 6px;
		background: var(--color-fc-field);
		transition: color 0.2s ease;
	}
	.due-on {
		font-weight: 700;
		color: var(--color-fc-fg);
	}
	.day {
		display: flex;
		flex-direction: column;
		gap: 8px;
		padding: 6px;
		border-radius: 8px;
		background: var(--color-fc-surface);
	}
	.event {
		overflow: hidden;
		padding: 4px 6px;
		border-radius: 6px;
		background: #8b5cf6;
		font-size: 10px;
		font-weight: 700;
		white-space: nowrap;
		text-overflow: ellipsis;
		color: white;
		opacity: 0;
		scale: 0.9;
		transform-origin: left center;
		transition:
			opacity 0.3s ease,
			scale 0.4s cubic-bezier(0.2, 0.8, 0.2, 1);
	}
	.event-on {
		opacity: 1;
		scale: 1;
	}
	@media (prefers-reduced-motion: reduce) {
		.event {
			transition: none;
		}
	}
</style>
