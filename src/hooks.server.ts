import type { Handle } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
import * as jose from 'jose';
import { hashString } from '$lib/helpers';

export const handle: Handle = async ({ event, resolve }) => {
	if (env.PASSWORD.length === 0) {
		event.locals.userAuthenticated = true;
		return resolve(event);
	}

	const token = event.cookies.get('token');
	if (!token) {
		return resolve(event);
	}

	const secret = env.SECRET_KEY?.length > 0 ? env.SECRET_KEY : await hashString(env.PASSWORD);
	await jose
		.jwtVerify(token, new TextEncoder().encode(secret))
		.then(() => {
			event.locals.userAuthenticated = true;
		})
		.catch(() => {
			event.cookies.delete('token', { path: '/' });
		});

	return resolve(event);
};
