import type { CliCommand } from '@facile/muse';

export const RELEASES_URL = 'https://github.com/FacileStudio/opus-cli/releases';
export const CHECKSUMS_URL = `${RELEASES_URL}/latest/download/checksums.txt`;

export const CLI_COMMANDS: CliCommand[] = [
	{
		label: 'install.sh',
		command: 'curl -fsSL https://raw.githubusercontent.com/FacileStudio/opus-cli/main/install.sh | bash'
	},
	{ label: 'Homebrew', command: 'brew install facilestudio/tap/opus' },
	{ label: 'facile', command: 'facile install opus' }
];
