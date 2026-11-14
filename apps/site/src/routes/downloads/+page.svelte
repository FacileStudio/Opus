<script lang="ts">
	import { DownloadSection } from '@facile/muse';
	import { CHECKSUMS_URL, CLI_COMMANDS, RELEASES_URL } from '$lib/downloads';
	import { t } from '$lib/i18n';
	import Shell from '$lib/landing/Shell.svelte';
</script>

<svelte:head><title>Opus · {t('downloads.title')}</title></svelte:head>

<Shell>
	{#snippet children()}
		<section class="mx-auto max-w-4xl px-4 pt-16 pb-20 sm:px-6 sm:pt-24">
			<div class="mb-8 text-center">
				<h1 class="font-fc-title text-fc-3xl leading-tight font-bold tracking-tight sm:text-5xl">
					{t('downloads.title')}
				</h1>
				<p class="mt-4 text-fc-lg text-fc-fg-muted">{t('downloads.text')}</p>
			</div>
			<DownloadSection
				commands={CLI_COMMANDS}
				releasesHref={RELEASES_URL}
				checksumsHref={CHECKSUMS_URL}
				headingLevel={2}
			/>
		</section>
	{/snippet}
</Shell>
