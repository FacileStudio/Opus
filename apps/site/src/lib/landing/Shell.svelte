<script lang="ts">
	import { LandingFooter, LandingHeader, SignInWithFacile } from '@facile/muse';
	import { onMount, type Snippet } from 'svelte';
	import { rememberLocale, t } from '$lib/i18n';
	import { OPUS_ICON } from '$lib/icons';
	import { elsewhere } from '$lib/session';

	let { children }: { children: Snippet<[string]> } = $props();

	const signInHref = '/api/auth/oidc';
	let ready = $state(false);

	const links = $derived([
		{ href: '/#features', label: t('nav.features') },
		{ href: '/downloads', label: t('nav.download') }
	]);
	const columns = $derived([
		{ title: 'Opus', links: links.map(({ href, label }) => ({ href, label })) },
		{
			title: 'Facile',
			links: [{ href: 'https://facile.studio', label: 'facile.studio', external: true }]
		}
	]);

	onMount(async () => {
		const target = await elsewhere();
		if (target) {
			location.replace(target);
			return;
		}
		ready = true;
	});
</script>

{#if ready}
	<div class="min-h-dvh bg-fc-page text-fc-fg">
		<LandingHeader appName="Opus" icon={OPUS_ICON} {links} skipHref="#main">
			{#snippet actions()}
				<SignInWithFacile href={signInHref} />
			{/snippet}
		</LandingHeader>

		<main id="main">{@render children(signInHref)}</main>

		<LandingFooter
			appName="Opus"
			icon={OPUS_ICON}
			tagline={t('footer.tagline')}
			{columns}
			legal={{ onLocaleChange: rememberLocale }}
		/>
	</div>
{/if}
