import { env } from '$env/dynamic/private';
import { hashString } from '$lib/helpers';
import currentPackage from '../../../package.json';

export const getSecretKey = async (password: string) => {
	return (env.SECRET_KEY?.length ?? 0 > 0) ? env.SECRET_KEY : await hashString(password);
};

export const getVersion = () => {
	return currentPackage.version || '0.0.0';
};
