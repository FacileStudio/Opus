<script lang="ts">
	import { type Act, Bar, Caret, grow, Stage, Tile } from '@facile/mise-en-scene';
	import { Icon, icons } from '@facile/muse';
	import { t } from '$lib/i18n';
	import { COMMENT_ICON, TASK_ICON } from '$lib/icons';

	const query = $derived(t('tiles.search.query'));

	let typed = $state('');
	let found = $state(false);

	function reset() {
		typed = '';
		found = false;
	}

	function poster() {
		typed = query;
		found = true;
	}

	async function script(act: Act) {
		await act.wait(700);
		await act.type(query, (v) => (typed = v), 110);
		await act.wait(300);
		found = true;
		await act.wait(3000);
		reset();
		await act.wait(600);
	}
</script>

{#snippet result(icon: string, title: string)}
	<div class="flex items-start gap-3 rounded-fc-md px-2 py-2">
		<Icon {icon} size={14} class="mt-0.5 shrink-0 text-fc-fg-muted" />
		<div class="flex min-w-0 flex-1 flex-col gap-1.5">
			<span class="truncate text-fc-xs font-bold text-fc-fg">{title}</span>
			<Bar w="55%" h={6} />
		</div>
	</div>
{/snippet}

<Tile title={t('tiles.search.title')} body={t('tiles.search.body')}>
	<Stage width={320} height={240} {script} {reset} {poster}>
		{#snippet children()}
			<div class="panel absolute top-5 right-5 bottom-5 left-5 flex flex-col rounded-fc-lg bg-fc-surface">
				<div class="flex h-10 shrink-0 items-center gap-2 px-4 text-fc-xs text-fc-fg">
					<Icon icon={icons.search} size={14} class="shrink-0 text-fc-fg-muted" />
					{#if typed}
						<span class="whitespace-nowrap"
							>{typed}{#if !found}<Caret />{/if}</span
						>
					{:else}
						<Bar w="62%" />
					{/if}
				</div>
				<div class="flex min-h-0 flex-1 flex-col rounded-fc-lg bg-fc-page">
					{#if found}
						<div class="flex flex-col px-1 pt-1" in:grow={{ duration: 400 }} out:grow={{ duration: 250 }}>
							<span class="px-2 pt-1 pb-1 text-[10px] font-bold text-fc-fg-muted">{t('tiles.search.tasks')}</span>
							<div class="rounded-fc-md bg-fc-surface">{@render result(TASK_ICON, t('demo.task'))}</div>
							<span class="px-2 pt-2 pb-1 text-[10px] font-bold text-fc-fg-muted">{t('tiles.search.comments')}</span>
							{@render result(COMMENT_ICON, t('tiles.search.comment'))}
						</div>
					{:else}
						<div class="flex flex-col gap-3 px-3 pt-4" out:grow={{ duration: 250 }}>
							<Bar w="30%" h={6} />
							<Bar w="74%" />
							<Bar w="58%" />
						</div>
					{/if}
				</div>
			</div>
		{/snippet}
	</Stage>
</Tile>

<style>
	.panel {
		box-shadow: 0 10px 24px -14px rgb(0 0 0 / 0.3);
	}
</style>
