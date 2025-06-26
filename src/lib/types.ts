export enum ActionType {
	CREATE = 'create',
	DELETE = 'delete',
	EDIT = 'edit'
}

export interface Item {
	id: string;
	title: string;
	description?: string;
	icon?: string;
	url: string;
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
