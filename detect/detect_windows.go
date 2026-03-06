//go:build windows

package detect

import (
	"github.com/gogrlx/snack"
	"github.com/gogrlx/snack/winget"
)

// candidates returns manager factories in probe order for Windows.
func candidates() []managerFactory {
	return []managerFactory{
		func() snack.Manager { return winget.New() },
	}
}

// allManagers returns all known manager factories (for ByName).
func allManagers() []managerFactory {
	return candidates()
}
