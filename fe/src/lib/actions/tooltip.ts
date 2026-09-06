export function tooltip(node: HTMLElement, content: string) {
	let value = content;
	let popup: HTMLDivElement | null = null;

	function place() {
		if (!popup) return;
		const anchor = node.getBoundingClientRect();
		const bounds = popup.getBoundingClientRect();
		const left = Math.min(
			window.innerWidth - bounds.width - 8,
			Math.max(8, anchor.left + anchor.width / 2 - bounds.width / 2)
		);
		const top = anchor.top - bounds.height - 8;
		popup.style.left = `${left}px`;
		popup.style.top = `${top >= 8 ? top : anchor.bottom + 8}px`;
	}

	function show() {
		if (!value.trim() || popup) return;
		popup = document.createElement('div');
		popup.setAttribute('role', 'tooltip');
		popup.className =
			'pointer-events-none fixed z-[100] max-w-md whitespace-pre-line rounded-md bg-[var(--color-text)] px-2 py-1 text-xs leading-snug text-white shadow-lg';
		popup.textContent = value;
		document.body.append(popup);
		place();
	}

	function hide() {
		popup?.remove();
		popup = null;
	}

	node.addEventListener('pointerenter', show);
	node.addEventListener('pointerleave', hide);
	node.addEventListener('focus', show);
	node.addEventListener('blur', hide);
	window.addEventListener('resize', place);
	document.addEventListener('scroll', place, true);

	return {
		update(nextContent: string) {
			value = nextContent;
			if (popup) {
				popup.textContent = value;
				place();
			}
		},
		destroy() {
			hide();
			node.removeEventListener('pointerenter', show);
			node.removeEventListener('pointerleave', hide);
			node.removeEventListener('focus', show);
			node.removeEventListener('blur', hide);
			window.removeEventListener('resize', place);
			document.removeEventListener('scroll', place, true);
		}
	};
}
