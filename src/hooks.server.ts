import type { Handle } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
import * as jose from 'jose';
import { hashString } from '$lib/helpers';
import { cookie_token_key } from '$lib';
import { getSecretKey } from '$lib/server/helper';

export const handle: Handle = async ({ event, resolve }) => {
	if (!env.PASSWORD || env.PASSWORD?.length === 0) {
		event.locals.userAuthenticated = true;
		return resolve(event);
	}

	const token = event.cookies.get(cookie_token_key);
	if (!token) {
		console.log('!!!!! not has token', token);
		return resolve(event);
	}

	const secret = await getSecretKey(env.PASSWORD);
	await jose
		.jwtVerify(token, new TextEncoder().encode(secret))
		.then(() => {
			event.locals.userAuthenticated = true;
		})
		.catch(() => {
			console.log('!!!!! cant check jwt', token);
			event.cookies.delete(cookie_token_key, { path: '/' });
		});

	return resolve(event);
};
