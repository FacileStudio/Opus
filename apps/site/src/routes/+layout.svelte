<script lang="ts">
	import '../app.css';
	import { provideLocale } from '@facile/muse';
	import { bindLocale, localeSource } from '$lib/i18n';
	import '$lib/icons';

	let { children } = $props();

	const i18n = provideLocale(localeSource);
	bindLocale(() => i18n.locale);
</script>

{@render children()}
