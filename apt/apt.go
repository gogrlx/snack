// Package apt provides Go bindings for APT (Advanced Packaging Tool) on Debian/Ubuntu.
package apt

import (
	"context"

	"github.com/gogrlx/snack"
)

// Apt implements the snack.Manager interface using apt-get and apt-cache.
type Apt struct{}

// New returns a new Apt manager.
func New() *Apt {
	return &Apt{}
}

// Name returns "apt".
func (a *Apt) Name() string { return "apt" }

// Install one or more packages.
func (a *Apt) Install(ctx context.Context, pkgs []string, opts ...snack.Option) error {
	return install(ctx, pkgs, opts...)
}

// Remove one or more packages.
func (a *Apt) Remove(ctx context.Context, pkgs []string, opts ...snack.Option) error {
	return remove(ctx, pkgs, opts...)
}

// Purge one or more packages including config files.
func (a *Apt) Purge(ctx context.Context, pkgs []string, opts ...snack.Option) error {
	return purge(ctx, pkgs, opts...)
}

// Upgrade all installed packages.
func (a *Apt) Upgrade(ctx context.Context, opts ...snack.Option) error {
	return upgrade(ctx, opts...)
}

// Update refreshes the package index.
func (a *Apt) Update(ctx context.Context) error {
	return update(ctx)
}

// List returns all installed packages.
func (a *Apt) List(ctx context.Context) ([]snack.Package, error) {
	return list(ctx)
}

// Search queries the package index.
func (a *Apt) Search(ctx context.Context, query string) ([]snack.Package, error) {
	return search(ctx, query)
}

// Info returns details about a specific package.
func (a *Apt) Info(ctx context.Context, pkg string) (*snack.Package, error) {
	return info(ctx, pkg)
}

// IsInstalled reports whether a package is currently installed.
func (a *Apt) IsInstalled(ctx context.Context, pkg string) (bool, error) {
	return isInstalled(ctx, pkg)
}

// Version returns the installed version of a package.
func (a *Apt) Version(ctx context.Context, pkg string) (string, error) {
	return version(ctx, pkg)
}

// Available reports whether apt-get is present on the system.
func (a *Apt) Available() bool {
	return available()
}
