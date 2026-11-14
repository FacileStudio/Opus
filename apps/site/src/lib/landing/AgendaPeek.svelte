<script lang="ts">
	import { Icon } from '@facile/muse';
	import { t } from '$lib/i18n';
	import { AGENDA_ICON } from '$lib/icons';

	let { shown, style }: { shown: boolean; style: string } = $props();
</script>

<div class="ap" class:ap-on={shown} {style}>
	<div class="ap-head">
		<Icon icon={AGENDA_ICON} size={16} />
		<span class="ap-app">Agenda</span>
		<span class="ap-day">{t('demo.agenda.day')}</span>
	</div>
	<div class="ap-event">
		<span class="ap-title">{t('demo.leaTask')}</span>
		<span class="ap-meta">{t('demo.agenda.allDay')} · {t('demo.agenda.calendar')}</span>
	</div>
</div>

<style>
	.ap {
		position: absolute;
		z-index: 20;
		display: grid;
		gap: 10px;
		padding: 12px;
		border-radius: 14px;
		background: var(--mes-w-body);
		box-shadow: 0 18px 40px -16px rgb(0 0 0 / 0.35);
		opacity: 0;
		translate: 0 14px;
		transition:
			opacity 0.35s ease,
			translate 0.5s cubic-bezier(0.2, 0.8, 0.2, 1);
	}
	.ap-on {
		opacity: 1;
		translate: 0 0;
	}
	.ap-head {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 12px;
	}
	.ap-app {
		font-weight: 700;
	}
	.ap-day {
		margin-left: auto;
		color: var(--color-fc-fg-muted);
	}
	.ap-event {
		display: grid;
		gap: 2px;
		padding: 8px 10px;
		border-radius: 8px;
		background: #8b5cf6;
		color: white;
	}
	.ap-title {
		overflow: hidden;
		font-size: 12px;
		font-weight: 700;
		white-space: nowrap;
		text-overflow: ellipsis;
	}
	.ap-meta {
		font-size: 10px;
		opacity: 0.85;
	}
	@media (prefers-reduced-motion: reduce) {
		.ap {
			transition: none;
		}
	}
</style>
