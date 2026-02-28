// Package detect provides auto-detection of the system's available package manager.
package detect

import (
	"os/exec"
	"sync/atomic"

	"github.com/gogrlx/snack"
)

type detected struct {
	mgr snack.Manager
	err error
}

var result atomic.Pointer[detected]

// Default returns the first available package manager on the system.
// The result is cached after the first successful detection.
// Returns ErrManagerNotFound if no supported manager is detected.
// Multiple goroutines may probe simultaneously on the first call; this is
// harmless because Available() is read-only and all goroutines store
// equivalent results.
func Default() (snack.Manager, error) {
	if r := result.Load(); r != nil {
		return r.mgr, r.err
	}
	for _, fn := range candidates() {
		m := fn()
		if m.Available() {
			result.Store(&detected{mgr: m})
			return m, nil
		}
	}
	r := &detected{err: snack.ErrManagerNotFound}
	result.Store(r)
	return r.mgr, r.err
}

// Reset clears the cached result of Default(), forcing re-detection on the
// next call. This is safe to call concurrently with Default().
func Reset() {
	result.Store(nil)
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
