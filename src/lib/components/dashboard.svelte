<script lang="ts">
	import { ActionType, ShowUrlType, type Group, type Item } from '$lib/types';
	import Icon from '@iconify/svelte';
	import { droppable, draggable, type DragDropState } from '@thisux/sveltednd';
	import { flip } from 'svelte/animate';
	import { fade } from 'svelte/transition';
	import ActionButtons from './actionButtons.svelte';
	import { getIds, hasField, isUrlString } from '$lib/helpers';
	import EmptyItem from './emptyItem.svelte';
	import EmptyGroup from './emptyGroup.svelte';
	import { newGroup, newItem } from '$lib/factory';
	import { on } from 'svelte/events';

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

	let hoveredOnActionsEnitytId = $state<string | undefined>(undefined); // if hover on actions buttons on edit mode
	let hoveredItemId = $state<string | undefined>(undefined); // if hover on item (not group) on not edit mode
	let disableGroupsDrag = $state(true);
	let disableItemDrag = $state(true);

	$effect(() => {
		disableGroupsDrag = !editMode;
		disableItemDrag = !editMode;
	});

	const getHoverDescription = (groupId: string, item: Item) => {
		if (!hoveredItemId) {
			return item.description;
		}
		const ids = getIds(hoveredItemId);
		if (!ids) {
			return item.description;
		}

		if (ids.groupId === groupId && ids.itemId === item.id) {
			if (item.showUrl === ShowUrlType.HOVER) {
				return item.url;
			}
		}
		return item.description;
	};

	const getDescription = (groupId: string, item: Item) => {
		switch (item.showUrl) {
			case ShowUrlType.NEVER:
				return item.description;
			case ShowUrlType.ALWAYS:
				return item.description ? item.description : item.url;
			case ShowUrlType.HOVER:
				return getHoverDescription(groupId, item);
			case ShowUrlType.DESC_EMPTY:
				return item.description ? item.description : item.url;
			default:
				return item.description ? item.description : item.url;
		}
	};

	const getUrl = (item: Item) => {
		switch (item.showUrl) {
			case ShowUrlType.NEVER:
				return undefined;
			case ShowUrlType.ALWAYS:
				return item.description ? item.url : undefined;
			case ShowUrlType.HOVER:
				return undefined;
			case ShowUrlType.DESC_EMPTY:
				return undefined;
			default:
				return undefined;
		}
	};

	const isDisabledGroupDrag = (id: string) => {
		if (hoveredOnActionsEnitytId) {
			const ids = getIds(hoveredOnActionsEnitytId);
			if (ids && ids.groupId == id) {
				return true;
			}
		}
		return disableGroupsDrag;
	};

	const isDisabledItemDrag = (id: string) => {
		if (id == hoveredOnActionsEnitytId) {
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

<div class="group-container grid grid-cols-1 gap-4">
	{#if editMode}
		<EmptyGroup handleClick={() => handleClickGroupAction(ActionType.CREATE, newGroup())} />
	{/if}
	{#each groups as group (`g_${group.id}`)}
		<div
			class:edit-mode={editMode}
			class="group rounded-md bg-gray-50 p-4 shadow-sm ring-1 ring-gray-200 dark:bg-slate-900 dark:ring-slate-800"
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
					<h2 class="title font-semibold text-gray-900 capitalize dark:text-gray-200">
						{group.title}
					</h2>
					{#if group.description}
						<p class="description text-xs text-gray-500">{group.description}</p>
					{/if}
				</div>
				{#if editMode}
					<ActionButtons
						id={`${group.id}-0`}
						handleHover={(id) => {
							hoveredOnActionsEnitytId = id;
						}}
						handleClick={(action) => handleClickGroupAction(action, group)}
					/>
				{/if}
			</div>

			<div
				class="item-container grid grid-cols-1 gap-2 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4"
			>
				{#each group.items as item (`i_${item.id}`)}
					<div
						tabindex="0"
						role="button"
						onkeyup={(e) => {
							if (e.key === 'Enter') {
								handleClickItem(item);
							}
						}}
						onmouseenter={() => {
							if (editMode) {
								hoveredItemId = undefined;
								return;
							}
							hoveredItemId = `${group.id}-${item.id}`;
						}}
						onmouseleave={() => {
							hoveredItemId = undefined;
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
								<div class="h-14 w-19">
									{#if isUrlString(item.icon)}
										<img
											src={item.icon}
											alt={item.title}
											class="h-full w-full rounded-full object-cover"
										/>
									{:else}
										<Icon
											color={item.iconColor ?? 'gray'}
											icon={item.icon}
											width={56}
											height={56}
										/>
									{/if}
								</div>
							{/if}
							<div class="w-full truncate">
								<h3 class="title font-medium text-gray-900 dark:text-gray-100">
									{item.title}
								</h3>
								{#if getDescription(group.id, item)}
									<p class="description text-sm text-gray-500">
										{getDescription(group.id, item)}
									</p>
								{/if}
								{#if getUrl(item)}
									<p class="url text-[10px] text-gray-400 dark:text-gray-500">
										{getUrl(item)}
									</p>
								{/if}
							</div>
						</div>
						<div class="absolute top-1 right-1">
							{#if editMode}
								<ActionButtons
									id={`${group.id}-${item.id}`}
									handleHover={(id) => {
										hoveredOnActionsEnitytId = id;
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
							hoveredOnActionsEnitytId = id;
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
		@apply !bg-blue-50 !ring-2 !ring-blue-400 dark:!bg-slate-800 dark:ring-blue-600;
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
