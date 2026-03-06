//go:build darwin

package detect

import (
	"github.com/gogrlx/snack"
	"github.com/gogrlx/snack/brew"
)

// candidates returns manager factories in probe order for macOS.
func candidates() []managerFactory {
	return []managerFactory{
		func() snack.Manager { return brew.New() },
	}
}

// allManagers returns all known manager factories (for ByName).
func allManagers() []managerFactory {
	return candidates()
}
