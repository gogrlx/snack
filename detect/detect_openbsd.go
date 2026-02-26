//go:build openbsd

package detect

// candidates returns manager factories in probe order for OpenBSD.
// TODO: wire up ports.New() once the ports package is implemented.
func candidates() []managerFactory {
	return nil
}

func allManagers() []managerFactory {
	return nil
}
