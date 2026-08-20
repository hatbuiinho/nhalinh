<script lang="ts">
	import { onMount } from 'svelte';
	import type { AppRoute } from '$lib/navigation/routes';
	import { router } from '$lib/navigation/router.svelte';
	import { authStore } from '$lib/auth/auth-store.svelte';
	import Logo from '$lib/ui/Logo.svelte';
	import { changePasswordPopupStore } from '$lib/auth/change-password-popup-store.svelte';

	let {
		route,
		collapsed = false,
		ontoggle
	}: {
		route: AppRoute;
		collapsed?: boolean;
		ontoggle?: () => void;
	} = $props();
	let profileMenuOpen = $state(false);
	let profileMenuRoot = $state<HTMLDivElement>();
	let showBack = $derived(route.name === 'profile');
	let desktopTitle = $derived(route.title);

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

<header
	class="z-20 border-b border-[var(--color-border)] bg-[rgb(255_255_255_/_0.96)] px-4 pt-[max(env(safe-area-inset-top),0.75rem)] backdrop-blur md:px-6 md:pt-0 lg:px-8"
>
		<div class="flex h-14 items-center justify-between gap-3 md:h-[72px]">
		<div class="flex min-w-0 items-center gap-2">
			{#if showBack}
				<button
					type="button"
					class="grid h-10 w-10 place-items-center rounded-full border border-[var(--color-border-strong)] md:hidden"
					aria-label="Quay lại"
					onclick={() => router.back()}
					><span class="icon-[lucide--chevron-left] h-5 w-5" aria-hidden="true"></span></button
				>
			{/if}
			<span class="md:hidden"><Logo compact /></span>
			<button
				type="button"
				class="hidden h-10 w-10 place-items-center rounded-md border border-[var(--color-border-strong)] text-[var(--color-text-secondary)] hover:bg-[var(--color-surface-muted)] hover:text-[var(--color-primary-dark)] md:grid"
				aria-label={collapsed ? 'Mở rộng thanh điều hướng' : 'Thu gọn thanh điều hướng'}
				title={collapsed ? 'Mở rộng thanh điều hướng' : 'Thu gọn thanh điều hướng'}
				onclick={ontoggle}
			>
				<span class="icon-[lucide--panel-left] h-5 w-5" aria-hidden="true"></span>
			</button>
			<p class="truncate text-base font-semibold md:text-lg">
				<span class="md:hidden">{route.title}</span>
				<span class="hidden md:inline">{desktopTitle}</span>
			</p>
		</div>
		<div class="flex shrink-0 items-center gap-2 md:hidden">
			<div class="relative" bind:this={profileMenuRoot}>
				<button
					type="button"
					class="grid h-10 w-10 place-items-center rounded-full border border-[var(--color-border-strong)] bg-[var(--color-primary-soft)] text-sm font-semibold text-[var(--color-primary-dark)]"
					aria-label="Mở menu cá nhân"
					aria-haspopup="menu"
					aria-expanded={profileMenuOpen}
					onclick={() => (profileMenuOpen = !profileMenuOpen)}
				>
					{#if authStore.user?.avatar_url}<img
							src={authStore.user.avatar_url}
							alt=""
							class="h-full w-full rounded-full object-cover"
						/>{:else}{authStore.user?.display_name.slice(0, 1).toUpperCase() || '?'}{/if}
				</button>
				{#if profileMenuOpen}
					<div
						class="absolute top-12 right-0 z-40 w-60 overflow-hidden rounded-md border border-[var(--color-border)] bg-[var(--color-surface)] py-1 shadow-[var(--shadow-popover)]"
						role="menu"
					>
						<div class="border-b border-[var(--color-border)] px-3 py-2.5">
							<p class="truncate text-sm font-semibold">{authStore.user?.display_name}</p>
							<p class="truncate text-xs text-[var(--color-text-secondary)]">
								@{authStore.user?.username}
							</p>
						</div>
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
			</div>
		</div>
	</div>
</header>
