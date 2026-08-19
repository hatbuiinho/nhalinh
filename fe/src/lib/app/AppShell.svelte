<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import { authStore } from '$lib/auth/auth-store.svelte';
	import { router } from '$lib/navigation/router.svelte';
	import LoginScreen from '$lib/screens/LoginScreen.svelte';
	import MemorialScreen from '$lib/screens/MemorialScreen.svelte';
	import StructureScreen from '$lib/screens/StructureScreen.svelte';
	import StatisticsScreen from '$lib/screens/StatisticsScreen.svelte';
	import UsersScreen from '$lib/screens/UsersScreen.svelte';
	import ProfileScreen from '$lib/screens/ProfileScreen.svelte';
	import LoadingIndicator from '$lib/ui/LoadingIndicator.svelte';
	import Popup from '$lib/ui/Popup.svelte';
	import { popupStore } from '$lib/ui/popup-store.svelte';
	import Toast from '$lib/ui/Toast.svelte';
	import { toastStore } from '$lib/ui/toast-store.svelte';
	import BottomNav from './BottomNav.svelte';
	import DesktopSidebar from './DesktopSidebar.svelte';
	import TopBar from './TopBar.svelte';
	import ChangePasswordPopup from '$lib/auth/ChangePasswordPopup.svelte';
	import { parseRoute, routePermission } from '$lib/navigation/routes';

	let requestedRoute = $derived(router.current);
	let requestedPermission = $derived(routePermission(requestedRoute));
	let routeAllowed = $derived(!requestedPermission || authStore.can(requestedPermission));
	let route = $derived(routeAllowed ? requestedRoute : parseRoute('/memorial'));

	onMount(async () => {
		router.init();
		await authStore.init();
	});

	onDestroy(() => router.destroy());

	$effect(() => {
		if (authStore.user && !routeAllowed) router.replace('/memorial');
	});
</script>

{#if authStore.initializing}
	<div class="grid min-h-screen place-items-center">
		<LoadingIndicator label="Đang khởi động..." />
	</div>
{:else if !authStore.user}
	<LoginScreen />
{:else}
	<div
		class="mx-auto flex h-dvh max-w-[1600px] bg-[var(--color-bg)] text-[var(--color-text)] md:border-x md:border-[var(--color-border)]"
	>
		<DesktopSidebar {route} />
		<div class="flex min-w-0 flex-1 flex-col">
			<TopBar {route} />
			<section class="relative min-h-0 flex-1 overflow-hidden">
				{#if route.name === 'memorial'}
					<MemorialScreen />
				{:else if route.name === 'structure'}
					<StructureScreen />
				{:else if route.name === 'statistics'}
					<StatisticsScreen />
				{:else if route.name === 'users'}
					<UsersScreen />
				{:else if route.name === 'profile'}
					<ProfileScreen />
				{/if}
			</section>
			<BottomNav {route} />
		</div>
	</div>
{/if}

<Toast
	open={toastStore.open}
	message={toastStore.message}
	tone={toastStore.tone}
	onClose={() => toastStore.close()}
/>
<ChangePasswordPopup />
<Popup open={popupStore.open} title={popupStore.title} onClose={() => popupStore.cancel()}>
	<p class="text-sm leading-6 text-[var(--color-text-secondary)]">{popupStore.message}</p>
	{#snippet footer()}
		<div class="grid grid-cols-2 gap-3">
			<button
				type="button"
				class="h-11 rounded-md border border-[var(--color-border-strong)] bg-[var(--color-surface)] text-sm font-semibold"
				onclick={() => popupStore.cancel()}>{popupStore.cancelLabel}</button
			>
			<button
				type="button"
				class={[
					'h-11 rounded-md text-sm font-semibold text-white',
					popupStore.tone === 'danger' ? 'bg-[var(--color-danger)]' : 'bg-[var(--color-primary)]'
				]}
				onclick={() => popupStore.accept()}>{popupStore.confirmLabel}</button
			>
		</div>
	{/snippet}
</Popup>
