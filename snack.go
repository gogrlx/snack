// Package snack provides idiomatic Go wrappers for system package managers.
//
// Each sub-package wraps a specific package manager's CLI, while the root
// package defines the common [Manager] interface that all providers implement.
// Use [detect.Default] to auto-detect the system's package manager.
package snack

import "context"

// Target represents a package to install, remove, or otherwise act on.
// At minimum, Name must be set. Version and other fields constrain the action.
//
// Modeled after SaltStack's pkgs list, which accepts both plain names
// and name:version mappings:
//
//	pkgs:
//	  - nginx
//	  - redis: ">=7.0"
//	  - curl: "8.5.0-1"
type Target struct {
	// Name is the package name (required).
	Name string

	// Version pins a specific version. If empty, the latest is used.
	// Comparison operators are supported where the backend allows them
	// (e.g. ">=1.2.3", "<2.0", "1.2.3-4").
	Version string

	// FromRepo constrains the install to a specific repository
	// (e.g. "unstable", "community", "epel").
	FromRepo string

	// Source is a local file path or URL for package files
	// (e.g. .deb, .rpm, .pkg.tar.zst). When set, Name is used only
	// for display/logging.
	Source string
}

// Targets is a convenience constructor for a slice of [Target] from
// plain package names (no version constraint).
func Targets(names ...string) []Target {
	targets := make([]Target, len(names))
	for i, name := range names {
		targets[i] = Target{Name: name}
	}
	return targets
}

// Manager is the common interface implemented by all package manager wrappers.
type Manager interface {
	// Install one or more packages.
	Install(ctx context.Context, pkgs []Target, opts ...Option) error

	// Remove one or more packages.
	Remove(ctx context.Context, pkgs []Target, opts ...Option) error

	// Purge one or more packages (remove including config files).
	Purge(ctx context.Context, pkgs []Target, opts ...Option) error

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
