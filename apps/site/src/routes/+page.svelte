<script lang="ts">
	import {
		FeatureSplit,
		ForDevelopers,
		LANDING_HEADING,
		LandingCta,
		PricingPlans,
		WorksWith,
		icons
	} from '@facile/muse';
	import { t } from '$lib/i18n';
	import AgendaTile from '$lib/landing/AgendaTile.svelte';
	import BoardDemo from '$lib/landing/BoardDemo.svelte';
	import { connections, developers, features } from '$lib/landing/sections';
	import SearchTile from '$lib/landing/SearchTile.svelte';
	import Shell from '$lib/landing/Shell.svelte';
	import TimelineTile from '$lib/landing/TimelineTile.svelte';

	const block = 'mx-auto w-full max-w-5xl px-4 sm:px-6';
</script>

<svelte:head>
	<title>Opus · {t('meta.title')}</title>
	<meta name="description" content={t('meta.description')} />
</svelte:head>

<Shell>
	{#snippet children(signInHref)}
		<section class="mx-auto flex max-w-5xl flex-col items-center px-4 pt-16 pb-12 text-center sm:px-6 sm:pt-24">
			<h1 class="max-w-3xl font-fc-title text-fc-3xl leading-tight font-bold tracking-tight sm:text-5xl">
				{t('hero.title')}
			</h1>
			<p class="mt-5 max-w-2xl text-fc-lg text-fc-fg-muted">{t('hero.text')}</p>
			<LandingCta {signInHref} align="center" class="mt-8 w-full" />
			<div class="mt-12 w-full sm:mt-16"><BoardDemo /></div>
		</section>

		<section class="{block} py-16 sm:py-20">
			<h2 class="mb-10 text-center {LANDING_HEADING}">{t('tiles.title')}</h2>
			<div class="grid gap-6 md:grid-cols-3">
				<TimelineTile />
				<SearchTile />
				<AgendaTile />
			</div>
		</section>

		<div class="{block} py-16 sm:py-20">
			<WorksWith class="max-w-3xl" apps={connections()} description={t('with.text')} />
		</div>

		<FeatureSplit
			id="features"
			class="{block} scroll-mt-20 py-16 sm:py-20"
			title={t('features.title')}
			description={t('features.text')}
			features={features()}
		/>

		<PricingPlans
			id="pricing"
			class="{block} scroll-mt-20 py-16 sm:py-20"
			title={t('plan.title')}
			headingClass={LANDING_HEADING}
			highlight={['ai.access']}
			highlightOnly
			notes={{ '*': t('plan.included') }}
			recommended="pro"
			suiteHref="https://suite.facile.studio/#tarifs"
			suiteNote={t('plan.note')}
		/>

		<div id="developers" class="{block} scroll-mt-20 py-16 sm:py-20">
			<ForDevelopers
				rows={developers()}
				description={t('dev.text')}
				selfHostHref="https://github.com/FacileStudio/Opus/blob/main/docs/deployment.md"
			/>
		</div>

		<section class="{block} flex flex-col items-center py-16 text-center sm:py-24">
			<h2 class={LANDING_HEADING}>{t('cta.title')}</h2>
			<p class="mt-3 text-fc-fg-muted">{t('cta.text')}</p>
			<LandingCta {signInHref} align="center" class="mt-8 w-full" />
		</section>
	{/snippet}
</Shell>
