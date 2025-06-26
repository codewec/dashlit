import type { Ids } from './types';

export const hasField = <T>(data: any, key: string): data is T => {
	return key in data;
};

export const getIds = (str: string): Ids | undefined => {
	const parts = str.split('-');
	if (parts.length !== 2) {
		return undefined;
	}
	return {
		groupId: parts[0],
		itemId: parts[1]
	};
};

export const hashString = async (message: string): Promise<string> => {
	const msgBuffer = new TextEncoder().encode(message);
	const hashBuffer = await crypto.subtle.digest('SHA-256', msgBuffer);
	const hashArray = Array.from(new Uint8Array(hashBuffer));
	const hashHex = hashArray.map((b) => b.toString(16).padStart(2, '0')).join('');
	return hashHex;
};

export const generateRandomString = (length: number) => {
	return Math.random()
		.toString(36)
		.substring(2, 2 + length);
};
