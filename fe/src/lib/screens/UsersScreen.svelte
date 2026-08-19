<script lang="ts">
	import { onMount, tick } from 'svelte';
	import { authStore } from '$lib/auth/auth-store.svelte';
	import { userStore } from '$lib/users/user-store.svelte';
	import PasswordInput from '$lib/ui/PasswordInput.svelte';
	import Popup from '$lib/ui/Popup.svelte';
	import LoadingIndicator from '$lib/ui/LoadingIndicator.svelte';
	import { changePasswordPopupStore } from '$lib/auth/change-password-popup-store.svelte';
	import type { AdminUser, UserRole } from '$lib/auth/auth-store.svelte';
	import { listHouses, type House } from '$lib/memorial/api';

	let editingUser = $state<AdminUser | null>(null);
	let formOpen = $state(false);
	let displayName = $state('');
	let username = $state('');
	let password = $state('');
	let confirmPassword = $state('');
	let role = $state<UserRole>('viewer');
	let allHouses = $state(true);
	let houseIds = $state<string[]>([]);
	let houses = $state<House[]>([]);
	let displayNameInput = $state<HTMLInputElement>();
	let canManage = $derived(authStore.can('user.manage'));
	let isEditing = $derived(editingUser !== null);
	let isEditingSelf = $derived(editingUser?.id === authStore.user?.id);
	let canEditPassword = $derived(!isEditing || !isEditingSelf);
	let passwordMismatch = $derived(
		canEditPassword && confirmPassword.length > 0 && password !== confirmPassword
	);
	let passwordValid = $derived(
		isEditing
			? isEditingSelf ||
					(password === '' && confirmPassword === '') ||
					(password.length >= 8 && password === confirmPassword)
			: password.length >= 8 && password === confirmPassword
	);
	let canSave = $derived(
		Boolean(
			displayName.trim() &&
			username.trim() &&
			passwordValid &&
			(role === 'admin' || allHouses || houseIds.length > 0) &&
			!userStore.isSaving
		)
	);

	onMount(() => {
		void userStore.refreshIfStale(45_000);
		void loadHouses();
	});

	async function loadHouses() {
		try {
			houses = await listHouses();
		} catch {
			houses = [];
		}
	}

	async function focusDisplayName() {
		await tick();
		displayNameInput?.focus();
		const end = displayNameInput?.value.length ?? 0;
		displayNameInput?.setSelectionRange(end, end);
	}

	function openCreate() {
		editingUser = null;
		displayName = '';
		username = '';
		password = '';
		confirmPassword = '';
		role = 'viewer';
		allHouses = true;
		houseIds = [];
		formOpen = true;
		void focusDisplayName();
	}

	function openEdit(user: AdminUser) {
		editingUser = user;
		displayName = user.display_name;
		username = user.username;
		password = '';
		confirmPassword = '';
		role = user.role;
		allHouses = user.role === 'admin' || user.all_houses;
		houseIds = [...(user.house_ids ?? [])];
		formOpen = true;
		void focusDisplayName();
	}

	function closeForm() {
		if (!userStore.isSaving) formOpen = false;
	}

	async function save(event: SubmitEvent) {
		event.preventDefault();
		const effectiveAllHouses = role === 'admin' || allHouses;
		const effectiveHouseIds = effectiveAllHouses ? [] : houseIds;
		if (editingUser) {
			const item = await userStore.update(
				editingUser.id,
				displayName,
				username,
				role,
				password,
				effectiveAllHouses,
				effectiveHouseIds
			);
			if (!item) return;
			authStore.syncUser(item);
		} else if (
			!(await userStore.create(
				displayName,
				username,
				password,
				role,
				effectiveAllHouses,
				effectiveHouseIds
			))
		) {
			return;
		}
		formOpen = false;
	}

	function toggleHouse(houseId: string, checked: boolean) {
		houseIds = checked
			? [...new Set([...houseIds, houseId])]
			: houseIds.filter((id) => id !== houseId);
	}

	function roleLabel(userRole: UserRole) {
		return userRole === 'admin'
			? 'Quản trị viên'
			: userRole === 'editor'
				? 'Biên tập viên'
				: 'Giám sát viên';
	}

	function scopeLabel(user: AdminUser) {
		if (user.role === 'admin' || user.all_houses) return 'Tất cả Nhà Linh';
		const selectedNames = (user.house_ids ?? [])
			.map((id) => houses.find((house) => house.id === id)?.name)
			.filter(Boolean);
		return selectedNames.length > 0 ? selectedNames.join(', ') : 'Chưa chọn Nhà Linh';
	}
</script>

<svelte:window onkeydown={(event) => event.key === 'Escape' && formOpen && closeForm()} />

<section class="h-full overflow-y-auto px-4 py-4 md:px-6 md:py-6 lg:px-8">
	<div class="mx-auto max-w-[1200px]">
		<div class="mb-4 flex items-center justify-between gap-3">
			<div>
				<p class="text-sm font-medium">Tài khoản</p>
				<p class="mt-0.5 text-xs text-[var(--color-text-secondary)]">
					{userStore.items.length} tài khoản
				</p>
			</div>
			{#if canManage}
				<button
					type="button"
					class="grid h-10 w-10 place-items-center rounded-md bg-[var(--color-primary)] text-white"
					aria-label="Thêm tài khoản"
					title="Thêm tài khoản"
					onclick={openCreate}
				>
					<span class="icon-[lucide--user-plus] h-5 w-5" aria-hidden="true"></span>
				</button>
			{/if}
		</div>

		{#if userStore.isLoading && userStore.items.length === 0}
			<div class="py-16"><LoadingIndicator label="Đang tải tài khoản..." /></div>
		{:else}
			<ul class="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
				{#each userStore.items as user (user.id)}
					<li
						class="flex items-center gap-3 rounded-md border border-[var(--color-border)] bg-[var(--color-surface)] p-3"
					>
						<span
							class="grid h-10 w-10 shrink-0 place-items-center overflow-hidden rounded-full bg-[var(--color-primary-soft)] text-sm font-semibold text-[var(--color-primary-dark)]"
						>
							{#if user.avatar_url}<img
									src={user.avatar_url}
									alt=""
									class="h-full w-full object-cover"
								/>{:else}{user.display_name.slice(0, 1).toUpperCase()}{/if}
						</span>
						<div class="min-w-0 flex-1">
							<div class="flex items-center gap-2">
								<p class="truncate text-sm font-semibold">{user.display_name}</p>
								{#if user.id === authStore.user?.id}
									<span class="shrink-0 text-xs text-[var(--color-primary)]">Bạn</span>
								{/if}
							</div>
							<p class="truncate text-xs text-[var(--color-text-secondary)]">@{user.username}</p>
							<p class="mt-0.5 text-xs text-[var(--color-text-muted)]">
								{roleLabel(user.role)} · {scopeLabel(user)}
							</p>
						</div>
						{#if canManage}
							<div class="flex shrink-0 items-center gap-1">
								{#if user.id === authStore.user?.id}
									<button
										type="button"
										class="grid h-9 w-9 place-items-center rounded-md text-[var(--color-text-muted)] hover:bg-[var(--color-surface-muted)] hover:text-[var(--color-primary-dark)]"
										aria-label="Đổi mật khẩu"
										title="Đổi mật khẩu"
										onclick={() => changePasswordPopupStore.show()}
									>
										<span class="icon-[lucide--key-round] h-4.5 w-4.5" aria-hidden="true"></span>
									</button>
								{/if}
								<button
									type="button"
									class="grid h-9 w-9 place-items-center rounded-md text-[var(--color-text-muted)] hover:bg-[var(--color-surface-muted)] hover:text-[var(--color-primary-dark)]"
									aria-label={`Sửa tài khoản ${user.display_name}`}
									title="Sửa tài khoản"
									onclick={() => openEdit(user)}
								>
									<span class="icon-[lucide--pencil] h-4.5 w-4.5" aria-hidden="true"></span>
								</button>
							</div>
						{/if}
					</li>
				{/each}
			</ul>
		{/if}
	</div>
</section>

<Popup open={formOpen} title={isEditing ? 'Sửa tài khoản' : 'Thêm tài khoản'} onClose={closeForm}>
	<form id="user-form" class="space-y-4" onsubmit={save}>
		<label class="block">
			<span class="mb-1.5 block text-sm font-medium">Tên hiển thị</span>
			<input
				bind:this={displayNameInput}
				bind:value={displayName}
				required
				maxlength="120"
				class="h-11 w-full rounded-md border-[var(--color-border-strong)]"
			/>
		</label>
		<label class="block">
			<span class="mb-1.5 block text-sm font-medium">Tên đăng nhập</span>
			<input
				bind:value={username}
				required
				autocomplete="off"
				class="h-11 w-full rounded-md border-[var(--color-border-strong)]"
			/>
		</label>
		{#if canEditPassword}
			<label class="block">
				<span class="mb-1.5 block text-sm font-medium">
					{isEditing ? 'Mật khẩu mới' : 'Mật khẩu'}
				</span>
				<PasswordInput
					bind:value={password}
					required={!isEditing}
					minlength={8}
					placeholder={isEditing ? 'Để trống nếu không đổi' : 'Ít nhất 8 ký tự'}
					ariaLabel={isEditing ? 'Mật khẩu mới' : 'Mật khẩu'}
					autocomplete="new-password"
				/>
			</label>
			<label class="block">
				<span class="mb-1.5 block text-sm font-medium">
					{isEditing ? 'Nhập lại mật khẩu mới' : 'Nhập lại mật khẩu'}
				</span>
				<PasswordInput
					bind:value={confirmPassword}
					required={!isEditing}
					minlength={8}
					ariaLabel={isEditing ? 'Nhập lại mật khẩu mới' : 'Nhập lại mật khẩu'}
					autocomplete="new-password"
				/>
			</label>
			{#if passwordMismatch}
				<p class="text-sm text-[var(--color-danger)]">Mật khẩu nhập lại không khớp</p>
			{/if}
		{/if}
		<label class="block">
			<span class="mb-1.5 block text-sm font-medium">Vai trò</span>
			<select
				bind:value={role}
				disabled={isEditingSelf}
				class="h-11 w-full rounded-md border-[var(--color-border-strong)] disabled:opacity-60"
			>
				<option value="viewer">Viewer · Chỉ xem</option>
				<option value="editor">Editor · Cập nhật dữ liệu</option>
				<option value="admin">Quản trị viên</option>
			</select>
		</label>
		{#if role !== 'admin'}
			<fieldset class="space-y-3 rounded-md border border-[var(--color-border)] p-3">
				<legend class="px-1 text-sm font-medium">Phạm vi Nhà Linh</legend>
				<label class="flex items-center gap-2 text-sm">
					<input
						type="radio"
						name="house-scope"
						checked={allHouses}
						onchange={() => (allHouses = true)}
					/>
					<span>Tất cả Nhà Linh</span>
				</label>
				<label class="flex items-center gap-2 text-sm">
					<input
						type="radio"
						name="house-scope"
						checked={!allHouses}
						onchange={() => (allHouses = false)}
					/>
					<span>Chỉ các Nhà Linh được chọn</span>
				</label>
				{#if !allHouses}
					<div
						class="max-h-44 space-y-1 overflow-y-auto rounded-md bg-[var(--color-surface-muted)] p-2"
					>
						{#each houses as house (house.id)}
							<label
								class="flex items-center gap-2 rounded px-2 py-2 text-sm hover:bg-[var(--color-surface)]"
							>
								<input
									type="checkbox"
									checked={houseIds.includes(house.id)}
									onchange={(event) => toggleHouse(house.id, event.currentTarget.checked)}
								/>
								<span>{house.name}</span>
							</label>
						{:else}
							<p class="px-2 py-3 text-xs text-[var(--color-text-secondary)]">Chưa có Nhà Linh</p>
						{/each}
					</div>
					{#if houseIds.length === 0}
						<p class="text-xs text-[var(--color-danger)]">Hãy chọn ít nhất một Nhà Linh</p>
					{/if}
				{/if}
			</fieldset>
		{/if}
	</form>

	{#snippet footer()}
		<div class="grid grid-cols-2 gap-3">
			<button
				type="button"
				disabled={userStore.isSaving}
				class="h-11 rounded-md border border-[var(--color-border-strong)] bg-[var(--color-surface)] text-sm font-semibold disabled:opacity-50"
				onclick={closeForm}>Huỷ</button
			>
			<button
				type="submit"
				form="user-form"
				disabled={!canSave}
				class="flex h-11 items-center justify-center gap-2 rounded-md bg-[var(--color-primary)] text-sm font-semibold text-white disabled:opacity-50"
			>
				{#if userStore.isSaving}
					<span class="icon-[lucide--loader-circle] h-4 w-4 animate-spin" aria-hidden="true"></span>
				{/if}
				{isEditing ? 'Lưu thay đổi' : 'Tạo tài khoản'}
			</button>
		</div>
	{/snippet}
</Popup>
