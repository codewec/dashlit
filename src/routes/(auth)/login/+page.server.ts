import type { PageServerLoad } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
import * as jose from 'jose';
import { hashString } from '$lib/helpers';

export const load: PageServerLoad = async (event) => {
	if (event.locals.userAuthenticated) {
		return redirect(302, '/');
	}

	return {};
};

export const actions = {
	login: async ({ request, cookies }) => {
		const form = await request.formData();
		let password = String(form.get('password'));

		if (!password) {
			return fail(400, { error: 'Password is required' });
		}

		if (password !== env.PASSWORD) {
			return fail(401, { error: 'Invalid password' });
		}

		const secret = env.SECRET_KEY?.length > 0 ? env.SECRET_KEY : await hashString(password);
		const secretKey = new TextEncoder().encode(secret);

		const token = await new jose.SignJWT()
			.setProtectedHeader({ alg: 'HS256' })
			.setIssuedAt()
			.setExpirationTime('4weeks')
			.sign(secretKey);

		cookies.set('token', token, {
			path: '/',
			httpOnly: true,
			sameSite: 'strict',
			secure: process.env.NODE_ENV === 'production',
			maxAge: 60 * 60 * 24 * 30
		});

		redirect(302, '/');
	}
};
