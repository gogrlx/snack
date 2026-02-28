// Package flatpak provides Go bindings for the Flatpak package manager.
package flatpak

import (
	"context"

	"github.com/gogrlx/snack"
)

// Flatpak wraps the flatpak CLI.
type Flatpak struct {
	snack.Locker
}

// New returns a new Flatpak manager.
func New() *Flatpak {
	return &Flatpak{}
}

// Name returns "flatpak".
func (f *Flatpak) Name() string { return "flatpak" }

// Available reports whether flatpak is present on the system.
func (f *Flatpak) Available() bool { return available() }

// Install one or more packages.
func (f *Flatpak) Install(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) (snack.InstallResult, error) {
	f.Lock()
	defer f.Unlock()
	return install(ctx, pkgs, opts...)
}

// Remove one or more packages.
func (f *Flatpak) Remove(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) (snack.RemoveResult, error) {
	f.Lock()
	defer f.Unlock()
	return remove(ctx, pkgs, opts...)
}

// Purge removes packages including their data.
func (f *Flatpak) Purge(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) error {
	f.Lock()
	defer f.Unlock()
	return purge(ctx, pkgs, opts...)
}

// Upgrade all installed packages to their latest versions.
func (f *Flatpak) Upgrade(ctx context.Context, opts ...snack.Option) error {
	f.Lock()
	defer f.Unlock()
	return upgrade(ctx, opts...)
}

// Update is a no-op for flatpak (auto-refreshes).
func (f *Flatpak) Update(ctx context.Context) error {
	return nil
}

// List returns all installed packages.
func (f *Flatpak) List(ctx context.Context) ([]snack.Package, error) {
	return list(ctx)
}

// Search queries for packages matching the query.
func (f *Flatpak) Search(ctx context.Context, query string) ([]snack.Package, error) {
	return search(ctx, query)
}

// Info returns details about a specific package.
func (f *Flatpak) Info(ctx context.Context, pkg string) (*snack.Package, error) {
	return info(ctx, pkg)
}

// IsInstalled reports whether a package is currently installed.
func (f *Flatpak) IsInstalled(ctx context.Context, pkg string) (bool, error) {
	return isInstalled(ctx, pkg)
}

// Version returns the installed version of a package.
func (f *Flatpak) Version(ctx context.Context, pkg string) (string, error) {
	return version(ctx, pkg)
}

// Verify interface compliance at compile time.
var _ snack.Manager = (*Flatpak)(nil)
