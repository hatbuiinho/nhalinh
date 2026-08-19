import { apiRequest } from '$lib/api/client';
import type { AdminUser, UserRole } from '$lib/auth/auth-store.svelte';

export async function listUsers() {
	return (await apiRequest<{ users: AdminUser[] }>('/api/users')).users;
}

export function createUser(
	displayName: string,
	username: string,
	password: string,
	role: UserRole,
	allHouses: boolean,
	houseIds: string[]
) {
	return apiRequest<AdminUser>('/api/users', {
		method: 'POST',
		body: JSON.stringify({
			username,
			display_name: displayName,
			password,
			role,
			all_houses: allHouses,
			house_ids: houseIds
		})
	});
}

export function updateUser(
	id: string,
	displayName: string,
	username: string,
	role: UserRole,
	password: string,
	allHouses: boolean,
	houseIds: string[]
) {
	return apiRequest<AdminUser>(`/api/users/${encodeURIComponent(id)}`, {
		method: 'PUT',
		body: JSON.stringify({
			username,
			display_name: displayName,
			role,
			password,
			all_houses: allHouses,
			house_ids: houseIds
		})
	});
}
