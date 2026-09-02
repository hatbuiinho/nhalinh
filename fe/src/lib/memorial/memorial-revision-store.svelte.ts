class MemorialRevisionStore {
	revision = $state(0);

	invalidate() {
		this.revision++;
	}
}

export const memorialRevisionStore = new MemorialRevisionStore();
