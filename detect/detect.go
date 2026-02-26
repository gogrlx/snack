// Package detect provides auto-detection of the system's available package manager.
package detect

import (
	"os/exec"

	"github.com/gogrlx/snack"
)

// Default returns the first available package manager on the system.
// Returns ErrManagerNotFound if no supported manager is detected.
func Default() (snack.Manager, error) {
	for _, fn := range candidates() {
		m := fn()
		if m.Available() {
			return m, nil
		}
	}
	return nil, snack.ErrManagerNotFound
}

// All returns all available package managers on the system.
func All() []snack.Manager {
	var out []snack.Manager
	for _, fn := range candidates() {
		m := fn()
		if m.Available() {
			out = append(out, m)
		}
	}
	return out
}

// ByName returns a specific manager by name, regardless of availability.
// Returns ErrManagerNotFound if the name is not recognized.
func ByName(name string) (snack.Manager, error) {
	for _, fn := range allManagers() {
		m := fn()
		if m.Name() == name {
			return m, nil
		}
	}
	return nil, snack.ErrManagerNotFound
}

// managerFactory is a function that returns a new Manager instance.
type managerFactory func() snack.Manager

// HasBinary reports whether a binary is available in PATH.
func HasBinary(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
