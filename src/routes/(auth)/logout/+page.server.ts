import type { PageServerLoad } from './$types';
import { redirect } from '@sveltejs/kit';

export const load: PageServerLoad = async (event) => {
	event.cookies.delete('token', { path: '/' });
	event.locals.userAuthenticated = false;
	return redirect(302, '/login');
};
