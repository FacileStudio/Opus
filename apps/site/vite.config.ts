import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	optimizeDeps: { exclude: ['@facile/muse', '@facile/mise-en-scene'] },
	server: {
		proxy: { '/api': { target: 'http://localhost:1337', changeOrigin: true } }
	}
});
