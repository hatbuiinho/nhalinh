<script lang="ts">
	type TabletType = 'ancestral' | 'spirit' | 'family';
	type TabletStatus = 'pending' | 'enshrined' | 'moving' | 'moved' | 'archived';

	let {
		code,
		tabletType = 'spirit',
		spiritCount,
		spiritNames = [],
		imageUrl = '',
		birthYear = '',
		deathYear = '',
		status = 'enshrined',
		fontSize = 9,
		codeScale = 1
	}: {
		code: string;
		tabletType?: TabletType;
		spiritCount: number;
		spiritNames?: string[];
		imageUrl?: string;
		birthYear?: string;
		deathYear?: string;
		status?: TabletStatus;
		fontSize?: number;
		codeScale?: number;
	} = $props();

	let imageFailed = $state(false);

	const statusText: Record<TabletStatus, string> = {
		pending: 'Chưa an vị', enshrined: 'Đã an vị', moving: 'Đang di dời', moved: 'Đã di dời', archived: 'Lưu trữ'
	};
	let variant = $derived(tabletType === 'ancestral' ? 'cuu-huyen' : spiritCount === 1 ? 'single' : 'multiple');
	let visibleSpiritLimit = $derived(fontSize >= 13 ? (imageUrl && !imageFailed ? 3 : 4) : 2);
	let representativeNames = $derived(spiritNames.slice(0, visibleSpiritLimit));
	let remaining = $derived(Math.max(0, spiritCount - representativeNames.length));
	let singleNameWords = $derived((spiritNames[0] ?? '').trim().split(/\s+/).filter(Boolean));
	let years = $derived(birthYear && deathYear ? `${birthYear} – ${deathYear}` : birthYear || deathYear || '');
	let tone = $derived(['gold', 'rose', 'jade', 'sky'][Array.from(code).reduce((sum, char) => sum + char.charCodeAt(0), 0) % 4]);
	function label() { return tabletType === 'ancestral' ? 'CỬU HUYỀN THẤT TỔ' : tabletType === 'family' ? 'GIA TIÊN' : 'HƯƠNG LINH'; }
</script>

<div class={`tablet-card ${tone}`} class:single={variant === 'single'} class:multiple={variant === 'multiple'} class:cuu-huyen={variant === 'cuu-huyen'} style={`font-size:${fontSize}px;--code-scale:${codeScale}`}>
	<span class="code">{code}</span>
	<span class="lotus icon-[lucide--flower-2]" aria-hidden="true"></span>
	<div class="content">
		{#if imageUrl && !imageFailed}
			<img class="tablet-image" src={imageUrl} alt={`Ảnh bài vị ${code}`} onerror={() => (imageFailed = true)} />
		{/if}
		<p class="type">{label()}</p>
		{#if variant === 'single'}
			<p class="single-name" title={spiritNames[0] ?? ''}>{#each singleNameWords as word}<span>{word}</span>{/each}</p>
			{#if years}<p class="years">{years}</p>{/if}
		{:else}
			<div class="names">{#each representativeNames as name}<p title={name}>{name}</p>{/each}</div>
			{#if remaining > 0}<p class="more">+{remaining} HL</p>{/if}
		{/if}
	</div>
	<span class:pending={status === 'pending'} class="status">{statusText[status]}</span>
</div>

<style>
	.tablet-card { position: relative; display: flex; height: 100%; width: 100%; flex-direction: column; overflow: hidden; border: 1px solid #c7ad78; border-radius: 5px; background: #fffdf7; box-shadow: 0 1px 2px rgb(73 53 28 / 10%); color: #382c22; padding: .25em .28em .24em; font-family: ui-sans-serif, system-ui, sans-serif; }
	.tablet-card::after { content: ''; position: absolute; inset: 3px; border: 1px solid rgb(199 173 120 / 32%); border-radius: 3px; pointer-events: none; }
	.code { position: relative; z-index: 1; align-self: flex-start; border-radius: 2px; background: #f4e5be; color: #604820; padding: .08em .35em; font-size: calc(.82em * var(--code-scale)); font-weight: 700; line-height: 1.15; }
	.content { position: relative; z-index: 1; display: flex; min-height: 0; flex: 1; flex-direction: column; align-items: center; justify-content: flex-start; gap: .17em; overflow: hidden; padding: .18em 0 .1em; text-align: center; }
	.tablet-image { width: 3.3em; height: 3.3em; flex: 0 0 auto; margin: 0 0 .21em; border: 1px solid color-mix(in srgb, var(--accent-strong) 38%, white); border-radius: 4px; object-fit: cover; box-shadow: 0 1px 2px rgb(73 53 28 / 14%); }
	.type { margin: 0; color: #776454; font-size: .68em; font-weight: 700; letter-spacing: .045em; line-height: 1.15; }
	.single-name { display: flex; min-height: 0; flex: 1; width: 100%; flex-direction: column; justify-content: space-evenly; margin: 0; overflow: hidden; color: #2f241c; font-size: 1.02em; font-weight: 700; line-height: 1.12; }
	.single-name span { display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
	.years { margin: 0; color: #837568; font-size: .76em; line-height: 1.15; }
	.names { display: flex; min-height: 0; flex: 1; width: 100%; flex-direction: column; justify-content: space-evenly; gap: .12em; overflow: hidden; color: #2f241c; font-size: .88em; font-weight: 600; line-height: 1.15; }
	.names p { margin: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
	.more { margin: 0; color: #80673c; font-size: .76em; font-weight: 700; }
	.status { position: relative; z-index: 1; align-self: stretch; margin-top: .19em; border-radius: 999px; background: var(--accent-strong); color: white; padding: .15em .18em; text-align: center; font-size: .72em; font-weight: 700; line-height: 1.05; white-space: nowrap; }
	.status.pending { background: #9a7a48; }
	.cuu-huyen .type { color: #785d31; }
	.lotus { position: absolute; z-index: 0; right: -7%; bottom: 12%; height: 46%; width: 72%; color: var(--accent); opacity: .14; }
	.gold { --accent: #d5a445; --accent-strong: #bc8426; border-color: #ddb96d; background: #fffdf7; }
	.rose { --accent: #db7780; --accent-strong: #c55b68; border-color: #e39aa1; background: #fffafa; }
	.jade { --accent: #66ae83; --accent-strong: #39885e; border-color: #8ac7a2; background: #fbfefb; }
	.sky { --accent: #60aee7; --accent-strong: #3587c8; border-color: #8cc7ee; background: #f9fcff; }
	.gold .code, .rose .code, .jade .code, .sky .code { align-self: center; background: var(--accent-strong); color: white; padding: .16em .65em; border-radius: 999px; box-shadow: 0 1px 1px rgb(0 0 0 / 10%); }
</style>
