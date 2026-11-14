import { type DeveloperRow, type FeatureItem, icons, type WorksWithApp } from '@facile/muse';
import { t } from '$lib/i18n';
import { AGENDA_ICON, BULLE_ICON, SABLIER_ICON, TOILE_ICON } from '$lib/icons';

/** The suite apps Opus exchanges data with, as copy.json lists them. */
export function connections(): WorksWithApp[] {
	return [
		{ name: 'Sablier', icon: SABLIER_ICON, line: t('with.sablier'), href: 'https://sablier.facile.studio' },
		{ name: 'Agenda', icon: AGENDA_ICON, line: t('with.agenda'), href: 'https://agenda.facile.studio' },
		{ name: 'Toile', icon: TOILE_ICON, line: t('with.toile'), href: 'https://toile.facile.studio' },
		{ name: 'Bulle', icon: BULLE_ICON, line: t('with.bulle'), href: 'https://bulle.facile.studio' }
	];
}

/** The plain feature list under the demos. */
export function features(): FeatureItem[] {
	return [
		{ icon: icons.layers, title: t('features.views.title'), text: t('features.views.text') },
		{ icon: icons.bolt, title: t('features.live.title'), text: t('features.live.text') },
		{ icon: icons.widget, title: t('features.links.title'), text: t('features.links.text') },
		{ icon: icons.eye, title: t('features.share.title'), text: t('features.share.text') }
	];
}

/** The CLI, API, MCP and GitHub rows of the developer block. */
export function developers(): DeveloperRow[] {
	return [
		{ kind: 'cli', text: t('dev.cli'), command: 'brew install facilestudio/tap/opus', href: '/downloads' },
		{ kind: 'api', text: t('dev.api'), href: 'https://github.com/FacileStudio/Opus/blob/main/docs/api.md' },
		{ kind: 'mcp', text: t('dev.mcp') },
		{ kind: 'source', text: t('dev.source'), href: 'https://github.com/FacileStudio/Opus' }
	];
}
