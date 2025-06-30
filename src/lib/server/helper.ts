import { env } from '$env/dynamic/private';
import { hashString } from '$lib/helpers';
import currentPackage from '../../../package.json';
import { default_dashboard, data_path } from '$lib';

export const getSecretKey = async (password: string) => {
	return (env.SECRET_KEY?.length ?? 0 > 0) ? env.SECRET_KEY : await hashString(password);
};

export const getVersion = () => {
	return currentPackage.version || '0.0.0';
};

export const dataPath = () => {
	return env.DATA_PATH ?? data_path;
};

export const filePath = () => {
	return `${dataPath()}/${default_dashboard}`;
};
