//go:build !linux && !freebsd && !openbsd && !darwin && !windows

package detect

// candidates returns an empty list on unsupported platforms.
func candidates() []managerFactory {
	return nil
}

// allManagers returns an empty list on unsupported platforms.
func allManagers() []managerFactory {
	return nil
}
