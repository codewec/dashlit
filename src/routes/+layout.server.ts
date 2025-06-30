import type { LayoutServerLoad } from './$types';
import { env } from '$env/dynamic/private';

export const load: LayoutServerLoad = async ({ locals }) => {
	return {
		envOrigin: env.ORIGIN ?? '',
		userAuthenticated: locals.userAuthenticated
	};
};
