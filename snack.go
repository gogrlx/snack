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

// Manager is the base interface implemented by all package manager wrappers.
// It covers the core operations every package manager supports.
//
// For extended capabilities, use type assertions against the optional
// interfaces below. Example:
//
//	if holder, ok := mgr.(snack.Holder); ok {
//	    holder.Hold(ctx, "nginx")
//	} else {
//	    log.Warn("hold not supported by", mgr.Name())
//	}
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

// VersionQuerier provides version comparison and upgrade availability checks.
// Supported by: apt, pacman, apk, dnf.
type VersionQuerier interface {
	// LatestVersion returns the latest available version of a package
	// from configured repositories.
	LatestVersion(ctx context.Context, pkg string) (string, error)

	// ListUpgrades returns packages that have newer versions available.
	ListUpgrades(ctx context.Context) ([]Package, error)

	// UpgradeAvailable reports whether a newer version of a package is
	// available in the repositories.
	UpgradeAvailable(ctx context.Context, pkg string) (bool, error)

	// VersionCmp compares two version strings using the package manager's
	// native version comparison logic. Returns -1, 0, or 1.
	VersionCmp(ctx context.Context, ver1, ver2 string) (int, error)
}

// Holder provides package version pinning (hold/unhold).
// Supported by: apt, pacman (with pacman-contrib), dnf.
type Holder interface {
	// Hold pins packages at their current version, preventing upgrades.
	Hold(ctx context.Context, pkgs []string) error

	// Unhold removes version pins, allowing packages to be upgraded.
	Unhold(ctx context.Context, pkgs []string) error

	// ListHeld returns all currently held/pinned packages.
	ListHeld(ctx context.Context) ([]Package, error)
}

// Cleaner provides orphan/cache cleanup operations.
// Supported by: apt, pacman, apk, dnf.
type Cleaner interface {
	// Autoremove removes packages that were installed as dependencies
	// but are no longer required by any installed package.
	Autoremove(ctx context.Context, opts ...Option) error

	// Clean removes cached package files from the local cache.
	Clean(ctx context.Context) error
}

// FileOwner provides file-to-package ownership queries.
// Supported by: apt/dpkg, pacman, rpm/dnf, apk.
type FileOwner interface {
	// FileList returns all files installed by a package.
	FileList(ctx context.Context, pkg string) ([]string, error)

	// Owner returns the package that owns a given file path.
	Owner(ctx context.Context, path string) (string, error)
}

// RepoManager provides repository configuration operations.
// Supported by: apt, dnf, pacman (partially).
type RepoManager interface {
	// ListRepos returns all configured package repositories.
	ListRepos(ctx context.Context) ([]Repository, error)

	// AddRepo adds a new package repository.
	AddRepo(ctx context.Context, repo Repository) error

	// RemoveRepo removes a configured repository.
	RemoveRepo(ctx context.Context, id string) error
}

// KeyManager provides GPG/signing key management for repositories.
// Supported by: apt, rpm/dnf.
type KeyManager interface {
	// AddKey imports a GPG key for package verification.
	// The key can be a URL, file path, or key ID.
	AddKey(ctx context.Context, key string) error

	// RemoveKey removes a GPG key.
	RemoveKey(ctx context.Context, keyID string) error

	// ListKeys returns all trusted package signing keys.
	ListKeys(ctx context.Context) ([]string, error)
}

// Grouper provides package group operations.
// Supported by: pacman, dnf/yum.
type Grouper interface {
	// GroupList returns all available package groups.
	GroupList(ctx context.Context) ([]string, error)

	// GroupInfo returns the packages in a group.
	GroupInfo(ctx context.Context, group string) ([]Package, error)

	// GroupInstall installs all packages in a group.
	GroupInstall(ctx context.Context, group string, opts ...Option) error
}

// GroupQuerier provides an efficient check for whether a package group is
// fully installed. This is an optional extension of [Grouper].
// Supported by: pacman, dnf/yum.
type GroupQuerier interface {
	// GroupIsInstalled reports whether all packages in the group are installed.
	GroupIsInstalled(ctx context.Context, group string) (bool, error)
}

// NormalizeName provides package name normalization.
// Supported by: apt (strips :arch suffixes), rpm.
type NameNormalizer interface {
	// NormalizeName returns the canonical form of a package name,
	// stripping architecture suffixes, epoch prefixes, etc.
	NormalizeName(name string) string

	// ParseArch extracts the architecture from a package name if present.
	// Returns the name without arch and the arch string.
	ParseArch(name string) (string, string)
}
