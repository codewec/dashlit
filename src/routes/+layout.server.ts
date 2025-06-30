import type { LayoutServerLoad } from './$types';
import { env } from '$env/dynamic/private';

export const load: LayoutServerLoad = async ({ locals, url }) => {
	const origin = env.ORIGIN ?? '';
	return {
		originWarning: {
			currentOrigin: url.origin,
			isWarning: origin !== url.origin
		},
		userAuthenticated: locals.userAuthenticated
	};
};
