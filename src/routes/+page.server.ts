import type { PageServerLoad } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import fs from 'node:fs/promises';
import { env } from '$env/dynamic/private';
import type { Dashboard } from '$lib/types';
import { default_dashboard, data_path } from '$lib';
import { getVersion } from '$lib/server/helper';

const dataPath = () => {
	return env.DATA_PATH ?? data_path;
};

const filePath = () => {
	return `${dataPath()}/${default_dashboard}`;
};

export const load: PageServerLoad = async (event) => {
	if (!event.locals.userAuthenticated) {
		return redirect(302, '/login');
	}
	const data = await fs.readFile(filePath(), { encoding: 'utf8' }).catch(() => {
		console.log(`File ${filePath()} not found`);
		return '{}';
	});
	const dashboard: Dashboard = JSON.parse(data);
	return {
		groups: dashboard.groups,
		canLogout: (env.PASSWORD?.length ?? 0) > 0,
		isDemoMode: env.DEMO_MODE === 'true'
	};
};

export const actions = {
	default: async ({ locals, request }) => {
		if (!locals.userAuthenticated) {
			return redirect(302, '/login');
		}
		const json = await request.json();

		await fs.mkdir(dataPath(), { recursive: true }).catch(console.error);
		await fs
			.writeFile(filePath(), JSON.stringify({ version: getVersion(), groups: json }))
			.catch((error) => {
				console.log(`Cant write ${filePath()}`);
				return fail(500);
			});
		return { status: 'ok' };
	}
};
