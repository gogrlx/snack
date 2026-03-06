//go:build linux

package detect

import (
	"github.com/gogrlx/snack"
	"github.com/gogrlx/snack/apk"
	"github.com/gogrlx/snack/apt"
	"github.com/gogrlx/snack/aur"
	"github.com/gogrlx/snack/brew"
	"github.com/gogrlx/snack/dnf"
	"github.com/gogrlx/snack/flatpak"
	"github.com/gogrlx/snack/pacman"
	"github.com/gogrlx/snack/snap"
)

// candidates returns manager factories in probe order for Linux.
// The first available manager wins for Default().
func candidates() []managerFactory {
	return []managerFactory{
		func() snack.Manager { return apt.New() },
		func() snack.Manager { return dnf.New() },
		func() snack.Manager { return pacman.New() },
		func() snack.Manager { return apk.New() },
		func() snack.Manager { return flatpak.New() },
		func() snack.Manager { return snap.New() },
		func() snack.Manager { return brew.New() },
		func() snack.Manager { return aur.New() },
	}
}

// allManagers returns all known manager factories (for ByName).
func allManagers() []managerFactory {
	return candidates()
}
