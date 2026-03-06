// Package winget provides Go bindings for the Windows Package Manager (winget).
package winget

import (
	"context"

	"github.com/gogrlx/snack"
)

// Winget wraps the winget CLI.
type Winget struct {
	snack.Locker
}

// New returns a new Winget manager.
func New() *Winget {
	return &Winget{}
}

// Name returns "winget".
func (w *Winget) Name() string { return "winget" }

// Available reports whether winget is present on the system.
func (w *Winget) Available() bool { return available() }

// Install one or more packages.
func (w *Winget) Install(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) (snack.InstallResult, error) {
	w.Lock()
	defer w.Unlock()
	return install(ctx, pkgs, opts...)
}

// Remove one or more packages.
func (w *Winget) Remove(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) (snack.RemoveResult, error) {
	w.Lock()
	defer w.Unlock()
	return remove(ctx, pkgs, opts...)
}

// Purge removes packages including configuration data.
func (w *Winget) Purge(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) error {
	w.Lock()
	defer w.Unlock()
	return purge(ctx, pkgs, opts...)
}

// Upgrade all installed packages.
func (w *Winget) Upgrade(ctx context.Context, opts ...snack.Option) error {
	w.Lock()
	defer w.Unlock()
	return upgrade(ctx, opts...)
}

// Update refreshes the winget source index.
func (w *Winget) Update(ctx context.Context) error {
	return update(ctx)
}

// List returns all installed packages.
func (w *Winget) List(ctx context.Context) ([]snack.Package, error) {
	return list(ctx)
}

// Search queries the winget repository for packages matching the query.
func (w *Winget) Search(ctx context.Context, query string) ([]snack.Package, error) {
	return search(ctx, query)
}

// Info returns details about a specific package.
func (w *Winget) Info(ctx context.Context, pkg string) (*snack.Package, error) {
	return info(ctx, pkg)
}

// IsInstalled reports whether a package is currently installed.
func (w *Winget) IsInstalled(ctx context.Context, pkg string) (bool, error) {
	return isInstalled(ctx, pkg)
}

// Version returns the installed version of a package.
func (w *Winget) Version(ctx context.Context, pkg string) (string, error) {
	return version(ctx, pkg)
}

// NormalizeName returns the canonical form of a winget package ID.
func (w *Winget) NormalizeName(name string) string {
	return normalizeName(name)
}

// ParseArch extracts the architecture from a package name if present.
func (w *Winget) ParseArch(name string) (string, string) {
	return parseArch(name)
}

// Verify interface compliance at compile time.
var _ snack.Manager = (*Winget)(nil)
var _ snack.PackageUpgrader = (*Winget)(nil)

// UpgradePackages upgrades specific installed packages.
func (w *Winget) UpgradePackages(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) (snack.InstallResult, error) {
	w.Lock()
	defer w.Unlock()
	return upgradePackages(ctx, pkgs, opts...)
}
