<script lang="ts">
	type TabletType =
		| 'ancestral'
		| 'spirit'
		| 'giac_linh'
		| 'family'
		| 'cuu_huyen'
		| 'clan'
		| 'collective'
		| 'fetus'
		| 'martyr'
		| 'victim'
		| 'childless'
		| 'unknown'
		| 'other';
	type TabletStatus = 'pending' | 'enshrined' | 'moving' | 'moved' | 'archived';

	let {
		code,
		tabletType = 'spirit',
		spiritCount,
		spiritNames = [],
		imageUrl = '',
		showDefaultPortrait = false,
		birthYear = '',
		deathYear = '',
		status = 'enshrined',
		fontSize = 9,
		codeScale = 1,
		maxNameWords = 0
	}: {
		code: string;
		tabletType?: TabletType;
		spiritCount: number;
		spiritNames?: string[];
		imageUrl?: string;
		showDefaultPortrait?: boolean;
		birthYear?: string;
		deathYear?: string;
		status?: TabletStatus;
		fontSize?: number;
		codeScale?: number;
		maxNameWords?: number;
	} = $props();

	let imageFailed = $state(false);

	const statusText: Record<TabletStatus, string> = {
		pending: 'Chưa an vị', enshrined: 'Đã an vị', moving: 'Đang di dời', moved: 'Đã di dời', archived: 'Lưu trữ'
	};
	let variant = $derived(tabletType === 'ancestral' ? 'cuu-huyen' : spiritCount === 1 ? 'single' : 'multiple');
	let visibleSpiritLimit = $derived(fontSize >= 13 ? (imageUrl && !imageFailed ? 3 : 4) : 2);
	let representativeNames = $derived(spiritNames.slice(0, visibleSpiritLimit));
	let remaining = $derived(Math.max(0, spiritCount - representativeNames.length));
	let rawSingleNameWords = $derived((spiritNames[0] ?? '').trim().split(/\s+/).filter(Boolean));
	let singleNameWords = $derived(maxNameWords > 0 ? rawSingleNameWords.slice(0, maxNameWords) : rawSingleNameWords);
	let singleNameTruncated = $derived(maxNameWords > 0 && rawSingleNameWords.length > maxNameWords);
	let years = $derived(birthYear && deathYear ? `${birthYear} – ${deathYear}` : birthYear || deathYear || '');
	let tone = $derived(['gold', 'rose', 'jade', 'sky'][Array.from(code).reduce((sum, char) => sum + char.charCodeAt(0), 0) % 4]);
	function label() {
		return {
			ancestral: 'CỬU HUYỀN THẤT TỔ', spirit: 'HƯƠNG LINH', giac_linh: 'GIÁC LINH', family: 'GIA TIÊN',
			cuu_huyen: 'CỬU HUYỀN THẤT TỔ', clan: 'TỘC HỌ', collective: 'CHƯ HƯƠNG LINH', fetus: 'THAI NHI',
			martyr: 'ANH HÙNG LIỆT SĨ', victim: 'HƯƠNG LINH TỬ NẠN', childless: 'HƯƠNG LINH VÔ TỰ',
			unknown: 'HƯƠNG LINH VÔ DANH', other: 'KHÁC'
		}[tabletType];
	}
	function visibleName(name: string) {
		const words = name.trim().split(/\s+/).filter(Boolean);
		if (!maxNameWords || words.length <= maxNameWords) return name;
		return `${words.slice(0, maxNameWords).join(' ')}…`;
	}
</script>

<div class={`tablet-card ${tone}`} class:single={variant === 'single'} class:multiple={variant === 'multiple'} class:cuu-huyen={variant === 'cuu-huyen'} style={`font-size:${fontSize}px;--code-scale:${codeScale}`}>
	<span class="code">{code}</span>
	<svg class="lotus" viewBox="0 0 100 100" aria-hidden="true">
		<!-- Hoa sen cách điệu: cánh giữa, cánh hai bên và đài sen. -->
		<path d="M50 78C31 66 25 43 33 19c13 9 19 23 17 39 2-16 8-30 17-39 8 24 2 47-17 59Z" />
		<path d="M48 77C29 76 14 61 12 38c18 3 30 15 36 32Z" />
		<path d="M52 77c19-1 34-16 36-39-18 3-30 15-36 32Z" />
		<path d="M50 86c-17 0-30-5-39-15 17 3 29 1 39-7 10 8 22 10 39 7-9 10-22 15-39 15Z" />
		<path d="M31 90h38" fill="none" stroke="currentColor" stroke-width="5" stroke-linecap="round" />
	</svg>
	<div class="content">
		{#if imageUrl && !imageFailed}
			<img class="tablet-image" src={imageUrl} alt={`Ảnh bài vị ${code}`} onerror={() => (imageFailed = true)} />
		{:else if showDefaultPortrait}
			<img class="portrait-placeholder" src="/icons/memorial-person-placeholder.png" alt="Chưa có ảnh bài vị" />
		{/if}
		<p class="type">{label()}</p>
		{#if variant === 'single'}
			<p class="single-name" title={spiritNames[0] ?? ''}>{#each singleNameWords as word, index}<span>{word}{#if singleNameTruncated && index === singleNameWords.length - 1}…{/if}</span>{/each}</p>
			{#if years}<p class="years">{years}</p>{/if}
		{:else}
			<div class="names">{#each representativeNames as name}<p title={name}>{visibleName(name)}</p>{/each}</div>
			{#if remaining > 0}<p class="more">+{remaining} HL</p>{/if}
		{/if}
	</div>
</div>

<style>
	.tablet-card { position: relative; display: flex; height: 100%; width: 100%; flex-direction: column; overflow: hidden; border: 1px solid #c7ad78; border-radius: 5px; background: #fffdf7; box-shadow: 0 1px 2px rgb(73 53 28 / 10%); color: #382c22; padding: .25em .28em .24em; font-family: ui-sans-serif, system-ui, sans-serif; }
	.tablet-card::after { content: ''; position: absolute; inset: 3px; border: 1px solid rgb(199 173 120 / 32%); border-radius: 3px; pointer-events: none; }
	.code { position: relative; z-index: 1; align-self: flex-start; border-radius: 2px; background: #f4e5be; color: #604820; padding: .08em .35em; font-size: calc(.82em * var(--code-scale)); font-weight: 700; line-height: 1.15; }
	.content { position: relative; z-index: 1; display: flex; min-height: 0; flex: 1; flex-direction: column; align-items: center; justify-content: flex-start; gap: .17em; overflow: hidden; padding: .18em 0 .1em; text-align: center; }
	.tablet-image { width: 3.3em; height: 3.3em; flex: 0 0 auto; margin: 0 0 .21em; border: 1px solid color-mix(in srgb, var(--accent-strong) 38%, white); border-radius: 4px; object-fit: cover; box-shadow: 0 1px 2px rgb(73 53 28 / 14%); }
	.portrait-placeholder { width: 2.45em; height: 2.45em; flex: 0 0 auto; margin: 0 0 .18em; object-fit: contain; opacity: .82; }
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
	.lotus { position: absolute; z-index: 0; right: -7%; bottom: 12%; height: 46%; width: 72%; fill: currentColor; color: var(--accent); opacity: .14; }
	.gold { --accent: #d5a445; --accent-strong: #bc8426; border-color: #ddb96d; background: #fffdf7; }
	.rose { --accent: #db7780; --accent-strong: #c55b68; border-color: #e39aa1; background: #fffafa; }
	.jade { --accent: #66ae83; --accent-strong: #39885e; border-color: #8ac7a2; background: #fbfefb; }
	.sky { --accent: #60aee7; --accent-strong: #3587c8; border-color: #8cc7ee; background: #f9fcff; }
	.gold .code, .rose .code, .jade .code, .sky .code { align-self: center; background: var(--accent-strong); color: white; padding: .16em .65em; border-radius: 999px; box-shadow: 0 1px 1px rgb(0 0 0 / 10%); }
</style>
