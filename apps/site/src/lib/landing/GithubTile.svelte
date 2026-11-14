<script lang="ts">
	import { type Act, Bar, Stage, Tile } from '@facile/mise-en-scene';
	import { Icon } from '@facile/muse';
	import { t } from '$lib/i18n';
	import { PR_MERGED_ICON, PR_OPEN_ICON } from '$lib/icons';

	let merged = $state(false);
	let done = $state(false);

	function reset() {
		merged = false;
		done = false;
	}

	function poster() {
		merged = true;
		done = true;
	}

	async function script(act: Act) {
		await act.wait(1600);
		merged = true;
		await act.wait(1000);
		done = true;
		await act.wait(2800);
		reset();
		await act.wait(1400);
	}
</script>

{#snippet swap(on: boolean, from: string, to: string)}
	<span class="swap">
		<span class="item" class:off={on}>{from}</span>
		<span class="item item-in" class:off={!on}>{to}</span>
	</span>
{/snippet}

<Tile title={t('tiles.github.title')} body={t('tiles.github.body')}>
	<Stage width={320} height={240} {script} {reset} {poster}>
		{#snippet children()}
			<div class="panel absolute top-5 right-5 bottom-5 left-5 flex flex-col gap-3 rounded-fc-lg bg-fc-page p-4">
				<span class="font-fc-mono text-[10px] text-fc-fg-muted">WEB-21</span>
				<p class="-mt-2 text-fc-sm leading-snug font-bold text-fc-fg">{t('tiles.github.task')}</p>
				<div class="flex flex-wrap items-center gap-1.5">
					<span class="chip">
						<span class="swap">
							<span class="dot" class:off={done}></span>
							<span class="dot dot-done" class:off={!done}></span>
						</span>
						{@render swap(done, t('demo.doing'), t('demo.done'))}
					</span>
					<span class="chip">
						<span class="swap">
							<span class="item" class:off={merged}>
								<Icon icon={PR_OPEN_ICON} size={13} class="text-fc-success" />
							</span>
							<span class="item item-in" class:off={!merged}>
								<Icon icon={PR_MERGED_ICON} size={13} class="text-fc-info" />
							</span>
						</span>
						<span>#42</span>
						{@render swap(merged, t('tiles.github.open'), t('tiles.github.merged'))}
					</span>
				</div>
				<div class="flex flex-col gap-2">
					<Bar w="88%" />
					<Bar w="64%" />
				</div>
				<div class="item item-in flex items-center gap-2" class:off={!merged}>
					<Icon icon={PR_MERGED_ICON} size={13} class="shrink-0 text-fc-info" />
					<Bar w="58%" />
				</div>
			</div>
		{/snippet}
	</Stage>
</Tile>

<style>
	.panel {
		box-shadow: 0 10px 24px -14px rgb(0 0 0 / 0.3);
	}
	.chip {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		height: 28px;
		padding: 0 8px;
		border-radius: 8px;
		background: var(--color-fc-surface);
		font-size: 11px;
		font-weight: 700;
		color: var(--color-fc-fg);
	}
	.swap {
		display: inline-grid;
		align-items: center;
		justify-items: start;
	}
	.swap > * {
		grid-area: 1 / 1;
	}
	.item {
		transition:
			opacity 320ms cubic-bezier(0.22, 1, 0.36, 1),
			transform 320ms cubic-bezier(0.22, 1, 0.36, 1);
	}
	.item.off {
		opacity: 0;
		transform: translateY(-5px);
	}
	.item-in.off {
		transform: translateY(5px);
	}
	.dot {
		width: 8px;
		height: 8px;
		border-radius: 999px;
		background: var(--color-fc-info);
		transition:
			opacity 320ms cubic-bezier(0.22, 1, 0.36, 1),
			transform 320ms cubic-bezier(0.22, 1, 0.36, 1);
	}
	.dot.off {
		opacity: 0;
		transform: scale(0.4);
	}
	.dot-done {
		background: var(--color-fc-success);
	}
</style>
