import type { PageServerLoad } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
import * as jose from 'jose';
import { cookie_token_key } from '$lib';
import { getSecretKey } from '$lib/server/helper';

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

		const secretKey = await getSecretKey(password);
		const sign = new TextEncoder().encode(secretKey);

		const token = await new jose.SignJWT()
			.setProtectedHeader({ alg: 'HS256' })
			.setIssuedAt()
			.setExpirationTime('4weeks')
			.sign(sign);

		cookies.set(cookie_token_key, token, {
			path: '/',
			httpOnly: true,
			sameSite: 'lax',
			secure: process.env.NODE_ENV === 'production',
			maxAge: 60 * 60 * 24 * 30
		});

		redirect(302, '/');
	}
};
