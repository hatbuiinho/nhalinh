import type { Permission } from '$lib/auth/auth-store.svelte';
export type MainRouteName = 'memorial' | 'structure' | 'statistics' | 'users';
export type RouteName = MainRouteName | 'profile';
export type AppRoute = { name: RouteName; path: string; title: string };
export const bottomNavItems = [
	{
		name: 'memorial' as const,
		path: '/memorial',
		label: 'Hương linh',
		icon: 'icon-[lucide--search]'
	},
	{
		name: 'structure' as const,
		path: '/structure',
		label: 'Cơ cấu tổ chức',
		icon: 'icon-[lucide--landmark]'
	},
	{
		name: 'statistics' as const,
		path: '/statistics',
		label: 'Thống kê',
		icon: 'icon-[lucide--chart-no-axes-combined]'
	},
	{ name: 'users' as const, path: '/users', label: 'Tài khoản', icon: 'icon-[lucide--shield-user]' }
];
export function parseRoute(pathname: string): AppRoute {
	const path = pathname.length > 1 && pathname.endsWith('/') ? pathname.slice(0, -1) : pathname;
	if (path === '/' || path === '/memorial')
		return { name: 'memorial', path: '/memorial', title: 'Tra cứu Hương linh' };
	if (path === '/structure') return { name: 'structure', path, title: 'Nhà Linh & bài vị' };
	if (path === '/statistics') return { name: 'statistics', path, title: 'Thống kê phân bổ' };
	if (path === '/users') return { name: 'users', path, title: 'Tài khoản' };
	if (path === '/profile') return { name: 'profile', path, title: 'Hồ sơ cá nhân' };
	return { name: 'memorial', path: '/memorial', title: 'Tra cứu Hương linh' };
}
export function mainRouteFor(route: AppRoute): MainRouteName {
	return route.name === 'profile' ? 'users' : route.name;
}
export function routePermission(route: AppRoute): Permission | null {
	return route.name === 'users' ? 'user.read' : route.name === 'profile' ? null : 'memorial.read';
}
