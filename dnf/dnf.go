// Package dnf provides Go bindings for the dnf package manager (Fedora/RHEL).
package dnf

import (
	"context"

	"github.com/gogrlx/snack"
)

// DNF wraps the dnf package manager CLI.
type DNF struct {
	snack.Locker
	v5    bool // true when the system dnf is dnf5
	v5Set bool // true after detection has run
}

// New returns a new DNF manager.
func New() *DNF {
	d := &DNF{}
	d.detectVersion()
	return d
}

// IsDNF5 reports whether the underlying dnf binary is dnf5.
func (d *DNF) IsDNF5() bool { return d.v5 }

// Name returns "dnf".
func (d *DNF) Name() string { return "dnf" }

// Available reports whether dnf is present on the system.
func (d *DNF) Available() bool { return available() }

// Install one or more packages.
func (d *DNF) Install(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) (snack.InstallResult, error) {
	d.Lock()
	defer d.Unlock()
	return install(ctx, d.v5, pkgs, opts...)
}

// Remove one or more packages.
func (d *DNF) Remove(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) (snack.RemoveResult, error) {
	d.Lock()
	defer d.Unlock()
	return remove(ctx, d.v5, pkgs, opts...)
}

// Purge removes packages including configuration files (same as Remove for dnf).
func (d *DNF) Purge(ctx context.Context, pkgs []snack.Target, opts ...snack.Option) error {
	d.Lock()
	defer d.Unlock()
	_, err := remove(ctx, d.v5, pkgs, opts...)
	return err
}

// Upgrade all installed packages to their latest versions.
func (d *DNF) Upgrade(ctx context.Context, opts ...snack.Option) error {
	d.Lock()
	defer d.Unlock()
	return upgrade(ctx, opts...)
}

// Update refreshes the package index/database.
func (d *DNF) Update(ctx context.Context) error {
	d.Lock()
	defer d.Unlock()
	return update(ctx)
}

// List returns all installed packages.
func (d *DNF) List(ctx context.Context) ([]snack.Package, error) {
	return list(ctx, d.v5)
}

// Search queries the repositories for packages matching the query.
func (d *DNF) Search(ctx context.Context, query string) ([]snack.Package, error) {
	return search(ctx, query, d.v5)
}

// Info returns details about a specific package.
func (d *DNF) Info(ctx context.Context, pkg string) (*snack.Package, error) {
	return info(ctx, pkg, d.v5)
}

// IsInstalled reports whether a package is currently installed.
func (d *DNF) IsInstalled(ctx context.Context, pkg string) (bool, error) {
	return isInstalled(ctx, pkg, d.v5)
}

// Version returns the installed version of a package.
func (d *DNF) Version(ctx context.Context, pkg string) (string, error) {
	return version(ctx, pkg, d.v5)
}

// Verify interface compliance at compile time.
var _ snack.Manager = (*DNF)(nil)
