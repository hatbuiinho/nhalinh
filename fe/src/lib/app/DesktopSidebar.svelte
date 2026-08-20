<script lang="ts">
	import { onMount } from 'svelte';
	import { authStore } from '$lib/auth/auth-store.svelte';
	import { changePasswordPopupStore } from '$lib/auth/change-password-popup-store.svelte';
	import { bottomNavItems, mainRouteFor, type AppRoute } from '$lib/navigation/routes';
	import { router } from '$lib/navigation/router.svelte';
	import Logo from '$lib/ui/Logo.svelte';

	let {
		route,
		collapsed = false
	}: {
		route: AppRoute;
		collapsed?: boolean;
	} = $props();
	let active = $derived(mainRouteFor(route));
	let profileMenuOpen = $state(false);
	let profileMenuRoot = $state<HTMLDivElement>();
	let navItems = $derived(
		bottomNavItems.filter((item) => item.name !== 'users' || authStore.can('user.read'))
	);

	$effect(() => {
		void route.path;
		profileMenuOpen = false;
	});

	onMount(() => {
		function closeOutside(event: PointerEvent) {
			if (event.target instanceof Node && !profileMenuRoot?.contains(event.target)) {
				profileMenuOpen = false;
			}
		}
		function closeOnEscape(event: KeyboardEvent) {
			if (event.key === 'Escape') profileMenuOpen = false;
		}
		document.addEventListener('pointerdown', closeOutside);
		document.addEventListener('keydown', closeOnEscape);
		return () => {
			document.removeEventListener('pointerdown', closeOutside);
			document.removeEventListener('keydown', closeOnEscape);
		};
	});

	function viewProfile() {
		profileMenuOpen = false;
		router.push('/profile');
	}

	function changePassword() {
		profileMenuOpen = false;
		changePasswordPopupStore.show();
	}

	function logout() {
		profileMenuOpen = false;
		void authStore.logout();
	}
</script>

<aside
	class={[
		'hidden h-dvh shrink-0 flex-col border-r border-[var(--color-border)] bg-[var(--color-surface)] transition-[width] duration-200 md:flex',
		collapsed ? 'w-20' : 'w-60'
	]}
>
	<div
		class={[
			'flex h-[73px] items-center border-b border-[var(--color-border)]',
			collapsed ? 'justify-center px-2' : 'px-5'
		]}
	>
		{#if !collapsed}
			<Logo />
		{:else}
			<Logo compact />
		{/if}
	</div>

	<nav class={['flex-1 space-y-1 py-5', collapsed ? 'px-2' : 'px-3']} aria-label="Điều hướng chính">
		{#each navItems as item (item.name)}
			<button
				type="button"
				class={[
					'flex h-11 w-full items-center rounded-md text-sm font-semibold',
					collapsed ? 'justify-center px-2' : 'gap-3 px-3',
					active === item.name
						? 'bg-[var(--color-primary-soft)] text-[var(--color-primary-dark)]'
						: 'text-[var(--color-text-secondary)] hover:bg-[var(--color-surface-muted)]'
				]}
				aria-current={active === item.name ? 'page' : undefined}
				aria-label={item.label}
				title={item.label}
				onclick={() => router.openMain(item.name)}
			>
				<span class={['h-5 w-5 shrink-0', item.icon]} aria-hidden="true"></span>
				{#if !collapsed}
					<span>{item.label}</span>
				{/if}
			</button>
		{/each}
	</nav>

	<div class={['border-t border-[var(--color-border)]', collapsed ? 'p-2' : 'p-3']}>
		<div class="relative" bind:this={profileMenuRoot}>
			{#if profileMenuOpen}
				<div
					class={[
						'absolute bottom-full z-40 mb-2 overflow-hidden rounded-md border border-[var(--color-border)] bg-[var(--color-surface)] py-1 shadow-[var(--shadow-popover)]',
						collapsed ? 'left-full ml-2 w-60' : 'left-0 w-full'
					]}
					role="menu"
				>
					<button
						type="button"
						role="menuitem"
						class="flex h-11 w-full items-center gap-3 px-3 text-left text-sm hover:bg-[var(--color-surface-muted)]"
						onclick={viewProfile}
					>
						<span class="icon-[lucide--user-round] h-4.5 w-4.5" aria-hidden="true"></span>
						Xem hồ sơ
					</button>
					<button
						type="button"
						role="menuitem"
						class="flex h-11 w-full items-center gap-3 px-3 text-left text-sm hover:bg-[var(--color-surface-muted)]"
						onclick={changePassword}
					>
						<span class="icon-[lucide--key-round] h-4.5 w-4.5" aria-hidden="true"></span>
						Đổi mật khẩu
					</button>
					<button
						type="button"
						role="menuitem"
						class="flex h-11 w-full items-center gap-3 border-t border-[var(--color-border)] px-3 text-left text-sm text-[var(--color-danger)] hover:bg-[var(--color-surface-muted)]"
						onclick={logout}
					>
						<span class="icon-[lucide--log-out] h-4.5 w-4.5" aria-hidden="true"></span>
						Đăng xuất
					</button>
				</div>
			{/if}
			<button
				type="button"
				class={[
					'flex min-h-14 w-full rounded-md px-2 py-2 text-left hover:bg-[var(--color-surface-muted)]',
					collapsed ? 'justify-center' : 'items-center gap-3'
				]}
				aria-label="Mở menu cá nhân"
				title={collapsed ? authStore.user?.display_name : undefined}
				aria-haspopup="menu"
				aria-expanded={profileMenuOpen}
				onclick={() => (profileMenuOpen = !profileMenuOpen)}
			>
				<span
					class="grid h-9 w-9 shrink-0 place-items-center rounded-full bg-[var(--color-primary-soft)] text-sm font-semibold text-[var(--color-primary-dark)]"
				>
					{#if authStore.user?.avatar_url}<img
							src={authStore.user.avatar_url}
							alt=""
							class="h-full w-full rounded-full object-cover"
						/>{:else}{authStore.user?.display_name.slice(0, 1).toUpperCase() || '?'}{/if}
				</span>
				{#if !collapsed}
					<span class="min-w-0 flex-1">
						<span class="block truncate text-sm font-semibold">{authStore.user?.display_name}</span>
						<span class="block truncate text-xs text-[var(--color-text-secondary)]">
							@{authStore.user?.username}
						</span>
					</span>
					<span
						class={profileMenuOpen
							? 'icon-[lucide--chevron-down] h-4 w-4 text-[var(--color-text-muted)]'
							: 'icon-[lucide--chevron-up] h-4 w-4 text-[var(--color-text-muted)]'}
						aria-hidden="true"
					></span>
				{/if}
			</button>
		</div>
	</div>
</aside>
