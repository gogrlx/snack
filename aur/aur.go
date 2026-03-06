// Package aur provides Go bindings for AUR (Arch User Repository) package building.
// AUR packages are built from source using makepkg.
package aur

import (
	"context"

	"github.com/gogrlx/snack"
)

// AUR wraps makepkg and AUR helper tools for building packages from the AUR.
type AUR struct {
	snack.Locker
}

// New returns a new AUR manager.
func New() *AUR {
	return &AUR{}
}

// Compile-time interface checks.
var (
	_ snack.Manager         = (*AUR)(nil)
	_ snack.VersionQuerier  = (*AUR)(nil)
	_ snack.Cleaner         = (*AUR)(nil)
	_ snack.PackageUpgrader = (*AUR)(nil)
	_ snack.NameNormalizer  = (*AUR)(nil)
)

// Name returns "aur".
func (a *AUR) Name() string { return "aur" }

// Available reports whether makepkg is present on the system.
func (a *AUR) Available() bool { return available() }

// Install one or more packages from the AUR.
func (a *AUR) Install(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) (snack.InstallResult, error) {
	a.Lock()
	defer a.Unlock()
	return install(ctx, pkgs, opts...)
}

// Remove is not directly supported by AUR (use pacman).
func (a *AUR) Remove(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) (snack.RemoveResult, error) {
	return remove(ctx, pkgs, opts...)
}

// Purge is not directly supported by AUR (use pacman).
func (a *AUR) Purge(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) error {
	return purge(ctx, pkgs, opts...)
}

// Upgrade all AUR packages (requires re-building from source).
func (a *AUR) Upgrade(ctx context.Context, opts ...snack.Option) error {
	a.Lock()
	defer a.Unlock()
	return upgrade(ctx, opts...)
}

// Update is a no-op for AUR (packages are fetched on demand).
func (a *AUR) Update(ctx context.Context) error {
	return update(ctx)
}

// List returns installed packages that came from the AUR.
func (a *AUR) List(ctx context.Context) ([]snack.Package, error) {
	return list(ctx)
}

// Search queries the AUR for packages matching the query.
func (a *AUR) Search(ctx context.Context, query string) ([]snack.Package, error) {
	return search(ctx, query)
}

// Info returns details about a specific AUR package.
func (a *AUR) Info(ctx context.Context, pkg string) (*snack.Package, error) {
	return info(ctx, pkg)
}

// IsInstalled reports whether a package from the AUR is currently installed.
func (a *AUR) IsInstalled(ctx context.Context, pkg string) (bool, error) {
	return isInstalled(ctx, pkg)
}

// Version returns the installed version of an AUR package.
func (a *AUR) Version(ctx context.Context, pkg string) (string, error) {
	return version(ctx, pkg)
}

// NormalizeName returns the canonical form of an AUR package name.
func (a *AUR) NormalizeName(name string) string {
	return normalizeName(name)
}

// ParseArch extracts the architecture from a package name if present.
func (a *AUR) ParseArch(name string) (string, string) {
	return parseArch(name)
}
