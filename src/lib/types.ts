export enum ActionType {
	CREATE = 'create',
	DELETE = 'delete',
	EDIT = 'edit'
}

export enum ShowUrlType {
	NEVER = 'never',
	ALWAYS = 'always',
	DESC_EMPTY = 'empty_desc',
	HOVER = 'hover'
}

export interface Item {
	id: string;
	title: string;
	url: string;
	showUrl?: ShowUrlType;
	target?: string;
	description?: string;
	icon?: string;
	iconColor?: string;
}

export interface Group {
	id: string;
	title: string;
	description?: string;
	items: Item[];
}

export interface Dashboard {
	version: string;
	groups: Group[];
}

export interface Ids {
	groupId: string;
	itemId: string;
}

export type DeletionEntity = {
	ids: Ids;
	element: Group | Item;
};

export type EditableItem = {
	groupId: string;
	item: Item;
};
