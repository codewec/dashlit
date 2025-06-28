<script lang="ts">
	import { ActionType, type Group, type Item } from '$lib/types';
	import Icon from '@iconify/svelte';
	import { droppable, draggable, type DragDropState } from '@thisux/sveltednd';
	import { flip } from 'svelte/animate';
	import { fade } from 'svelte/transition';
	import ActionButtons from './actionButtons.svelte';
	import { getIds, hasField } from '$lib/helpers';
	import EmptyItem from './emptyItem.svelte';
	import EmptyGroup from './emptyGroup.svelte';
	import { newGroup, newItem } from '$lib/factory';

	const {
		editMode,
		groups,
		handleClickItem,
		handleClickItemAction,
		handleClickGroupAction
	}: {
		editMode: boolean;
		groups: Group[];
		handleClickItem: (item: Item) => void;
		handleClickItemAction: (type: ActionType, groupId: string, item: Item) => void;
		handleClickGroupAction: (type: ActionType, group: Group) => void;
	} = $props();

	let hoveredId = $state<string | undefined>(undefined);
	let disableGroupsDrag = $state(true);
	let disableItemDrag = $state(true);

	$effect(() => {
		disableGroupsDrag = !editMode;
		disableItemDrag = !editMode;
	});

	const isDisabledGroupDrag = (id: string) => {
		if (hoveredId) {
			const ids = getIds(hoveredId);
			if (ids && ids.groupId == id) {
				return true;
			}
		}
		return disableGroupsDrag;
	};

	const isDisabledItemDrag = (id: string) => {
		if (id == hoveredId) {
			return true;
		}
		return disableItemDrag;
	};

	const isDisabledGroupDrop = (group: Group) => {
		if (group.items.length === 0) {
			return false;
		}
		return disableGroupsDrag;
	};

	const onDropInGroup = (state: DragDropState<Item | Group>) => {
		const { draggedItem, sourceContainer, targetContainer } = state;
		if (hasField(draggedItem, 'url')) {
			state.targetContainer = `${targetContainer}-0`;
			onDropInItem(state as DragDropState<Item>);
		} else {
			if (!targetContainer) {
				console.log('Target container not found');
				return;
			}

			const sourceIndex = groups.findIndex((t) => t.id === sourceContainer);
			const targetIndex = groups.findIndex((t) => t.id === targetContainer);

			if (sourceIndex === undefined || targetIndex === undefined) {
				console.log('Source or target index not found');
				return;
			}

			groups.splice(sourceIndex, 1);
			groups.splice(targetIndex, 0, draggedItem);
		}
	};

	function onDropInItem(state: DragDropState<Item>) {
		const { draggedItem, sourceContainer, targetContainer } = state;
		if (!targetContainer) {
			console.log('Target container not found');
			return;
		}

		const sourceIds = getIds(sourceContainer);
		if (!sourceIds) {
			console.log('Source IDs not found');
			return;
		}

		const targetIds = getIds(targetContainer);
		if (!targetIds) {
			console.log('Target IDs not found');
			return;
		}

		const sourceGroup = groups.find((g) => g.id === sourceIds.groupId);
		const sourceIndex = sourceGroup?.items.findIndex((t) => t.id === sourceIds.itemId);

		const targetGroup = groups.find((g) => g.id === targetIds.groupId);
		const targetIndex = targetGroup?.items.findIndex((t) => t.id === targetIds.itemId);

		if (sourceIndex === undefined || targetIndex === undefined) {
			console.log('Source or target index not found');
			return;
		}

		sourceGroup?.items.splice(sourceIndex, 1);
		targetGroup?.items.splice(targetIndex, 0, draggedItem);
	}
</script>

<div class="grid grid-cols-1 gap-4">
	{#if editMode}
		<EmptyGroup handleClick={() => handleClickGroupAction(ActionType.CREATE, newGroup())} />
	{/if}
	{#each groups as group (`g_${group.id}`)}
		<div
			class:edit-mode={editMode}
			class="rounded-md bg-gray-50 p-4 shadow-sm ring-1 ring-gray-200 dark:bg-slate-900 dark:ring-slate-800"
			use:draggable={{
				container: group.id,
				dragData: group,
				disabled: isDisabledGroupDrag(group.id),
				callbacks: {
					onDragStart: () => (disableItemDrag = true),
					onDragEnd: () => (disableItemDrag = false)
				}
			}}
			use:droppable={{
				dragData: group,
				container: `${group.id}`,
				disabled: isDisabledGroupDrop(group),
				callbacks: {
					onDrop: onDropInGroup
				}
			}}
		>
			<div class="mb-4 flex items-center justify-between">
				<div class="inline-flex gap-2">
					<h2 class="font-semibold text-gray-900 capitalize dark:text-gray-200">
						{group.title}
					</h2>
					{#if group.description}
						<p class="text-xs text-gray-500">{group.description}</p>
					{/if}
				</div>
				{#if editMode}
					<ActionButtons
						id={`${group.id}-0`}
						handleHover={(id) => {
							hoveredId = id;
						}}
						handleClick={(action) => handleClickGroupAction(action, group)}
					/>
				{/if}
			</div>

			<div class="grid grid-cols-1 gap-2 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
				{#each group.items as item (`i_${item.id}`)}
					<div
						tabindex="0"
						role="button"
						onkeyup={(e) => {
							if (e.key === 'Enter') {
								handleClickItem(item);
							}
						}}
						onclick={() => {
							handleClickItem(item);
						}}
						use:draggable={{
							container: `${group.id}-${item.id}`,
							dragData: item,
							disabled: isDisabledItemDrag(`${group.id}-${item.id}`),
							callbacks: {
								onDragStart: () => (disableGroupsDrag = true),
								onDragEnd: () => (disableGroupsDrag = false)
							}
						}}
						use:droppable={{
							dragData: item,
							disabled: disableItemDrag,
							container: `${group.id}-${item.id}`,
							callbacks: {
								onDrop: onDropInItem
							}
						}}
						animate:flip={{ duration: 200 }}
						in:fade={{ duration: 150 }}
						out:fade={{ duration: 150 }}
						class="item svelte-dnd-touch-feedback"
					>
						<div class="flex items-center gap-2">
							{#if item.icon}
								<div class="h-14 w-14">
									<Icon color="gray" icon={item.icon} height={56} />
								</div>
							{/if}
							<div>
								<h3 class="font-medium text-gray-900 dark:text-gray-100">
									{item.title}
								</h3>
								<p class="text-sm text-gray-500">
									{item.description}
								</p>
							</div>
						</div>
						<div class="absolute top-1 right-1">
							{#if editMode}
								<ActionButtons
									id={`${group.id}-${item.id}`}
									handleHover={(id) => {
										hoveredId = id;
									}}
									handleClick={(action) => handleClickItemAction(action, group.id, item)}
								/>
							{/if}
						</div>
					</div>
				{/each}
				{#if editMode}
					<EmptyItem
						id={`${group.id}-0`}
						handleHover={(id) => {
							hoveredId = id;
						}}
						handleClick={() => handleClickItemAction(ActionType.CREATE, group.id, newItem())}
					/>
				{/if}
			</div>
		</div>
	{/each}
</div>

<style lang="postcss">
	@reference "$lib/../app.css";
	:global(.dragging) {
		@apply !opacity-50 !shadow-lg !ring-2 !ring-blue-400;
	}

	:global(.drag-over) {
		@apply !bg-blue-50 !ring-2 !ring-blue-400;
	}

	.item {
		@apply relative rounded-lg bg-white p-3 shadow-sm ring-1 ring-gray-200 transition-all duration-200 dark:bg-black dark:ring-gray-800;
	}

	.item:not(.edit-mode) {
		@apply cursor-pointer hover:shadow-md hover:ring-2 hover:ring-blue-300 dark:hover:ring-blue-900;
	}

	.edit-mode {
		@apply cursor-move hover:shadow-md hover:ring-2 hover:ring-blue-200 dark:hover:ring-blue-900;

		.item {
			@apply cursor-move hover:shadow-md hover:ring-2 hover:ring-blue-200 dark:hover:ring-blue-900;
		}
	}
</style>
