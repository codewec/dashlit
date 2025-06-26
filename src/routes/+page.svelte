<script lang="ts">
	import '$lib/styles/dnd.css';
	import {
		type Item,
		type Group,
		type DeletionEntity,
		ActionType,
		type EditableItem
	} from '$lib/types';
	import Header from '$lib/components/header.svelte';
	import Dashboard from '$lib/components/dashboard.svelte';
	import ModalDelete from '$lib/components/modalDelete.svelte';
	import ModalFormItem from '$lib/components/modalFormItem.svelte';
	import ModalFormGroup from '$lib/components/modalFormGroup.svelte';
	import { newGroup, newItem } from '$lib/factory.js';

	let { data } = $props();

	let groups = $state(data.groups ?? []);
	let editMode = $state(false);
	let deletionEntity = $state<DeletionEntity | undefined>(undefined);
	let editableGroup = $state<Group | undefined>(undefined);
	let editableItem = $state<EditableItem | undefined>(undefined);

	$effect(() => {
		if (groups.length === 0) {
			editMode = true;
		}
	});

	// dashboard
	const handleSaveDashboard = async () => {
		await fetch('', {
			method: 'POST',
			body: JSON.stringify(groups)
		}).then(() => {
			editMode = false;
		});
	};

	const handleDeleteEntity = () => {
		if (deletionEntity === undefined) {
			return;
		}

		if (deletionEntity.ids.groupId.length === 0 && deletionEntity.ids.itemId.length === 0) {
			deletionEntity = undefined;
			return;
		}

		const isGroup = deletionEntity.ids.itemId.length === 0;

		if (isGroup) {
			const groupIndex = groups.findIndex((s) => s.id === deletionEntity?.ids.groupId);
			if (groupIndex === undefined) {
				console.log('Group index not found');
				return;
			}
			groups.splice(groupIndex, 1);
		} else {
			const group = groups.find((g) => g.id === deletionEntity?.ids.groupId);
			const itemIndex = group?.items.findIndex((i) => i.id === deletionEntity?.ids.itemId);
			if (itemIndex === undefined) {
				console.log('Item index not found');
				return;
			}
			group?.items.splice(itemIndex, 1);
		}
		deletionEntity = undefined;
	};

	const handleSaveGroup = (group: Group) => {
		let g = groups.find((g) => g.id === group.id);
		if (g) {
			g.title = group.title;
			g.description = group.description;
		} else {
			groups.unshift(group);
		}
		editableGroup = undefined;
	};

	const handleSaveItem = (groupId: string, item: Item) => {
		let group = groups.find((g) => g.id === groupId);
		const i = group?.items.find((i) => i.id === item.id);
		if (i) {
			i.title = item.title;
			i.description = item.description;
			i.url = item.url;
			i.icon = item.icon;
		} else {
			group?.items.push(item);
		}
		editableItem = undefined;
	};

	// groups
	const handleActionGroup = (type: ActionType, group: Group) => {
		switch (type) {
			case ActionType.CREATE:
				editableGroup = group;
				break;
			case ActionType.EDIT:
				editableGroup = { ...group };
				break;
			case ActionType.DELETE:
				deletionEntity = {
					ids: { groupId: group.id, itemId: '' },
					element: { ...group }
				};
				break;
		}
	};

	// items
	const handleActionItem = (type: ActionType, groupId: string, item: Item) => {
		switch (type) {
			case ActionType.CREATE:
				editableItem = {
					groupId: groupId,
					item: item
				};
				break;
			case ActionType.EDIT:
				editableItem = {
					groupId: groupId,
					item: { ...item }
				};
				break;
			case ActionType.DELETE:
				deletionEntity = {
					ids: { groupId: groupId, itemId: item.id },
					element: { ...item }
				};
				break;
		}
	};

	const handleClickItem = (item: Item) => {
		if (editMode) {
			return;
		}
		window.open(item.url, '_blank');
	};
</script>

<div class="bg-gray-50 p-4">
	<Header
		{editMode}
		canLogout={data.canLogout}
		handleSave={handleSaveDashboard}
		handleEdit={() => (editMode = !editMode)}
	/>

	<Dashboard
		{editMode}
		{groups}
		{handleClickItem}
		handleClickItemAction={handleActionItem}
		handleClickGroupAction={handleActionGroup}
	/>
</div>

<ModalDelete
	entity={deletionEntity}
	handleClose={() => (deletionEntity = undefined)}
	handleConfirm={handleDeleteEntity}
/>

<ModalFormGroup
	isOpen={editableGroup !== undefined}
	group={editableGroup ?? newGroup()}
	handleClose={(group) => {
		if (group) {
			handleSaveGroup(group);
		} else {
			editableGroup = undefined;
		}
	}}
/>

<ModalFormItem
	isOpen={editableItem !== undefined}
	item={editableItem?.item ?? newItem()}
	handleClose={(item) => {
		if (item && editableItem) {
			handleSaveItem(editableItem.groupId, item);
		} else {
			editableGroup = undefined;
		}
	}}
/>
