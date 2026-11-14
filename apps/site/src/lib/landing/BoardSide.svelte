<script lang="ts">
	import { Bar } from '@facile/mise-en-scene';
	import { t } from '$lib/i18n';
</script>

<div class="bs-side">
	<div class="bs-space"><span class="bs-logo"></span><Bar w="55%" /></div>
	<div class="bs-nav">
		<Bar w="62%" />
		<Bar w="48%" />
	</div>
	<div class="bs-nav">
		<Bar w={56} h={6} />
		<span class="bs-project"><span class="bs-dot bs-blue"></span>{t('demo.project')}</span>
		<span class="bs-row"><span class="bs-dot"></span><Bar w="58%" /></span>
		<span class="bs-row"><span class="bs-dot"></span><Bar w="44%" /></span>
	</div>
</div>

<style>
	.bs-side {
		display: flex;
		flex-direction: column;
		gap: 22px;
		padding: 10px 16px;
	}
	.bs-space {
		display: flex;
		align-items: center;
		gap: 10px;
	}
	.bs-logo {
		width: 22px;
		height: 22px;
		border-radius: 6px;
		background: var(--color-fc-field);
	}
	.bs-nav {
		display: grid;
		gap: 14px;
	}
	.bs-row,
	.bs-project {
		display: flex;
		align-items: center;
		gap: 8px;
	}
	.bs-project {
		margin: -6px -8px;
		padding: 6px 8px;
		border-radius: 8px;
		background: var(--color-fc-field);
		font-size: 13px;
		font-weight: 700;
	}
	.bs-dot {
		flex: none;
		width: 8px;
		height: 8px;
		border-radius: 999px;
		background: var(--color-fc-fg-muted);
	}
	.bs-blue {
		background: var(--color-fc-info);
	}
</style>
