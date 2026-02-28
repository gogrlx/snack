package snack

// Capabilities reports which optional interfaces a Manager implements.
// Useful for grlx to determine what operations are available before
// attempting them.
type Capabilities struct {
	VersionQuery   bool
	Hold           bool
	Clean          bool
	FileOwnership  bool
	RepoManagement bool
	KeyManagement  bool
	Groups         bool
	NameNormalize  bool
	DryRun         bool
}

// GetCapabilities probes a Manager for all optional interface support.
func GetCapabilities(m Manager) Capabilities {
	_, vq := m.(VersionQuerier)
	_, h := m.(Holder)
	_, c := m.(Cleaner)
	_, fo := m.(FileOwner)
	_, rm := m.(RepoManager)
	_, km := m.(KeyManager)
	_, g := m.(Grouper)
	_, nn := m.(NameNormalizer)
	_, dr := m.(DryRunner)
	return Capabilities{
		VersionQuery:   vq,
		Hold:           h,
		Clean:          c,
		FileOwnership:  fo,
		RepoManagement: rm,
		KeyManagement:  km,
		Groups:         g,
		NameNormalize:  nn,
		DryRun:         dr,
	}
}
