import { Capacitor } from '@capacitor/core';

let accessToken = '';

export function setAccessToken(token: string) {
	accessToken = token;
}

export async function apiFetch(path: string, init: RequestInit = {}) {
	const headers = new Headers(init.headers);
	if (accessToken && !headers.has('Authorization')) {
		headers.set('Authorization', `Bearer ${accessToken}`);
	}
	const body = init.body;
	if (body !== undefined && !(body instanceof FormData) && !headers.has('Content-Type')) {
		headers.set('Content-Type', 'application/json');
	}
	return fetch(`${apiBaseURL()}${path}`, {
		...init,
		headers
	});
}

export async function apiRequest<T>(path: string, init: RequestInit = {}): Promise<T> {
	const response = await apiFetch(path, init);

	if (!response.ok) {
		let message = `Yêu cầu thất bại (${response.status})`;
		try {
			const body = (await response.json()) as { error?: { message?: string } };
			message = body.error?.message ?? message;
		} catch {
			// Keep the fallback message for non-JSON responses.
		}
		throw new Error(message);
	}
	if (response.status === 204) return undefined as T;
	return response.json() as Promise<T>;
}

function apiBaseURL(): string {
	const configuredURL = import.meta.env.VITE_API_BASE_URL;
	if (configuredURL) return configuredURL;
	if (Capacitor.getPlatform() === 'android') return 'http://10.0.2.2:8080';
	if (typeof window === 'undefined') return 'http://localhost:8080';
	return `${window.location.protocol}//${window.location.hostname}:8080`;
}
