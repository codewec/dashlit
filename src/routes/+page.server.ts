import type { PageServerLoad } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import fs from 'node:fs/promises';
import { env } from '$env/dynamic/private';
import type { Dashboard } from '$lib/types';
import { file_path, version } from '$lib';

export const load: PageServerLoad = async (event) => {
	if (!event.locals.userAuthenticated) {
		return redirect(302, '/login');
	}
	const path = `${file_path}/dashboard.json`;
	const data = await fs.readFile(path, { encoding: 'utf8' }).catch((error) => {
		console.log(error);
		return '{}';
	});
	const dashboard: Dashboard = JSON.parse(data);
	return { groups: dashboard.groups, canLogout: env.PASSWORD?.length > 0 };
};

export const actions = {
	default: async ({ locals, request }) => {
		if (!locals.userAuthenticated) {
			return redirect(302, '/login');
		}
		const json = await request.json();

		const path = `${file_path}/dashboard.json`;
		await fs.mkdir(file_path, { recursive: true }).catch(console.error);
		await fs.writeFile(path, JSON.stringify({ version: version, groups: json })).catch((error) => {
			console.log(error);
			return fail(500);
		});
		return { status: 'ok' };
	}
};
