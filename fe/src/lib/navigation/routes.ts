import type { Permission } from '$lib/auth/auth-store.svelte';
export type MainRouteName = 'memorial' | 'urns' | 'structure' | 'statistics' | 'users';
export type RouteName = MainRouteName | 'profile';
export type AppRoute = { name: RouteName; path: string; title: string };
export const bottomNavItems = [
	{
		name: 'memorial' as const,
		path: '/memorial',
		label: 'Hương Linh & Bài Vị',
		icon: 'icon-[lucide--search]'
	},
	{ name: 'urns' as const, path: '/urns', label: 'QL Hũ Cốt', icon: 'icon-[lucide--archive]'},
	{
		name: 'structure' as const,
		path: '/structure',
		label: 'QL Khu Vực & Vị Trí',
		icon: 'icon-[lucide--landmark]'
	},
	{
		name: 'statistics' as const,
		path: '/statistics',
		label: 'Thống Kê & Báo Cáo',
		icon: 'icon-[lucide--chart-no-axes-combined]'
	},
	{ name: 'users' as const, path: '/users', label: 'QL Tài Khoản', icon: 'icon-[lucide--shield-user]' }
];
export function parseRoute(pathname: string): AppRoute {
	const path = pathname.length > 1 && pathname.endsWith('/') ? pathname.slice(0, -1) : pathname;
	if (path === '/' || path === '/memorial')
		return { name: 'memorial', path: '/memorial', title: 'Hương Linh & Bài Vị' };
	if (path === '/urns') return { name: 'urns', path, title: 'Quản lý Hũ Cốt' };
	if (path === '/structure') return { name: 'structure', path, title: 'QL Khu Vực & Vị Trí' };
	if (path === '/statistics') return { name: 'statistics', path, title: 'Thống Kê & Báo Cáo' };
	if (path === '/users') return { name: 'users', path, title: 'Tài khoản' };
	if (path === '/profile') return { name: 'profile', path, title: 'Hồ sơ cá nhân' };
	return { name: 'memorial', path: '/memorial', title: 'Hương Linh & Bài Vị' };
}
export function mainRouteFor(route: AppRoute): MainRouteName {
	return route.name === 'profile' ? 'users' : route.name;
}
export function routePermission(route: AppRoute): Permission | null {
	return route.name === 'users' ? 'user.read' : route.name === 'profile' ? null : 'memorial.read';
}
