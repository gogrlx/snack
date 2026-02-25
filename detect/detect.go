// Package detect provides auto-detection of the system's available package manager.
package detect

import (
	"os/exec"

	"github.com/gogrlx/snack"
)

// probeOrder defines the order in which package managers are probed.
// The first available manager wins.
var probeOrder = []struct {
	name string
	bin  string
}{
	{"pacman", "pacman"},
	{"apk", "apk"},
	{"apt", "apt-get"},
	{"dnf", "dnf"},
	{"rpm", "rpm"},
	{"flatpak", "flatpak"},
	{"snap", "snap"},
	{"pkg", "pkg"},
}

// Default returns the first available package manager on the system.
// Returns ErrManagerNotFound if no supported manager is detected.
func Default() (snack.Manager, error) {
	// TODO: implement — probe for each manager in order, return first match
	return nil, snack.ErrManagerNotFound
}

// All returns all available package managers on the system.
func All() []snack.Manager {
	// TODO: implement
	return nil
}

// HasBinary reports whether a binary is available in PATH.
func HasBinary(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
