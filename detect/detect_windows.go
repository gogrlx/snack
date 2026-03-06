//go:build windows

package detect

// candidates returns manager factories in probe order for Windows.
// Currently no Windows package managers are supported.
func candidates() []managerFactory {
	return nil
}

// allManagers returns all known manager factories (for ByName).
func allManagers() []managerFactory {
	return nil
}
