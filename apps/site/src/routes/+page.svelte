<script lang="ts">
	import { DownloadSection, Icon, icons, SignInWithFacile } from '@facile/muse';
	import { CHECKSUMS_URL, CLI_COMMANDS, RELEASES_URL } from '$lib/downloads';
	import { t } from '$lib/i18n';
	import BoardDemo from '$lib/landing/BoardDemo.svelte';
	import Shell from '$lib/landing/Shell.svelte';

	const features = $derived([
		{
			icon: icons.layers,
			title: t('features.views.title'),
			text: t('features.views.text')
		},
		{
			icon: icons.bolt,
			title: t('features.live.title'),
			text: t('features.live.text')
		},
		{
			icon: icons.settings,
			title: t('features.rules.title'),
			text: t('features.rules.text')
		},
		{
			icon: icons.code,
			title: t('features.github.title'),
			text: t('features.github.text')
		}
	]);
</script>

<svelte:head><title>Opus · {t('hero.title')}</title></svelte:head>

<Shell>
	{#snippet children(signInHref)}
		<section class="mx-auto flex max-w-5xl flex-col items-center px-4 pt-16 pb-12 text-center sm:px-6 sm:pt-24">
			<h1 class="max-w-3xl font-fc-title text-fc-3xl leading-tight font-bold tracking-tight sm:text-5xl">
				{t('hero.title')}
			</h1>
			<p class="mt-5 max-w-xl text-fc-lg text-fc-fg-muted">{t('hero.text')}</p>
			<div class="mt-8 flex flex-wrap items-center justify-center gap-3">
				<SignInWithFacile href={signInHref} size="lg" />
				<a
					href="#download"
					class="inline-flex min-h-11 items-center gap-2 rounded-fc-pill bg-fc-surface px-5 text-fc-sm font-bold transition-colors hover:bg-fc-field focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
				>
					<Icon icon={icons.code} size={16} />{t('hero.cli')}
				</a>
			</div>
			<div class="mt-12 w-full sm:mt-16"><BoardDemo /></div>
		</section>

		<section id="features" class="mx-auto max-w-5xl scroll-mt-20 px-4 py-16 sm:px-6 sm:py-20">
			<h2 class="text-center font-fc-title text-fc-2xl font-bold tracking-tight sm:text-fc-3xl">
				{t('features.title')}
			</h2>
			<div class="mt-10 grid gap-4 sm:grid-cols-2">
				{#each features as feature (feature.icon)}
					<div class="rounded-fc-lg bg-fc-component p-6">
						<span class="flex size-10 items-center justify-center rounded-fc-md bg-fc-surface">
							<Icon icon={feature.icon} size={20} />
						</span>
						<h3 class="mt-4 text-fc-md font-bold">{feature.title}</h3>
						<p class="mt-1.5 text-fc-sm text-fc-fg-muted">{feature.text}</p>
					</div>
				{/each}
			</div>
		</section>

		<section id="download" class="mx-auto max-w-4xl scroll-mt-20 px-4 py-16 sm:px-6 sm:py-20">
			<div class="mb-8 text-center">
				<h2 class="font-fc-title text-fc-2xl font-bold tracking-tight sm:text-fc-3xl">{t('download.title')}</h2>
				<p class="mt-3 text-fc-fg-muted">{t('download.text')}</p>
			</div>
			<DownloadSection
				mode="device"
				allHref="/downloads"
				commands={CLI_COMMANDS}
				releasesHref={RELEASES_URL}
				checksumsHref={CHECKSUMS_URL}
				headingLevel={3}
			/>
		</section>

		<section class="mx-auto max-w-5xl px-4 py-16 text-center sm:px-6 sm:py-24">
			<h2 class="font-fc-title text-fc-2xl font-bold tracking-tight sm:text-fc-3xl">{t('cta.title')}</h2>
			<p class="mt-3 text-fc-fg-muted">{t('cta.text')}</p>
			<div class="mt-8 flex justify-center">
				<SignInWithFacile href={signInHref} size="lg" />
			</div>
		</section>
	{/snippet}
</Shell>
