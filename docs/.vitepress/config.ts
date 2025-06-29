import { defineConfig } from 'vitepress';
import { getVersion } from './utils';

// https://vitepress.dev/reference/site-config
export default defineConfig({
	title: 'DashLit',
	description: 'DashLit - A simple solution for self-hosting your home page.',
	head: [['link', { rel: 'icon', href: '/favicon.png' }]],
	themeConfig: {
		search: {
			provider: 'local'
		},

		logo: {
			src: '/logo.png',
			innerWidth: 50,
			height: 50
		},
		nav: [
			{ text: 'Home', link: '/' },
			{ text: 'Getting Started', link: '/guide/getting-started' },
			{ text: 'Demo', link: 'https://demo.dashlit.cwec.dev' },
			{
				text: getVersion(),
				items: [
					{ text: 'Changelog', link: '/changelog' },
					{ text: 'Contributing', link: '/contributing' }
				]
			}
		],
		socialLinks: [{ icon: 'github', link: 'https://github.com/codewec/dashlit' }],
		sidebar: [
			{
				text: 'Guide',
				base: '/guide',
				items: [
					{ text: 'What is DashLit?', link: '/what-is' },
					{ text: 'Getting Started', link: '/getting-started' }
				]
			},
			{ text: 'Contributing', link: '/contributing' },
			{ text: 'Changelog', link: '/changelog' },
			{ text: 'License', link: '/license' }
		],

		editLink: {
			pattern: 'https://github.com/codewec/dashlit/edit/main/docs/:path',
			text: 'Edit this page on GitHub'
		},

		lastUpdated: {
			text: 'Last updated',
			formatOptions: {
				dateStyle: 'short',
				timeStyle: 'medium'
			}
		},
		footer: {
			message: 'Released under the MIT License.',
			copyright: 'Copyright © 2025 CodeWec'
		}
	}
});
