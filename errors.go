package snack

import "errors"

var (
	// ErrNotInstalled is returned when a queried package is not installed.
	ErrNotInstalled = errors.New("package is not installed")

	// ErrNotFound is returned when a package cannot be found in any repository.
	ErrNotFound = errors.New("package not found")

	// ErrUnsupportedPlatform is returned when a package manager is not
	// available on the current platform.
	ErrUnsupportedPlatform = errors.New("package manager not available on this platform")

	// ErrPermissionDenied is returned when an operation requires elevated
	// privileges that were not provided.
	ErrPermissionDenied = errors.New("permission denied; try WithSudo()")

	// ErrAlreadyInstalled is returned when attempting to install a package
	// that is already present.
	ErrAlreadyInstalled = errors.New("package is already installed")

	// ErrDependencyConflict is returned when a package has unresolvable
	// dependency conflicts.
	ErrDependencyConflict = errors.New("dependency conflict")

	// ErrManagerNotFound is returned by detect when no supported package
	// manager can be found on the system.
	ErrManagerNotFound = errors.New("no supported package manager found")

	// ErrDaemonNotRunning is returned when a package manager's required
	// daemon (e.g. snapd) is not running.
	ErrDaemonNotRunning = errors.New("package manager daemon is not running")
)
