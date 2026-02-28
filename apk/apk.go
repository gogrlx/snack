// Package apk provides Go bindings for apk-tools (Alpine Linux package manager).
package apk

import (
	"context"

	"github.com/gogrlx/snack"
)

// Apk wraps apk-tools operations.
type Apk struct {
	snack.Locker
}

// New returns a new Apk manager.
func New() *Apk {
	return &Apk{}
}

// compile-time check
var _ snack.Manager = (*Apk)(nil)

// Name returns "apk".
func (a *Apk) Name() string { return "apk" }

// Available reports whether apk is present on the system.
func (a *Apk) Available() bool { return available() }

// Install one or more packages.
func (a *Apk) Install(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) (snack.InstallResult, error) {
	a.Lock()
	defer a.Unlock()
	return install(ctx, pkgs, opts...)
}

// Remove one or more packages.
func (a *Apk) Remove(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) (snack.RemoveResult, error) {
	a.Lock()
	defer a.Unlock()
	return remove(ctx, pkgs, opts...)
}

// Purge removes packages including config files.
func (a *Apk) Purge(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) error {
	a.Lock()
	defer a.Unlock()
	return purge(ctx, pkgs, opts...)
}

// Upgrade all installed packages.
func (a *Apk) Upgrade(ctx context.Context, opts ...snack.Option) error {
	a.Lock()
	defer a.Unlock()
	return upgrade(ctx, opts...)
}

// Update refreshes the package index.
func (a *Apk) Update(ctx context.Context) error {
	a.Lock()
	defer a.Unlock()
	return update(ctx)
}

// List returns all installed packages.
func (a *Apk) List(ctx context.Context) ([]snack.Package, error) {
	return list(ctx)
}

// Search queries the index for matching packages.
func (a *Apk) Search(ctx context.Context, query string) ([]snack.Package, error) {
	return search(ctx, query)
}

// Info returns details about a package.
func (a *Apk) Info(ctx context.Context, pkg string) (*snack.Package, error) {
	return info(ctx, pkg)
}

// IsInstalled reports whether a package is installed.
func (a *Apk) IsInstalled(ctx context.Context, pkg string) (bool, error) {
	return isInstalled(ctx, pkg)
}

// Version returns the installed version of a package.
func (a *Apk) Version(ctx context.Context, pkg string) (string, error) {
	return version(ctx, pkg)
}
