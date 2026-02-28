// Package detect provides auto-detection of the system's available package manager.
package detect

import (
	"os/exec"
	"sync"

	"github.com/gogrlx/snack"
)

var (
	defaultOnce sync.Once
	defaultMgr  snack.Manager
	defaultErr  error
)

// Default returns the first available package manager on the system.
// The result is cached after the first call.
// Returns ErrManagerNotFound if no supported manager is detected.
func Default() (snack.Manager, error) {
	defaultOnce.Do(func() {
		for _, fn := range candidates() {
			m := fn()
			if m.Available() {
				defaultMgr = m
				return
			}
		}
		defaultErr = snack.ErrManagerNotFound
	})
	return defaultMgr, defaultErr
}

// Reset clears the cached result of Default(), forcing re-detection on the
// next call. This is intended for use in tests or dynamic environments where
// the available package managers may change.
// Reset is not safe to call concurrently with Default().
func Reset() {
	defaultOnce = sync.Once{}
	defaultMgr = nil
	defaultErr = nil
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
