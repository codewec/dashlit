import { generateRandomString } from './helpers';
import type { Group, Item } from './types';

export const newGroup = (): Group => {
	return {
		id: generateRandomString(10),
		title: '',
		items: []
	};
};

export const newItem = (): Item => {
	return {
		id: generateRandomString(10),
		title: '',
		description: '',
		url: ''
	};
};
