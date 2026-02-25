// Package snack provides idiomatic Go wrappers for system package managers.
//
// Each sub-package wraps a specific package manager's CLI, while the root
// package defines the common [Manager] interface that all providers implement.
// Use [detect.Default] to auto-detect the system's package manager.
package snack

import "context"

// Manager is the common interface implemented by all package manager wrappers.
type Manager interface {
	// Install one or more packages.
	Install(ctx context.Context, pkgs []string, opts ...Option) error

	// Remove one or more packages.
	Remove(ctx context.Context, pkgs []string, opts ...Option) error

	// Purge one or more packages (remove including config files).
	Purge(ctx context.Context, pkgs []string, opts ...Option) error

	// Upgrade all installed packages to their latest versions.
	Upgrade(ctx context.Context, opts ...Option) error

	// Update refreshes the package index/database.
	Update(ctx context.Context) error

	// List returns all installed packages.
	List(ctx context.Context) ([]Package, error)

	// Search queries the package index for packages matching the query.
	Search(ctx context.Context, query string) ([]Package, error)

	// Info returns details about a specific package.
	Info(ctx context.Context, pkg string) (*Package, error)

	// IsInstalled reports whether a package is currently installed.
	IsInstalled(ctx context.Context, pkg string) (bool, error)

	// Version returns the installed version of a package.
	Version(ctx context.Context, pkg string) (string, error)

	// Available reports whether this package manager is present on the system.
	Available() bool

	// Name returns the package manager's identifier (e.g. "apt", "pacman").
	Name() string
}
