<script lang="ts">
	import { type Act, Bar, grow, Roll, Stage, Tile } from '@facile/mise-en-scene';
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
		await act.wait(700);
		done = true;
		await act.wait(2800);
		reset();
		await act.wait(900);
	}
</script>

<Tile title={t('tiles.github.title')} body={t('tiles.github.body')}>
	<Stage width={320} height={240} {script} {reset} {poster}>
		{#snippet children()}
			<div class="panel absolute top-5 right-5 bottom-5 left-5 flex flex-col gap-3 rounded-fc-lg bg-fc-page p-4">
				<span class="font-fc-mono text-[10px] text-fc-fg-muted">WEB-21</span>
				<p class="-mt-2 text-fc-sm leading-snug font-bold text-fc-fg">{t('tiles.github.task')}</p>
				<div class="flex flex-wrap items-center gap-1.5">
					<span class="chip">
						<span class="dot" class:dot-done={done}></span>
						<Roll value={done ? t('demo.done') : t('demo.doing')} class="h-4 leading-4" />
					</span>
					<span class="chip">
						<Icon
							icon={merged ? PR_MERGED_ICON : PR_OPEN_ICON}
							size={13}
							class={merged ? 'text-fc-info' : 'text-fc-success'}
						/>
						<span>#42</span>
						<Roll value={merged ? t('tiles.github.merged') : t('tiles.github.open')} class="h-4 leading-4" />
					</span>
				</div>
				<div class="flex flex-col gap-2">
					<Bar w="88%" />
					<Bar w="64%" />
				</div>
				{#if merged}
					<div class="flex items-center gap-2" in:grow={{ duration: 400 }} out:grow={{ duration: 250 }}>
						<Icon icon={PR_MERGED_ICON} size={13} class="shrink-0 text-fc-info" />
						<Bar w="58%" />
					</div>
				{/if}
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
	.dot {
		flex: none;
		width: 8px;
		height: 8px;
		border-radius: 999px;
		background: var(--color-fc-info);
		transition: background 0.3s ease;
	}
	.dot-done {
		background: var(--color-fc-success);
	}
</style>
