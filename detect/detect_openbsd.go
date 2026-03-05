//go:build openbsd

package detect

import (
	"github.com/gogrlx/snack"
	"github.com/gogrlx/snack/ports"
)

// candidates returns manager factories in probe order for OpenBSD.
func candidates() []managerFactory {
	return []managerFactory{
		func() snack.Manager { return ports.New() },
	}
}

func allManagers() []managerFactory {
	return candidates()
}
