import { cookie_token_key } from '$lib';
import type { PageServerLoad } from './$types';
import { redirect } from '@sveltejs/kit';

export const load: PageServerLoad = async (event) => {
	event.cookies.delete(cookie_token_key, { path: '/' });
	event.locals.userAuthenticated = false;
	return redirect(302, '/login');
};
