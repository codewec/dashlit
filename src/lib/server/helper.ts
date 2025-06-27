import { env } from '$env/dynamic/private';
import { hashString } from '$lib/helpers';

export const getSecretKey = async (password: string) => {
	return (env.SECRET_KEY?.length ?? 0 > 0) ? env.SECRET_KEY : await hashString(password);
};
